package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStateStore struct {
	pool *pgxpool.Pool
}

func NewPGStateStore(ctx context.Context, databaseURL string) (*PGStateStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PGStateStore{pool: pool}, nil
}

func (s *PGStateStore) Load(ctx context.Context) ([]byte, error) {
	var payload []byte
	err := s.pool.QueryRow(ctx, `SELECT payload::text FROM app_runtime_state WHERE singleton = TRUE`).Scan(&payload)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return payload, nil
}

func (s *PGStateStore) Save(ctx context.Context, payload []byte) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO app_runtime_state (singleton, payload, updated_at)
		VALUES (TRUE, $1::jsonb, NOW())
		ON CONFLICT (singleton)
		DO UPDATE SET payload = EXCLUDED.payload, updated_at = NOW()
	`, string(payload))
	return err
}

func (s *PGStateStore) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
