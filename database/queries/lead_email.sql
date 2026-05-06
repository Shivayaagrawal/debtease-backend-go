-- name: CreateLeadEmail :one
INSERT INTO lead_emails (email, source, session_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLeadEmailByID :one
SELECT * FROM lead_emails WHERE id = $1;

-- name: GetLeadEmailByEmail :one
SELECT * FROM lead_emails WHERE email = $1 LIMIT 1;

-- name: ListLeadEmails :many
SELECT * FROM lead_emails
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListLeadEmailsBySource :many
SELECT * FROM lead_emails
WHERE source = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- name: CountLeadEmailsBySource :one
SELECT 
    COUNT(*) FILTER (WHERE source = 'waitlist')::int AS waitlist_count,
    COUNT(*) FILTER (WHERE source = 'pdf')::int AS pdf_count,
    COUNT(*)::int AS total_count
FROM lead_emails;

-- name: CheckEmailExists :one
SELECT EXISTS(SELECT 1 FROM lead_emails WHERE email = $1) AS exists;

