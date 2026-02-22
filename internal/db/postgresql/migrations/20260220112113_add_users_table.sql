-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS shurl_users (
    id              UUID PRIMARY KEY
  , username        TEXT NOT NULL UNIQUE
  , email           TEXT NOT NULL UNIQUE
  , password_hash   TEXT NOT NULL
  , created_at      TIMESTAMPTZ NOT NULL
  , updated_at      TIMESTAMPTZ NOT NULL
);

ALTER TABLE short_urls
ADD COLUMN 
IF NOT EXISTS user_id UUID 
REFERENCES shurl_users(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE short_urls
DROP COLUMN 
IF EXISTS user_id;

DROP TABLE shurl_users;
-- +goose StatementEnd
