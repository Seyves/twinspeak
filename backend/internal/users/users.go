package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twinspeak/backend/internal/auth"
	"github.com/twinspeak/backend/internal/billing"
	"github.com/twinspeak/backend/internal/db"
	"github.com/twinspeak/backend/internal/email"
	"github.com/twinspeak/backend/internal/googleauth"
	"github.com/twinspeak/backend/internal/metrics"
	"github.com/twinspeak/backend/internal/preferences"
)

type Service struct {
	db          *pgxpool.Pool
	queries     *db.Queries
	auth        *auth.Module
	googleAuth  *googleauth.Module
	preferences *preferences.Module
	billing     *billing.Module
	metrics     *metrics.Module
	email       *email.Module
}

func (s *Service) RotateSession(ctx context.Context, now time.Time, refreshToken string) (
	accessToken *auth.Token, newRefreshToken *auth.Token, userId uuid.UUID, err error,
) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("cannot start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	userId, err = s.auth.ValidateSession(ctx, qtx, refreshToken)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("during checking session: %w", err)
	}

	err = s.auth.RevokeSession(ctx, qtx, refreshToken)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("cannot revoke session: %w", err)
	}

	accessToken, newRefreshToken, err = s.auth.StartSession(ctx, qtx, userId, now)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("cannot start session: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("cannot commit db transaction: %w", err)
	}

	return accessToken, newRefreshToken, userId, nil
}

func (s *Service) GetWSTicket(ctx context.Context, now time.Time, userId uuid.UUID) (wsToken *auth.Token, err error) {
	return s.auth.IssueWSTicket(ctx, now, userId)
}

func (s *Service) SignIn(ctx context.Context, now time.Time, email string, password string) (
	accessToken *auth.Token, refreshToken *auth.Token, userId uuid.UUID, err error,
) {
	userId, err = s.auth.ValidatePassword(ctx, s.queries, email, password)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("validating credentials: %w", err)
	}

	accessToken, refreshToken, err = s.auth.StartSession(ctx, s.queries, userId, now)
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("cannot start session: %w", err)
	}

	return accessToken, refreshToken, userId, nil
}

func (s *Service) SignUp(ctx context.Context, now time.Time, email string, password string) (
	accessToken *auth.Token, refreshToken *auth.Token, err error,
) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)
	userId, err := s.auth.CreateUser(ctx, qtx, email, password)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create user: %w", err)
	}

	err = s.preferences.CreatePreferences(ctx, qtx, userId)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create user preferences: %w", err)
	}

	err = s.billing.StartSubscription(ctx, qtx, userId, now)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot start subscription: %w", err)
	}

	// Send verification email
	err = s.email.SendVerificationEmail(ctx, qtx, userId, email)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot send verification email: %w", err)
	}

	accessToken, refreshToken, err = s.auth.StartSession(ctx, qtx, userId, now)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot start session: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot commit db transaction: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) GetPreferences(ctx context.Context, userId uuid.UUID) (*db.Preference, error) {
	return s.preferences.GetPreferences(ctx, s.queries, userId)
}

func (s *Service) UpdatePreferences(ctx context.Context, params db.UpdateUserPrefsParams) error {
	return s.preferences.UpdatePreferences(ctx, s.queries, params)
}

func (s *Service) GoogleCallback(ctx context.Context, now time.Time, code string, sessionState string, state string) (
	accessToken *auth.Token, refreshToken *auth.Token, err error,
) {
	openIdInfo, err := s.googleAuth.Callback(ctx, code, sessionState, state)
	if err != nil {
		return nil, nil, fmt.Errorf("during google callback: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// Trying to find existing google integrated account
	userId, err := s.googleAuth.FindGoogleAccount(ctx, qtx, openIdInfo.Id)
	if errors.Is(err, googleauth.ErrGoogleAccountNotFound) {
		// If there is no google integrated account - trying to create new
		userId, err = s.googleAuth.CreateUser(ctx, qtx, openIdInfo)
		// If there is an account with the same email - integrating it with the google account
		if errors.Is(err, googleauth.ErrEmailAlreadyTaken) {
			userId, err = s.googleAuth.LinkExistingUser(ctx, qtx, openIdInfo)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot link existing user account to google: %w", err)
			}
		} else if err != nil {
			return nil, nil, fmt.Errorf("cannot create user: %w", err)
		} else {
			err = s.preferences.CreatePreferences(ctx, qtx, userId)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot create user preferences: %w", err)
			}

			err = s.billing.StartSubscription(ctx, qtx, userId, now)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot start subscription: %w", err)
			}
		}
	} else if err != nil {
		return nil, nil, fmt.Errorf("cannot find google account: %w", err)
	}

	accessToken, refreshToken, err = s.auth.StartSession(ctx, qtx, userId, now)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot start session: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot commit db transaction: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) BuyTopup(ctx context.Context, now time.Time, userId uuid.UUID, amount int32) error {
	err := s.billing.BuyTopup(ctx, s.queries, userId, now, amount)
	if err != nil {
		return fmt.Errorf("cannot buy topup: %w", err)
	}
	return nil
}

