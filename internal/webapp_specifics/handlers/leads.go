package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	webapp_models "DebtEase/internal/webapp_specifics/models"
	webapp_service "DebtEase/internal/webapp_specifics/service"
	webapp_utils "DebtEase/internal/webapp_specifics/utils"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func HandlePdfRequest(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("PDF request", "path", r.URL.Path, "method", r.Method)

		// Get session_id from cookie
		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			logger.Logger.Warnw("Missing or invalid session_id in cookie")
			utils.RespondWithError(w, http.StatusBadRequest, "Missing or invalid session_id")
			return
		}

		// Validate session exists
		exists, err := webapp_service.SessionExists(ctx, cfg.DB, sessionID)
		if err != nil {
			logger.Logger.Errorw("Failed to check session existence", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to validate session")
			return
		}
		if !exists {
			logger.Logger.Warnw("Session does not exist", "session_id", sessionID)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid session_id")
			return
		}

		// Parse request
		var req webapp_models.PdfRequestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON for PDF request", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validate email
		if req.Email == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email is required")
			return
		}

		// Basic email validation
		if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
			return
		}

		// Emit pdf_requested event
		if err := webapp_service.EmitEvent(ctx, cfg.DB, sessionID, "pdf_requested", nil); err != nil {
			logger.Logger.Warnw("Failed to emit pdf_requested event", "error", err)
			// Don't fail the request, just log
		}

		// Create lead email
		if err := webapp_service.CreateLeadEmail(ctx, cfg.DB, req.Email, "pdf", sessionID); err != nil {
			logger.Logger.Errorw("Failed to create lead email", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save email")
			return
		}

		// Emit email_submitted event
		if err := webapp_service.EmitEvent(ctx, cfg.DB, sessionID, "email_submitted", map[string]interface{}{
			"email": req.Email,
		}); err != nil {
			logger.Logger.Warnw("Failed to emit email_submitted event", "error", err)
			// Don't fail the request, just log
		}

		// Return success
		utils.RespondWithJSON(w, http.StatusOK, webapp_models.PdfRequestResponse{
			Success: true,
			Message: "Email saved successfully",
		})
	}
}

func HandleWaitlist(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("Waitlist request", "path", r.URL.Path, "method", r.Method)

		// Get session_id from cookie (optional)
		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			logger.Logger.Infow("No session_id in cookie for waitlist request (optional)")
			sessionID = uuid.Nil // Will be handled as null by CreateLeadEmail
		} else {
			exists, err := webapp_service.SessionExists(ctx, cfg.DB, sessionID)
			if err != nil {
				// log, and probably treat as no session instead of failing waitlist
				sessionID = uuid.Nil
			} else if !exists {
				// stale/invalid cookie; ignore it to avoid FK error
				sessionID = uuid.Nil
			}
		}

		// Parse request
		var req webapp_models.WaitlistRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON for waitlist request", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		// Validate email
		if req.Email == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email is required")
			return
		}

		// Basic email validation
		if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
			return
		}

		// Create lead email with source='waitlist'
		if err := webapp_service.CreateLeadEmail(ctx, cfg.DB, req.Email, "waitlist", sessionID); err != nil {
			logger.Logger.Errorw("Failed to create lead email", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save email")
			return
		}

		// Return success
		utils.RespondWithJSON(w, http.StatusOK, webapp_models.WaitlistResponse{
			Success: true,
			Message: "Email saved successfully",
		})
	}
}
