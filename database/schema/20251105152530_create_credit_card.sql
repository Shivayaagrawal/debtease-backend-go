-- +goose Up
-- +goose StatementBegin
CREATE TABLE credit_cards (
	debt_id UUID PRIMARY KEY REFERENCES debts(id) ON DELETE CASCADE,
	billing_day INTEGER NOT NULL CHECK (billing_day BETWEEN 1 AND 31),
	grace_period_days INTEGER NOT NULL DEFAULT 18,
	min_payment_percent DECIMAL(5,2) NOT NULL DEFAULT 5.00,
	last_statement_date DATE,
	payment_due_date DATE,
	billed_balance DECIMAL(12,2) NOT NULL DEFAULT 0,
	unbilled_balance DECIMAL(12,2) NOT NULL DEFAULT 0,
	in_grace_period BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ DEFAULT now(),
	updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_credit_cards_payment_due ON credit_cards(payment_due_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS credit_cards;
-- +goose StatementEnd
