package services

import (
	"context"
	"errors"
	"time"

	internalgrpc "github.com/axelburling/dfs/internal/grpc"
	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/pubsub/apiv1/message"
	"github.com/axelburling/dfs/pkg/pubsub/apiv1/scheduler"
	"github.com/axelburling/dfs/pkg/pubsub/apiv1/wildcard"
	pb "github.com/axelburling/dfs/pkg/pubsub/grpc/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const MAX_WAIT_TIME_SEC = 10

type SubscriberService struct {
	pb.UnimplementedSubscriberServer

	Scheduler *scheduler.Scheduler
	log       *log.Logger
	err       *internalgrpc.Error
}

func NewSubscriberService(scheduler *scheduler.Scheduler, log *log.Logger, errSrv *internalgrpc.Error) pb.SubscriberServer {
	return &SubscriberService{
		Scheduler: scheduler,
		log:       log,
		err:       errSrv,
	}
}

func (s *SubscriberService) Pull(ctx context.Context, req *pb.PullRequest) (*pb.PullResponse, error) {
	var messages []*pb.ReceivedMessage
	if req.ReturnImmediately != nil && *req.ReturnImmediately == true {
		msgs, err := s.Scheduler.GetMessages(req.Topic, req.Subscription, int(req.MaxMessages))

		if err != nil {
			return nil, s.err.New(codes.NotFound, "could not get messages", err)
		}

		for _, msg := range msgs {
			mg := pb.ReceivedMessage{
				AckId: &msg.AckID,
				Message: &pb.PubsubMessage{
					MessageId:   msg.ID,
					Data:        msg.Data,
					PublishTime: timestamppb.New(msg.PublishTime),
					Attributes:  make(map[string]string),
				},
			}

			messages = append(messages, &mg)
		}

		return &pb.PullResponse{
			ReceivedMessages: messages,
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, MAX_WAIT_TIME_SEC*time.Second)
	defer cancel()

	var mss []*message.Message

	for {
		time.Sleep(500 * time.Millisecond)
		msgs, err := s.Scheduler.GetMessages(req.Topic, req.Subscription, int(req.MaxMessages))

		if err != nil {
			return nil, s.err.New(codes.NotFound, "could not get messages", err)
		}

		mss = append(mss, msgs...)

		select {
		case <-ctx.Done():
			var messages []*pb.ReceivedMessage
			for _, msg := range mss {
				mg := pb.ReceivedMessage{
					AckId: &msg.AckID,
					Message: &pb.PubsubMessage{
						MessageId:   msg.ID,
						Data:        msg.Data,
						PublishTime: timestamppb.New(msg.PublishTime),
						Attributes:  make(map[string]string),
					},
				}

				messages = append(messages, &mg)
			}
			return &pb.PullResponse{
				ReceivedMessages: messages,
			}, nil

		default:
			// Continue the loop if the context has not expired
		}
	}
}

func (s *SubscriberService) Acknowledge(ctx context.Context, req *pb.AcknowledgeRequest) (*emptypb.Empty, error) {
	err := s.Scheduler.AckMessages(req.Topic, req.Subscription, req.AckIds)

	if err != nil {
		return nil, s.err.New(codes.Internal, err.Error(), err)
	}

	return &emptypb.Empty{}, nil
}

func (s *SubscriberService) StreamingPull(srv grpc.BidiStreamingServer[pb.StreamingPullRequest, pb.StreamingPullResponse]) error {
	ctx, cancel := context.WithCancel(srv.Context())
	defer cancel()
	var (
		subscription   string
		topic          string
		isFirstRequest = true
		messageChan    = make(chan *pb.StreamingPullResponse, 10)
		filters        = [][]byte{}
		sender         = ""
	)

	defer close(messageChan)

	// Goroutine: Sends messages to the client
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.log.Debug("Stopping message sender", zap.String("topic", topic), zap.String("subscription", subscription))
				return
			case response := <-messageChan:
				if response != nil {
					if err := srv.Send(response); err != nil {
						s.log.Warn("Failed to send message", zap.String("topic", topic), zap.String("subscription", subscription), zap.Error(err))
						return
					}

				}
			}
		}
	}()

	// Goroutine: Fetches messages from the scheduler
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond) // Adjust interval as needed
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.log.Debug("Stopping message fetcher", zap.String("topic", topic), zap.String("subscription", subscription))
				return
			case <-ticker.C:
				if !isFirstRequest {
					messages, err := s.Scheduler.GetMessages(topic, subscription, 10)
					if err != nil {
						s.log.Warn("Failed to fetch messages", zap.String("topic", topic), zap.String("subscription", subscription), zap.Error(err))
						return
					}

					if len(messages) > 0 {
						var receivedMessages []*pb.ReceivedMessage
						unUsedMessages := make([]string, 0)
						for _, msg := range messages {
							if msg.Sender == sender {
								unUsedMessages = append(unUsedMessages, msg.ID)
								continue
							}


							if len(filters) == 0 {
								receivedMessages = append(receivedMessages, convertMessage(msg))
								continue
							}

							for _, filter := range filters {
								if wildcard.Matches(filter, msg.Action) {
									receivedMessages = append(receivedMessages, convertMessage(msg))
								} else {
									unUsedMessages = append(unUsedMessages, msg.ID)
								}
							}
						}

						if len(receivedMessages) > 0 {
							if len(unUsedMessages) > 0 {
								s.Scheduler.AckMessages(topic, subscription, unUsedMessages)
							}

							select {
							case messageChan <- &pb.StreamingPullResponse{ReceivedMessages: receivedMessages}:
							default: // Avoid blocking
								s.log.Warn("Message queue full, dropping messages")
							}
						}
					}
				}
			}
		}
	}()

	// Main loop: Receives messages from the client
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Client disconnected", zap.String("topic", topic), zap.String("subscription", subscription))

			s.Scheduler.RemoveSubscription(topic, subscription)
			if sub, _ := s.Scheduler.GetSubscription(topic, subscription); sub != nil {
				sub.Active = false
			}

			return ctx.Err()
		default:
			req, err := srv.Recv()
			if err != nil {
				s.Scheduler.RemoveSubscription(topic, subscription)
				if status.Code(err) == codes.Canceled {
					s.log.Debug("Client closed the stream", zap.String("topic", topic), zap.String("subscription", subscription))
					cancel()
					return nil
				}

				return s.err.New(codes.Internal, "failed to receive request", err)
			}

			if isFirstRequest {
				if req.Subscription == "" || req.Topic == "" {
					return s.err.New(codes.InvalidArgument, "subscription and topic must be provided in the first request", nil)
				}

				subscription = req.Subscription
				topic = req.Topic
				for _, fi := range req.Filter {
					filters = append(filters, []byte(fi))
				}

				if req.IgnoreFrom != nil {
					sender = *req.IgnoreFrom
				}

				isFirstRequest = false

				if _, err := s.Scheduler.GetTopic(topic); errors.Is(err, scheduler.ErrTopicCreate) {
					return s.err.New(codes.NotFound, "topic does not exist", err)
				}

				sub, err := s.Scheduler.GetSubscription(topic, subscription)
				if errors.Is(err, scheduler.ErrSubscriptionGet) {
					return s.err.New(codes.NotFound, "subscription does not exist", err)
				}

				if sub.Active {
					return s.err.New(codes.AlreadyExists, "Subscription already active", nil)
				}

				s.log.Info("Streaming Pull started", zap.String("topic", topic), zap.String("subscription", subscription))
				continue
			}

			if len(req.AckIds) > 0 {
				if err := s.Scheduler.AckMessages(topic, subscription, req.AckIds); err != nil {
					s.log.Warn("Failed to acknowledge messages", zap.String("topic", topic), zap.String("subscription", subscription), zap.Error(err))
					return s.err.New(codes.Internal, "failed to acknowledge messages", err)
				}
			}
		}
	}
}

func (s *SubscriberService) CreateSubscription(ctx context.Context, req *pb.Subscription) (*pb.Subscription, error) {
	_, err := s.Scheduler.GetTopic(req.Topic)

	if errors.Is(err, scheduler.ErrTopicNotExist) {
		return nil, status.Error(codes.AlreadyExists, "Topic already exists")
	}

	_, err = s.Scheduler.GetSubscription(req.Topic, req.Name)

	if errors.Is(err, scheduler.ErrMessagesGet) {
		return nil, status.Error(codes.AlreadyExists, "Topic already exists")
	}

	_, err = s.Scheduler.CreateSubscription(req.Topic, req.Name)

	if err != nil {
		return nil, status.Error(codes.Internal, "Could not create sub")
	}

	return req, nil
}

func convertMessage(msg *message.Message) *pb.ReceivedMessage {
	return &pb.ReceivedMessage{
		AckId: &msg.AckID,
		Message: &pb.PubsubMessage{
			MessageId:   msg.ID,
			Data:        msg.Data,
			Action:      msg.Action,
			PublishTime: timestamppb.New(msg.PublishTime),
			Attributes:  make(map[string]string),
		},
	}
}
