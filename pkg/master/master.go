package master

import (
	"context"
	"sync"

	"github.com/axelburling/dfs/internal/config"
	internalgrpc "github.com/axelburling/dfs/internal/grpc"
	internalhttp "github.com/axelburling/dfs/internal/http"
	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/internal/server"
	"github.com/axelburling/dfs/pkg/db"
	"github.com/axelburling/dfs/pkg/db/generated"
	"github.com/axelburling/dfs/pkg/master/cache"
	"github.com/axelburling/dfs/pkg/master/grpc"
	"github.com/axelburling/dfs/pkg/master/grpc/pb"
	internalps "github.com/axelburling/dfs/pkg/master/pubsub"
	"github.com/axelburling/dfs/pkg/master/queue"
	"github.com/axelburling/dfs/pkg/master/router"
	"github.com/axelburling/dfs/pkg/master/utils"
	"github.com/axelburling/dfs/pkg/node/client"
	pubsub "github.com/axelburling/dfs/pkg/pubsub/client"
	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/message"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

type Master struct {
	id  string
	ctx context.Context

	log      *log.Logger
	config   *config.MasterConfig
	server   *server.Server
	app      *fiber.App
	router   *router.Router
	services services
	pubsub   pubsub.PubSubInfo

	mu    sync.RWMutex
	nodes []*client.Node

	uploadQueue *queue.GlobalQueue

	nodeAddChan chan generated.Node
	db          *db.DB

	cache cache.Cache
}

type services struct {
	grpc       *internalgrpc.GRPC
	errSrv     *internalhttp.Error
	statusChan chan bool
}

func New(log *log.Logger) *Master {
	ctx := context.Background()
	conf := config.New[config.MasterConfig](config.Master, log)

	m := &Master{
		ctx: ctx,
		log: log,
		config: conf,
	}

	m.WithID()
	m.WithDB()
	m.WithApp()
	m.WithQueue()
	m.WithGrpc()
	m.WithServer()
	m.WithPubSub()


	if v := m.isValid(); !v {
		log.Fatal("master initialization did not succeed, missing fields", zap.String("id", m.id))
	}

	log.Info("master initialization successful", zap.String("id", m.id))


	return m
}

func (m *Master) isValid() bool {

	if len(m.id) == 0 {
		return false
	}

	if m.config == nil {
		return false
	}

	if m.log == nil {
		return false
	}

	if m.server == nil {
		return false
	}

	if m.db == nil {
		return false
	}

	return true
}

func (m *Master) Start() error {
	return m.server.Start()
}

func (m *Master) WithID() *Master {
	masterId := utils.GenerateMasterId()
	ctx := context.Background()

	m.id = masterId
	m.ctx = ctx

	return m
}

func (m *Master) WithDB() *Master {
	db, err := db.New(m.ctx, m.config.DatabaseUrl)

	if err != nil {
		m.log.Fatal("failed to initialize database pool", zap.Error(err))
	}
	m.db = db

	m.log.Info("database initialization successful")

	return m
}

func (m *Master) WithApp() *Master {
	app := fiber.New(fiber.Config{
		BodyLimit: 70 * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	m.app = app

	return m
}

func (m *Master) WithQueue() *Master {
	globalQueue := queue.NewGlobalQueue(m.log)

	m.uploadQueue = globalQueue

	m.nodeAddChan = make(chan generated.Node, 100)

	return m
}

func (m *Master) WithServer() *Master {	
	errSrv := internalhttp.NewError(m.log, m.id)

	router := router.New(m.app, errSrv, m.uploadQueue, m.db)

	m.router = router

	server := server.New(m.id, m.config.Addr, server.Graceful, m.app, m.services.grpc, m.log, m.monitorAddNode)

	m.server = server

	return m
}

func (m *Master) WithGrpc() *Master {
	masterService := grpc.NewMasterService(m.log, m.db, m.nodeAddChan)

	registries := internalgrpc.NewServiceRegistries(
		internalgrpc.NewServiceRegistry(masterService, pb.RegisterMasterServiceServer),
	)

	grpcServer := internalgrpc.NewGrpc(m.config.GrpcAddr, m.log, registries)

	m.services.grpc = grpcServer

	return m
}

func (m *Master) WithPubSub() *Master {
	pubsub, err := pubsub.NewClient(m.config.PubSub.Addr, m.config.PubSub.PoolSize, m.id)

	if err != nil {
		m.log.Fatal("failed to initialize pubsub client", zap.Error(err))
	}

	topic, err := pubsub.CreateTopic(m.ctx, m.config.PubSub.Topic)

	if err != nil {
		m.log.Fatal("failed to create pubsub topic", zap.Error(err))
	}

	subscription, err := topic.CreateSubscription(m.ctx, m.id)

	if err != nil {
		m.log.Fatal("failed to create pubsub subscription", zap.Error(err))
	}

	m.pubsub.PubSub = pubsub
	m.pubsub.Topic = topic
	m.pubsub.Subscription = subscription
	m.pubsub.MessageChan = make(chan *message.Message, 50)

	ps := internalps.NewInternal(m.ctx, m.log, m.pubsub, m.db.Conv, m.nodeAddChan)

	m.server.AddStopMonitoringFunc(ps.ReceiveMessage)
	m.server.AddStopMonitoringFunc(ps.MonitorMessage)

	return m
}
