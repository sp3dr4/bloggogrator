-- +goose Up
CREATE TABLE IF NOT EXISTS posts (
    id              UUID PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    url             TEXT NOT NULL UNIQUE,
    title           VARCHAR(255),
    description     TEXT,
    published_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    feed_id         UUID NOT NULL,
    FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS posts;
