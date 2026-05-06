-- +goose Up
-- +goose StatementBegin

CREATE TABLE finance_steps (
    session_id UUID PRIMARY KEY,
    monthly_income NUMERIC(18,2) NULL,
    fixed_obligations NUMERIC(18,2) NULL,
    loans JSONB NOT NULL DEFAULT '[]'::jsonb,
    engine_version TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_finance_steps_session
        FOREIGN KEY (session_id)
        REFERENCES calculation_sessions (session_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_finance_steps_session_id ON finance_steps (session_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS finance_steps;
-- +goose StatementEnd
