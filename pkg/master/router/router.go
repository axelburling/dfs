package router

import (
	internalhttp "github.com/axelburling/dfs/internal/http"
	"github.com/axelburling/dfs/pkg/db"
	"github.com/axelburling/dfs/pkg/master/constants"
	"github.com/axelburling/dfs/pkg/master/queue"
	"github.com/gofiber/fiber/v3"
)

type Router struct {
	app *fiber.App

	errSrv *internalhttp.Error

	queue *queue.GlobalQueue

	db *db.DB
}

func New(app *fiber.App, errSrv *internalhttp.Error, queue *queue.GlobalQueue, db *db.DB) *Router {
	r := &Router{
		app:    app,
		errSrv: errSrv,
		queue:  queue,
		db:     db,
	}

	r.setupRoutes()

	return r
}

func (r *Router) setupRoutes() {
	r.app.Put(constants.HTTP_UPLOAD_FILE, r.handleUpload)
}
