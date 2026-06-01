package billing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twinspeak/backend/internal/db"
)

// All tests use this initial time: Mon, 1 Jan 2024 08:00:00 +0000

func prepare(t *testing.T) (pgx.Tx, *db.Queries, uuid.UUID) {
	dbUrl := os.Getenv("DB_URL")
	require.NotZero(t, dbUrl)

	pool, err := pgxpool.New(context.Background(), dbUrl)
	require.NoError(t, err)
	queries := db.New(pool)

	tx, err := pool.Begin(t.Context())
	require.NoError(t, err)
	qtx := queries.WithTx(tx)

	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@gmail.com",
		PasswordHash: []byte{},
	})
	require.NoError(t, err)

	return tx, qtx, userId
}

func TestStartSubscription(t *testing.T) {
	billing := New()
	tx, qtx, userId := prepare(t)
	defer tx.Rollback(t.Context())

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	err := billing.StartSubscription(t.Context(), qtx, userId, now)
	require.NoError(t, err)

	err = billing.SpendCredits(t.Context(), qtx, userId, now, MaxCreditsPerSession)
	require.NoError(t, err)
}

func TestExpiredSubscription(t *testing.T) {
	billing := New()
	tx, qtx, userId := prepare(t)
	defer tx.Rollback(t.Context())

	// Mon, 1 Jan 2024 08:00:00 +0000
	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	err := billing.StartSubscription(t.Context(), qtx, userId, now)
	require.NoError(t, err)

	// Mon, 1 Feb 2024 07:00:00 +0000
	now = time.Date(2024, time.February, 1, 7, 0, 0, 0, time.UTC)
	err = billing.SpendCredits(t.Context(), qtx, userId, now, MaxCreditsPerSession)
	require.NoError(t, err)

	// Mon, 1 Feb 2024 08:00:00 +0000
	now = time.Date(2024, time.February, 1, 8, 0, 0, 0, time.UTC)
	err = billing.SpendCredits(t.Context(), qtx, userId, now, MaxCreditsPerSession)
	require.Error(t, err)
}

func TestRenewSubscription(t *testing.T) {
	billing := New()
	tx, qtx, userId := prepare(t)
	defer tx.Rollback(t.Context())

	// Mon, 1 Jan 2024 08:00:00 +0000
	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	err := billing.StartSubscription(t.Context(), qtx, userId, now)
	require.NoError(t, err)

	// Mon, 1 Feb 2024 08:00:00 +0000
	now = time.Date(2024, time.February, 1, 8, 0, 0, 0, time.UTC)
	err = billing.SpendCredits(t.Context(), qtx, userId, now, MaxCreditsPerSession)
	require.Error(t, err)

	expired, err := billing.GetExpiredSubscriptions(t.Context(), qtx, now)
	require.NoError(t, err)

	require.Len(t, expired, 1)
	assert.Contains(t, expired, userId)

	err = billing.RenewSubscription(t.Context(), qtx, userId, now)
	require.NoError(t, err)

	err = billing.SpendCredits(t.Context(), qtx, userId, now, MaxCreditsPerSession)
	require.NoError(t, err)
}

func TestSpendAllSubscription(t *testing.T) {
	billing := New()
	tx, qtx, userId := prepare(t)
	defer tx.Rollback(t.Context())

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	err := billing.StartSubscription(t.Context(), qtx, userId, now)
	require.NoError(t, err)

	for range MonthlyCredits / MaxCreditsPerSession {
		err = billing.SpendCredits(t.Context(), qtx, userId, now, MaxCreditsPerSession)
		require.NoError(t, err)
	}
	err = billing.SpendCredits(t.Context(), qtx, userId, now, MaxCreditsPerSession)
	require.Error(t, err)
}

func TestSpendCreditsNoGrants(t *testing.T) {
	billing := New()
	tx, qtx, userId := prepare(t)
	defer tx.Rollback(t.Context())

	err := billing.SpendCredits(t.Context(), qtx, userId, time.Now(), MaxCreditsPerSession)
	require.ErrorIs(t, err, ErrInsufficientCredits)
}

func TestBuyTopup(t *testing.T) {
	billing := New()
	tx, qtx, userId := prepare(t)
	defer tx.Rollback(t.Context())

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	err := billing.StartSubscription(t.Context(), qtx, userId, now)
	require.NoError(t, err)

	for range MonthlyCredits / MaxCreditsPerSession {
		err = billing.SpendCredits(t.Context(), qtx, userId, now, MaxCreditsPerSession)
		require.NoError(t, err)
	}
	err = billing.SpendCredits(t.Context(), qtx, userId, now, MaxCreditsPerSession)
	require.Error(t, err)

	err = billing.BuyTopup(t.Context(), qtx, userId, now, 20, now.AddDate(0, 1, 0))
	require.NoError(t, err)

	err = billing.SpendCredits(t.Context(), qtx, userId, now, 20)
	require.NoError(t, err)

	err = billing.SpendCredits(t.Context(), qtx, userId, now, 20)
	require.Error(t, err)
}
