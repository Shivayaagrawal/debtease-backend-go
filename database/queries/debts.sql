-- name: CreateDebt :one
INSERT INTO debts (
  user_id, name, type, lender, principal_decimal,
  outstanding_balance, interest_rate, min_payment,
  due_date, payment_frequency, debt_taken,
  risk_level, priority, notes,
  payment_due_day, emi_amount, tenure_months, months_paid, moratorium_until, in_moratorium
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
  $15, $16, $17, $18, $19, $20
)
RETURNING *;

-- name: ListUserDebts :many
SELECT *
FROM debts
WHERE user_id = $1
ORDER BY priority DESC, due_date ASC;

-- name: GetDebtByID :one
SELECT *
FROM debts
WHERE id = $1 AND user_id = $2;

-- name: UpdateDebtBalance :exec
UPDATE debts
SET outstanding_balance = $2,
    is_overdue = $3,
    updated_at = now()
WHERE id = $1;

-- name: UpdateDebtHealthScore :exec
UPDATE debts
SET debt_health_score = $2, updated_at = now()
WHERE id = $1;

-- name: MarkDebtOverdue :exec
UPDATE debts
SET is_overdue = TRUE, updated_at = now()
WHERE id = $1;

-- name: AccrueInterest :exec
UPDATE debts
SET accrued_interest = accrued_interest + $2,
    last_accrual_date = $3,
    updated_at = now()
WHERE id = $1;

-- name: ResetAccruedInterest :exec
UPDATE debts
SET accrued_interest = 0,
    updated_at = now()
WHERE id = $1;

-- name: GetDebtForUpdate :one
SELECT *
FROM debts
WHERE id = $1 AND user_id = $2
FOR UPDATE;

-- name: UpdateDebtAfterPayment :exec
UPDATE debts
SET outstanding_balance = $2,
    accrued_interest = $3,
    last_accrual_date = $4,
    updated_at = now()
WHERE id = $1;

-- name: ListAllActiveDebts :many
SELECT *
FROM debts
WHERE outstanding_balance > 0 OR accrued_interest > 0;

-- name: UpdateLoanPayment :exec
UPDATE debts
SET months_paid = months_paid + 1,
	updated_at = now()
WHERE id = $1;

-- name: ExitMoratorium :exec
UPDATE debts
SET in_moratorium = FALSE,
	emi_amount = $2,
	updated_at = now()
WHERE id = $1;

-- name: ListLoansNeedingEMI :many
SELECT *
FROM debts
WHERE type IN ('personal_loan', 'student_loan', 'mortgage')
	AND payment_due_day IS NOT NULL
	AND outstanding_balance > 0
	AND EXTRACT(DAY FROM CURRENT_DATE) = payment_due_day;



-- name: GetDebtBreakdownByType :many
-- Breakdown of debts by type for visualization
SELECT 
    type,
    COUNT(*) as count,
    COALESCE(SUM(outstanding_balance + accrued_interest), 0) as total_outstanding,
    COALESCE(SUM(min_payment), 0) as total_min_payment,
    CASE 
        WHEN SUM(outstanding_balance) > 0 
        THEN ROUND((SUM(interest_rate * outstanding_balance) / SUM(outstanding_balance))::numeric, 5)
        ELSE 0
    END as avg_interest_rate
FROM debts
WHERE user_id = $1 
    AND outstanding_balance > 0
GROUP BY type
ORDER BY total_outstanding DESC;

-- name: GetDebtsByRiskLevel :many
-- Group debts by risk level
SELECT 
    risk_level,
    COUNT(*) as count,
    COALESCE(SUM(outstanding_balance + accrued_interest), 0) as total_outstanding,
    COALESCE(AVG(debt_health_score), 0) as avg_health_score
FROM debts
WHERE user_id = $1 
    AND outstanding_balance > 0
    AND risk_level IS NOT NULL
GROUP BY risk_level
ORDER BY 
    CASE risk_level
        WHEN 'high' THEN 1
        WHEN 'medium' THEN 2
        WHEN 'low' THEN 3
    END;