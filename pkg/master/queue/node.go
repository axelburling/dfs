package queue

import (
	"container/list"
	"io"
	"sync"
	"time"

	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/node/client"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NodeJob struct {
	ID         string
	Chunk      io.ReadCloser
	ChunkIndex int64
	Time       time.Time
	Status     bool
	attempts   int
}

type NodeQueue struct {
	Node         *client.Node
	nodeJobQueue chan NodeJob
	log          *log.Logger
	done         chan struct{}
	id           string

	mu       sync.Mutex
	jobsList *list.List
	jobs     map[string]*NodeJob
}

func NewNodeQueue(node *client.Node, log *log.Logger, nodeJobQueue chan NodeJob, done chan struct{}) *NodeQueue {
	nq := &NodeQueue{
		Node:         node,
		nodeJobQueue: nodeJobQueue,
		done:         done,
		log:          log,
		id:           node.ID,
		jobsList:     list.New(),
		jobs:         make(map[string]*NodeJob),
	}

	for range numNodeWorker {
		log.Debug("started node worker", zap.String("id", node.ID))
		go nq.worker()
	}

	go nq.checkLoop()

	return nq
}

func (nq *NodeQueue) worker() {
	for {
		select {
		case <-nq.done:
			return
		case job := <-nq.nodeJobQueue:
			nq.log.Debug("got node job", zap.String("id", job.ID), zap.String("node id", nq.id), zap.Int("attempts", job.attempts))

			nq.jobs[job.ID] = &job

			_, err := nq.Node.SendChunk(job.Chunk, job.ChunkIndex, uuid.NewString())

			if err != nil {
				nq.log.Debug("got error while sending chunk", zap.String("node id", nq.id), zap.String("job id", job.ID), zap.Error(err))
				job.Status = false

				nq.jobs[job.ID] = &job
				continue
			}

			delete(nq.jobs, job.ID)
		}
	}
}

func (nq *NodeQueue) checkLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-nq.done:
			return
		case t := <-ticker.C:
			nq.mu.Lock()
			var toRequeue []*NodeJob

			for id, job := range nq.jobs {
				if job.attempts >= 5 {
					nq.log.Warn("job failed too many times, dropping", zap.String("id", id))
					delete(nq.jobs, id)
					continue
				}

				if job.Time.Add(5*time.Second).Before(t) || !job.Status {
					nq.log.Debug("requeuing job", zap.String("id", id), zap.Int("attempts", job.attempts))

					job.Time = time.Now()
					job.attempts++
					toRequeue = append(toRequeue, job)
					delete(nq.jobs, id)
				}
			}

			nq.mu.Unlock()

			// Requeue jobs outside the lock to avoid deadlocks
			for _, job := range toRequeue {
				nq.nodeJobQueue <- *job
			}
		}
	}
}
