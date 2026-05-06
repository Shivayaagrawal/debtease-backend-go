-- name: UpsertFinanceSteps :one
INSERT INTO finance_steps (session_id, monthly_income, fixed_obligations, loans, engine_version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, now(), now())
ON CONFLICT (session_id) DO UPDATE SET
    monthly_income = EXCLUDED.monthly_income,
    fixed_obligations = EXCLUDED.fixed_obligations,
    loans = EXCLUDED.loans,
    engine_version = EXCLUDED.engine_version,
    updated_at = now()
RETURNING *;

-- name: GetFinanceStepsBySessionID :one
SELECT * FROM finance_steps WHERE session_id = $1;

-- name: DeleteFinanceStepsBySessionID :exec
DELETE FROM finance_steps WHERE session_id = $1;
