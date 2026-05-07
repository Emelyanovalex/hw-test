-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id            TEXT PRIMARY KEY,
    title         TEXT        NOT NULL,
    start_time    TIMESTAMPTZ NOT NULL,
    duration      BIGINT      NOT NULL, -- nanoseconds
    description   TEXT        NOT NULL DEFAULT '',
    user_id       TEXT        NOT NULL,
    notify_before BIGINT      NOT NULL DEFAULT 0 -- nanoseconds
);

CREATE INDEX IF NOT EXISTS events_user_start_idx ON events (user_id, start_time);
CREATE INDEX IF NOT EXISTS events_start_idx      ON events (start_time);

-- +goose Down
DROP INDEX IF EXISTS events_start_idx;
DROP INDEX IF EXISTS events_user_start_idx;
DROP TABLE IF EXISTS events;
