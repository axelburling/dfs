package utils

import (
	"github.com/axelburling/dfs/pkg/db/generated"
	"github.com/axelburling/dfs/pkg/pubsub/client/msg"
)

type Conv struct {
	Node node

	ty Types
}

type node struct {
	ty Types
}

func NewConv(ty Types) Conv {
	return Conv{
		Node: node{
			ty: ty,
		},
		ty: ty,
	}
}

func (n *node) ConvertPubSub(in *msg.NodeAddReq) (*generated.Node, error) {
	id, err := n.ty.UUID.ConvertToUUIDFromString(in.ID)

	if err != nil {
		return nil, err
	}

	return &generated.Node{
		ID:          id,
		Address:     in.Address,
		GrpcAddress: in.GrpcAddress,
		Hostname:    in.Hostname,
		TotalSpace:  in.TotalSpace,
		FreeSpace:   in.FreeSpace,
		IsHealthy:   n.ty.Bool.ConvertToBool(in.IsHealthy),
	}, nil
}
