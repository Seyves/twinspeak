package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/netip"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/db"
)

const (
	accessTokenLifetime  = time.Minute * 10
	refreshTokenLifetime = time.Hour * 24 * 14
	wsTicketLifetime     = time.Second * 10
)

type token struct {
	Value     string
	ExpiresAt time.Time
}

func createRefreshToken(ctx context.Context, now time.Time, queries *db.Queries, userId uuid.UUID, userAgent string, ip *netip.Addr) (*token, error) {
	b := make([]byte, 32)
	rand.Read(b)

	tok := base64.URLEncoding.EncodeToString(b)
	hash := sha256.Sum256([]byte(tok))
	expiresAt := now.Add(refreshTokenLifetime)

	_, err := queries.CreateRefreshSession(ctx, db.CreateRefreshSessionParams{
		UserID:    userId,
		TokenHash: hash[:],
		UserAgent: &userAgent,
		ExpiresAt: expiresAt,
		Ip:        ip,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot insert refresh session into db: %w", err)
	}
	return &token{Value: tok, ExpiresAt: expiresAt}, nil
}

func createAccessToken(ctx context.Context, now time.Time, hmacSecret string, userId uuid.UUID) (*token, error) {
	expiresAt := now.Add(accessTokenLifetime)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId.String(),
		"exp": expiresAt.Unix(),
	})

	tokenString, err := tok.SignedString([]byte(hmacSecret))
	if err != nil {
		return nil, fmt.Errorf("cannot sign token: %w", err)
	}
	return &token{Value: tokenString, ExpiresAt: expiresAt}, nil
}

func createWSTicket(ctx context.Context, now time.Time, hmacSecret string, userId uuid.UUID) (*token, error) {
	expiresAt := now.Add(wsTicketLifetime)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId.String(),
		"ws":     true,
		"exp":    expiresAt.Unix(),
	})

	tokenString, err := tok.SignedString([]byte(hmacSecret))
	if err != nil {
		return nil, fmt.Errorf("cannot sign ws token: %w", err)
	}
	return &token{Value: tokenString, ExpiresAt: expiresAt}, nil
}
