package router

import (
	"os"
	"sync"

	internalhttp "github.com/axelburling/dfs/internal/http"
	internalnode "github.com/axelburling/dfs/internal/node"
	"github.com/axelburling/dfs/pkg/node/constants"
	"github.com/axelburling/dfs/pkg/node/storage"
	"github.com/gofiber/fiber/v3"
)

type Router struct {
	app  *fiber.App

	storage storage.Storage
	errSrv *internalhttp.Error
	diskSpace *internalnode.DiskSpace

	// node data
	mu sync.Mutex
	readonly bool

	StatusChan chan bool
	stopChan chan os.Signal

}

func New(app *fiber.App, storage storage.Storage, ds *internalnode.DiskSpace, errSrv *internalhttp.Error, statusChan chan bool) *Router {
	r := &Router{
		app: app,
		readonly: false,
		StatusChan: statusChan,
		storage: storage,
		errSrv: errSrv,
		diskSpace: ds,
	}

	r.setupRoutes()

	return r
}

func (r *Router) setupRoutes() {
	r.app.Post(constants.HTTP_RECEIVE_CHUNK, r.handleReceive)
}

func (r *Router) Start(stopChan chan os.Signal) {
	r.stopChan = stopChan

	go r.monitorStatusChange()
}

func (r *Router) monitorStatusChange() {
	for {
		select {
		case <- r.stopChan:
			return
		case status := <- r.StatusChan:
			r.setReadOnly(status)
		}
	}
}

func (n *Router) setReadOnly(status bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.readonly = status
}

func (n *Router) isReadOnly() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.readonly
}
