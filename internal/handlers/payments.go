package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"DebtEase/internal/models"
	"DebtEase/internal/service"
	"DebtEase/internal/utils"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func HandleRecordPayment(cfg *api.Config, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		debtIDStr := r.PathValue("id")
		debtID, err := uuid.Parse(debtIDStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}

		var req models.RecordPaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		if req.Type == "extra" && (req.Amount == nil || req.Amount.LessThanOrEqual(decimal.Zero)) {
			utils.RespondWithError(w, http.StatusBadRequest, "amount required for extra payments")
			return
		}
		if req.PaymentDate.IsZero() {
			utils.RespondWithError(w, http.StatusBadRequest, "payment_date is required")
			return
		}

		// Determine amount based on type if omitted
		ctx := context.Background()
		q := cfg.DB
		debt, err := q.GetDebtForUpdate(ctx, database.GetDebtForUpdateParams{ID: debtID, UserID: userID})
		if err != nil {
			if err == sql.ErrNoRows {
				utils.RespondWithError(w, http.StatusNotFound, "Debt not found")
				return
			}
			logger.Logger.Errorw("GetDebtByID failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load debt")
			return
		}

		// Accrue to payment date for auto amount computation context
		accrued := debt.AccruedInterest

		// If last accrual date before payment_date, estimate additional days for auto amount calc (service will persist real accrual)
		last := req.PaymentDate
		if !debt.LastAccrualDate.IsZero() {
			last = debt.LastAccrualDate
		}
		start := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, time.UTC)
		end := time.Date(req.PaymentDate.Year(), req.PaymentDate.Month(), req.PaymentDate.Day(), 0, 0, 0, 0, time.UTC)
		days := int(end.Sub(start).Hours() / 24)
		if days > 0 {
			dailyRate := debt.InterestRate.Div(decimal.NewFromInt(36500))
			accrued = accrued.Add(debt.OutstandingBalance.Mul(dailyRate).Mul(decimal.NewFromInt(int64(days))))
		}

		totalDue := debt.OutstandingBalance.Add(accrued)
		amount := decimal.Zero
		if req.Amount != nil {
			amount = *req.Amount
			if amount.GreaterThan(totalDue) {
				amount = totalDue
			}
		} else {
			switch req.Type {
			case "minimum":
				amount = debt.MinPayment
				if amount.LessThan(accrued) {
					amount = accrued
				}
				if amount.GreaterThan(totalDue) {
					amount = totalDue
				}
			case "full":
				amount = totalDue
			default:
				utils.RespondWithError(w, http.StatusBadRequest, "amount required for extra payments")
				return
			}
		}
		if amount.LessThanOrEqual(decimal.Zero) {
			utils.RespondWithError(w, http.StatusBadRequest, "amount must be > 0")
			return
		}

		svc := service.NewPaymentService(cfg, db)
		payment, updated, err := svc.RecordPayment(ctx, service.RecordPaymentParams{
			UserID:        userID,
			DebtID:        debtID,
			Amount:        amount,
			PaymentDate:   req.PaymentDate,
			Type:          database.PaymentType(req.Type),
			Notes:         req.Notes,
			PaymentMethod: req.PaymentMethod,
		})
		if err != nil {
			logger.Logger.Errorw("RecordPayment failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to record payment")
			return
		}

		resp := models.PaymentResponse{
			ID:               payment.ID.String(),
			DebtID:           payment.DebtID.String(),
			Amount:           payment.Amount,
			PaymentDate:      payment.PaymentDate.Format("2006-01-02"),
			Type:             string(payment.Type),
			Notes:            toPtrIfValid(payment.Notes),
			PrincipalApplied: payment.PrincipalApplied,
			InterestApplied:  payment.InterestApplied,
			PaymentMethod:    toPtrIfValid(payment.PaymentMethod),
			CreatedAt:        firstValidTime(payment.CreatedAt, time.Now()),
		}

		utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"payment": resp,
			"debt":    toDebtResponse(updated),
		})
	}
}

func HandleListPayments(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromContext(r.Context())
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		debtID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		// Verify ownership
		ctx := context.Background()
		if _, err := cfg.DB.GetDebtForUpdate(ctx, database.GetDebtForUpdateParams{ID: debtID, UserID: userID}); err != nil {
			if err == sql.ErrNoRows {
				utils.RespondWithError(w, http.StatusNotFound, "Debt not found")
				return
			}
			logger.Logger.Errorw("GetDebtByID failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load debt")
			return
		}

		limit := 50
		offset := 0
		items, err := cfg.DB.ListPaymentsByDebt(ctx, database.ListPaymentsByDebtParams{DebtID: debtID, Limit: int32(limit), Offset: int32(offset)})
		if err != nil {
			logger.Logger.Errorw("ListPaymentsByDebt failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list payments")
			return
		}
		out := make([]models.PaymentResponse, 0, len(items))
		for _, p := range items {
			out = append(out, models.PaymentResponse{
				ID:               p.ID.String(),
				DebtID:           p.DebtID.String(),
				Amount:           p.Amount,
				PaymentDate:      p.PaymentDate.Format("2006-01-02"),
				Type:             string(p.Type),
				Notes:            toPtrIfValid(p.Notes),
				PrincipalApplied: p.PrincipalApplied,
				InterestApplied:  p.InterestApplied,
				PaymentMethod:    toPtrIfValid(p.PaymentMethod),
				CreatedAt:        firstValidTime(p.CreatedAt, time.Now()),
			})
		}
		utils.RespondWithJSON(w, http.StatusOK, out)
	}
}

func toPtrIfValid(ns sql.NullString) *string {
	if ns.Valid {
		s := ns.String
		return &s
	}
	return nil
}

func firstValidTime(nt sql.NullTime, fallback time.Time) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return fallback
}
