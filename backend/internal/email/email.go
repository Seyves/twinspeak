package email

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/resend/resend-go/v3"
	"github.com/twinspeak/backend/internal/db"
)

//go:embed templates/*.html
var templatesFS embed.FS

const verificationTokenLifetime = time.Hour * 24

var ErrInvalidEmail = errors.New("invalid email")

type Module struct {
	resendClient *resend.Client
	fromEmail    string
	publicUrl    string
	template     *template.Template
}

type VerificationEmailData struct {
	VerificationURL string
}

func (m *Module) SendVerificationEmail(ctx context.Context, tx *db.Queries, userId uuid.UUID, email string) error {
	// Generate random token (32 bytes = 256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("cannot generate random token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	tokenHash := sha256.Sum256([]byte(token))
	expiresAt := time.Now().Add(verificationTokenLifetime)

	// Store token in database
	_, err := tx.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{
		UserID:    userId,
		TokenHash: tokenHash[:],
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("cannot store verification token: %w", err)
	}

	// Render email template
	verificationURL := fmt.Sprintf("%s/verify-email/callback?token=%s", m.publicUrl, token)
	var emailBody bytes.Buffer
	err = m.template.Execute(&emailBody, VerificationEmailData{
		VerificationURL: verificationURL,
	})
	if err != nil {
		return fmt.Errorf("cannot render email template: %w", err)
	}

	// Send email via Resend
	params := &resend.SendEmailRequest{
		From:    m.fromEmail,
		To:      []string{email},
		Subject: "Verify your TwinSpeak email",
		Html:    emailBody.String(),
	}

	_, err = m.resendClient.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("cannot send email via Resend: %w", err)
	}

	return nil
}

func (m *Module) VerifyEmail(ctx context.Context, tx *db.Queries, token string) (uuid.UUID, error) {
	tokenHash := sha256.Sum256([]byte(token))

	// Get token from database
	verificationToken, err := tx.GetVerificationToken(ctx, tokenHash[:])
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// Mark user as verified
	err = tx.VerifyUserEmail(ctx, verificationToken.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot verify user email: %w", err)
	}

	// Delete token (one-time use)
	err = tx.DeleteVerificationToken(ctx, tokenHash[:])
	if err != nil {
		// Log but don't fail - verification already succeeded
		log.Warnf("Warning: could not delete verification token: %v\n", err)
	}

	return verificationToken.UserID, nil
}

func New(apiKey string, fromEmail string, publicUrl string) (*Module, error) {
	client := resend.NewClient(apiKey)

	tmpl, err := template.ParseFS(templatesFS, "templates/verification.html")
	if err != nil {
		return nil, fmt.Errorf("cannot parse email template: %w", err)
	}

	return &Module{
		resendClient: client,
		fromEmail:    fromEmail,
		publicUrl:    publicUrl,
		template:     tmpl,
	}, nil
}
