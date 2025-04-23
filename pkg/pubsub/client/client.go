package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/message"
	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/pool"
	"github.com/axelburling/dfs/pkg/pubsub/client/msg"
	pb "github.com/axelburling/dfs/pkg/pubsub/grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

type PubSubInfo struct {
	PubSub       *PubSub
	Topic        *Topic
	Subscription *Subscription
	MessageChan  chan *message.Message
}

type PubSub struct {
	pool *pool.GrpcPool
	id   string
}

type Topic struct {
	pool *pool.GrpcPool
	Name string
}

type Subscription struct {
	pool  *pool.GrpcPool
	Name  string
	Topic string
}

func NewClient(target string, size int, id string) (*PubSub, error) {
	p, err := pool.New(target, size, pool.CreateGRPCConn)
	if err != nil {
		return nil, err
	}

	return &PubSub{
		pool: p,
		id:   id,
	}, nil
}

func (p *PubSub) CreateTopic(ctx context.Context, name string) (*Topic, error) {
	conn, err := p.pool.GetStream()
	if err != nil {
		return nil, err
	}

	defer p.pool.ReleaseConnection(conn)

	top, err := conn.PublisherClient.CreateTopic(ctx, &pb.Topic{
		Name:                     name,
		State:                    pb.Topic_STATE_UNSPECIFIED,
		MessageRetentionDuration: durationpb.New(2 * time.Minute),
	})

	if err != nil {
		return nil, err
	}

	return &Topic{Name: top.Name, pool: p.pool}, nil
}

func (t *Topic) Publish(ctx context.Context, msgs []*message.Message, deadlinesDur []time.Duration) ([]string, error) {
	conn, err := t.pool.GetStream()
	if err != nil {
		return nil, err
	}
	defer t.pool.ReleaseConnection(conn)

	messages := make([]*pb.PubsubMessage, 0, len(msgs))

	for _, ms := range msgs {
		messages = append(messages, &pb.PubsubMessage{
			Action:     ms.Action,
			Data:       ms.Data,
			Attributes: ms.Attributes,
		})
	}

	var deadlines []*durationpb.Duration

	if len(deadlinesDur) > 0 {
		if len(deadlinesDur) != len(msgs) {
			return nil, errors.New("length of deadlines must match length of messages or be empty")
		}

		deadlines = make([]*durationpb.Duration, len(deadlinesDur))
		for _, deadline := range deadlinesDur {
			deadlines = append(deadlines, durationpb.New(deadline))
		}
	}

	res, err := conn.PublisherClient.Publish(ctx, &pb.PublishRequest{
		Topic:     t.Name,
		Deadlines: deadlines,
		Messages:  messages,
	})

	if err != nil {
		return nil, err
	}

	if len(res.MessageIds) != len(msgs) {
		return nil, errors.New("something went wrong when publishing")
	}

	return res.MessageIds, nil
}

func (t *Topic) CreateSubscription(ctx context.Context, name string) (*Subscription, error) {
	conn, err := t.pool.GetStream()
	defer t.pool.ReleaseConnection(conn)

	if err != nil {
		return nil, err
	}

	sub, err := conn.SubscriberClient.CreateSubscription(ctx, &pb.Subscription{
		Name:  name,
		Topic: t.Name,
	})

	if err != nil {
		return nil, err
	}

	return &Subscription{Name: sub.Name, Topic: t.Name, pool: t.pool}, nil
}

func (t *Topic) NewMessage(action string, data []byte) *message.Message {
	return message.NewMessage(action, data)
}

func (s *Subscription) Receive(ctx context.Context, messageChan chan *message.Message, filters []string) error {
	ackChan := make(chan string, 500)
	conn, err := s.pool.GetStream()

	if err != nil {
		return err
	}

	defer s.pool.ReleaseConnection(conn)

	actions, err := msg.ParseAction(filters)

	pull, err := conn.SubscriberClient.StreamingPull(ctx)

	if err != nil {
		return err
	}
	resp := &pb.StreamingPullRequest{
		Topic:        s.Topic,
		Subscription: s.Name,
		Filter:       actions,
		AckIds:       []string{},
	}

	pull.Send(resp)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ackID := <-ackChan:
				ackReq := &pb.StreamingPullRequest{AckIds: []string{ackID}}
				if err := pull.Send(ackReq); err != nil {
					fmt.Printf("Error sending acknowledgment: %v\n", err)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			data, err := pull.Recv()

			if err != nil {
				if status.Code(err) == codes.Unavailable {
					return fmt.Errorf("Stream closed by server")
				}

				return fmt.Errorf("Error receiving messages: %v", err)
			}

			for _, msg := range data.ReceivedMessages {
				messageChan <- &message.Message{
					AckID:         *msg.AckId,
					ID:            msg.Message.MessageId,
					Action:        msg.Message.Action,
					Data:          msg.Message.Data,
					PublishedTime: msg.Message.PublishTime.AsTime(),
					AckChan:       ackChan,
				}
			}
		}
	}
}
