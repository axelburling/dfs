package config

import (
	"strconv"

	"github.com/axelburling/dfs/internal/log"
	"go.uber.org/zap"
)

type PubSubConfig struct {
	Addr        string
	MaxAttempts int
}

func newPubSubConfig(log *log.Logger) *PubSubConfig {
	addr := getEnv("pubsub_address", log)
	maxStr := getEnv("pubsub_max_attempts", log)

	max, err := strconv.Atoi(maxStr)

	if err != nil {
		log.Fatal("pubsub_max_attempts is not a number", zap.Error(err))
	}

	return &PubSubConfig{
		Addr:        addr,
		MaxAttempts: max,
	}
}
