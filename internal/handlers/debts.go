package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"DebtEase/internal/middleware"
	"DebtEase/internal/models"
	"DebtEase/internal/utils"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func getUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(middleware.ContextUserIDKey)
	if val == nil {
		return uuid.Nil, false
	}
	uid, ok := val.(uuid.UUID)
	return uid, ok
}

func HandleCreateDebt(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		var req models.CreateDebtRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		if req.Name == "" || req.Type == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing required fields")
			return
		}

		// req.DueDate is time.Time per model; JSON must be RFC3339. We accept as-is.
		dueDate := req.DueDate

		var paymentFreq database.NullPaymentFreq
		if req.PaymentFrequency != nil && *req.PaymentFrequency != "" {
			paymentFreq = database.NullPaymentFreq{PaymentFreq: database.PaymentFreq(*req.PaymentFrequency), Valid: true}
		}

		var debtTaken sql.NullTime
		if req.DebtTaken != nil {
			debtTaken = sql.NullTime{Time: *req.DebtTaken, Valid: true}
		}

		var lender sql.NullString
		if req.Lender != nil {
			lender = sql.NullString{String: *req.Lender, Valid: *req.Lender != ""}
		}

		var risk database.NullRiskLevel
		if req.RiskLevel != nil && *req.RiskLevel != "" {
			risk = database.NullRiskLevel{RiskLevel: database.RiskLevel(*req.RiskLevel), Valid: true}
		}

		var priority sql.NullInt16
		if req.Priority != nil {
			priority = sql.NullInt16{Int16: *req.Priority, Valid: true}
		}

		var notes sql.NullString
		if req.Notes != nil {
			notes = sql.NullString{String: *req.Notes, Valid: *req.Notes != ""}
		}

		// Handle loan-specific fields
		var paymentDueDay sql.NullInt32
		if req.PaymentDueDay != nil {
			paymentDueDay = sql.NullInt32{Int32: *req.PaymentDueDay, Valid: true}
		}

		var emiAmount sql.NullString
		if req.EMIAmount != nil {
			emiAmount = sql.NullString{String: req.EMIAmount.String(), Valid: true}
		} else if req.Type == "personal_loan" || req.Type == "student_loan" || req.Type == "mortgage" {
			// Calculate EMI if not provided for loan types
			if req.TenureMonths != nil && *req.TenureMonths > 0 {
				emi := calculateEMI(req.Principal, req.InterestRate, int(*req.TenureMonths))
				emiAmount = sql.NullString{String: emi.String(), Valid: true}
			}
		}

		var tenureMonths sql.NullInt32
		if req.TenureMonths != nil {
			tenureMonths = sql.NullInt32{Int32: *req.TenureMonths, Valid: true}
		}

		var moratoriumUntil sql.NullTime
		if req.MoratoriumUntil != nil {
			moratoriumUntil = sql.NullTime{Time: *req.MoratoriumUntil, Valid: true}
		}

		inMoratorium := req.MoratoriumUntil != nil && req.MoratoriumUntil.After(time.Now())

		params := database.CreateDebtParams{
			UserID:             userID,
			Name:               req.Name,
			Type:               database.DebtType(req.Type),
			Lender:             lender,
			PrincipalDecimal:   req.Principal,
			OutstandingBalance: req.OutstandingBalance,
			InterestRate:       req.InterestRate,
			MinPayment:         req.MinPayment,
			DueDate:            dueDate,
			PaymentFrequency:   paymentFreq,
			DebtTaken:          debtTaken,
			RiskLevel:          risk,
			Priority:           priority,
			Notes:              notes,
			PaymentDueDay:      paymentDueDay,
			EmiAmount:          emiAmount,
			TenureMonths:       tenureMonths,
			MonthsPaid:         sql.NullInt32{Int32: 0, Valid: true},
			MoratoriumUntil:    moratoriumUntil,
			InMoratorium:       sql.NullBool{Bool: inMoratorium, Valid: true},
		}

		ctx := context.Background()
		debt, err := cfg.DB.CreateDebt(ctx, params)
		if err != nil {
			logger.Logger.Errorw("CreateDebt failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create debt")
			return
		}

		// If credit card, create credit_cards entry
		if req.Type == "credit_card" {
			if req.BillingDay == nil || req.GracePeriodDays == nil {
				utils.RespondWithError(w, http.StatusBadRequest, "billing_day and grace_period_days required for credit cards")
				return
			}

			minPayPercent := decimal.NewFromFloat(5.00)
			if req.MinPaymentPercent != nil {
				minPayPercent = *req.MinPaymentPercent
			}

			ccParams := database.CreateCreditCardParams{
				DebtID:            debt.ID,
				BillingDay:        *req.BillingDay,
				GracePeriodDays:   *req.GracePeriodDays,
				MinPaymentPercent: minPayPercent,
				BilledBalance:     decimal.Zero,
				UnbilledBalance:   req.OutstandingBalance,
				InGracePeriod:     true,
			}

			_, err := cfg.DB.CreateCreditCard(ctx, ccParams)
			if err != nil {
				logger.Logger.Errorw("CreateCreditCard failed", "error", err)
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create credit card")
				return
			}
		}

		utils.RespondWithJSON(w, http.StatusCreated, toDebtResponse(debt))
	}
}

