-- +goose Up
-- +goose StatementBegin
ALTER TABLE debts 
	ADD COLUMN payment_due_day INTEGER CHECK (payment_due_day BETWEEN 1 AND 31),
	ADD COLUMN emi_amount DECIMAL(12,2),
	ADD COLUMN tenure_months INTEGER,
	ADD COLUMN months_paid INTEGER DEFAULT 0,
	ADD COLUMN moratorium_until DATE,
	ADD COLUMN in_moratorium BOOLEAN DEFAULT FALSE;

CREATE INDEX idx_debts_payment_due_day ON debts(payment_due_day) WHERE payment_due_day IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_debts_payment_due_day;

ALTER TABLE debts 
	DROP COLUMN IF EXISTS payment_due_day,
	DROP COLUMN IF EXISTS emi_amount,
	DROP COLUMN IF EXISTS tenure_months,
	DROP COLUMN IF EXISTS months_paid,
	DROP COLUMN IF EXISTS moratorium_until,
	DROP COLUMN IF EXISTS in_moratorium;
-- +goose StatementEnd

