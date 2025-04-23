package pubsub

import (
	"github.com/axelburling/dfs/internal/config"
	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/internal/server"
	"github.com/axelburling/dfs/pkg/pubsub/grpc"
	"github.com/google/uuid"
)

type PubSub struct {
	id     string
	conf   *config.PubSubConfig
	log    *log.Logger
	server *server.Server
}

func New(log *log.Logger) *PubSub {
	conf := config.New[config.PubSubConfig](config.PubSub, log)

	grpcServer, scheduler := grpc.New(log, conf)

	pubSubId := uuid.NewString()

	srv := server.New(pubSubId, "", server.Direct, nil, grpcServer, log, nil)

	srv.AddStopMonitoringFunc(scheduler.Start)

	return &PubSub{
		id:     pubSubId,
		log:    log,
		conf:   conf,
		server: srv,
	}
}

func (pb *PubSub) Start() error {
	return pb.server.Start()
}
