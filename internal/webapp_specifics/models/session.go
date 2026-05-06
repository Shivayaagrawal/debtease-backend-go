package models

import "github.com/google/uuid"

type SessionResponse struct {
	SessionID     uuid.UUID `json:"session_id"`
	EngineVersion string    `json:"engine_version"`
}
