package pubsub

import (
	"github.com/axelburling/dfs/internal/config"
	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/pubsub/client"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func setupPubSub(log *log.Logger, conf *config.PubSubConfig) *client.PubSub {
	pubsub, err := client.NewClient(conf.Addr, 5, uuid.NewString())

	if err != nil {
		log.Fatal("could not connect to pubsub server", zap.String("address", conf.Addr), zap.Error(err))
	}

	return pubsub
}
