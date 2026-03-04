-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_expires_at 
ON idempotency_keys (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_idempotency_keys_expires_at;
-- +goose StatementEnd
