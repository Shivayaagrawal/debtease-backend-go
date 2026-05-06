package api


import (
	"DebtEase/internal/database"
	"sync/atomic"
)

type Config struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	JWTSecret	   string
	RabbitMQURL    string
}