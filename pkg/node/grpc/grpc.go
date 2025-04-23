package grpc

import (
	"context"
	"os"

	grpc_ "github.com/axelburling/dfs/internal/grpc"
	"github.com/axelburling/dfs/internal/log"
	internalnode "github.com/axelburling/dfs/internal/node"
	"github.com/axelburling/dfs/pkg/node/grpc/pb"
	"github.com/axelburling/dfs/pkg/node/storage"
	"google.golang.org/grpc/codes"
)

type Service struct {
	pb.UnimplementedNodeServiceServer
	
	log *log.Logger
	err *grpc_.Error
	ds *internalnode.DiskSpace
	storage storage.Storage
	statusChan chan bool

	hostname *string
	readonly bool
}

func (s *Service) DiskSpace(ctx context.Context, req *pb.Empty) (*pb.DiskSpaceResponse, error) {
	return &pb.DiskSpaceResponse{
		Total: s.ds.Total(),
		Free: s.ds.Free(),
	}, nil
}

func (s *Service) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	if s.hostname == nil {
		return &pb.HealthResponse{}, s.err.New(codes.Internal, "failed to get storage node hostname", nil)
	}

	if req.ReadOnly != nil {
		s.readonly = *req.ReadOnly
		s.statusChan <- *req.ReadOnly
	}

	return &pb.HealthResponse{
		Hostname: *s.hostname,
		Alive: true,
		ReadOnly: s.readonly,
	}, nil
}

func (s *Service) Init(ctx context.Context, req *pb.Empty) (*pb.InitResponse, error) {
	chunks, err := s.storage.GetChunks()

	if err != nil {
		return &pb.InitResponse{}, s.err.New(codes.Internal, "failed to get chunks", err)
	}

	var grpcChunks []*pb.Chunk

	for _, chunk := range *chunks {
		grpcChunks = append(grpcChunks, &pb.Chunk{
			Id: chunk.ID,
			Hash: chunk.Hash,
			Size: chunk.Size,
		})
	}

	return &pb.InitResponse{Chunks: grpcChunks, Health: &pb.HealthResponse{
		Hostname: *s.hostname, 
		Alive: true, 
		ReadOnly: s.readonly,
	}}, nil
}

func NewNodeService(log *log.Logger, ds *internalnode.DiskSpace, storage storage.Storage, statusChan chan bool) pb.NodeServiceServer {
	var hn *string
	hostname, err := os.Hostname()

	if err != nil {
		hn = nil
	} else {
		hn = &hostname
	}

	errSrv := grpc_.NewError(log)

	return &Service{
		log: log,
		err: errSrv,
		ds: ds,
		hostname: hn,
		storage: storage,
		statusChan: statusChan,
	}
}