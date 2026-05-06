-- name: CreateCalculationSession :one
INSERT INTO calculation_sessions (session_id, engine_version)
VALUES ($1, $2)
RETURNING *;

-- name: GetCalculationSessionByID :one
SELECT * FROM calculation_sessions WHERE session_id = $1;

-- name: SessionExists :one
SELECT EXISTS(SELECT 1 FROM calculation_sessions WHERE session_id = $1) AS exists;