func (s *Service) RenewSubscriptionsWorker(ctx context.Context, now time.Time) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("cannot start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	expiredSubs, err := s.billing.GetExpiredSubscriptions(ctx, qtx, now)
	for _, expiredSubs := range expiredSubs {
		err := s.billing.RenewSubscription(ctx, qtx, expiredSubs, now)
		if err != nil {
			return fmt.Errorf("cannot renew subscription: %w", err)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("cannot commit db transaction: %w", err)
	}

	return nil
}

func (s *Service) StartSpeech(
	ctx context.Context,
	now time.Time,
	userId uuid.UUID,
) error {
	err := s.billing.CheckAvailableCredits(ctx, s.queries, userId, now)
	if err != nil {
		return fmt.Errorf("cannot start speech: %w", err)
	}
	return nil
}

func (s *Service) EndSpeech(ctx context.Context, now time.Time, params db.InsertSpeechParams) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("cannot start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	seconds := params.EndedAt.Sub(params.StartedAt).Seconds()
	err = s.billing.SpendCredits(ctx, qtx, params.UserID, now, int32(seconds))
	if err != nil {
		return fmt.Errorf("cannot spend credits: %w", err)
	}
	err = s.metrics.CreateSpeechMetric(context.Background(), qtx, params)
	if err != nil {
		return fmt.Errorf("cannot create metric: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("cannot commit db transaction: %w", err)
	}
	return nil
}

func (s *Service) GetSpeeches(ctx context.Context, userId uuid.UUID) ([]db.Speech, error) {
	return s.metrics.GetSpeeches(ctx, s.queries, userId)
}

func (s *Service) CreateHttpRequestMetric(ctx context.Context, data db.InsertHttpRequestParams) error {
	return s.metrics.CreateHttpRequestMetric(ctx, s.queries, data)
}

func (s *Service) ValidateAccessToken(ctx context.Context, now time.Time, accessToken string) (uuid.UUID, error) {
	_, userId, err := s.auth.ValidateAccessToken(ctx, now, accessToken)
	return userId, err
}

func (s *Service) ValidateWSTicket(ctx context.Context, now time.Time, ticket string) (uuid.UUID, error) {
	_, userId, err := s.auth.ValidateWSTicket(ctx, now, ticket)
	return userId, err
}

func (s *Service) GoogleRedirect() (url string, state string, err error) {
	return s.googleAuth.Redirect()
}

// TODO
func (s *Service) GetCurrentUser(ctx context.Context, userId uuid.UUID) (db.User, error) {
	return s.queries.GetUserByID(ctx, userId)
}

func (s *Service) GetCreditGrants(ctx context.Context, userId uuid.UUID, now time.Time) ([]db.CreditGrant, error) {
	return s.queries.GetUserCreditGrants(ctx, db.GetUserCreditGrantsParams{
		UserID:    userId,
		ExpiresAt: &now,
	})
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.auth.RevokeSession(ctx, s.queries, refreshToken)
}

func (s *Service) GetQueries() *db.Queries {
	return s.queries
}

func New(
	db *pgxpool.Pool,
	queries *db.Queries,
	auth *auth.Module,
	googleauth *googleauth.Module,
	billing *billing.Module,
	emailModule *email.Module,
) *Service {
	return &Service{
		db:         db,
		queries:    queries,
		auth:       auth,
		googleAuth: googleauth,
		billing:    billing,
		email:      emailModule,
	}
}
