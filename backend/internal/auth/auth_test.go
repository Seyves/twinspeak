package auth

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

const testSecret = "test-hmac-secret-key-minimum-32-characters-long"

func prepare(t *testing.T) (pgx.Tx, *db.Queries) {
	dbUrl := os.Getenv("DB_URL")
	require.NotZero(t, dbUrl)

	pool, err := pgxpool.New(context.Background(), dbUrl)
	require.NoError(t, err)
	queries := db.New(pool)

	tx, err := pool.Begin(t.Context())
	require.NoError(t, err)
	qtx := queries.WithTx(tx)

	return tx, qtx
}

// --- Pure JWT tests (no DB required) ---

func TestValidateAccessToken(t *testing.T) {
	auth := New(testSecret)
	userId := uuid.New()

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	token, err := createAccessToken(now, testSecret, userId)
	require.NoError(t, err)

	_, gotUserId, err := auth.ValidateAccessToken(t.Context(), now, token.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId)
}

func TestValidateAccessToken_Expired(t *testing.T) {
	auth := New(testSecret)
	userId := uuid.New()

	// Mon, 1 Jan 2024 08:00:00 +0000
	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	token, err := createAccessToken(now, testSecret, userId)
	require.NoError(t, err)

	future := now.Add(accessTokenLifetime + time.Second)
	_, _, err = auth.ValidateAccessToken(t.Context(), future, token.Value)
	require.Error(t, err)
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	auth := New(testSecret)
	userId := uuid.New()

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	token, err := createAccessToken(now, "wrong-secret-key-minimum-32-characters-long", userId)
	require.NoError(t, err)

	_, _, err = auth.ValidateAccessToken(t.Context(), now, token.Value)
	require.Error(t, err)
}

func TestValidateWSTicket(t *testing.T) {
	auth := New(testSecret)
	userId := uuid.New()

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	ticket, err := createWSTicket(now, testSecret, userId)
	require.NoError(t, err)

	_, gotUserId, err := auth.ValidateWSTicket(t.Context(), now, ticket.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId)
}

func TestValidateWSTicket_Expired(t *testing.T) {
	auth := New(testSecret)
	userId := uuid.New()

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	ticket, err := createWSTicket(now, testSecret, userId)
	require.NoError(t, err)

	future := now.Add(wsTicketLifetime + time.Second)
	_, _, err = auth.ValidateWSTicket(t.Context(), future, ticket.Value)
	require.Error(t, err)
}

func TestValidateWSTicket_AccessToken(t *testing.T) {
	// A regular access token (no ws claim) should fail WS ticket validation
	auth := New(testSecret)
	userId := uuid.New()

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	accessToken, err := createAccessToken(now, testSecret, userId)
	require.NoError(t, err)

	_, _, err = auth.ValidateWSTicket(t.Context(), now, accessToken.Value)
	require.Error(t, err)
}

func TestIssueWSTicket(t *testing.T) {
	auth := New(testSecret)
	userId := uuid.New()

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	ticket, err := auth.IssueWSTicket(t.Context(), now, userId)
	require.NoError(t, err)
	require.NotNil(t, ticket)
	assert.NotEmpty(t, ticket.Value)
	assert.True(t, ticket.ExpiresAt.After(now))

	_, gotUserId, err := auth.ValidateWSTicket(t.Context(), now, ticket.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId)
}

// --- DB-dependent tests ---

func TestCreateUser(t *testing.T) {
	auth := New(testSecret)
	tx, qtx := prepare(t)
	defer tx.Rollback(t.Context())

	userId, err := auth.CreateUser(t.Context(), qtx, "newuser@gmail.com", "password123")
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, userId)
}

func TestValidatePassword(t *testing.T) {
	auth := New(testSecret)
	tx, qtx := prepare(t)
	defer tx.Rollback(t.Context())

	_, err := auth.CreateUser(t.Context(), qtx, "user@gmail.com", "correctpassword")
	require.NoError(t, err)

	userId, err := auth.ValidatePassword(t.Context(), qtx, "user@gmail.com", "correctpassword")
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, userId)
}

func TestValidatePassword_WrongPassword(t *testing.T) {
	auth := New(testSecret)
	tx, qtx := prepare(t)
	defer tx.Rollback(t.Context())

	_, err := auth.CreateUser(t.Context(), qtx, "user@gmail.com", "correctpassword")
	require.NoError(t, err)

	_, err = auth.ValidatePassword(t.Context(), qtx, "user@gmail.com", "wrongpassword")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestValidatePassword_NoUser(t *testing.T) {
	auth := New(testSecret)
	tx, qtx := prepare(t)
	defer tx.Rollback(t.Context())

	_, err := auth.ValidatePassword(t.Context(), qtx, "nobody@gmail.com", "password")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestValidatePassword_NilHash(t *testing.T) {
	// Google OAuth users with no password hash should return ErrInvalidCredentials
	auth := New(testSecret)
	tx, qtx := prepare(t)
	defer tx.Rollback(t.Context())

	_, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "google@gmail.com",
		PasswordHash: nil,
	})
	require.NoError(t, err)

	_, err = auth.ValidatePassword(t.Context(), qtx, "google@gmail.com", "anypassword")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestStartSession(t *testing.T) {
	auth := New(testSecret)
	tx, qtx := prepare(t)
	defer tx.Rollback(t.Context())

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	userId, err := auth.CreateUser(t.Context(), qtx, "user@gmail.com", "anypassword")
	require.NoError(t, err)

	accessToken, refreshToken, err := auth.StartSession(t.Context(), qtx, userId, now)
	require.NoError(t, err)
	require.NotNil(t, accessToken)
	require.NotNil(t, refreshToken)
	assert.NotEmpty(t, accessToken.Value)
	assert.NotEmpty(t, refreshToken.Value)
}

func TestValidateSession(t *testing.T) {
	auth := New(testSecret)
	tx, qtx := prepare(t)
	defer tx.Rollback(t.Context())

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	userId, err := auth.CreateUser(t.Context(), qtx, "user@gmail.com", "anypassword")
	require.NoError(t, err)

	_, refreshToken, err := auth.StartSession(t.Context(), qtx, userId, now)
	require.NoError(t, err)

	gotUserId, err := auth.ValidateSession(t.Context(), qtx, refreshToken.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId)
}

func TestRevokeSession(t *testing.T) {
	auth := New(testSecret)
	tx, qtx := prepare(t)
	defer tx.Rollback(t.Context())

	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	userId, err := auth.CreateUser(t.Context(), qtx, "user@gmail.com", "anypassword")
	require.NoError(t, err)

	_, refreshToken, err := auth.StartSession(t.Context(), qtx, userId, now)
	require.NoError(t, err)

	err = auth.RevokeSession(t.Context(), qtx, refreshToken.Value)
	require.NoError(t, err)

	_, err = auth.ValidateSession(t.Context(), qtx, refreshToken.Value)
	require.ErrorIs(t, err, ErrMaliciousSuspicion)
}
