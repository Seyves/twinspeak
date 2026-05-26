package billing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/twinspeak/backend/auth"
	"github.com/twinspeak/backend/db"
)

type testApp struct {
	Billing *Billing
	Auth    *auth.Auth
}

func setupTestApp(t *testing.T) (testApp, func(context.Context)) {
	dbUrl := os.Getenv("DB_URL")
	require.NotZero(t, dbUrl)
	pool, err := pgxpool.New(context.Background(), dbUrl)
	require.NoError(t, err)
	queries := db.New(pool)

	billing := NewBilling(pool, queries)
	auth := auth.NewAuth(pool, queries, "secret")

	cleanup := func(ctx context.Context) {
		truncate := `TRUNCATE TABLE http_requests, speeches, credit_expenses, credit_grants, subscriptions, refresh_sessions, users CASCADE;`
		if _, err := pool.Exec(ctx, truncate); err != nil {
			t.Fatalf("failed to cleanup db: %v", err)
		}
	}

	return testApp{
		Billing: billing,
		Auth:    auth,
	}, cleanup
}

func TestSpendCredits(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup(t.Context())

	now := time.Now()
	accessToken, _, err := app.Auth.SignUp(t.Context(), now, "test@gmail.com", "password", "", nil)
	require.NoError(t, err)

	_, userId, err := app.Auth.ValidateAccessToken(t.Context(), now, accessToken.Value)
	require.NoError(t, err)

	err = app.Billing.BuyMonthly(t.Context(), now, userId, 100, now.Add(time.Hour))
	require.NoError(t, err)

	err = app.Billing.SpendCredits(t.Context(), now, userId, MaxCreditsPerSession)
	require.NoError(t, err)

	err = app.Billing.SpendCredits(t.Context(), now, userId, MaxCreditsPerSession)
	require.NoError(t, err)

	err = app.Billing.SpendCredits(t.Context(), now, userId, MaxCreditsPerSession)
	require.ErrorIs(t, err, ErrInsufficientCredits)
}

func TestSpendCreditsNoGrants(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup(t.Context())

	accessToken, _, err := app.Auth.SignUp(t.Context(), time.Now(), "test@gmail.com", "password", "", nil)
	require.NoError(t, err)

	_, userId, err := app.Auth.ValidateAccessToken(t.Context(), time.Now(), accessToken.Value)
	require.NoError(t, err)

	err = app.Billing.SpendCredits(t.Context(), time.Now(), userId, 20)
	require.ErrorIs(t, err, ErrInsufficientCredits)
}

func TestSpendCreditsAllGrantsExpired(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup(t.Context())

	now := time.Now()
	accessToken, _, err := app.Auth.SignUp(t.Context(), now, "test@gmail.com", "password", "", nil)
	require.NoError(t, err)

	_, userId, err := app.Auth.ValidateAccessToken(t.Context(), now, accessToken.Value)
	require.NoError(t, err)

	err = app.Billing.BuyMonthly(t.Context(), now, userId, 200, now.Add(time.Hour))
	require.NoError(t, err)

	now = now.Add(time.Hour + time.Minute)
	err = app.Billing.SpendCredits(t.Context(), now, userId, 20)
	require.ErrorIs(t, err, ErrInsufficientCredits)
}

