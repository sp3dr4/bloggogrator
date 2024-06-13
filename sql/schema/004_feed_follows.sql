-- +goose Up
CREATE TABLE IF NOT EXISTS feed_follows (
    id UUID     PRIMARY KEY,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    user_id     UUID NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    feed_id     UUID NOT NULL,
    FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS feed_follows;
