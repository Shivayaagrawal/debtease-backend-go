package events

import (
	"time"
	"github.com/google/uuid"
)

const (
	EventTypeUserRegistered = "user.registered"
)

type BaseEvent struct {
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    uuid.UUID `json:"user_id"`
}
func (b BaseEvent) GetEventType() string {
	return b.EventType
}

type UserRegisteredEvent struct {
	BaseEvent
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}


// NewUserRegisteredEvent creates a new user registration event
func NewUserRegisteredEvent(userID uuid.UUID, email, name string, createdAt time.Time) *UserRegisteredEvent {
	return &UserRegisteredEvent{
		BaseEvent: BaseEvent{
			EventType: EventTypeUserRegistered,
			Timestamp: time.Now(),
			UserID:    userID,
		},
		Email:     email,
		Name:      name,
		CreatedAt: createdAt,
	}
}
