-- +goose Up
CREATE TABLE IF NOT EXISTS notifications_sent (
    id       SERIAL PRIMARY KEY,
    event_id TEXT        NOT NULL,
    sent_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS notifications_sent_event_idx ON notifications_sent (event_id);

-- +goose Down
DROP INDEX IF EXISTS notifications_sent_event_idx;
DROP TABLE IF EXISTS notifications_sent;
