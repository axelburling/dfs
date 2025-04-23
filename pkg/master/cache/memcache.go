package cache

import (
	"context"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type Memcache struct {
	client *memcache.Client
}

func NewMemcache(client *memcache.Client) Cache {
	return &Memcache{
		client: client,
	}
}

func (r *Memcache) Set(ctx context.Context, key string, val string, expiration time.Duration) error {
	return r.client.Set(&memcache.Item{
		Key: key,
		Value: []byte(val),
		Expiration: int32(expiration.Seconds()),
	})
}

func (r *Memcache) Get(ctx context.Context, key string) (string, error) {
	item, err := r.client.Get(key)

	if err != nil {
		return "", ErrCacheMiss
	}

	return string(item.Value), nil
}