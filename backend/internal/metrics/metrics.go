package metrics

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twinspeak/backend/internal/db"
)

type Module struct{}

func (m *Module) CreateHttpRequestMetric(ctx context.Context, tx *db.Queries, params db.InsertHttpRequestParams) error {
	err := tx.InsertHttpRequest(ctx, params)
	if err != nil {
		return fmt.Errorf("cannot insert http request metric into db: %w", err)
	}
	return nil
}

func (m *Module) CreateSpeechMetric(ctx context.Context, tx *db.Queries, params db.InsertSpeechParams) error {
	err := tx.InsertSpeech(ctx, params)
	if err != nil {
		return fmt.Errorf("cannot insert speech metric into db: %w", err)
	}
	return nil
}

func New(db *pgxpool.Pool, queries *db.Queries) *Module {
	return &Module{}
}
