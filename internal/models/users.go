package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=1"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type CreateUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// Pointer fields allow partial updates (nil = not provided)
type UpdateUserRequest struct {
	Name          *string          `json:"name,omitempty"`
	Email         *string          `json:"email,omitempty"`
	Password      *string          `json:"password,omitempty"`
	MonthlyIncome *decimal.Decimal `json:"monthly_income,omitempty"`
}

type UpdateUserResponse struct {
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Email         string          `json:"email"`
	MonthlyIncome decimal.Decimal `json:"monthly_income"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}
