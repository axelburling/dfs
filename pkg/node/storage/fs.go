package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/node/storage/compression"
	"github.com/axelburling/dfs/pkg/node/storage/crypto"
	"go.uber.org/zap"
)

type FileStorage struct {
	logger      *log.Logger
	crypto      *crypto.Crypto
	compression *compression.Compression
	basePath    string
}

func NewFileStorage(basePath string, logger *log.Logger, crypto *crypto.Crypto) (Storage, error) {
	if err := os.Mkdir(basePath, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	logger.Debug("initialized FileStorage")
	return &FileStorage{
		basePath: basePath,
		logger: logger,
		crypto: crypto,
		compression: compression.NewCompression(crypto, logger),
	}, nil
}

func (fs *FileStorage) SaveChunk(chunkPath string, reader io.Reader) (*ChunkInfo, error) {
	path := filepath.Join(fs.basePath, chunkPath)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		fs.logger.Error("failed to create directories", zap.String("path", path), zap.Error(err))
		return nil, fmt.Errorf("failed to create directories: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		fs.logger.Error("failed to create file", zap.String("path", path), zap.Error(err))
		return nil, err
	}
	defer file.Close()

	hash := sha256.New()

	// writer := io.MultiWriter(file, hash)
	writer := io.MultiWriter(fs.compression.NewCompress(file), hash)

	size, err := io.Copy(writer, reader)

	if err != nil {
		os.Remove(path)
		fs.logger.Error("failed to read chunk", zap.String("path", path))
		return nil, err
	}

	hashStr := hex.EncodeToString(hash.Sum(nil))

	fs.logger.Info("saved chunk to file system", zap.String("path", path), zap.String("hash", hashStr), zap.Int64("size", size))
	return &ChunkInfo{
		Path: path,
		Hash: hashStr,
		Size: size,
	}, nil
}

func (fs *FileStorage) GetChunk(chunkPath string) (io.ReadCloser, string, error) {
	path := filepath.Join(fs.basePath, chunkPath)
	file, err := os.Open(path)
	if err != nil {
		fs.logger.Error("failed to open chunk file", zap.String("path", path), zap.Error(err))
		return nil, "", err
	}

	reader, err := fs.compression.Decompress(file)

	if err != nil {
		fs.logger.Error("failed to decompress or encrypt file", zap.String("path", path), zap.Error(err))
		return nil, "", err
	}

	// Calculate hash without loading entire file into memory
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		file.Close()
		fs.logger.Error("failed to calculate chunk hash", zap.String("path", path), zap.Error(err))
		return nil, "", err
	}

	// Seek back to start of file
	if _, err := file.Seek(0, 0); err != nil {
		file.Close()
		fs.logger.Error("failed to seek chunk file", zap.String("path", path), zap.Error(err))
		return nil, "", err
	}

	return file, hex.EncodeToString(hash.Sum(nil)), nil
}

func (fs *FileStorage) GetChunks() (*[]Chunk, error) {
	entries, err := os.ReadDir(fs.basePath)
	if err != nil {
		fs.logger.Error("failed to read dir", zap.String("path", fs.basePath), zap.Error(err))
		return nil, err
	}

	var wg sync.WaitGroup
	chunksChan := make(chan Chunk, len(entries))
	errChan := make(chan error, len(entries))

	for _, entry := range entries {
		wg.Add(1)
		name := entry.Name()
		path := filepath.Join(fs.basePath, name)

		go func(name, path string) {
			defer wg.Done()
			fs.getChunkData(path, name, chunksChan, errChan)
		}(name, path) // Pass variables explicitly to avoid race conditions
	}

	// Close channels when all goroutines finish
	go func() {
		wg.Wait()
		close(chunksChan)
		close(errChan)
	}()

	// Collect results
	var chunks []Chunk
	for chunk := range chunksChan {
		chunks = append(chunks, chunk)
	}


	// Check for errors after collecting chunks
	select {
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
	default:
	}

	return &chunks, nil
}

func (fs *FileStorage) getChunkData(path, name string, chunkChan chan Chunk, errChan chan error) {
	file, err := os.Open(path)
	if err != nil {
		errChan <- err
		return
	}
	defer file.Close() // Ensure file is closed

	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		errChan <- err
		return
	}

	id := strings.Split(name, "_")[0]

	chunkChan <- Chunk{
		ID:   id,
		Hash: hex.EncodeToString(hash.Sum(nil)),
		Size: size,
	}
}


func (fs *FileStorage) GetChunkContent(chunkPath string) ([]byte, string, error) {
	path := filepath.Join(fs.basePath, chunkPath)
	file, err := os.Open(path)
	if err != nil {
		fs.logger.Error("failed to open chunk file", zap.String("path", path), zap.Error(err))
		return nil, "", err
	}
	defer file.Close()

	hash := sha256.New()
	data, err := io.ReadAll(io.TeeReader(file, hash))

	if err != nil {
		fs.logger.Error("failed to read file", zap.String("path", path), zap.Error(err))
		return nil, "", err
	}
	
	hashStr := hex.EncodeToString(hash.Sum(nil))

	fs.logger.Info("got chunk content", zap.String("path", path), zap.String("hash", hashStr))
	return data, hashStr, nil
}

func (fs *FileStorage) DeleteChunk(chunkPath string) error {
	path := filepath.Join(fs.basePath, chunkPath)
	err := os.Remove(path)

	if err != nil {
		fs.logger.Error("failed to delete chunk", zap.String("path", path))
	}
	return err
}

func (fs *FileStorage) Purge() error {
	entries, err := os.ReadDir(fs.basePath)
	if err != nil {
		fs.logger.Error("failed to read dir", zap.String("path", fs.basePath), zap.Error(err))
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(entries))

	for _, entry := range entries {
		wg.Add(1)
		path := filepath.Join(fs.basePath, entry.Name())

		go func(path string) {
			defer wg.Done()
			fs.removeChunk(path, errChan)
		}(path) // Pass variables explicitly to avoid race conditions
	}

	wg.Wait()
	close(errChan)


	// Check for errors after collecting chunks
	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	default:
	}

	return nil
}

func (fs *FileStorage) removeChunk(path string, errChan chan error) {
	err := os.Remove(path)

	if err != nil {
		fs.logger.Error("failed to delete chunk", zap.String("path", path))
		errChan <- err
	}
}