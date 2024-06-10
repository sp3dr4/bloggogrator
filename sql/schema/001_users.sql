-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    name        VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS users;
