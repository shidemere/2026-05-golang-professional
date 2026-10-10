-- +goose Up
CREATE TABLE events (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ NOT NULL,
    description TEXT,
    user_id UUID NOT NULL,
    notify_before_seconds BIGINT
);

-- +goose Down
DROP TABLE events;
