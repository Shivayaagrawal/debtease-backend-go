-- +goose Up
-- +goose StatementBegin
CREATE TYPE payment_type AS ENUM ('minimum', 'extra', 'full', 'late');

CREATE TABLE payments_history (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    debt_id           UUID NOT NULL REFERENCES debts(id) ON DELETE CASCADE,
    amount            DECIMAL(12,2) NOT NULL,
    payment_date      DATE NOT NULL,
    type              payment_type NOT NULL,
    notes             TEXT,
    principal_applied DECIMAL(12,2) NOT NULL DEFAULT 0,
    interest_applied  DECIMAL(12,2) NOT NULL DEFAULT 0,
    payment_method    VARCHAR(100),
    created_at        TIMESTAMPTZ DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payments_history;
DROP TYPE IF EXISTS payment_type;
-- +goose StatementEnd
