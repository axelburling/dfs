package client

import (
	"context"

	internalgrpc "github.com/axelburling/dfs/internal/grpc"
	"github.com/axelburling/dfs/pkg/master/grpc/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	cc pb.MasterServiceClient
}

func NewClient(target string, customOptions ...grpc.DialOption) (*Client, error) {
	conn, err := internalgrpc.NewPreClient(target, customOptions)

	if err != nil {
		return nil, err
	}

	cc := pb.NewMasterServiceClient(conn)

	return &Client{
		conn: conn,
		cc: cc,
	}, nil
}

func (c *Client) Register(ctx context.Context, node *pb.Node) (*pb.RegistrationResponse, error) {
	return c.cc.Register(ctx, node)
}