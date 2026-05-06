package models

import (
	"github.com/google/uuid"
)
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Token 	  string    `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
