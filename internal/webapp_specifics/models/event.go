package models

import (
	"time"

	"github.com/google/uuid"
)

type EventRequest struct {
	EventType string                 `json:"event_type" validate:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type EventResponse struct {
	ID        uuid.UUID              `json:"id"`
	SessionID uuid.UUID              `json:"session_id"`
	EventType string                 `json:"event_type"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}
