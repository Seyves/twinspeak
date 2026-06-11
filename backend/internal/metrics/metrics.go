package metrics

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (m *Module) GetSpeeches(ctx context.Context, tx *db.Queries, userId uuid.UUID) ([]db.Speech, error) {
	speeches, err := tx.GetSpeeches(ctx, db.GetSpeechesParams{
		UserID: userId,
		Limit: 20,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []db.Speech{}, nil
		}
		return []db.Speech{}, fmt.Errorf("cannot select speeche metrics from db: %w", err)
	}
	return speeches, nil
}

func New(db *pgxpool.Pool, queries *db.Queries) *Module {
	return &Module{}
}
