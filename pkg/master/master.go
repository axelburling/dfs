package master

import (
	"context"
	"os"
	"sync"
	"time"

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
	"github.com/axelburling/dfs/pkg/master/queue"
	"github.com/axelburling/dfs/pkg/master/router"
	"github.com/axelburling/dfs/pkg/master/utils"
	"github.com/axelburling/dfs/pkg/node/client"
	pubsub "github.com/axelburling/dfs/pkg/pubsub/client"
	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/message"
	ms "github.com/axelburling/dfs/pkg/pubsub/client/msg"
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
	stopChan    chan os.Signal
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
	app := fiber.New(
		fiber.Config{
			BodyLimit: 5000000000,
		},
	)

	conf := config.New[config.MasterConfig](config.Master, log)

	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	db, err := db.New(ctx, conf.DatabaseUrl)

	if err != nil {
		log.Fatal("failed to initialize database pool")
	}

	log.Debug("database initialization successful")

	nodeAddChan := make(chan generated.Node)

	masterService := grpc.NewMasterService(log, db, nodeAddChan)

	registries := internalgrpc.NewServiceRegistries(
		internalgrpc.NewServiceRegistry(masterService, pb.RegisterMasterServiceServer),
	)

	grpcServer := internalgrpc.NewGrpc(conf.GrpcAddr, log, registries, nil)

	masterId := utils.GenerateMasterId()

	errSrv := internalhttp.NewError(log, masterId)

	server := server.New(masterId, conf.Addr, server.Graceful, app, grpcServer, log, nil)

	globalQueue := queue.NewGlobalQueue(log)

	route := router.New(app, errSrv, globalQueue, db)

	m := &Master{
		log:         log,
		id:          masterId,
		nodeAddChan: nodeAddChan,
		uploadQueue: globalQueue,
		server:      server,
		config:      conf,
		db:          db,
		router:      route,
	}

	if v := m.isValid(); !v {
		log.Fatal("master initialization did not succeed, missing fields", zap.String("id", masterId))
	}

	log.Info("master initialization successful", zap.String("id", masterId))

	server.AddStopMonitoringFunc(m.monitorAddNode)

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

func (m *Master) WithGrpc() *Master {
	masterService := grpc.NewMasterService(m.log, m.db, make(chan generated.Node))

	registries := internalgrpc.NewServiceRegistries(
		internalgrpc.NewServiceRegistry(masterService, pb.RegisterMasterServiceServer),
	)

	grpcServer := internalgrpc.NewGrpc(n.config.GrpcAddr, n.log, registries)

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

	subscription, err := topic.CreateSubscription(m.ctx, m.config.PubSub.Subscription)

	if err != nil {
		m.log.Fatal("failed to create pubsub subscription", zap.Error(err))
	}

	m.pubsub.PubSub = pubsub
	m.pubsub.Topic = topic
	m.pubsub.Subscription = subscription
	m.pubsub.MessageChan = make(chan *message.Message, 50)

	m.server.AddStopMonitoringFunc(m.receiveMessage)

	return m
}

func (m *Master) receiveMessage(stop chan os.Signal) {
	if m.pubsub.MessageChan == nil {
		m.pubsub.MessageChan = make(chan *message.Message, 50) // Initialize message channel
	}

	go func() {
		for {
			select {
			case <-stop:
				close(m.pubsub.MessageChan) // Close channel on stop signal
				return
			default:
				if m.pubsub.Subscription != nil { // Ensure subscription is set
					// TODO:
					m.pubsub.Subscription.Receive(m.ctx, m.pubsub.MessageChan, nil)
				}
			}
		}
	}()
}

func (m *Master) monitorMessages(stop chan os.Signal) {
	if m.pubsub.MessageChan == nil {
		return
	}

	go func() {
		for {
			select {
				case <-stop:
					return
				case msg := <- m.pubsub.MessageChan:
					switch ms.Action(msg.Action) {
						case ms.NodeAdd:
							message, err := ms.Decode[ms.NodeAddReq](msg)
							if err != nil {
								m.log.Warn("could not decode pubsub message", zap.String("action", ms.NodeAdd.String()))
							}



					}

			}
		}
	}()
}

func (m *Master) monitorAddNode(stop chan os.Signal) {
	m.stopChan = stop
	go func() {
		for {
			select {
			case <-m.stopChan:
				return
			case n := <-m.nodeAddChan:
				m.log.Info("Adding node", zap.Any("node id", n.ID))

				var node *client.Node
				var err error
				maxRetries := 5
				for i := range maxRetries {
					node, err = client.NewNode(n.Address, n.GrpcAddress)
					if err == nil {
						break // Successfully connected
					}
					m.log.Warn("Failed to connect to node, retrying...", zap.Int("attempt", i+1), zap.Error(err))
					time.Sleep(2 * time.Second) // Wait before retrying
				}

				if err != nil {
					m.log.Error("Failed to connect after retries, marking node as unhealthy", zap.String("address", n.Address))
					m.db.DeleteNode(context.Background(), n.ID)
					continue
				}

				m.mu.Lock()
				m.nodes = append(m.nodes, node) // Add node even if unhealthy
				m.mu.Unlock()
				m.uploadQueue.AddNode(node)
			}
		}
	}()
}
