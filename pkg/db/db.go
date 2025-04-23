package db

import (
	"context"

	"github.com/axelburling/dfs/pkg/db/generated"
	"github.com/axelburling/dfs/pkg/db/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*generated.Queries
	ctx context.Context

	Types utils.Types
}

func New(ctx context.Context, connStr string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connStr)

	if err != nil {
		return nil, err
	}

	q := generated.New(pool)

	_, err = pool.Exec(ctx, "SELECT 1+1")

	if err != nil {
		return nil, err
	}

	return &DB{
		Queries: q,
		ctx:     ctx,
		Types:   utils.Types{},
	}, nil
}

func (db *DB) RegisterNode(ctx context.Context, node *generated.Node) {
	db.InsertNode(ctx, generated.InsertNodeParams{
		Address:     node.Address,
		GrpcAddress: node.GrpcAddress,
		Hostname:    node.Hostname,
		IsHealthy: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
		TotalSpace: node.TotalSpace,
		FreeSpace:  node.FreeSpace,
		Readonly:   node.Readonly,
	})
}
