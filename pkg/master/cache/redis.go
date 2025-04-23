package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(opt *redis.Options) Cache {
	client := redis.NewClient(opt)

	return &Redis{
		client: client,
	}
}

func (r *Redis) Set(ctx context.Context, key string, val string, expiration time.Duration) error {
	 return r.client.Set(ctx, key, val, expiration).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()

	if err != nil {
		return "", ErrCacheMiss
	}

	return val, nil
}