-- +goose Up
-- +goose StatementBegin
ALTER TABLE short_urls
ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES shurl_users(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE short_urls
DROP COLUMN IF EXISTS user_id;
-- +goose StatementEnd
