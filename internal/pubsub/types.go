// internal/pubsub/types.go
package pubsub

type QueueType string

const (
	QueueTypeDurable   QueueType = "durable"
	QueueTypeTransient QueueType = "transient"
)