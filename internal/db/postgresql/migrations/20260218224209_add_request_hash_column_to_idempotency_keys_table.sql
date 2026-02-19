-- +goose Up
-- +goose StatementBegin
ALTER TABLE idempotency_keys
ADD COLUMN IF NOT EXISTS request_hash text NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE idempotency_keys
DROP COLUMN IF EXISTS request_hash;
-- +goose StatementEnd
