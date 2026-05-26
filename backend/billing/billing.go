package billing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twinspeak/backend/db"
)

const (
	MaxCreditsPerSession = 60
	UnreechableCredits = 70
)

var ErrInsufficientCredits = errors.New("insufficient credits")

type Billing struct {
	db      *pgxpool.Pool
	queries *db.Queries
}

func (b *Billing) SpendCredits(ctx context.Context, now time.Time, userId uuid.UUID, amountSpending int32) error {
	// TODO: This is pretty stupid and dangerous
	if amountSpending > UnreechableCredits {
		panic("Max credits exceeded")
	}

	tx, err := b.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("cannot start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := b.queries.WithTx(tx)

	grant, err := qtx.FindCreditGrantForSpend(ctx, db.FindCreditGrantForSpendParams{
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
	fmt.Println(grant.RemainingAmount)
	// We assume that no session will exceed ~60 seconds, so we round up in the favor of user.
	// That prevents search for the next grant.
	remainingCredits := math.Max(float64(grant.RemainingAmount-amountSpending), 0)
	err = qtx.UpdateGrant(ctx, db.UpdateGrantParams{
		ID:              grant.ID,
		RemainingAmount: int32(remainingCredits),
	})
	if err != nil {
		return fmt.Errorf("cannot update credit grant in db: %w", err)
	}

	err = qtx.CreateCreditExpenses(ctx, db.CreateCreditExpensesParams{
		UserID:  userId,
		GrantID: grant.ID,
		Spent:   amountSpending,
		SpentAt: now,
	})
	if err != nil {
		return fmt.Errorf("cannot crete credit expense in db: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("cannot commit db transaction: %w", err)
	}

	return nil
}

func (b *Billing) BuyTopup(
	ctx context.Context,
	now time.Time,
	userId uuid.UUID,
	amount int32,
	expiresAt time.Time,
) error {
	err := b.queries.CreateCreditGrant(ctx, db.CreateCreditGrantParams{
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

func (b *Billing) BuyMonthly(
	ctx context.Context,
	now time.Time,
	userId uuid.UUID,
	amount int32,
	expiresAt time.Time,
) error {
	err := b.queries.CreateCreditGrant(ctx, db.CreateCreditGrantParams{
		UserID:          userId,
		Amount:          amount,
		RemainingAmount: amount,
		Type:            db.CreditGrantTypeMonthly,
		ExpiresAt:       &expiresAt,
	})
	if err != nil {
		return fmt.Errorf("cannot insert credit grant to db: %w", err)
	}
	return nil
}

func NewBilling(db *pgxpool.Pool, queries *db.Queries) *Billing {
	return &Billing{
		db:      db,
		queries: queries,
	}
}
