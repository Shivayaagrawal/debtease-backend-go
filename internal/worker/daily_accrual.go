package worker

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"
	"time"

	"github.com/shopspring/decimal"
)

func StartDailyAccrual(ctx context.Context, q *database.Queries, tick time.Duration) {
	if tick <= 0 {
		tick = 24 * time.Hour
	}
	ticker := time.NewTicker(tick)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				processAccrual(ctx, q, t)
			}
		}
	}()
}

func processAccrual(ctx context.Context, q *database.Queries, now time.Time) {
	// Use date-only for comparisons
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Get all credit cards that are out of grace period
	creditCards, err := q.ListCreditCardsOutOfGrace(ctx)
	if err != nil {
		logger.Logger.Errorw("ListCreditCardsOutOfGrace failed", "error", err)
		return
	}

	// Accrue daily interest only for credit cards out of grace period
	for _, cc := range creditCards {
		// LastAccrualDate is a direct time.Time field from the joined debts table
		last := cc.LastAccrualDate
		lastDate := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, time.UTC)

		days := int(today.Sub(lastDate).Hours() / 24)
		if days <= 0 {
			continue
		}

		// daily_rate = interest_rate / 36500
		dailyRate := cc.InterestRate.Div(decimal.NewFromInt(36500))
		delta := cc.OutstandingBalance.Mul(dailyRate).Mul(decimal.NewFromInt(int64(days)))

		// Accrue interest on debt
		if err := q.AccrueInterest(ctx, database.AccrueInterestParams{
			ID:              cc.DebtID,
			AccruedInterest: delta,
			LastAccrualDate: today,
		}); err != nil {
			logger.Logger.Errorw("AccrueInterest failed", "debt_id", cc.DebtID, "error", err)
			continue
		}

		// Add to unbilled balance on credit card
		if err := q.AddDailyInterestToUnbilled(ctx, database.AddDailyInterestToUnbilledParams{
			DebtID:          cc.DebtID,
			UnbilledBalance: delta,
		}); err != nil {
			logger.Logger.Errorw("AddDailyInterestToUnbilled failed", "debt_id", cc.DebtID, "error", err)
		}
	}

	// NOTE: Loans do NOT accrue daily. They accrue monthly on EMI due date.
	// Loan interest accrual is handled in the payment/EMI processing logic.
}
