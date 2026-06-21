package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/twinspeak/backend/internal/db"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenLifetime  = time.Minute * 10
	refreshTokenLifetime = time.Hour * 24 * 14
	wsTicketLifetime     = time.Second * 10
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrMaliciousSuspicion = errors.New("malicious suspicion")
var ErrEmailAlreadyTaken = errors.New("email already taken")

type Module struct {
	hmacSecret string
}

type Token struct {
	Value     string
	ExpiresAt time.Time
}

func (m *Module) CreateUser(ctx context.Context, tx *db.Queries, email string, password string) (uuid.UUID, error) {
	if _, err := tx.GetUserByEmail(ctx, email); err == nil {
		return uuid.Nil, ErrEmailAlreadyTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("cannot select user from db by email: %w", err)
	}
	salted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot hash the password: %w", err)
	}
	userId, err := tx.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: salted,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot insert user in db: %w", err)
	}
	return userId, nil
}

func (m *Module) ValidateSession(ctx context.Context, tx *db.Queries, refreshToken string) (uuid.UUID, error) {
	hash := sha256.Sum256([]byte(refreshToken))
	session, err := tx.GetRefreshSessionForUpdate(ctx, hash[:])
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot select refresh session from db: %w", err)
	}
	if session.RevokedAt != nil {
		return uuid.Nil, ErrMaliciousSuspicion
	}
	return session.UserID, nil
}

func (m *Module) RevokeSession(ctx context.Context, tx *db.Queries, refreshToken string) error {
	hash := sha256.Sum256([]byte(refreshToken))
	err := tx.RevokeRefreshSession(ctx, hash[:])
	if err != nil {
		return fmt.Errorf("cannot update refresh session in db: %w", err)
	}
	return nil
}

func (m *Module) StartSession(ctx context.Context, tx *db.Queries, userId uuid.UUID, now time.Time) (
	accessToken *Token,
	refreshToken *Token,
	err error,
) {
	refreshToken, err = createRefreshToken(ctx, now, tx, userId)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot insert refresh token in db: %w", err)
	}
	accessToken, err = createAccessToken(now, m.hmacSecret, userId)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create access token: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (m *Module) IssueWSTicket(ctx context.Context, now time.Time, userId uuid.UUID) (wsToken *Token, err error) {
	wsToken, err = createWSTicket(now, m.hmacSecret, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot create ws token: %w", err)
	}

	return wsToken, nil
}

func (m *Module) ValidatePassword(ctx context.Context, tx *db.Queries, email string, password string) (uuid.UUID, error) {
	user, err := tx.GetUser(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrInvalidCredentials
	} else if err != nil {
		return uuid.Nil, fmt.Errorf("cannot select user from db: %w", err)
	}

	// If user signed up using google and hasn't set the password
	if user.PasswordHash == nil {
		return uuid.Nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return uuid.Nil, ErrInvalidCredentials
	} else if err != nil {
		return uuid.Nil, fmt.Errorf("cannot compare passwords: %w", err)
	}
	return user.ID, nil
}

func (m *Module) UpdatePassword(ctx context.Context, tx *db.Queries, userId uuid.UUID, newPassword string) error {
	salted, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("cannot hash the password: %w", err)
	}
	err = tx.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userId,
		PasswordHash: salted,
	})
	if err != nil {
		return fmt.Errorf("cannot update user password in db: %w", err)
	}
	return nil
}

func (m *Module) ValidateAccessToken(ctx context.Context, now time.Time, accessToken string) (*jwt.Token, uuid.UUID, error) {
	token, err := jwt.Parse(
		accessToken,
		func(t *jwt.Token) (any, error) {
			return []byte(m.hmacSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
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

func (m *Module) ValidateWSTicket(ctx context.Context, now time.Time, ticket string) (*jwt.Token, uuid.UUID, error) {
	token, err := jwt.Parse(
		ticket,
		func(t *jwt.Token) (any, error) {
			return []byte(m.hmacSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
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

func createRefreshToken(ctx context.Context, now time.Time, queries *db.Queries, userId uuid.UUID) (*Token, error) {
	b := make([]byte, 32)
	rand.Read(b)

	tok := base64.URLEncoding.EncodeToString(b)
	hash := sha256.Sum256([]byte(tok))
	expiresAt := now.Add(refreshTokenLifetime)

	ip, _ := ctx.Value("ip").(*netip.Addr)
	userAgent, _ := ctx.Value("userAgent").(*string)

	_, err := queries.CreateRefreshSession(ctx, db.CreateRefreshSessionParams{
		UserID:    userId,
		TokenHash: hash[:],
		ExpiresAt: expiresAt,
		UserAgent: userAgent,
		Ip:        ip,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot insert refresh session into db: %w", err)
	}
	return &Token{Value: tok, ExpiresAt: expiresAt}, nil
}

func createAccessToken(now time.Time, hmacSecret string, userId uuid.UUID) (*Token, error) {
	expiresAt := now.Add(accessTokenLifetime)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId.String(),
		"exp": expiresAt.Unix(),
	})

	tokenString, err := tok.SignedString([]byte(hmacSecret))
	if err != nil {
		return nil, fmt.Errorf("cannot sign token: %w", err)
	}
	return &Token{Value: tokenString, ExpiresAt: expiresAt}, nil
}

func createWSTicket(now time.Time, hmacSecret string, userId uuid.UUID) (*Token, error) {
	expiresAt := now.Add(wsTicketLifetime)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId.String(),
		"ws":  true,
		"exp": expiresAt.Unix(),
	})

	tokenString, err := tok.SignedString([]byte(hmacSecret))
	if err != nil {
		return nil, fmt.Errorf("cannot sign ws token: %w", err)
	}
	return &Token{Value: tokenString, ExpiresAt: expiresAt}, nil
}

func New(hmacSecret string) *Module {
	return &Module{
		hmacSecret: hmacSecret,
	}
}
