package node

import (
	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/db/generated"
	"github.com/axelburling/dfs/pkg/db/utils"
	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/message"
	ms "github.com/axelburling/dfs/pkg/pubsub/client/msg"
	"go.uber.org/zap"
)

type Node struct {
	log         *log.Logger
	conv        utils.Conv
	nodeAddChan chan generated.Node
}

func NewNode(log *log.Logger, conv utils.Conv, nodeAddChan chan generated.Node) *Node {
	return &Node{
		log:         log,
		conv:        conv,
		nodeAddChan: nodeAddChan,
	}
}

func (n *Node) HandleNodeAction(action ms.Action, msg *message.Message) {
	switch action {
	case ms.NodeAdd:
		n.HandleNodeAdd(msg)
	case ms.NodeUpdate:

	}
}

func (n *Node) HandleNodeAdd(msg *message.Message) {
	message, err := decode[ms.NodeAddReq](n.log, msg)

	if err != nil {
		return
	}

	node, err := n.conv.Node.ConvertPubSub(message)

	if err != nil {
		n.log.Warn("could not convert to database object", zap.Error(err))
		return
	}

	n.nodeAddChan <- *node
}

func decode[T any](log *log.Logger, msg *message.Message) (*T, error) {
	message, err := ms.Decode[T](msg)
	if err != nil {
		log.Warn("could not decode pubsub message", zap.String("action", msg.Action), zap.Error(err))
		return nil, err
	}

	return message, nil
}
