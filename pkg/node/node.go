package node

import (
	"context"
	"os"
	"sync"

	"github.com/axelburling/dfs/internal/config"
	internalgrpc "github.com/axelburling/dfs/internal/grpc"
	internalhttp "github.com/axelburling/dfs/internal/http"
	"github.com/axelburling/dfs/internal/log"
	internalnode "github.com/axelburling/dfs/internal/node"
	"github.com/axelburling/dfs/internal/server"
	"github.com/axelburling/dfs/pkg/node/grpc"
	"github.com/axelburling/dfs/pkg/node/grpc/pb"
	"github.com/axelburling/dfs/pkg/node/router"
	"github.com/axelburling/dfs/pkg/node/storage"
	"github.com/axelburling/dfs/pkg/node/utils"
	"github.com/axelburling/dfs/pkg/pubsub/client"
	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/message"
	"github.com/axelburling/dfs/pkg/pubsub/client/msg"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

type Node struct {
	id  string
	ctx context.Context

	mu       sync.Mutex
	readonly bool

	stopChan    chan os.Signal
	messageChan chan *message.Message

	server  *server.Server
	config  *config.NodeConfig
	router  *router.Router
	log     *log.Logger
	app     *fiber.App
	service services

	pubsub  client.PubSubInfo
	filters []string
}

type services struct {
	disk       *internalnode.DiskSpace
	storage    storage.Storage
	grpc       *internalgrpc.GRPC
	errSrv     *internalhttp.Error
	statusChan chan bool
}

func New(conf *config.NodeConfig, volumePath string, log *log.Logger, storage storage.Storage) *Node {
	ctx := context.Background()
	node := &Node{
		ctx:    ctx,
		log:    log,
		config: conf,
		service: services{
			storage:    storage,
			statusChan: make(chan bool),
		},
	}

	node.WithID()
	node.WithApp()
	node.WithDiskSpace(volumePath)
	node.WithGrpc()
	node.WithServer()
	node.WithPubSub()

	node.Register()

	return node
}

func (n *Node) Start() error {
	return n.server.Start()
}

func (n *Node) WithID() *Node {
	nodeId := utils.GenerateNodeId()

	n.id = nodeId

	return n
}

func (n *Node) WithApp() *Node {
	app := fiber.New(fiber.Config{
		BodyLimit: 70 * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	n.app = app

	return n
}

func (n *Node) WithDiskSpace(volumePath string) *Node {
	diskSpace := internalnode.NewDiskSpace(volumePath)
	n.service.disk = diskSpace

	return n
}

func (n *Node) WithGrpc() *Node {
	nodeService := grpc.NewNodeService(n.log, n.service.disk, n.service.storage, n.service.statusChan)

	registries := internalgrpc.NewServiceRegistries(
		internalgrpc.NewServiceRegistry(nodeService, pb.RegisterNodeServiceServer),
	)

	grpcServer := internalgrpc.NewGrpc(n.config.GrpcAddr, n.log, registries)

	n.service.grpc = grpcServer

	return n
}

func (n *Node) WithServer() *Node {
	errSrv := internalhttp.NewError(n.log, n.id)

	router := router.New(n.app, n.service.storage, n.service.disk, errSrv, n.service.statusChan)

	n.router = router

	server := server.New(n.id, n.config.Addr, server.Graceful, n.app, n.service.grpc, n.log, router.Start)

	n.server = server

	return n
}

func (n *Node) WithPubSub() *Node {
	pubsub, err := client.NewClient(n.config.PubSub.Addr, n.config.PubSub.PoolSize, n.id)

	if err != nil {
		n.log.Fatal("failed to initialize pubsub client", zap.Error(err))
	}

	topic, err := pubsub.CreateTopic(n.ctx, n.config.PubSub.Topic)

	if err != nil {
		n.log.Fatal("failed to create pubsub topic", zap.Error(err))
	}

	subscription, err := topic.CreateSubscription(n.ctx, n.config.PubSub.Subscription)

	if err != nil {
		n.log.Fatal("failed to create pubsub subscription", zap.Error(err))
	}

	//	subscription.Receive(ctx, messageChan chan *message.Message, filters []string)

	n.server.AddStopMonitoringFunc(n.receiveMessage)

	n.pubsub.PubSub = pubsub
	n.pubsub.Topic = topic
	n.pubsub.Subscription = subscription
	n.pubsub.MessageChan = make(chan *message.Message, 50)

	return n
}

func (n *Node) receiveMessage(stop chan os.Signal) {
	if n.pubsub.MessageChan == nil {
		n.pubsub.MessageChan = make(chan *message.Message, 10) // Initialize message channel
	}

	go func() {
		for {
			select {
			case <-stop:
				close(n.pubsub.MessageChan) // Close channel on stop signal
				return
			default:
				if n.pubsub.Subscription != nil { // Ensure subscription is set
					n.pubsub.Subscription.Receive(n.ctx, n.pubsub.MessageChan, n.filters)
				}
			}
		}
	}()
}

func (n *Node) Register() {
	hostname, err := os.Hostname()

	if err != nil {
		n.log.Fatal("failed to get computer hostname", zap.Error(err))
	}

	ms, err := msg.Create(msg.NodeAdd, msg.NodeAddReq{
		ID:          n.id,
		Address:     n.config.Addr,
		GrpcAddress: n.config.GrpcAddr,
		Hostname:    hostname,
		IsHealthy:   true,
		Readonly:    n.readonly,
		TotalSpace:  int64(n.service.disk.Total()),
		FreeSpace:   int64(n.service.disk.Free()),
	})

	if err != nil {
		n.log.Fatal("failed to create message", zap.Error(err))
	}

	ids, err := n.topic.Publish(n.ctx, []*message.Message{
		ms,
	}, nil)

	if err != nil || len(ids) != 1 {
		n.log.Fatal("failed to register node", zap.String("id", n.id), zap.Error(err))
	}
}
