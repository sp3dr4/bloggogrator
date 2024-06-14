-- +goose Up
CREATE TABLE IF NOT EXISTS feeds (
    id              UUID PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    name            VARCHAR(255) NOT NULL,
    url             TEXT NOT NULL UNIQUE,
    last_fetched_at TIMESTAMP WITH TIME ZONE,
    user_id         UUID NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS feeds;
