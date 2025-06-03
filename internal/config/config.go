package config

import (
	"os"
	"strings"

	"github.com/axelburling/dfs/internal/log"
	"go.uber.org/zap"
)

type Type int

const (
	Master Type = iota
	Node
	PubSub
)

type Config interface{}

func New[T Config](ty Type, log *log.Logger) *T {
	var cfg any

	switch ty {
	case Master:
		cfg = newMasterConfig(log)
	case Node:
		cfg = newNodeConfig(log)
	case PubSub:
		cfg = newPubSubConfig(log)
	default:
		log.Fatal("invalid config type", zap.Any("type", ty))
		return nil
	}

	result, ok := cfg.(*T)
	if !ok {
		log.Fatal("failed to cast config type")
		return nil
	}

	return result
}

func getEnv(key string, log *log.Logger) string {
	val, exists := os.LookupEnv(strings.ToUpper(key))

	if !exists {
		log.Fatal("environment variable does not exist", zap.String("key", key))
	}

	return val
}

type pubsubClientConfig struct {
	Addr         string
	PoolSize     int
	Topic        string
}
