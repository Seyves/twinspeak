package preferences

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/twinspeak/backend/internal/db"
)

type Module struct{}

func (m *Module) CreatePreferences(ctx context.Context, tx *db.Queries, userId uuid.UUID) error {
	err := tx.InsertUserPrefs(ctx, userId)
	if err != nil {
		return fmt.Errorf("cannot insert preferences into db: %w", err)
	}
	return err
}

func (m *Module) UpdatePreferences(ctx context.Context, tx *db.Queries, params db.UpdateUserPrefsParams) error {
	err := tx.UpdateUserPrefs(ctx, params)
	if err != nil {
		return fmt.Errorf("cannot update preferences in db: %w", err)
	}
	return err
}

func (m *Module) GetPreferences(ctx context.Context, tx *db.Queries, userId uuid.UUID) (*db.Preference, error) {
	prefs, err := tx.GetUserPrefs(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot select preferences from db: %w", err)
	}
	return &prefs, err
}

func (m *Module) SetHideMessagesTimestamp(ctx context.Context, tx *db.Queries, userId uuid.UUID, timestamp time.Time) error {
	err := tx.SetHideMessagesTimestamp(ctx, db.SetHideMessagesTimestampParams{
		UserID:             userId,
		HideMessagesBefore: &timestamp,
	})
	if err != nil {
		return fmt.Errorf("cannot update hide messages in db: %w", err)
	}
	return nil
}
