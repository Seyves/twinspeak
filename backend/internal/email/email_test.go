package email

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twinspeak/backend/internal/db"
)

func prepare(t *testing.T) (pgx.Tx, *db.Queries, *Module) {
	dbUrl := os.Getenv("DB_URL")
	require.NotZero(t, dbUrl)

	pool, err := pgxpool.New(context.Background(), dbUrl)
	require.NoError(t, err)
	queries := db.New(pool)

	tx, err := pool.Begin(t.Context())
	require.NoError(t, err)
	qtx := queries.WithTx(tx)

	// Create module with dummy values (won't actually send emails in tests)
	module, err := New("test-api-key", "test@twinspeak.com", "https://test.twinspeak.com")
	require.NoError(t, err)

	return tx, qtx, module
}

// --- Pure token generation tests (no DB required) ---

func TestGenerateToken(t *testing.T) {
	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	lifetime := time.Hour * 24

	token, err := GenerateToken(now, lifetime)
	require.NoError(t, err)
	require.NotNil(t, token)

	// Check token value is not empty
	assert.NotEmpty(t, token.Value)

	// Check token hash is generated
	assert.NotEqual(t, [32]byte{}, token.Hash)

	// Check expiration is correct
	expectedExpiry := now.Add(lifetime)
	assert.Equal(t, expectedExpiry, token.ExpiresAt)
}

func TestGenerateToken_UniquenessExpectation(t *testing.T) {
	// Generate multiple tokens and verify they're different
	now := time.Date(2024, time.January, 1, 8, 0, 0, 0, time.UTC)
	lifetime := time.Hour * 24

	token1, err := GenerateToken(now, lifetime)
	require.NoError(t, err)

	token2, err := GenerateToken(now, lifetime)
	require.NoError(t, err)

	// Tokens should be different
	assert.NotEqual(t, token1.Value, token2.Value)
	assert.NotEqual(t, token1.Hash, token2.Hash)
}

func TestHashToken(t *testing.T) {
	token := "test-token-value"
	hash1 := HashToken(token)
	hash2 := HashToken(token)

	// Same input should produce same hash
	assert.Equal(t, hash1, hash2)

	// Different input should produce different hash
	hash3 := HashToken("different-token")
	assert.NotEqual(t, hash1, hash3)
}

// --- Verification token tests (DB-dependent) ---

func TestCreateVerificationToken(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user first
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	token, err := module.CreateVerificationToken(t.Context(), qtx, userId, time.Now())
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.NotEmpty(t, token.Value)
	assert.NotEqual(t, [32]byte{}, token.Hash)
	assert.True(t, token.ExpiresAt.After(time.Now()))

	// Verify token was stored in DB
	storedToken, err := qtx.GetVerificationToken(t.Context(), token.Hash[:])
	require.NoError(t, err)
	assert.Equal(t, userId, storedToken.UserID)
}

func TestValidateVerificationToken_Valid(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	// Create token
	token, err := module.CreateVerificationToken(t.Context(), qtx, userId, time.Now())
	require.NoError(t, err)

	// Validate token
	gotUserId, err := module.ValidateVerificationToken(t.Context(), qtx, token.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId)
}

func TestValidateVerificationToken_Invalid(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Try to validate non-existent token
	_, err := module.ValidateVerificationToken(t.Context(), qtx, "invalid-token")
	require.ErrorIs(t, err, ErrInvalidVerificationToken)
}

func TestValidateVerificationToken_Expired(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	// Create token that expires immediately
	pastTime := time.Now().Add(-time.Hour * 25) // 25 hours ago
	token, err := module.CreateVerificationToken(t.Context(), qtx, userId, pastTime)
	require.NoError(t, err)

	// Try to validate expired token - should fail because DB query filters by expires_at > now()
	_, err = module.ValidateVerificationToken(t.Context(), qtx, token.Value)
	require.ErrorIs(t, err, ErrInvalidVerificationToken)
}

func TestVerifyEmail_Success(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	// Create verification token
	token, err := module.CreateVerificationToken(t.Context(), qtx, userId, time.Now())
	require.NoError(t, err)

	// Verify email
	gotUserId, err := module.VerifyEmail(t.Context(), qtx, token.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId)

	// Check user is marked as verified
	user, err := qtx.GetUserByID(t.Context(), userId)
	require.NoError(t, err)
	assert.True(t, user.EmailVerified)

	// Token should be deleted (trying to use it again should fail)
	_, err = module.ValidateVerificationToken(t.Context(), qtx, token.Value)
	require.ErrorIs(t, err, ErrInvalidVerificationToken)
}

