-- +goose Up
-- +goose StatementBegin
ALTER TABLE short_urls
ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ NOT NULL;
CREATE INDEX IF NOT EXISTS idx_short_urls_expires_at ON short_urls (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE short_urls
DROP COLUMN IF EXISTS expires_at;
DROP INDEX idx_short_urls_expires_at;
-- +goose StatementEnd
