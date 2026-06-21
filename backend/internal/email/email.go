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
	"github.com/jackc/pgx/v5"
	"github.com/resend/resend-go/v3"
	"github.com/twinspeak/backend/internal/db"
)

//go:embed templates/*.html
var templatesFS embed.FS

const verificationTokenLifetime = time.Hour * 24
const passwordResetTokenLifetime = time.Hour * 24

var ErrInvalidResetToken = errors.New("invalid or expired reset token")
var ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
var ErrUserNotFound = errors.New("user not found")

type Module struct {
	resendClient          *resend.Client
	fromEmail             string
	publicUrl             string
	verificationTemplate  *template.Template
	passwordResetTemplate *template.Template
}

type VerificationEmailData struct {
	VerificationURL string
}

type PasswordResetEmailData struct {
	ResetURL string
}

type Token struct {
	Value     string
	Hash      [32]byte
	ExpiresAt time.Time
}

func GenerateToken(now time.Time, lifetime time.Duration) (*Token, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("cannot generate random token: %w", err)
	}

	value := base64.URLEncoding.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(value))
	expiresAt := now.Add(lifetime)

	return &Token{
		Value:     value,
		Hash:      hash,
		ExpiresAt: expiresAt,
	}, nil
}

func HashToken(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}

func (m *Module) CreateVerificationToken(ctx context.Context, tx *db.Queries, userId uuid.UUID, now time.Time) (*Token, error) {
	token, err := GenerateToken(now, verificationTokenLifetime)
	if err != nil {
		return nil, err
	}

	_, err = tx.CreateVerificationToken(ctx, db.CreateVerificationTokenParams{
		UserID:    userId,
		TokenHash: token.Hash[:],
		ExpiresAt: token.ExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot store verification token: %w", err)
	}

	return token, nil
}

func (m *Module) SendVerificationEmail(ctx context.Context, tx *db.Queries, userId uuid.UUID, email string) error {
	token, err := m.CreateVerificationToken(ctx, tx, userId, time.Now())
	if err != nil {
		return err
	}

	verificationURL := fmt.Sprintf("%s/verify-email/callback?token=%s", m.publicUrl, token.Value)
	return m.sendEmail(email, "Verify your TwinSpeak email", m.verificationTemplate, VerificationEmailData{
		VerificationURL: verificationURL,
	})
}

func (m *Module) ValidateVerificationToken(ctx context.Context, tx *db.Queries, token string) (uuid.UUID, error) {
	tokenHash := HashToken(token)

	verificationToken, err := tx.GetVerificationToken(ctx, tokenHash[:])
	if err != nil {
		return uuid.Nil, ErrInvalidVerificationToken
	}

	return verificationToken.UserID, nil
}

func (m *Module) VerifyEmail(ctx context.Context, tx *db.Queries, token string) (uuid.UUID, error) {
	userId, err := m.ValidateVerificationToken(ctx, tx, token)
	if err != nil {
		return uuid.Nil, err
	}

	err = tx.VerifyUserEmail(ctx, userId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot verify user email: %w", err)
	}

	tokenHash := HashToken(token)
	err = tx.DeleteVerificationToken(ctx, tokenHash[:])
	if err != nil {
		// Log but don't fail - verification already succeeded
		log.Warnf("Warning: could not delete verification token: %v\n", err)
	}

	return userId, nil
}

func (m *Module) CreatePasswordResetToken(ctx context.Context, tx *db.Queries, userId uuid.UUID, now time.Time) (*Token, error) {
	token, err := GenerateToken(now, passwordResetTokenLifetime)
	if err != nil {
		return nil, err
	}

	_, err = tx.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		UserID:    userId,
		TokenHash: token.Hash[:],
		ExpiresAt: token.ExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot insert token into db: %w", err)
	}

	return token, nil
}

func (m *Module) SendPasswordResetEmail(ctx context.Context, tx *db.Queries, email string) error {
	user, err := tx.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	} else if err != nil {
		return fmt.Errorf("cannot select user from db: %w", err)
	}

	token, err := m.CreatePasswordResetToken(ctx, tx, user.ID, time.Now())
	if err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/auth/recovery?token=%s", m.publicUrl, token.Value)
	return m.sendEmail(email, "Reset your TwinSpeak password", m.passwordResetTemplate, PasswordResetEmailData{
		ResetURL: resetURL,
	})
}

func (m *Module) ValidatePasswordResetToken(ctx context.Context, tx *db.Queries, token string) (uuid.UUID, error) {
	tokenHash := HashToken(token)

	resetToken, err := tx.GetPasswordResetToken(ctx, tokenHash[:])
	if err != nil {
		return uuid.Nil, ErrInvalidResetToken
	}

	return resetToken.UserID, nil
}

func (m *Module) DeletePasswordResetToken(ctx context.Context, tx *db.Queries, token string) error {
	tokenHash := HashToken(token)
	err := tx.DeletePasswordResetToken(ctx, tokenHash[:])
	if err != nil {
		log.Warnf("Could not delete password reset token: %v\n", err)
	}
	return nil
}

func (m *Module) sendEmail(to string, subject string, tmpl *template.Template, data any) error {
	var emailBody bytes.Buffer
	err := tmpl.Execute(&emailBody, data)
	if err != nil {
		return fmt.Errorf("cannot render email template: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    m.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    emailBody.String(),
	}

	_, err = m.resendClient.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("cannot send email via Resend: %w", err)
	}

	return nil
}

func New(apiKey string, fromEmail string, publicUrl string) (*Module, error) {
	client := resend.NewClient(apiKey)

	verificationTmpl, err := template.ParseFS(templatesFS, "templates/verification.html")
	if err != nil {
		return nil, fmt.Errorf("cannot parse verification email template: %w", err)
	}

	passwordResetTmpl, err := template.ParseFS(templatesFS, "templates/password-reset.html")
	if err != nil {
		return nil, fmt.Errorf("cannot parse password reset email template: %w", err)
	}

	return &Module{
		resendClient:          client,
		fromEmail:             fromEmail,
		publicUrl:             publicUrl,
		verificationTemplate:  verificationTmpl,
		passwordResetTemplate: passwordResetTmpl,
	}, nil
}