// --- Password reset token tests (DB-dependent) ---

func TestCreatePasswordResetToken(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	token, err := module.CreatePasswordResetToken(t.Context(), qtx, userId, time.Now())
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.NotEmpty(t, token.Value)
	assert.NotEqual(t, [32]byte{}, token.Hash)
	assert.True(t, token.ExpiresAt.After(time.Now()))

	// Verify token was stored in DB
	storedToken, err := qtx.GetPasswordResetToken(t.Context(), token.Hash[:])
	require.NoError(t, err)
	assert.Equal(t, userId, storedToken.UserID)
}

func TestValidatePasswordResetToken_Valid(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	// Create token
	token, err := module.CreatePasswordResetToken(t.Context(), qtx, userId, time.Now())
	require.NoError(t, err)

	// Validate token
	gotUserId, err := module.ValidatePasswordResetToken(t.Context(), qtx, token.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId)
}

func TestValidatePasswordResetToken_Invalid(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Try to validate non-existent token
	_, err := module.ValidatePasswordResetToken(t.Context(), qtx, "invalid-token")
	require.ErrorIs(t, err, ErrInvalidResetToken)
}

func TestValidatePasswordResetToken_Expired(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	// Create token that expires immediately
	pastTime := time.Now().Add(-time.Hour * 25) // 25 hours ago
	token, err := module.CreatePasswordResetToken(t.Context(), qtx, userId, pastTime)
	require.NoError(t, err)

	// Try to validate expired token
	_, err = module.ValidatePasswordResetToken(t.Context(), qtx, token.Value)
	require.ErrorIs(t, err, ErrInvalidResetToken)
}

func TestDeletePasswordResetToken(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	// Create token
	token, err := module.CreatePasswordResetToken(t.Context(), qtx, userId, time.Now())
	require.NoError(t, err)

	// Delete token
	err = module.DeletePasswordResetToken(t.Context(), qtx, token.Value)
	require.NoError(t, err)

	// Token should no longer be valid
	_, err = module.ValidatePasswordResetToken(t.Context(), qtx, token.Value)
	require.ErrorIs(t, err, ErrInvalidResetToken)
}

func TestDeletePasswordResetToken_NonExistent(t *testing.T) {
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Deleting non-existent token should not error (just log warning)
	err := module.DeletePasswordResetToken(t.Context(), qtx, "non-existent-token")
	require.NoError(t, err)
}

// --- Multiple tokens tests ---

func TestMultiplePasswordResetTokens_AllValid(t *testing.T) {
	// User can request multiple password resets and all tokens should be valid
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	// Create first token
	token1, err := module.CreatePasswordResetToken(t.Context(), qtx, userId, time.Now())
	require.NoError(t, err)

	// Create second token
	token2, err := module.CreatePasswordResetToken(t.Context(), qtx, userId, time.Now())
	require.NoError(t, err)

	// Both tokens should be valid
	gotUserId1, err := module.ValidatePasswordResetToken(t.Context(), qtx, token1.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId1)

	gotUserId2, err := module.ValidatePasswordResetToken(t.Context(), qtx, token2.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId2)
}

func TestTokenReuseAfterDeletion(t *testing.T) {
	// Once a token is used (deleted), it cannot be reused
	tx, qtx, module := prepare(t)
	defer tx.Rollback(t.Context())

	// Create user
	userId, err := qtx.CreateUser(t.Context(), db.CreateUserParams{
		Email:        "test@example.com",
		PasswordHash: []byte("hashed"),
	})
	require.NoError(t, err)

	// Create token
	token, err := module.CreatePasswordResetToken(t.Context(), qtx, userId, time.Now())
	require.NoError(t, err)

	// First use - should work
	gotUserId, err := module.ValidatePasswordResetToken(t.Context(), qtx, token.Value)
	require.NoError(t, err)
	assert.Equal(t, userId, gotUserId)

	// Delete token (simulate consumption)
	err = module.DeletePasswordResetToken(t.Context(), qtx, token.Value)
	require.NoError(t, err)

	// Second use - should fail
	_, err = module.ValidatePasswordResetToken(t.Context(), qtx, token.Value)
	require.ErrorIs(t, err, ErrInvalidResetToken)
}
