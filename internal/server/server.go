package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	internalgrpc "github.com/axelburling/dfs/internal/grpc"
	"github.com/axelburling/dfs/internal/log"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type StoppingStrategy int

const (
	Graceful StoppingStrategy = iota
	Direct
)

type StopMonitoringFunc func(chan os.Signal)

type Server struct {
	id string

	app      *fiber.App
	grpc     *internalgrpc.GRPC
	log      *log.Logger
	strategy StoppingStrategy

	addr string

	stopMonitoringFuncs []StopMonitoringFunc
}

func New(id, addr string, strategy StoppingStrategy, app *fiber.App, grpc *internalgrpc.GRPC, log *log.Logger, stopMonitoring StopMonitoringFunc) *Server {
	funcs := make([]StopMonitoringFunc, 0)

	if stopMonitoring != nil {
		funcs = append(funcs, stopMonitoring)
	}

	return &Server{
		id:                  id,
		app:                 app,
		grpc:                grpc,
		addr:                addr,
		log:                 log,
		strategy:            strategy,
		stopMonitoringFuncs: funcs,
	}
}

func (s *Server) print() {
	fmt.Println("------------------------------------------------------------------------------------------------------")
	if len(s.addr) > 0 {
		s.log.Info(fmt.Sprintf("HTTP listening on: http://%s/", s.addr))
	}
	if len(s.grpc.Addr()) > 0 {
		s.log.Info(fmt.Sprintf("gRPC listening on: %s", s.grpc.Addr()))
	}
	fmt.Println("------------------------------------------------------------------------------------------------------")
}

func (s *Server) AddStopMonitoringFunc(stopMonitoring StopMonitoringFunc) {
	s.stopMonitoringFuncs = append(s.stopMonitoringFuncs, stopMonitoring)
}

func (s *Server) Start() error {
	s.print()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for _, fn := range s.stopMonitoringFuncs {
		fn(stop)
	}

	errChan := make(chan error, 2)

	if s.grpc != nil {
		go func() {
			if err := s.grpc.Start(); err != nil {
				s.log.Error("failed to start grpc server", zap.String("id", s.id), zap.String("address", s.grpc.Addr()))
				errChan <- fmt.Errorf("gRPC server error: %v", err)
			}
		}()
	}

	if s.app != nil {
		go func() {
			if err := s.app.Listen(s.addr, fiber.ListenConfig{DisableStartupMessage: true}); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.log.Error("failed to start http server", zap.String("id", s.id), zap.String("address", s.addr))
				errChan <- fmt.Errorf("http server error: %v", err)
			}
		}()
	}

	select {
	case <-stop:
		s.log.Info("Shutting down servers")

		if s.app != nil {
			if err := s.app.Shutdown(); err != nil {
				s.log.Error("HTTP server shutdown error", zap.Error(err))
				errChan <- err
			}
		}

		if s.grpc != nil {
			if s.strategy == Direct {
				s.grpc.Stop()
			} else {
				s.grpc.GracefulStop()
			}

			if err := s.grpc.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				s.log.Error("gRPC Listener close error", zap.Error(err))
				errChan <- err
			}
		}

		s.log.Info("Servers shut down cleanly.")
		return nil
	case err := <-errChan:
		return err
	}
}
