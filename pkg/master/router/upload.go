package router

import (
	"fmt"
	"io"
	"net/http"

	"github.com/axelburling/dfs/pkg/db/generated"
	"github.com/axelburling/dfs/pkg/master/queue"
	"github.com/axelburling/dfs/pkg/master/reader"
	"github.com/gofiber/fiber/v3"
)

const GB = 1000000000

func (r *Router) handleUpload(c fiber.Ctx) error {
	multi, err := c.MultipartForm()

	if err != nil {
		return r.errSrv.New(c, "failed to parse multipart form", http.StatusInternalServerError, err)
	}

	files := multi.File["file"]

	if len(files) != 1 {
		return r.errSrv.New(c, "can only one file per upload", http.StatusBadRequest, nil)
	}

	mFile := files[0]

	file, err := mFile.Open()

	if err != nil {
		file.Close()
		return r.errSrv.New(c, "could not open uploaded file", http.StatusNoContent, err)
	}

	if mFile.Size > 5*GB {
		file.Close()
		return r.errSrv.New(c, "file is to big, consider using multipart upload", http.StatusBadRequest, nil)
	}

	fileHeader := make([]byte, 512)

	file.Read(fileHeader)

	contentType := http.DetectContentType(fileHeader)

	file.Seek(0, 0)

	chunkIndex := 0

	chunkReader := reader.NewChunkReader(file, 1024*10)

	for range 1 {
		chunk, err := chunkReader.NextChunk()

		if err == io.EOF {
			break
		}

		if err != nil {
			file.Close()
			return r.errSrv.New(c, "failed to read file", http.StatusInternalServerError, err)
		}

		r.queue.Enqueue(queue.NewJob(chunk, int64(chunkIndex), 3))

		data, _ := io.ReadAll(chunk)
		fmt.Println("Chunk:", len(data), "bytes")
		chunkIndex++
	}

	r.db.InsertObject(c.Context(), generated.InsertObjectParams{
		Name:        mFile.Filename,
		TotalSize:   mFile.Size,
		ContentType: contentType,
		TotalChunks: int32(chunkIndex),
	})

	return nil
}
