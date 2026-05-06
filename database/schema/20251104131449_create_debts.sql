-- +goose Up
-- +goose StatementBegin
CREATE TYPE debt_type AS ENUM ('credit_card', 'personal_loan', 'student_loan', 'mortgage', 'medical', 'other');
CREATE TYPE payment_freq AS ENUM ('monthly', 'biweekly', 'weekly');
CREATE TYPE risk_level AS ENUM ('low', 'medium', 'high');

CREATE TABLE debts (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                 VARCHAR(255) NOT NULL,
    type                 debt_type NOT NULL,
    lender               VARCHAR(255),
    principal_decimal    DECIMAL(12,2) NOT NULL,
    outstanding_balance  DECIMAL(12,2) NOT NULL CHECK (outstanding_balance >= 0),
    interest_rate        DECIMAL(6,5) NOT NULL,
    min_payment          DECIMAL(12,2) NOT NULL,
    due_date             DATE NOT NULL,
    payment_frequency    payment_freq DEFAULT 'monthly',
    is_overdue           BOOLEAN DEFAULT FALSE,
    late_fees            DECIMAL(12,2) DEFAULT 0,
    other_charges        TEXT,
    debt_taken           TIMESTAMPTZ,
    risk_level           risk_level,
    debt_health_score    DECIMAL(5,2) CHECK (debt_health_score BETWEEN 0 AND 100),
    priority             SMALLINT DEFAULT 0,
    notes                TEXT,
    created_at           TIMESTAMPTZ DEFAULT now(),
    updated_at           TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_debts_user ON debts(user_id);
CREATE INDEX idx_debts_due   ON debts(due_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS debts;
DROP TYPE IF EXISTS debt_type;
DROP TYPE IF EXISTS payment_freq;
DROP TYPE IF EXISTS risk_level;
-- +goose StatementEnd
