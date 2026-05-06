package worker

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"
	"database/sql"
	"time"
)

// ProcessCreditCardStatements generates monthly statements for credit cards on their billing day
func ProcessCreditCardStatements(ctx context.Context, q *database.Queries) {
	// Get all credit cards needing statement generation
	cards, err := q.ListCreditCardsNeedingStatement(ctx)
	if err != nil {
		logger.Logger.Errorw("ListCreditCardsNeedingStatement failed", "error", err)
		return
	}

	today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)

	for _, card := range cards {
		// Calculate payment due date: billing_day + grace_period_days
		paymentDueDate := today.AddDate(0, 0, int(card.GracePeriodDays))

		// Generate statement: move unbilled to billed, reset unbilled, set grace period
		err := q.GenerateStatement(ctx, database.GenerateStatementParams{
			DebtID:            card.DebtID,
			LastStatementDate: sql.NullTime{Time: today, Valid: true},
			PaymentDueDate:    sql.NullTime{Time: paymentDueDate, Valid: true},
		})

		if err != nil {
			logger.Logger.Errorw("GenerateStatement failed",
				"debt_id", card.DebtID,
				"card_name", card.Name,
				"error", err)
			continue
		}

		logger.Logger.Infow("Credit card statement generated",
			"debt_id", card.DebtID,
			"card_name", card.Name,
			"billing_day", card.BillingDay,
			"payment_due_date", paymentDueDate.Format("2006-01-02"))
	}
}

// StartCreditCardStatementWorker starts a worker that runs daily to generate statements
func StartCreditCardStatementWorker(ctx context.Context, q *database.Queries, tick time.Duration) {
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
			case <-ticker.C:
				ProcessCreditCardStatements(ctx, q)
			}
		}
	}()
}
