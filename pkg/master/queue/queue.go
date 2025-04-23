package queue

import (
	"io"

	"github.com/google/uuid"
)

const (
	numGlobalWorkers = 10
	numNodeWorker    = 5
)

type Job struct {
	ID                string
	Chunk             io.ReadCloser
	ChunkIndex        int64
	ReplicationFactor int64
}

func NewJob(chunk io.ReadCloser, index int64, replicationFactor int64) Job {
	return Job{
		ID:                uuid.NewString(),
		Chunk:             chunk,
		ChunkIndex:        index,
		ReplicationFactor: replicationFactor,
	}
}
