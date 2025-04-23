package cache

import (
	"context"
	"errors"
	"time"
)

var ErrCacheMiss = errors.New("cache: miss")

type Cache interface {
	Set(ctx context.Context, key string, val string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}