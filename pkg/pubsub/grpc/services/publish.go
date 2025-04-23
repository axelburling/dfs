package services

import (
	"context"
	"fmt"
	"time"

	internalgrpc "github.com/axelburling/dfs/internal/grpc"
	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/pubsub/apiv1/message"
	"github.com/axelburling/dfs/pkg/pubsub/apiv1/scheduler"
	pb "github.com/axelburling/dfs/pkg/pubsub/grpc/pb"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/durationpb"
)

type PublisherService struct {
	pb.UnimplementedPublisherServer

	Scheduler *scheduler.Scheduler
	log       *log.Logger

	err *internalgrpc.Error
}

func NewPublisherService(scheduler *scheduler.Scheduler, log *log.Logger, errSrv *internalgrpc.Error) pb.PublisherServer {
	return &PublisherService{
		Scheduler: scheduler,
		log:       log,
		err:       errSrv,
	}
}

func (p *PublisherService) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {

	if p.Scheduler.Empty(req.Topic) {
		return nil, p.err.New(codes.Internal, "there are no active subscribers", nil)
	}

	var messages []*message.Message
	ids := make([]string, 0, len(req.Messages))

	if len(req.Deadlines) != 0 && (len(req.Deadlines) != len(req.Messages)) {
		return nil, p.err.New(codes.InvalidArgument, "the deadline array needs to the same length as the message array or 0", nil)
	}

	for i, msg := range req.Messages {
		var deadline time.Duration = 10 * time.Second // Default deadline

		// If deadlines are provided, use them
		if len(req.Deadlines) > 0 {
			deadline = req.Deadlines[i].AsDuration()

			// Cap deadline to max 5 minutes
			if deadline > 5*time.Minute {
				deadline = 5 * time.Minute
			}
		}

		id := uuid.New().String()
		ms := message.NewMessage(id,
			uuid.NewString(),
			msg.Action,
			req.Sender,
			msg.Data,
			msg.Attributes,
			deadline,
		)
		messages = append(messages, ms)
		ids = append(ids, id)
	}

	err := p.Scheduler.Publish(req.Topic, messages)

	if err != nil {
		return nil, p.err.New(codes.Internal, fmt.Sprintf("could not publish for topic: %s", req.Topic), err)
	}

	return &pb.PublishResponse{
		MessageIds: ids,
	}, nil
}

func (p *PublisherService) CreateTopic(ctx context.Context, req *pb.Topic) (*pb.Topic, error) {
	p.log.Debug("creating topic", zap.String("topic", req.Name))
	_, err := p.Scheduler.CreateTopic(req.Name)

	if err != nil {
		return nil, p.err.New(codes.Internal, fmt.Sprintf("could not create topic: %s", req.Name), err)
	}

	return &pb.Topic{
		Name:                     req.Name,
		State:                    pb.Topic_STATE_UNSPECIFIED,
		MessageRetentionDuration: durationpb.New(2 * time.Hour),
	}, nil
}
