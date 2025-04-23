package client

import (
	"context"

	internalgrpc "github.com/axelburling/dfs/internal/grpc"
	"github.com/axelburling/dfs/pkg/node/grpc/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	cc pb.NodeServiceClient
}

func NewClient(target string, customOptions ...grpc.DialOption) (*Client, error) {
	conn, err := internalgrpc.NewPreClient(target, customOptions)

	if err != nil {
		return nil, err
	}

	cc := pb.NewNodeServiceClient(conn)

	return &Client{
		conn: conn,
		cc: cc,
	}, nil
}

func (c *Client) Health(ctx context.Context, readonly *bool) (*pb.HealthResponse, error) {
	return c.cc.Health(ctx, &pb.HealthRequest{
		ReadOnly: readonly,
	})
}

func (c *Client) DiskSpace(ctx context.Context) (*pb.DiskSpaceResponse, error) {
	return c.cc.DiskSpace(ctx, &pb.Empty{})
}

func (c *Client) Init(ctx context.Context) (*pb.InitResponse, error) {
	return c.cc.Init(ctx, &pb.Empty{})
}