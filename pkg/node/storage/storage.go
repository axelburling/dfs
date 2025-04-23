package storage

import (
	"io"
)

type ChunkInfo struct {
	Path string
	Hash string
	Size int64
}

type Chunk struct {
	ID   string
	Hash string
	Size int64
}

type Storage interface {
	SaveChunk(path string, reader io.Reader) (*ChunkInfo, error)
	GetChunk(path string) (io.ReadCloser, string, error)
	GetChunkContent(path string) ([]byte, string, error)
	GetChunks() (*[]Chunk, error)
	DeleteChunk(path string) error
	Purge() error
}
