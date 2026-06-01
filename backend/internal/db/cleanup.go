package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const cleanup = `TRUNCATE TABLE 
	users, refresh_sessions, subscriptions,
	credit_grants, credit_expenses, speeches,
	http_requests
	CASCADE;
`

// ONLY FOR TESTS PURPOSES
func Cleanup(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, cleanup); err != nil {
		return fmt.Errorf("failed to cleanup db: %w", err)
	}
	return nil
}
