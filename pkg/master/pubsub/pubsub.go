package pubsub

import (
	"context"
	"os"

	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/db/generated"
	"github.com/axelburling/dfs/pkg/db/utils"
	"github.com/axelburling/dfs/pkg/master/pubsub/node"
	"github.com/axelburling/dfs/pkg/pubsub/client"
	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/message"
	ms "github.com/axelburling/dfs/pkg/pubsub/client/msg"
)

type Internal struct {
	ctx context.Context
	log *log.Logger

	pubsub client.PubSubInfo
	conv   utils.Conv

	node *node.Node
}

func NewInternal(ctx context.Context, log *log.Logger, pubsub client.PubSubInfo, conv utils.Conv, nodeAddChan chan generated.Node) *Internal {
	n := node.NewNode(log, conv, nodeAddChan)
	return &Internal{
		ctx:    ctx,
		log:    log,
		pubsub: pubsub,
		conv:   conv,
		node:   n,
	}
}

func (i *Internal) ReceiveMessage(stop chan os.Signal) {
	if i.pubsub.MessageChan == nil {
		i.pubsub.MessageChan = make(chan *message.Message, 100)
	}

	go func() {
		for {
			select {
			case <-stop:
				close(i.pubsub.MessageChan) // Close channel on stop signal
				return
			default:
				if i.pubsub.Subscription != nil { // Ensure subscription is set
					i.pubsub.Subscription.Receive(i.ctx, i.pubsub.MessageChan, nil)
				}
			}
		}
	}()
}

func (i *Internal) MonitorMessage(stop chan os.Signal) {
	if i.pubsub.MessageChan == nil {
		return
	}

	go func() {
		for {
			select {
			case <-stop:
				return
			case msg := <-i.pubsub.MessageChan:
				switch action := ms.Action(msg.Action); action {
				// Handle All Node actions
				case ms.NodeAdd, ms.NodeUpdate, ms.NodeDelete:
					i.node.HandleNodeAction(action, msg)
				}

			}
		}
	}()
}
