package router

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/axelburling/dfs/pkg/node/utils"
	"github.com/gofiber/fiber/v3"
)

func (r *Router) handleReceive(c fiber.Ctx) error {
	if r.isReadOnly() {
		return 	r.errSrv.New(c, "node is in readonly mode", http.StatusServiceUnavailable, nil)
	}
	
	form, err := c.MultipartForm()
	
	if err != nil {
		return r.errSrv.New(c, "failed to parse multipart form", http.StatusServiceUnavailable, err)
	}
	
	files, ok := form.File["chunk"]
	
	if !ok {
		return r.errSrv.New(c, "failed to get chunk file from multipart form", http.StatusBadRequest, nil)
	}
	
	if len(files) != 1 {
		return r.errSrv.New(c, "only one chunk should be uploaded at a time", http.StatusBadRequest, nil)
	}
	
	fileHeader := files[0]
	
	chunkIndexes, ok := form.Value["index"]
	
	if !ok {
		return r.errSrv.New(c, "failed to get chunk index from multipart form", http.StatusBadRequest, nil)
	}
	
	if len(chunkIndexes) != 1 {
		return r.errSrv.New(c, "only one chunk index should be uploaded at a time", http.StatusBadRequest, nil)
	}
	
	chunkIndex := chunkIndexes[0]
	
	re := regexp.MustCompile(`^[0-9]+$`)
	
	if !re.MatchString(chunkIndex) {
		return r.errSrv.New(c, "chunk index include non integer characters", http.StatusBadRequest, nil)
	}
	
	free := r.diskSpace.Free()
	
	requiredSpace := uint64(float64(fileHeader.Size) * 1.1)
	
	if free < requiredSpace {
		r.readonly = true
		return r.errSrv.New(c, "insufficient storage space", http.StatusInsufficientStorage, nil)
	}
	
	file, err := fileHeader.Open()
	
	if err != nil {
		return r.errSrv.New(c, "failed to open chunk file", http.StatusInternalServerError, err)
	}
	defer file.Close()
	
	chunkId := utils.GenerateChunkId()
	
	path := fmt.Sprintf("%s_%s", chunkId, chunkIndex)
	
	info, err := r.storage.SaveChunk(path, file)

	if err != nil {
		return r.errSrv.New(c, "failed to save chunk", http.StatusInternalServerError, err)
	}
	
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"hash": info.Hash,
		"size": info.Size,
	})
}