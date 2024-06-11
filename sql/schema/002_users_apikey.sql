-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN api_key VARCHAR(64) NOT NULL UNIQUE DEFAULT encode(sha256(random()::text::bytea), 'hex');
ALTER TABLE users
    ALTER COLUMN api_key DROP DEFAULT;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE users
    DROP COLUMN api_key;
