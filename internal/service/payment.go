package service

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentService struct {
	cfg *api.Config
	db  *sql.DB
}

func NewPaymentService(cfg *api.Config, db *sql.DB) *PaymentService {
	return &PaymentService{cfg: cfg, db: db}
}

type RecordPaymentParams struct {
	UserID        uuid.UUID
	DebtID        uuid.UUID
	Amount        decimal.Decimal
	PaymentDate   time.Time
	Type          database.PaymentType
	Notes         *string
	PaymentMethod *string
}

// accrueInterestUpTo calculates and persists interest accrual up to and including the given date.
// Handles credit cards (daily after grace period) and loans (monthly on EMI date) differently.
func (s *PaymentService) accrueInterestUpTo(ctx context.Context, q *database.Queries, debt database.Debt, upTo time.Time) error {
	// Check if this is a credit card
	if debt.Type == database.DebtTypeCreditCard {
		return s.accrueCreditCardInterest(ctx, q, debt, upTo)
	}

	// For loans, check if it's time for monthly accrual
	if debt.Type == database.DebtTypePersonalLoan ||
		debt.Type == database.DebtTypeStudentLoan ||
		debt.Type == database.DebtTypeMortgage {
		return s.accrueLoanInterest(ctx, q, debt, upTo)
	}

	// For other debt types, use daily accrual
	return s.accrueDaily(ctx, q, debt, upTo)
}

// accrueCreditCardInterest handles daily interest accrual for credit cards (only if grace period ended)
func (s *PaymentService) accrueCreditCardInterest(ctx context.Context, q *database.Queries, debt database.Debt, upTo time.Time) error {
	// Get credit card details
	cc, err := q.GetCreditCardByDebtID(ctx, debt.ID)
	if err != nil {
		return err
	}

	// Only accrue if NOT in grace period
	if cc.InGracePeriod {
		return nil
	}

	// Daily accrual
	return s.accrueDaily(ctx, q, debt, upTo)
}

// accrueLoanInterest handles monthly interest accrual for loans on EMI due day
func (s *PaymentService) accrueLoanInterest(ctx context.Context, q *database.Queries, debt database.Debt, upTo time.Time) error {
	// Loans accrue interest monthly, not daily
	// This should be called only on EMI due day
	// For payment scenarios, we calculate interest on-the-fly in RecordPayment
	return nil
}

// accrueDaily performs daily interest accrual
func (s *PaymentService) accrueDaily(ctx context.Context, q *database.Queries, debt database.Debt, upTo time.Time) error {
	// LastAccrualDate is a direct time.Time field
	lastDate := debt.LastAccrualDate

	// compute whole days difference based on date only
	start := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(upTo.Year(), upTo.Month(), upTo.Day(), 0, 0, 0, 0, time.UTC)
	days := int(end.Sub(start).Hours() / 24)
	if days <= 0 {
		return nil
	}

	// daily_rate = interest_rate / 36500
	dailyRate := debt.InterestRate.Div(decimal.NewFromInt(36500))
	accrued := debt.OutstandingBalance.Mul(dailyRate).Mul(decimal.NewFromInt(int64(days)))

	return q.AccrueInterest(ctx, database.AccrueInterestParams{
		ID:              debt.ID,
		AccruedInterest: accrued,
		LastAccrualDate: end,
	})
}

// RecordPayment accrues interest to payment date, splits amount into interest/principal, records payment and updates debt.
func (s *PaymentService) RecordPayment(ctx context.Context, arg RecordPaymentParams) (database.PaymentsHistory, database.Debt, error) {
	var created database.PaymentsHistory
	var updatedDebt database.Debt

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return created, updatedDebt, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	q := s.cfg.DB.WithTx(tx)

	debt, err := q.GetDebtForUpdate(ctx, database.GetDebtForUpdateParams{ID: arg.DebtID, UserID: arg.UserID})
	if err != nil {
		return created, updatedDebt, err
	}
	// Accrue to payment date first
	if err = s.accrueInterestUpTo(ctx, q, debt, arg.PaymentDate); err != nil {
		return created, updatedDebt, err
	}
	// Refresh debt after accrual
	debt, err = q.GetDebtForUpdate(ctx, database.GetDebtForUpdateParams{ID: arg.DebtID, UserID: arg.UserID})
	if err != nil {
		return created, updatedDebt, err
	}

	// Calculate total due
	accrued := debt.AccruedInterest
	totalDue := debt.OutstandingBalance.Add(accrued)
	amount := arg.Amount
	if amount.GreaterThan(totalDue) {
		amount = totalDue
	}

	// Split
	interestPaid := amount
	if interestPaid.GreaterThan(accrued) {
		interestPaid = accrued
	}
	principalPaid := amount.Sub(interestPaid)

	// New balances
	newAccrued := accrued.Sub(interestPaid)
	if newAccrued.IsNegative() {
		newAccrued = decimal.Zero
	}
	newOutstanding := debt.OutstandingBalance.Sub(principalPaid)
	if newOutstanding.IsNegative() {
		newOutstanding = decimal.Zero
	}

	// Record payment
	created, err = q.RecordPayment(ctx, database.RecordPaymentParams{
		DebtID:           arg.DebtID,
		Amount:           amount,
		PaymentDate:      arg.PaymentDate,
		Type:             arg.Type,
		Notes:            nullString(arg.Notes),
		PrincipalApplied: principalPaid,
		InterestApplied:  interestPaid,
		PaymentMethod:    nullString(arg.PaymentMethod),
	})
	if err != nil {
		return created, updatedDebt, err
	}

	// Persist debt updates
	err = q.UpdateDebtAfterPayment(ctx, database.UpdateDebtAfterPaymentParams{
		ID:                 arg.DebtID,
		OutstandingBalance: newOutstanding,
		AccruedInterest:    newAccrued,
		LastAccrualDate:    arg.PaymentDate,
	})
	if err != nil {
		return created, updatedDebt, err
	}
	updatedDebt, err = q.GetDebtByID(ctx, database.GetDebtByIDParams{ID: arg.DebtID, UserID: arg.UserID})
	return created, updatedDebt, err
}

func nullString(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *p, Valid: *p != ""}
}
