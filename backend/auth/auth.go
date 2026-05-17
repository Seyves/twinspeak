package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twinspeak/backend/db"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenLifetime  = time.Minute * 10
	refreshTokenLifetime = time.Hour * 24 * 14
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrMaliciousSuspicion = errors.New("malicious suspicion")

type Auth struct {
	hmacSecret string
	db         *pgxpool.Pool
	queries    *db.Queries
}

func (a *Auth) RotateAccessToken(
	ctx context.Context,
	now time.Time,
	refreshToken string,
	userAgent string,
	ip *netip.Addr,
) (accessToken string, newRefreshToken string, err error) {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("cannot start db transaction: %w", err)
	}

	txq := a.queries.WithTx(tx)

	hash := sha256.Sum256([]byte(refreshToken))
	session, err := txq.GetRefreshSessionForUpdate(ctx, hash[:])
	if err != nil {
		return "", "", fmt.Errorf("cannot select refresh session from db: %w", err)
	}

	if session.RevokedAt != nil {
		// TODO: Send email notification if token was revoked before
	}

	err = txq.RevokeRefreshSession(ctx, hash[:])
	if err != nil {
		return "", "", fmt.Errorf("cannot update refresh session in db: %w", err)
	}

	newRefreshToken, err = createRefreshToken(ctx, now, txq, session.UserID, userAgent, ip)

	err = tx.Commit(ctx)
	if err != nil {
		return "", "", fmt.Errorf("cannot commit db transaction: %w", err)
	}

	accessToken, err = createAccessToken(ctx, now, a.hmacSecret, session.UserID)
	if err != nil {
		return "", "", fmt.Errorf("cannot create access token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

func (a *Auth) ValidateAccessToken(ctx context.Context, now time.Time, accessToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (any, error) {
		return []byte(a.hmacSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, fmt.Errorf("cannot parse token: %w", err)
	}
	return token, nil
}

func (a *Auth) SignIn(ctx context.Context, now time.Time, email string, password string, userAgent string, ip *netip.Addr) (
	accessToken string,
	refreshToken string,
	err error,
) {
	user, err := a.queries.GetUser(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrInvalidCredentials
	} else if err != nil {
		return "", "", fmt.Errorf("cannot select user from db: %w", err)
	}

	// If user signed up using google and hasn't set the password
	if user.PasswordHash == nil {
		return "", "", ErrInvalidCredentials
	}

	salted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("cannot hash the password: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(salted, user.PasswordHash)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return "", "", ErrInvalidCredentials
	} else if err != nil {
		return "", "", fmt.Errorf("cannot compare passwords: %w", err)
	}

	refreshToken, err = createRefreshToken(ctx, now, a.queries, user.ID, userAgent, ip)
	if err != nil {
		return "", "", fmt.Errorf("cannot create refresh token: %w", err)
	}

	accessToken, err = createAccessToken(ctx, now, a.hmacSecret, user.ID)
	if err != nil {
		return "", "", fmt.Errorf("cannot create access token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (a *Auth) SignUp(ctx context.Context, now time.Time, email string, password string, userAgent string, ip *netip.Addr) (
	accessToken string,
	refreshToken string,
	err error,
) {
	salted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("cannot hash the password: %w", err)
	}

	userId, err := a.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: salted,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrInvalidCredentials
	} else if err != nil {
		return "", "", fmt.Errorf("cannot create user in db: %w", err)
	}

	refreshToken, err = createRefreshToken(ctx, now, a.queries, userId, userAgent, ip)
	if err != nil {
		return "", "", fmt.Errorf("cannot create refresh token: %w", err)
	}

	accessToken, err = createAccessToken(ctx, now, a.hmacSecret, userId)
	if err != nil {
		return "", "", fmt.Errorf("cannot create access token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func NewAuth(db *pgxpool.Pool, queries *db.Queries, hmacSecret string) *Auth {
	return &Auth{
		hmacSecret: hmacSecret,
		db:         db,
		queries:    queries,
	}
}
