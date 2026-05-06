-- +goose Up
-- +goose StatementBegin
CREATE TABLE notification_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    channel       VARCHAR(50) NOT NULL DEFAULT 'email',
    recipient     VARCHAR(255) NOT NULL,
    status        VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    metadata      JSONB,
    sent_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at    TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX idx_notification_events_user_id ON notification_events(user_id);
CREATE INDEX idx_notification_events_status ON notification_events(status);
CREATE INDEX idx_notification_events_event_type ON notification_events(event_type);
CREATE INDEX idx_notification_events_channel ON notification_events(channel);
CREATE INDEX idx_notification_events_created_at ON notification_events(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_notification_events_created_at;
DROP INDEX IF EXISTS idx_notification_events_channel;
DROP INDEX IF EXISTS idx_notification_events_event_type;
DROP INDEX IF EXISTS idx_notification_events_status;
DROP INDEX IF EXISTS idx_notification_events_user_id;
DROP TABLE IF EXISTS notification_events;
-- +goose StatementEnd
