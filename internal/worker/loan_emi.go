package worker

import (
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// ProcessLoanEMI handles EMI accrual and moratorium for loans on their due day
func ProcessLoanEMI(ctx context.Context, q *database.Queries) {
	// Get all loans with payment due today
	loans, err := q.ListLoansNeedingEMI(ctx)
	if err != nil {
		logger.Logger.Errorw("ListLoansNeedingEMI failed", "error", err)
		return
	}

	today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)

	for _, loan := range loans {
		if loan.OutstandingBalance.LessThanOrEqual(decimal.Zero) {
			continue
		}

		// Check if in moratorium
		if loan.InMoratorium.Valid && loan.InMoratorium.Bool {
			// Check if moratorium ended
			if loan.MoratoriumUntil.Valid && today.After(loan.MoratoriumUntil.Time) {
				// Exit moratorium: recalculate EMI with capitalized interest
				remainingMonths := int32(0)
				if loan.TenureMonths.Valid && loan.MonthsPaid.Valid {
					remainingMonths = loan.TenureMonths.Int32 - loan.MonthsPaid.Int32
				}

				if remainingMonths > 0 {
					// Recalculate EMI
					emi := calculateEMI(loan.OutstandingBalance, loan.InterestRate, int(remainingMonths))

					err := q.ExitMoratorium(ctx, database.ExitMoratoriumParams{
						ID:        loan.ID,
						EmiAmount: sql.NullString{String: emi.String(), Valid: true},
					})

					if err != nil {
						logger.Logger.Errorw("ExitMoratorium failed",
							"debt_id", loan.ID,
							"loan_name", loan.Name,
							"error", err)
						continue
					}

					logger.Logger.Infow("Loan exited moratorium",
						"debt_id", loan.ID,
						"loan_name", loan.Name,
						"new_emi", emi.String(),
						"remaining_months", remainingMonths)
				}
				continue
			} else {
				// Still in moratorium: capitalize interest
				monthlyRate := loan.InterestRate.Div(decimal.NewFromInt(1200))
				interest := loan.OutstandingBalance.Mul(monthlyRate)

				// Add interest to outstanding balance (capitalize)
				newBalance := loan.OutstandingBalance.Add(interest)

				err := q.UpdateDebtAfterPayment(ctx, database.UpdateDebtAfterPaymentParams{
					ID:                 loan.ID,
					OutstandingBalance: newBalance,
					AccruedInterest:    decimal.Zero,
					LastAccrualDate:    today,
				})

				if err != nil {
					logger.Logger.Errorw("Capitalize moratorium interest failed",
						"debt_id", loan.ID,
						"loan_name", loan.Name,
						"error", err)
					continue
				}

				logger.Logger.Infow("Moratorium interest capitalized",
					"debt_id", loan.ID,
					"loan_name", loan.Name,
					"interest", interest.String(),
					"new_balance", newBalance.String())
				continue
			}
		}

		// Normal EMI processing: calculate monthly interest
		monthlyRate := loan.InterestRate.Div(decimal.NewFromInt(1200))
		interestComponent := loan.OutstandingBalance.Mul(monthlyRate)

		// Get EMI amount
		var emiAmount decimal.Decimal
		if loan.EmiAmount.Valid {
			var err error
			emiAmount, err = decimal.NewFromString(loan.EmiAmount.String)
			if err != nil {
				logger.Logger.Errorw("Invalid EMI amount",
					"debt_id", loan.ID,
					"loan_name", loan.Name,
					"error", err)
				continue
			}
		} else {
			logger.Logger.Warnw("Loan has no EMI amount set",
				"debt_id", loan.ID,
				"loan_name", loan.Name)
			continue
		}

		principalComponent := emiAmount.Sub(interestComponent)

		// Handle last payment
		if principalComponent.GreaterThanOrEqual(loan.OutstandingBalance) {
			principalComponent = loan.OutstandingBalance
		}

		newBalance := loan.OutstandingBalance.Sub(principalComponent)
		if newBalance.IsNegative() {
			newBalance = decimal.Zero
		}

		// Update debt: reduce balance, increment months_paid
		err := q.UpdateDebtAfterPayment(ctx, database.UpdateDebtAfterPaymentParams{
			ID:                 loan.ID,
			OutstandingBalance: newBalance,
			AccruedInterest:    decimal.Zero,
			LastAccrualDate:    today,
		})

		if err != nil {
			logger.Logger.Errorw("UpdateDebtAfterPayment failed",
				"debt_id", loan.ID,
				"loan_name", loan.Name,
				"error", err)
			continue
		}

		// Increment months_paid
		err = q.UpdateLoanPayment(ctx, loan.ID)
		if err != nil {
			logger.Logger.Errorw("UpdateLoanPayment failed",
				"debt_id", loan.ID,
				"loan_name", loan.Name,
				"error", err)
		}

		logger.Logger.Infow("Loan EMI processed",
			"debt_id", loan.ID,
			"loan_name", loan.Name,
			"emi", emiAmount.String(),
			"interest", interestComponent.String(),
			"principal", principalComponent.String(),
			"new_balance", newBalance.String())
	}
}

// calculateEMI calculates EMI using reducing balance formula
func calculateEMI(principal, annualRate decimal.Decimal, months int) decimal.Decimal {
	if months == 0 {
		return decimal.Zero
	}

	// Convert annual rate to monthly rate: APR / 1200
	monthlyRate := annualRate.Div(decimal.NewFromInt(1200))

	if monthlyRate.IsZero() {
		return principal.Div(decimal.NewFromInt(int64(months)))
	}

	// Calculate (1 + r)
	onePlusR := decimal.NewFromInt(1).Add(monthlyRate)

	// Calculate (1 + r)^n using Pow
	onePlusRPowN := onePlusR.Pow(decimal.NewFromInt(int64(months)))

	// EMI = P * r * (1+r)^n / ((1+r)^n - 1)
	numerator := principal.Mul(monthlyRate).Mul(onePlusRPowN)
	denominator := onePlusRPowN.Sub(decimal.NewFromInt(1))

	emi := numerator.Div(denominator)
	return emi.Round(2)
}

// StartLoanEMIWorker starts a worker that runs daily to process loan EMIs
func StartLoanEMIWorker(ctx context.Context, q *database.Queries, tick time.Duration) {
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
				ProcessLoanEMI(ctx, q)
			}
		}
	}()
}
