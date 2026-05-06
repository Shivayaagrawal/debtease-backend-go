package middleware

import (
	"DebtEase/internal/api"
	"DebtEase/internal/auth"
	"DebtEase/internal/logger"
	"DebtEase/internal/utils"
	"context"
	"net/http"
)

type contextKey string

const ContextUserIDKey contextKey = "userID"

// AuthRequired validates the Authorization bearer token and injects userID into context.
func AuthRequired(cfg *api.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			logger.Logger.Infow("JWT validation failed", "error", err)
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
