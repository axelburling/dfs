package config

import (
	"strconv"

	"github.com/axelburling/dfs/internal/log"
	"go.uber.org/zap"
)

type MasterConfig struct {
	Addr        string
	GrpcAddr    string
	ChunkSize   int
	DatabaseUrl string
	PubSub      pubsubClientConfig
}

func newMasterConfig(log *log.Logger) *MasterConfig {
	addr := getEnv("address", log)
	gaddr := getEnv("grpc", log)
	sizeStr := getEnv("chunk_size", log)
	dbConn := getEnv("database_url", log)
	pubsubAddr := getEnv("pubsub_address", log)
	pubsubTopic := getEnv("pubsub_topic", log)
	pubsubPoolSizeStr := getEnv("master_pubsub_pool_size", log)

	size, err := strconv.Atoi(sizeStr)

	if err != nil {
		log.Fatal("chunk size is not a number", zap.Error(err))
	}

	pubsubPoolSize, err := strconv.Atoi(pubsubPoolSizeStr)

	if err != nil {
		log.Fatal("pubsub pool size is not a number", zap.Error(err))
	}

	return &MasterConfig{
		Addr:        addr,
		GrpcAddr:    gaddr,
		ChunkSize:   size,
		DatabaseUrl: dbConn,
		PubSub: pubsubClientConfig{
			Addr:         pubsubAddr,
			PoolSize:     pubsubPoolSize,
			Topic:        pubsubTopic,
		},
	}
}
