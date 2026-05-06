-- name: CreateCalculationEvent :one
INSERT INTO calculation_events (session_id, event_type, metadata)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListEventsBySession :many
SELECT * FROM calculation_events
WHERE session_id = $1
ORDER BY created_at ASC;

