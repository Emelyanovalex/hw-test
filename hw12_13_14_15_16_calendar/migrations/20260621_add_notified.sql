-- +goose Up
ALTER TABLE events ADD COLUMN IF NOT EXISTS notified BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS events_notify_idx ON events (notified, notify_before, start_time);

-- +goose Down
DROP INDEX IF EXISTS events_notify_idx;
ALTER TABLE events DROP COLUMN IF EXISTS notified;
