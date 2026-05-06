-- name: RecordPayment :one
INSERT INTO payments_history (
  debt_id, amount, payment_date, type,
  principal_applied, interest_applied,
  notes, payment_method
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListPaymentsByDebt :many
SELECT *
FROM payments_history
WHERE debt_id = $1
ORDER BY payment_date DESC
LIMIT $2 OFFSET $3;

-- name: GetPaymentByID :one
SELECT *
FROM payments_history
WHERE id = $1 AND debt_id = $2;

-- name: GetRecentPayments :many
SELECT p.*, d.name AS debt_name
FROM payments_history p
JOIN debts d ON p.debt_id = d.id
WHERE p.payment_date >= $1
  AND d.user_id = $2
ORDER BY p.payment_date DESC;


-- name: GetUpcomingEMIs :many
-- Returns upcoming EMI/payments within the next 30 days
SELECT 
    id,
    name,
    type,
    lender,
    outstanding_balance,
    accrued_interest,
    COALESCE(emi_amount, min_payment) as payment_amount,
    due_date,
    payment_due_day,
    is_overdue,
    late_fees,
    CASE 
        WHEN is_overdue THEN 0
        WHEN due_date <= CURRENT_DATE THEN 0
        ELSE (due_date - CURRENT_DATE)
    END as days_until_due
FROM debts
WHERE user_id = $1 
    AND outstanding_balance > 0
    AND (
        (due_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days')
        OR is_overdue = TRUE
    )
ORDER BY 
    is_overdue DESC,
    due_date ASC;


-- name: GetPaymentHistorySummary :one
-- Payment statistics for the last 30, 90, 365 days
SELECT 
    COALESCE(SUM(CASE WHEN p.payment_date >= CURRENT_DATE - INTERVAL '30 days' THEN p.amount ELSE 0 END), 0) as paid_last_30_days,
    COALESCE(SUM(CASE WHEN p.payment_date >= CURRENT_DATE - INTERVAL '90 days' THEN p.amount ELSE 0 END), 0) as paid_last_90_days,
    COALESCE(SUM(CASE WHEN p.payment_date >= CURRENT_DATE - INTERVAL '365 days' THEN p.amount ELSE 0 END), 0) as paid_last_year,
    COUNT(CASE WHEN p.payment_date >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as payments_last_30_days
FROM payments_history p
JOIN debts d ON p.debt_id = d.id
WHERE d.user_id = $1;
