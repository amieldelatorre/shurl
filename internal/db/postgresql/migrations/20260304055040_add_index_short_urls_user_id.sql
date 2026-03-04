-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_short_urls_user_id
ON short_urls(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_short_urls_user_id;
-- +goose StatementEnd
