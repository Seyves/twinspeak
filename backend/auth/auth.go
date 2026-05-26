package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twinspeak/backend/db"
	"golang.org/x/crypto/bcrypt"
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
) (accessToken *token, newRefreshToken *token, err error) {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txq := a.queries.WithTx(tx)

	hash := sha256.Sum256([]byte(refreshToken))
	session, err := txq.GetRefreshSessionForUpdate(ctx, hash[:])
	if err != nil {
		return nil, nil, fmt.Errorf("cannot select refresh session from db: %w", err)
	}

	if session.RevokedAt != nil {
		// TODO: Send email notification if token was revoked before
		return nil, nil, ErrMaliciousSuspicion
	}

	err = txq.RevokeRefreshSession(ctx, hash[:])
	if err != nil {
		return nil, nil, fmt.Errorf("cannot update refresh session in db: %w", err)
	}

	newRefreshToken, err = createRefreshToken(ctx, now, txq, session.UserID, userAgent, ip)

	err = tx.Commit(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot commit db transaction: %w", err)
	}

	accessToken, err = createAccessToken(ctx, now, a.hmacSecret, session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create access token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

func (a *Auth) GetWSTicket(
	ctx context.Context,
	now time.Time,
	userId uuid.UUID,
) (wsToken *token, err error) {
	wsToken, err = createWSTicket(ctx, now, a.hmacSecret, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot create ws token: %w", err)
	}

	return wsToken, nil
}

func (a *Auth) ValidateAccessToken(ctx context.Context, now time.Time, accessToken string) (*jwt.Token, uuid.UUID, error) {
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (any, error) {
		return []byte(a.hmacSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("cannot parse token: %w", err)
	}
	claims := token.Claims.(jwt.MapClaims)
	subStr, ok := claims["sub"].(string)
	if !ok {
		return nil, uuid.Nil, fmt.Errorf("cannot get token sub")
	}
	userId, err := uuid.Parse(subStr)
	if !ok {
		return nil, uuid.Nil, fmt.Errorf("cannot parse sub: %w", err)
	}
	return token, userId, nil
}

func (a *Auth) ValidateWSTicket(ctx context.Context, now time.Time, ticket string) (*jwt.Token, uuid.UUID, error) {
	token, err := jwt.Parse(ticket, func(t *jwt.Token) (any, error) {
		return []byte(a.hmacSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("cannot parse token: %w", err)
	}
	claims := token.Claims.(jwt.MapClaims)
	subStr, ok := claims["sub"].(string)
	if !ok {
		return nil, uuid.Nil, fmt.Errorf("cannot get token sub: %w", err)
	}
	userId, err := uuid.Parse(subStr)
	if !ok {
		return nil, uuid.Nil, fmt.Errorf("cannot parse sub: %w", err)
	}
	ws, ok := claims["ws"].(bool)
	if ok && ws == true {
		return token, userId, nil
	}
	return nil, uuid.Nil, fmt.Errorf("cannot assert claims ws: %w", err)
}

func (a *Auth) SignIn(ctx context.Context, now time.Time, email string, password string, userAgent string, ip *netip.Addr) (
	accessToken *token,
	refreshToken *token,
	err error,
) {
	user, err := a.queries.GetUser(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, nil, fmt.Errorf("cannot select user from db: %w", err)
	}

	// If user signed up using google and hasn't set the password
	if user.PasswordHash == nil {
		return nil, nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, nil, fmt.Errorf("cannot compare passwords: %w", err)
	}

	refreshToken, err = createRefreshToken(ctx, now, a.queries, user.ID, userAgent, ip)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create refresh token: %w", err)
	}

	accessToken, err = createAccessToken(ctx, now, a.hmacSecret, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create access token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (a *Auth) SignUp(ctx context.Context, now time.Time, email string, password string, userAgent string, ip *netip.Addr) (
	accessToken *token,
	refreshToken *token,
	err error,
) {
	salted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot hash the password: %w", err)
	}

	userId, err := a.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: salted,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create user in db: %w", err)
	}

	refreshToken, err = createRefreshToken(ctx, now, a.queries, userId, userAgent, ip)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create refresh token: %w", err)
	}

	accessToken, err = createAccessToken(ctx, now, a.hmacSecret, userId)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create access token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (a *Auth) GetCurrentUser(ctx context.Context, userId uuid.UUID) (db.User, error) {
	return a.queries.GetUserByID(ctx, userId)
}

func (a *Auth) Logout(ctx context.Context, refreshToken string) error {
	hash := sha256.Sum256([]byte(refreshToken))
	return a.queries.RevokeRefreshSession(ctx, hash[:])
}

func NewAuth(db *pgxpool.Pool, queries *db.Queries, hmacSecret string) *Auth {
	return &Auth{
		hmacSecret: hmacSecret,
		db:         db,
		queries:    queries,
	}
}
