package grpc

import (
	"time"

	"github.com/axelburling/dfs/internal/config"
	internalgrpc "github.com/axelburling/dfs/internal/grpc"
	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/pubsub/apiv1/scheduler"
	"github.com/axelburling/dfs/pkg/pubsub/grpc/pb"
	"github.com/axelburling/dfs/pkg/pubsub/grpc/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	Time:    5 * time.Second, // Ping the client if it is idle for 5 seconds to ensure the connection is still active
	Timeout: 1 * time.Second, // Wait 1 second for the ping ack before assuming the connection is dead
}

func New(log *log.Logger, conf *config.PubSubConfig) (*internalgrpc.GRPC, *scheduler.Scheduler) {
	schedule := scheduler.NewScheduler(200 * time.Millisecond)

	errSrv := internalgrpc.NewError(log)

	publish := services.NewPublisherService(schedule, log, errSrv)
	subscribe := services.NewSubscriberService(schedule, log, errSrv)

	registries := internalgrpc.NewServiceRegistries(
		internalgrpc.NewServiceRegistry[pb.PublisherServer](publish, pb.RegisterPublisherServer),
		internalgrpc.NewServiceRegistry[pb.SubscriberServer](subscribe, pb.RegisterSubscriberServer),
	)

	return internalgrpc.NewGrpc(conf.Addr, log, registries, grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp)), schedule
}
