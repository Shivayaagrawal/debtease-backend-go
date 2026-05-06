-- name: CreateNotificationEvent :one
INSERT INTO notification_events (user_id, event_type, channel, recipient, status, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, event_type, channel, recipient, status, error_message, metadata, sent_at, created_at, updated_at;


-- name: UpdateNotificationEventStatus :one
UPDATE notification_events
SET status = $2,
    error_message = $3,
    sent_at = CASE WHEN $2::varchar = 'sent' THEN now() ELSE sent_at END,
    updated_at = now()
WHERE id = $1
RETURNING id, user_id, event_type, channel, recipient, status, error_message, metadata, sent_at, created_at, updated_at;

-- name: GetNotificationEventByID :one
SELECT id, user_id, event_type, channel, recipient, status, error_message, metadata, sent_at, created_at, updated_at
FROM notification_events
WHERE id = $1;

-- name: ListNotificationEventsByUser :many
SELECT id, user_id, event_type, channel, recipient, status, error_message, metadata, sent_at, created_at, updated_at
FROM notification_events
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;