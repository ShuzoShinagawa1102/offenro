package postgres

import (
	"context"
	"fmt"
	"strings"

	postgresdb "github.com/ShuzoShinagawa1102/offenro/internal/platform/postgres/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Store struct {
	pool    transactionStarter
	queries *postgresdb.Queries
}

func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

// NewTransactionalStore is useful for integration tests that roll back every
// write while still exercising Store-managed nested transactions.
func NewTransactionalStore(tx pgx.Tx) *Store {
	return &Store{pool: tx, queries: postgresdb.New(tx)}
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		queries: postgresdb.New(pool),
	}
}
