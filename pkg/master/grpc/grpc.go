package grpc

import (
	"context"
	"database/sql"
	"errors"
	"os"

	grpc_ "github.com/axelburling/dfs/internal/grpc"
	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/db"
	"github.com/axelburling/dfs/pkg/db/generated"
	"github.com/axelburling/dfs/pkg/master/grpc/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

type Service struct {
	pb.UnimplementedMasterServiceServer

	log *log.Logger
	err *grpc_.Error
	db  *db.DB

	stopChan chan os.Signal
	nodeAddChan chan generated.Node
}

func NewMasterService(log *log.Logger, db *db.DB,  nodeAddChan chan generated.Node) pb.MasterServiceServer {
	return &Service{
		log: log,
		err: grpc_.NewError(log),
		nodeAddChan: nodeAddChan,
		db: db,
	}
}

func (s *Service) Register(ctx context.Context, req *pb.Node) (*pb.RegistrationResponse, error) {
	node, err := s.addNode(ctx, req)

	if err != nil {
		return nil, s.err.New(codes.Internal, "could not register node", err)
	}

	return &pb.RegistrationResponse{
		Registered: true,
		NodeId: node.ID.String(),
	}, nil
}

func (s *Service) addNode(ctx context.Context, req *pb.Node) (*generated.Node, error) {
	id, err := uuid.FromBytes(req.Id.Value)

	if err != nil {
		return nil, s.err.New(codes.InvalidArgument, "node id is not a valid uuid", err)
	}

	node, err := s.db.GetNodeByID(ctx, s.db.Types.UUID.ConvertToUUID(id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dbNode, err := s.db.InsertNode(ctx, generated.InsertNodeParams{
				ID: s.db.Types.UUID.ConvertToUUID(id),
				Address: req.Addr,
				GrpcAddress: req.GrpcAddr,
				Hostname: req.Hostname,
				IsHealthy: s.db.Types.Bool.ConvertToBool(true),
				TotalSpace: req.TotalSpace,
				FreeSpace: req.FreeSpace,
				Readonly: s.db.Types.Bool.ConvertToBool(req.ReadOnly),
			})
		
			if err != nil {
				return nil, err
			}
	
			s.nodeAddChan <- dbNode
			return &dbNode, nil
		} else {
			return nil, s.err.New(codes.Internal, "failed to get node", err)
		}
	}

	s.nodeAddChan <- node
	return &node, nil
}