package reader

import "io"

type ChunkReader struct {
	reader io.Reader
	size   int64
}

func NewChunkReader(r io.Reader, chunkSize int64) *ChunkReader {
	return &ChunkReader{
		reader: r,
		size:   chunkSize,
	}
}

func (cr *ChunkReader) NextChunk() (io.ReadCloser, error) {
	limitReader := &io.LimitedReader{
		R: cr.reader,
		N: cr.size,
	}

	if limitReader.N <= 0 {
		return nil, io.EOF
	}

	return io.NopCloser(limitReader), nil
}
