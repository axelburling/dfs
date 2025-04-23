package grpc

import (
	"github.com/axelburling/dfs/internal/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	log *log.Logger
}

func NewError(log *log.Logger) *Error {
	return &Error{log}
}

func (e *Error) New(code codes.Code, msg string, err error, a ...any) error {
	grpcerr := status.Errorf(code, msg, a...)
	e.log.Warn(msg, zap.String("code", code.String()), zap.Error(err), zap.NamedError("Grpc error", grpcerr))
	return grpcerr
}