func HandleListDebts(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		ctx := context.Background()
		items, err := cfg.DB.ListUserDebts(ctx, userID)
		if err != nil {
			logger.Logger.Errorw("ListUserDebts failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list debts")
			return
		}
		resp := make([]models.DebtResponse, 0, len(items))
		for _, d := range items {
			resp = append(resp, toDebtResponse(d))
		}
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func HandleGetDebt(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		idStr := r.PathValue("id")
		debtID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		ctx := context.Background()
		debt, err := cfg.DB.GetDebtByID(ctx, database.GetDebtByIDParams{ID: debtID, UserID: userID})
		if err != nil {
			if err == sql.ErrNoRows {
				utils.RespondWithError(w, http.StatusNotFound, "Debt not found")
				return
			}
			logger.Logger.Errorw("GetDebtByID failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch debt")
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, toDebtResponse(debt))
	}
}

func HandleUpdateDebtBalance(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getUserIDFromContext(r.Context()); !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		idStr := r.PathValue("id")
		debtID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		var req models.UpdateDebtBalanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		// OutstandingBalance is required decimal
		if req.OutstandingBalance.IsZero() && !req.OutstandingBalance.GreaterThan(decimal.Zero) {
			utils.RespondWithError(w, http.StatusBadRequest, "outstanding_balance is required")
			return
		}

		isOverdue := sql.NullBool{Bool: req.IsOverdue, Valid: true}

		ctx := context.Background()
		if err := cfg.DB.UpdateDebtBalance(ctx, database.UpdateDebtBalanceParams{
			ID:                 debtID,
			OutstandingBalance: req.OutstandingBalance,
			IsOverdue:          isOverdue,
		}); err != nil {
			logger.Logger.Errorw("UpdateDebtBalance failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update balance")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleUpdateDebtHealth(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getUserIDFromContext(r.Context()); !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		idStr := r.PathValue("id")
		debtID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		var req models.UpdateDebtHealthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		var score sql.NullString
		if req.DebtHealthScore != nil {
			score = sql.NullString{String: req.DebtHealthScore.StringFixed(2), Valid: true}
		}

		ctx := context.Background()
		if err := cfg.DB.UpdateDebtHealthScore(ctx, database.UpdateDebtHealthScoreParams{
			ID:              debtID,
			DebtHealthScore: score,
		}); err != nil {
			logger.Logger.Errorw("UpdateDebtHealthScore failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update health score")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleMarkDebtOverdue(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getUserIDFromContext(r.Context()); !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		idStr := r.PathValue("id")
		debtID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		ctx := context.Background()
		if err := cfg.DB.MarkDebtOverdue(ctx, debtID); err != nil {
			logger.Logger.Errorw("MarkDebtOverdue failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to mark overdue")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// func HandleDeleteDebt(cfg *api.Config) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if _, ok := getUserIDFromContext(r.Context()); !ok {
// 			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
// 			return
// 		}

// 		idStr := r.PathValue("id")
// 		debtID, err := uuid.Parse(idStr)
// 		if err != nil {
// 			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
// 			return
// 		}

// 		ctx := context.Background()
// 		if err := cfg.DB.DeleteDebt(ctx, debtID); err != nil {
// 			logger.Logger.Errorw("DeleteDebt failed", "error", err)
// 			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete debt")
// 			return
// 		}
// 		w.WriteHeader(http.StatusNoContent)
// 	}
// }

func toDebtResponse(d database.Debt) models.DebtResponse {
	var lender *string
	if d.Lender.Valid {
		lender = &d.Lender.String
	}
	var pf *string
	if d.PaymentFrequency.Valid {
		s := string(d.PaymentFrequency.PaymentFreq)
		pf = &s
	}
	isOverdue := false
	if d.IsOverdue.Valid {
		isOverdue = d.IsOverdue.Bool
	}
	var lateFees *decimal.Decimal
	if d.LateFees.Valid {
		if v, err := decimal.NewFromString(d.LateFees.String); err == nil {
			lateFees = &v
		}
	}
	var otherCharges *string
	if d.OtherCharges.Valid {
		otherCharges = &d.OtherCharges.String
	}
	var debtTaken *time.Time
	if d.DebtTaken.Valid {
		debtTaken = &d.DebtTaken.Time
	}
	var risk *string
	if d.RiskLevel.Valid {
		s := string(d.RiskLevel.RiskLevel)
		risk = &s
	}
	var score *decimal.Decimal
	if d.DebtHealthScore.Valid {
		if v, err := decimal.NewFromString(d.DebtHealthScore.String); err == nil {
			score = &v
		}
	}
	var priority *int16
	if d.Priority.Valid {
		p := d.Priority.Int16
		priority = &p
	}
	var notes *string
	if d.Notes.Valid {
		notes = &d.Notes.String
	}

	createdAt := time.Time{}
	if d.CreatedAt.Valid {
		createdAt = d.CreatedAt.Time
	}
	updatedAt := time.Time{}
	if d.UpdatedAt.Valid {
		updatedAt = d.UpdatedAt.Time
	}

	// AccruedInterest and LastAccrualDate are direct fields, not Null types
	accruedInterest := d.AccruedInterest
	lastAccrualDate := d.LastAccrualDate

	// Parse loan fields
	var paymentDueDay *int32
	if d.PaymentDueDay.Valid {
		v := d.PaymentDueDay.Int32
		paymentDueDay = &v
	}
	var emiAmount *decimal.Decimal
	if d.EmiAmount.Valid {
		if v, err := decimal.NewFromString(d.EmiAmount.String); err == nil {
			emiAmount = &v
		}
	}
	var tenureMonths *int32
	if d.TenureMonths.Valid {
		v := d.TenureMonths.Int32
		tenureMonths = &v
	}
	var monthsPaid *int32
	if d.MonthsPaid.Valid {
		v := d.MonthsPaid.Int32
		monthsPaid = &v
	}
	var moratoriumUntil *time.Time
	if d.MoratoriumUntil.Valid {
		moratoriumUntil = &d.MoratoriumUntil.Time
	}
	var inMoratorium *bool
	if d.InMoratorium.Valid {
		v := d.InMoratorium.Bool
		inMoratorium = &v
	}

	return models.DebtResponse{
		ID:                 d.ID.String(),
		UserID:             d.UserID.String(),
		Name:               d.Name,
		Type:               string(d.Type),
		Lender:             lender,
		Principal:          d.PrincipalDecimal,
		OutstandingBalance: d.OutstandingBalance,
		InterestRate:       d.InterestRate,
		MinPayment:         d.MinPayment,
		DueDate:            d.DueDate.Format("2006-01-02"),
		PaymentFrequency:   pf,
		IsOverdue:          isOverdue,
		LateFees:           lateFees,
		OtherCharges:       otherCharges,
		DebtTaken:          debtTaken,
		RiskLevel:          risk,
		DebtHealthScore:    score,
		Priority:           priority,
		Notes:              notes,
		AccruedInterest:    accruedInterest,
		LastAccrualDate:    lastAccrualDate,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		PaymentDueDay:      paymentDueDay,
		EMIAmount:          emiAmount,
		TenureMonths:       tenureMonths,
		MonthsPaid:         monthsPaid,
		MoratoriumUntil:    moratoriumUntil,
		InMoratorium:       inMoratorium,
	}
}

// calculateEMI calculates EMI using reducing balance formula
// EMI = P * r * (1+r)^n / ((1+r)^n - 1)
// where P = principal, r = monthly interest rate, n = number of months
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
