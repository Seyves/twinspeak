package metrics

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twinspeak/backend/db"
)

type Metrics struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func (m *Metrics) CreateHttpRequestMetric(ctx context.Context, params db.InsertHttpRequestParams) error {
	err := m.queries.InsertHttpRequest(ctx, params)
	if err != nil {
		return fmt.Errorf("cannot insert http request metric into db: %w", err)
	}
	return nil
}

func (m *Metrics) CreateSpeechMetric(ctx context.Context, params db.InsertSpeechParams) error {
	err := m.queries.InsertSpeech(ctx, params)
	if err != nil {
		return fmt.Errorf("cannot insert speech metric into db: %w", err)
	}
	return nil
}

func NewMetrics(db *pgxpool.Pool, queries *db.Queries) *Metrics {
	return &Metrics{
		db:      db,
		queries: queries,
	}
}
