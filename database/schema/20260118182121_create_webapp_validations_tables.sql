-- +goose Up
-- +goose StatementBegin

-- Calculations Session

CREATE TABLE calculation_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL UNIQUE,
    engine_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_calculation_sessions_created_at
    ON calculation_sessions (created_at);

-- Calculations Events 
CREATE TABLE calculation_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_calculation_session
        FOREIGN KEY (session_id)
        REFERENCES calculation_sessions (session_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_calculation_events_session_id
    ON calculation_events (session_id);

CREATE INDEX idx_calculation_events_event_type
    ON calculation_events (event_type);

CREATE INDEX idx_calculation_events_created_at
    ON calculation_events (created_at);

-- Calculations Session Metrics
CREATE TABLE calculation_session_metrics (
    session_id UUID PRIMARY KEY,
    engine_version TEXT NOT NULL,

    -- core behavioral signals
    calculation_runs INT NOT NULL DEFAULT 0,
    strategy_switches INT NOT NULL DEFAULT 0,
    input_modifications INT NOT NULL DEFAULT 0,

    -- commitment signals
    pdf_requested BOOLEAN NOT NULL DEFAULT FALSE,
    email_submitted BOOLEAN NOT NULL DEFAULT FALSE,

    -- timing
    first_seen_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_calculation_session_metrics_session_id
    ON calculation_session_metrics (session_id);

-- Lead Emails
CREATE TABLE lead_emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('waitlist', 'pdf')),
    session_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_lead_session
        FOREIGN KEY (session_id)
        REFERENCES calculation_sessions (session_id)
        ON DELETE SET NULL
);

CREATE INDEX idx_lead_emails_email
    ON lead_emails (email);

CREATE INDEX idx_lead_emails_source
    ON lead_emails (source);

CREATE INDEX idx_lead_emails_created_at
    ON lead_emails (created_at);



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS calculation_sessions;
DROP TABLE IF EXISTS calculation_events;
DROP TABLE IF EXISTS calculation_session_metrics;
DROP TABLE IF EXISTS lead_emails;
-- +goose StatementEnd
