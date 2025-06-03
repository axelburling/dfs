package queue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/node/client"
	"go.uber.org/zap"
)

type GlobalQueue struct {
	globalJobQueue chan Job
	nodeWorkers    map[string]chan NodeJob
	log            *log.Logger
	done           chan struct{}
	nodeQueues     []*NodeQueue
	nextNodeIndex  atomic.Uint64
	mu             sync.Mutex
	nodes          []*client.Node
	nodeChangeChan chan nodeChange
}

type nodeChange struct {
	node *client.Node
	ok bool
}

func NewGlobalQueue(log *log.Logger) *GlobalQueue {
	globalJobQueue := make(chan Job, 1000) // Global queue
	nodeWorkers := make(map[string]chan NodeJob)

	gq := &GlobalQueue{
		globalJobQueue: globalJobQueue,
		nodeWorkers:    nodeWorkers,
		done:           make(chan struct{}),
		log:            log,
	}

	for range numGlobalWorkers {
		log.Debug("started global worker")
		go gq.worker()
	}

	go gq.monitorNodeChange()

	return gq
}

func (gq *GlobalQueue) pickNodes(replicationFactor int64) []*client.Node {
	totalNodes := len(gq.nodes)

	if totalNodes == 0 {
		return nil
	}

	selectedNodes := make([]*client.Node, 0, replicationFactor)

	for i := 0; len(selectedNodes) < int(replicationFactor) && i < totalNodes; i++ {
		idx := int(gq.nextNodeIndex.Add(1)-1) % totalNodes

		node := gq.nodes[idx]

		if node.ReadOnly || !node.Alive {
			continue
		}

		selectedNodes = append(selectedNodes, node)
	}

	if len(selectedNodes) == 0 {
		return nil
	}

	return selectedNodes
}

func (gq *GlobalQueue) AddNode(node *client.Node) {
	gq.mu.Lock()
	defer gq.mu.Unlock()
	gq.nodes = append(gq.nodes, node)
	gq.nodeWorkers[node.ID] = make(chan NodeJob, 500)

	nq := NewNodeQueue(node, gq.log, gq.nodeWorkers[node.ID], gq.done)

	gq.nodeQueues = append(gq.nodeQueues, nq)
}

func (gq *GlobalQueue) MarkNode(node *client.Node, ok bool) {

}

func (gq *GlobalQueue) worker() {
	for {
		select {
		case <-gq.done:
			return
		case job := <-gq.globalJobQueue:
			nodes := gq.pickNodes(job.ReplicationFactor)

			for _, node := range nodes {
				gq.enqueueJobToNode(job, node.ID)
			}
			gq.log.Debug("got global job", zap.String("id", job.ID))
		}
	}
}

func (gq *GlobalQueue) monitorNodeChange() {
	for {
		select {
		case <-gq.done:
			return
		case ch := <-gq.nodeChangeChan:
			for _, n := range gq.nodes {
				if n.ID == ch.node.ID {
					gq.mu.Lock()
					n.Alive = ch.ok
					gq.mu.Unlock()
				}
			}
		}
	}
}

func (gq *GlobalQueue) Enqueue(job Job) {
	gq.globalJobQueue <- job
}
func (gq *GlobalQueue) enqueueJobToNode(job Job, id string) {
	nodeJob := NodeJob{
		ID:         job.ID,
		Chunk:      job.Chunk,
		ChunkIndex: job.ChunkIndex,
		Time:       time.Now(),
		Status:     true,
	}

	gq.nodeWorkers[id] <- nodeJob
}
