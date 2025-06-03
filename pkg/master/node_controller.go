package master

import (
	"context"
	"os"
	"slices"
	"time"

	"github.com/axelburling/dfs/pkg/db/generated"
	"github.com/axelburling/dfs/pkg/node/client"
	"go.uber.org/zap"
)

func (m *Master) monitorAddNode(stop chan os.Signal) {
	if m.nodeAddChan == nil {
		m.nodeAddChan = make(chan generated.Node, 100)
	}

	go func() {
		for {
			select {
			case <-stop:
				close(m.nodeAddChan)
				return
			case n := <-m.nodeAddChan:
				m.log.Info("Adding node", zap.Any("node id", n.ID))

				var node *client.Node
				var err error
				maxRetries := 5
				for i := range maxRetries {
					node, err = client.NewNode(n.Address, n.GrpcAddress)
					if err == nil {
						break // Successfully connected
					}
					m.log.Warn("Failed to connect to node, retrying...", zap.Int("attempt", i+1), zap.Error(err))
					time.Sleep(2 * time.Second) // Wait before retrying
				}

				if err != nil {
					m.log.Error("Failed to connect after retries, marking node as unhealthy", zap.String("address", n.Address))
					m.db.DeleteNode(context.Background(), n.ID)
					continue
				}

				m.mu.Lock()
				m.nodes = append(m.nodes, node) // Add node even if unhealthy
				m.mu.Unlock()
				m.uploadQueue.AddNode(node)

				m.log.Info("Successfully added node", zap.String("node id", node.ID))
			}
		}
	}()
}

func (m *Master) monitorNodeHeartbeat(stop chan os.Signal) {
	go func() {
		ticker := time.NewTicker(time.Second*1)

		select {
		case <-stop:
			return
		case <-ticker.C:
			for _, node := range  m.nodes {
				res, err := node.Health(m.ctx, nil)

				if err != nil || !res.Alive {
					m.log.Warn("trouble connecting to node retrying")

					m.reQueueNode(node)
				}
			}
		}
	}()
}

func (m *Master) reQueueNode(node *client.Node) {
	index := 0
	for i, n := range m.nodes {
		if n.ID == node.ID {
			index = i
		}
	}

	m.nodes = slices.Delete(m.nodes, index, index+1)

	m.uploadQueue.MarkNode(node, false)
}