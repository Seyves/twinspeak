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

func createRefreshToken(ctx context.Context, now time.Time, queries *db.Queries, userId uuid.UUID, userAgent string, ip *netip.Addr) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)

	token := base64.URLEncoding.EncodeToString(b)
	hash := sha256.Sum256([]byte(token))

	_, err := queries.CreateRefreshSession(ctx, db.CreateRefreshSessionParams{
		UserID:    userId,
		TokenHash: hash[:],
		UserAgent: &userAgent,
		ExpiresAt: now.Add(refreshTokenLifetime),
		Ip:        ip,
	})
	if err != nil {
		return "", fmt.Errorf("cannot insert refresh session into db: %w", err)
	}
	return token, nil
}

func createAccessToken(ctx context.Context, now time.Time, hmacSecret string, userId uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId.String(),
		"exp":    now.Add(accessTokenLifetime).Unix(),
	})

	tokenString, err := token.SignedString([]byte(hmacSecret))
	if err != nil {
		return "", fmt.Errorf("cannot sign token: %w", err)
	}
	return tokenString, nil
}
