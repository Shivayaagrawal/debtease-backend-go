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

	"github.com/google/uuid"
)

func HandlePostFinanceSteps(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("POST finance steps", "path", r.URL.Path, "method", r.Method)

		var req webapp_models.FinanceStepsRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req) // optional body; empty is ok
		}

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			if req.SessionID != nil {
				if id, parseErr := uuid.Parse(*req.SessionID); parseErr == nil {
					sessionID = id
				}
			}
			if sessionID == uuid.Nil {
				sessionID = uuid.New()
			}
		}

		engineVersion := "v1"
		if req.EngineVersion != nil && *req.EngineVersion != "" {
			engineVersion = *req.EngineVersion
		}
		sessionID, err = webapp_service.GetOrCreateSession(ctx, cfg.DB, sessionID, engineVersion)
		if err != nil {
			logger.Logger.Errorw("Failed to get or create session", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process session")
			return
		}

		resp, err := webapp_service.UpsertFinanceSteps(ctx, cfg.DB, sessionID, &req, engineVersion)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save finance steps")
			return
		}
		webapp_utils.SetSessionCookie(w, sessionID)
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func HandleGetFinanceSteps(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("GET finance steps", "path", r.URL.Path, "method", r.Method)

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			if q := r.URL.Query().Get("session_id"); q != "" {
				if id, parseErr := uuid.Parse(q); parseErr == nil {
					sessionID = id
				}
			}
		}
		if sessionID == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing or invalid session_id (cookie or query)")
			return
		}

		resp, err := webapp_service.GetFinanceSteps(ctx, cfg.DB, sessionID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to load finance steps")
			return
		}
		if resp == nil {
			utils.RespondWithError(w, http.StatusNotFound, "No finance steps found for this session")
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

// HandlePutFinanceSteps full-replace finance steps (PUT). Same session resolution as POST.
func HandlePutFinanceSteps(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("PUT finance steps", "path", r.URL.Path, "method", r.Method)

		var req webapp_models.FinanceStepsRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			if req.SessionID != nil {
				if id, parseErr := uuid.Parse(*req.SessionID); parseErr == nil {
					sessionID = id
				}
			}
			if sessionID == uuid.Nil {
				sessionID = uuid.New()
			}
		}

		engineVersion := "v1"
		if req.EngineVersion != nil && *req.EngineVersion != "" {
			engineVersion = *req.EngineVersion
		}
		sessionID, err = webapp_service.GetOrCreateSession(ctx, cfg.DB, sessionID, engineVersion)
		if err != nil {
			logger.Logger.Errorw("Failed to get or create session", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process session")
			return
		}

		resp, err := webapp_service.ReplaceFinanceSteps(ctx, cfg.DB, sessionID, &req, engineVersion)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to replace finance steps")
			return
		}
		webapp_utils.SetSessionCookie(w, sessionID)
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

// HandlePatchFinanceSteps partial-update finance steps (PATCH). Merges with existing.
func HandlePatchFinanceSteps(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("PATCH finance steps", "path", r.URL.Path, "method", r.Method)

		var req webapp_models.FinanceStepsRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			if req.SessionID != nil {
				if id, parseErr := uuid.Parse(*req.SessionID); parseErr == nil {
					sessionID = id
				}
			}
			if sessionID == uuid.Nil {
				utils.RespondWithError(w, http.StatusBadRequest, "Missing or invalid session_id (cookie or query)")
				return
			}
		}

		engineVersion := "v1"
		if req.EngineVersion != nil && *req.EngineVersion != "" {
			engineVersion = *req.EngineVersion
		}
		sessionID, err = webapp_service.GetOrCreateSession(ctx, cfg.DB, sessionID, engineVersion)
		if err != nil {
			logger.Logger.Errorw("Failed to get or create session", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process session")
			return
		}

		resp, err := webapp_service.UpsertFinanceSteps(ctx, cfg.DB, sessionID, &req, engineVersion)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update finance steps")
			return
		}
		webapp_utils.SetSessionCookie(w, sessionID)
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

// HandleDeleteFinanceSteps deletes saved finance steps for the session.
func HandleDeleteFinanceSteps(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Logger.Infow("DELETE finance steps", "path", r.URL.Path, "method", r.Method)

		sessionID, err := webapp_utils.GetSessionID(r)
		if err != nil || sessionID == uuid.Nil {
			if q := r.URL.Query().Get("session_id"); q != "" {
				if id, parseErr := uuid.Parse(q); parseErr == nil {
					sessionID = id
				}
			}
		}
		if sessionID == uuid.Nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing or invalid session_id (cookie or query)")
			return
		}

		err = webapp_service.DeleteFinanceSteps(ctx, cfg.DB, sessionID)
		if err != nil {
			logger.Logger.Errorw("Failed to delete finance steps", "error", err, "session_id", sessionID)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete finance steps")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
