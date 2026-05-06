package handlers

import (
	"DebtEase/internal/api"
	"DebtEase/internal/auth"
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"DebtEase/internal/models"
	"DebtEase/internal/utils"
	"DebtEase/internal/events"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func HandleCreateUser(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.Infow("Create user request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		var req models.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON for user creation",
				"error", err,
			)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		if req.Email == "" {
			logger.Logger.Warnw("Empty email supplied")
			utils.RespondWithError(w, http.StatusBadRequest, "Email is required")
			return
		}

		if req.Password == "" {
			logger.Logger.Warnw("Empty Password Supplied")
			utils.RespondWithError(w, http.StatusBadRequest, "Password is required")
		}

		hash, err := auth.HashPassword(req.Password)

		if err != nil {
			logger.Logger.Errorw("Failed to hash password", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process password")
			return
		}

		ctx := context.Background()
		user, err := cfg.DB.CreateUser(ctx, database.CreateUserParams{
			Email:        req.Email,
			PasswordHash: hash,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				logger.Logger.Warnw("Duplicate email attempt",
					"email", req.Email,
				)
				utils.RespondWithError(w, http.StatusConflict, "Email already exists")
				return
			}
			logger.Logger.Errorw("Failed to insert user into DB",
				"email", req.Email,
				"error", err,
			)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}
			// Publish user registered event (non-blocking, fire-and-forget)
		go func() {
			event := events.NewUserRegisteredEvent(
				user.ID,
				user.Email,
				user.Name,
				user.CreatedAt,
			)
			if err := events.PublishEvent(cfg.RabbitMQURL, events.ExchangeNotifications, events.RoutingUserRegistered, event); err != nil {
				logger.Logger.Warnw("Failed to publish user registered event (non-critical)", 
					"error", err,
					"user_id", user.ID)
			}
		}()
		resp := models.CreateUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		}

		logger.Logger.Infow("User created successfully",
			"user_id", user.ID,
			"email", user.Email,
		)

		utils.RespondWithJSON(w, http.StatusCreated, resp)
	}
}

func HandleUpdateUser(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// === 1. Extract and validate JWT ===
		tokenStr, err := auth.GetBearerToken(r.Header)
		if err != nil {
			logger.Logger.Warnw("Missing or malformed Authorization header",
				"error", err,
				"path", r.URL.Path,
			)
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		userID, err := auth.ValidateJWT(tokenStr, cfg.JWTSecret)
		if err != nil {
			logger.Logger.Infow("Invalid or expired access token",
				"error", err,
				"token_preview", auth.TruncateToken(tokenStr),
			)
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// === 2. Parse request body ===
		var req models.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Logger.Warnw("Invalid JSON payload for user update",
				"error", err,
			)
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// === 3. Validate provided fields (partial allowed) ===
		if req.Name == nil && req.Email == nil && req.Password == nil && req.MonthlyIncome == nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Provide at least one field to update")
			return
		}
		if req.Email != nil && strings.TrimSpace(*req.Email) == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email cannot be empty")
			return
		}
		if req.Password != nil && strings.TrimSpace(*req.Password) == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Password cannot be empty")
			return
		}

		// === 4. Load current user (for partial updates) ===
		ctx := context.Background()
		existing, err := cfg.DB.GetUserByID(ctx, userID)
		if err != nil {
			logger.Logger.Errorw("Failed to load existing user for update", "error", err, "user_id", userID)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}

		// === 5. Prepare updated fields ===
		finalName := existing.Name
		if req.Name != nil {
			finalName = strings.TrimSpace(*req.Name)
		}
		finalEmail := existing.Email
		if req.Email != nil {
			finalEmail = strings.TrimSpace(*req.Email)
		}
		finalPasswordHash := existing.PasswordHash
		if req.Password != nil {
			hashedPassword, hashErr := auth.HashPassword(*req.Password)
			if hashErr != nil {
				logger.Logger.Errorw("Failed to hash password", "error", hashErr)
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
				return
			}
			finalPasswordHash = hashedPassword
		}
		finalMonthlyIncome := existing.MonthlyIncome
		if req.MonthlyIncome != nil {
			finalMonthlyIncome = *req.MonthlyIncome
		}

		// === 6. Update user in DB ===
		updatedUser, err := cfg.DB.UpdateUser(ctx, database.UpdateUserParams{
			ID:            userID,
			Name:          finalName,
			Email:         finalEmail,
			PasswordHash:  finalPasswordHash,
			MonthlyIncome: finalMonthlyIncome,
		})
		if err != nil {
			logger.Logger.Errorw("Failed to update user in database",
				"error", err,
				"user_id", userID,
			)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}

		// === 7. Build response (omit password) ===
		resp := models.UpdateUserResponse{
			ID:            updatedUser.ID,
			Name:          updatedUser.Name,
			Email:         updatedUser.Email,
			MonthlyIncome: updatedUser.MonthlyIncome,
			CreatedAt:     updatedUser.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     updatedUser.UpdatedAt.Format(time.RFC3339),
		}

		// === 8. Log success ===
		logger.Logger.Infow("User updated successfully",
			"user_id", updatedUser.ID,
			"new_email", updatedUser.Email,
		)

		utils.RespondWithJSON(w, http.StatusOK, resp)

	}
}
