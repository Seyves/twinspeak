package billing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/twinspeak/backend/internal/db"
)

const (
	MaxCreditsPerSession = 60
	UnreechableCredits   = 70

	// 30 minutes
	MonthlyCredits = 30 * 60
)

var ErrInsufficientCredits = errors.New("insufficient credits")

type Module struct{}

func (m *Module) SpendCredits(ctx context.Context, tx *db.Queries, userId uuid.UUID, now time.Time, amountSpending int32) error {
	// TODO: This is pretty stupid and dangerous
	if amountSpending > UnreechableCredits {
		panic("Max credits exceeded")
	}
	grant, err := tx.FindCreditGrantForSpend(ctx, db.FindCreditGrantForSpendParams{
		UserID:    userId,
		ExpiresAt: &now,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInsufficientCredits
		} else {
			return fmt.Errorf("cannot select credit grant from db: %w", err)
		}
	}
	// We assume that no session will exceed ~60 seconds, so we round up in the favor of user.
	// That prevents search for the next grant.
	remainingCredits := math.Max(float64(grant.RemainingAmount-amountSpending), 0)
	err = tx.UpdateGrant(ctx, db.UpdateGrantParams{
		ID:              grant.ID,
		RemainingAmount: int32(remainingCredits),
	})
	if err != nil {
		return fmt.Errorf("cannot update credit grant in db: %w", err)
	}

	err = tx.CreateCreditExpenses(ctx, db.CreateCreditExpensesParams{
		UserID:  userId,
		GrantID: grant.ID,
		Spent:   amountSpending,
		SpentAt: now,
	})
	if err != nil {
		return fmt.Errorf("cannot insert credit expense in db: %w", err)
	}
	return nil
}

func (m *Module) CheckAvailableCredits(ctx context.Context, tx *db.Queries, userId uuid.UUID, now time.Time) error {
	_, err := tx.FindCreditGrantForSpend(ctx, db.FindCreditGrantForSpendParams{
		UserID:    userId,
		ExpiresAt: &now,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInsufficientCredits
		} else {
			return fmt.Errorf("cannot select credit grant from db: %w", err)
		}
	}
	return nil
}

func (m *Module) StartSubscription(ctx context.Context, tx *db.Queries, userId uuid.UUID, now time.Time) error {
	expiresAt := now.AddDate(0, 1, 0)
	_, err := tx.CreateSubscription(ctx, db.CreateSubscriptionParams{
		UserID:             userId,
		NextMonthlyGrantAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("cannot insert subscription to db: %w", err)
	}
	err = tx.CreateCreditGrant(ctx, db.CreateCreditGrantParams{
		UserID:          userId,
		Amount:          MonthlyCredits,
		RemainingAmount: MonthlyCredits,
		Type:            db.CreditGrantTypeMonthly,
		ExpiresAt:       &expiresAt,
	})
	if err != nil {
		return fmt.Errorf("cannot insert credit grant to db: %w", err)
	}
	return nil
}

func (m *Module) RenewSubscription(ctx context.Context, tx *db.Queries, userId uuid.UUID, now time.Time) error {
	expiresAt := now.AddDate(0, 1, 0)
	err := tx.UpdateSubscription(ctx, db.UpdateSubscriptionParams{
		UserID:             userId,
		NextMonthlyGrantAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("cannot update subscription in db: %w", err)
	}
	err = tx.CreateCreditGrant(ctx, db.CreateCreditGrantParams{
		UserID:          userId,
		Amount:          MonthlyCredits,
		RemainingAmount: MonthlyCredits,
		Type:            db.CreditGrantTypeMonthly,
		ExpiresAt:       &expiresAt,
	})
	if err != nil {
		return fmt.Errorf("cannot insert credit grant to db: %w", err)
	}
	return nil
}

func (m *Module) GetExpiredSubscriptions(ctx context.Context, tx *db.Queries, now time.Time) ([]uuid.UUID, error) {
	subs, err := tx.GetExpiredSubscriptions(ctx, now)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subs, err
		}
		return subs, fmt.Errorf("cannot select expired subscriptions from db: %w", err)
	}
	return subs, err
}

func (m *Module) BuyTopup(
	ctx context.Context,
	tx *db.Queries,
	userId uuid.UUID,
	now time.Time,
	amount int32,
	expiresAt time.Time,
) error {
	err := tx.CreateCreditGrant(ctx, db.CreateCreditGrantParams{
		UserID:          userId,
		Amount:          amount,
		RemainingAmount: amount,
		Type:            db.CreditGrantTypeTopup,
		ExpiresAt:       &expiresAt,
	})
	if err != nil {
		return fmt.Errorf("cannot insert credit grant to db: %w", err)
	}
	return nil
}

func New() *Module {
	return &Module{}
}
