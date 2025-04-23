package client

import (
	"context"
	"errors"

	"github.com/axelburling/dfs/pkg/node/grpc/client"
	"github.com/google/uuid"
)

type Node struct {
	*client.Client
	*http
	ID       string
	addr     string
	grpcAddr string
	ReadOnly bool
}

func NewNode(addr, grpcAddr string) (*Node, error) {
	grpcClient, err := client.NewClient(grpcAddr)

	if err != nil {
		return nil, err
	}

	res, err := grpcClient.Health(context.Background(), nil)

	if err != nil {
		return nil, err
	}

	if !res.Alive {
		return nil, errors.New("client: node is dead")
	}

	return &Node{
		ID:       uuid.NewString(),
		Client:   grpcClient,
		http:     newHttp(addr),
		addr:     addr,
		grpcAddr: grpcAddr,
		ReadOnly: res.ReadOnly,
	}, nil
}
