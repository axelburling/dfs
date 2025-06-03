package config

import (
	"encoding/base64"
	"strconv"

	"github.com/axelburling/dfs/internal/log"
	"go.uber.org/zap"
)

type NodeConfig struct {
	Addr          string
	GrpcAddr      string
	EncryptionKey []byte
	PubSub        pubsubClientConfig
}

func newNodeConfig(log *log.Logger) *NodeConfig {
	addr := getEnv("node_address", log)
	gaddr := getEnv("node_grpc", log)
	keyStr := getEnv("encryption_key", log)
	pubsubAddr := getEnv("pubsub_address", log)
	pubsubTopic := getEnv("pubsub_topic", log)
	pubsubPoolSizeStr := getEnv("node_pubsub_pool_size", log)

	key, err := base64.StdEncoding.DecodeString(keyStr)

	if err != nil {
		log.Fatal("encryption key is not in base64 format use node crypto to generate a new key", zap.Error(err))
	}

	if len(key) != 32 {
		log.Fatal("encryption key is not not 32 bytes long to use the AES-256", zap.Int("length", len(key)))
	}

	pubsubPoolSize, err := strconv.Atoi(pubsubPoolSizeStr)

	if err != nil {
		log.Fatal("pubsub pool size is not a number", zap.Error(err))
	}

	return &NodeConfig{
		Addr:          addr,
		GrpcAddr:      gaddr,
		EncryptionKey: key,
		PubSub: pubsubClientConfig{
			Addr:         pubsubAddr,
			PoolSize:     pubsubPoolSize,
			Topic:        pubsubTopic,
		},
	}
}
