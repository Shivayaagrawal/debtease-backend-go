-- name: CreateCreditCard :one
INSERT INTO credit_cards (
	debt_id, billing_day, grace_period_days, min_payment_percent,
	billed_balance, unbilled_balance, in_grace_period
) VALUES (
	$1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetCreditCardByDebtID :one
SELECT * FROM credit_cards
WHERE debt_id = $1;

-- name: UpdateCreditCardBalances :exec
UPDATE credit_cards
SET billed_balance = $2,
	unbilled_balance = $3,
	in_grace_period = $4,
	updated_at = now()
WHERE debt_id = $1;

-- name: GenerateStatement :exec
UPDATE credit_cards
SET billed_balance = unbilled_balance,
	unbilled_balance = 0,
	last_statement_date = $2,
	payment_due_date = $3,
	in_grace_period = TRUE,
	updated_at = now()
WHERE debt_id = $1;

-- name: ApplyCreditCardPayment :exec
UPDATE credit_cards
SET billed_balance = GREATEST(0, billed_balance - $2),
	in_grace_period = CASE 
		WHEN (billed_balance - $2) <= 10 THEN TRUE 
		ELSE FALSE 
	END,
	updated_at = now()
WHERE debt_id = $1;

-- name: AddDailyInterestToUnbilled :exec
UPDATE credit_cards
SET unbilled_balance = unbilled_balance + $2,
	updated_at = now()
WHERE debt_id = $1;

-- name: ListCreditCardsNeedingStatement :many
SELECT cc.*, d.user_id, d.name, d.outstanding_balance
FROM credit_cards cc
JOIN debts d ON cc.debt_id = d.id
WHERE EXTRACT(DAY FROM CURRENT_DATE) = cc.billing_day
	AND (cc.last_statement_date IS NULL OR cc.last_statement_date < CURRENT_DATE);

-- name: ListCreditCardsOutOfGrace :many
SELECT cc.*, d.user_id, d.name, d.outstanding_balance, d.interest_rate, d.last_accrual_date
FROM credit_cards cc
JOIN debts d ON cc.debt_id = d.id
WHERE cc.in_grace_period = FALSE
	AND d.outstanding_balance > 0;

