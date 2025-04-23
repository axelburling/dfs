package grpc

import (
	"net"
	"time"

	"github.com/axelburling/dfs/internal/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPC struct {
	*grpc.Server
	addr string
	log  *log.Logger
	ln   *net.Listener
}

type RegisterFunc[T any] func(grpc.ServiceRegistrar, T)

type ServiceRegistry[T any] struct {
	Service   T
	Registrar RegisterFunc[T]
}

func NewGrpc[T any](addr string, log *log.Logger, registries []ServiceRegistry[T], options ...grpc.ServerOption) *GRPC {
	server := grpc.NewServer(options...)

	for _, registry := range registries {
		registry.Registrar(server, registry.Service)
	}

	reflection.Register(server)

	return &GRPC{
		Server: server,
		log:    log,
		addr:   addr,
	}
}

func (g *GRPC) Start() error {
	ln, err := net.Listen("tcp", g.addr)
	if err != nil {
		g.log.Fatal("failed to start tcp listener", zap.String("address", g.addr), zap.Error(err))
	}
	g.ln = &ln
	return g.Server.Serve(ln)
}

func (g *GRPC) Stop() {
	time.Sleep(2 * time.Second)
	g.Server.Stop()
}

func (g *GRPC) GracefulStop() {
	g.Server.GracefulStop()
}

func (g *GRPC) Close() error {
	if g.ln == nil {
		return nil
	}
	return (*g.ln).Close()
}

func (g *GRPC) Addr() string {
	return g.addr
}

func NewServiceRegistries(registrations ...ServiceRegistry[any]) []ServiceRegistry[any] {
	return registrations
}

func NewServiceRegistry[T any](service T, registrar RegisterFunc[T]) ServiceRegistry[any] {
	wrappedRegistrar := func(s grpc.ServiceRegistrar, _ any) {
		registrar(s, service)
	}

	return ServiceRegistry[any]{
		Service:   service,
		Registrar: wrappedRegistrar,
	}
}
