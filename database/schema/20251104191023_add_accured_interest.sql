-- +goose Up
-- +goose StatementBegin
ALTER TABLE debts 
  ADD COLUMN accrued_interest DECIMAL(12,2) DEFAULT 0 NOT NULL,
  ADD COLUMN last_accrual_date DATE DEFAULT '2025-11-04' NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE debts 
  DROP COLUMN accrued_interest,
  DROP COLUMN last_accrual_date;
-- +goose StatementEnd
