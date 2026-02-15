-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS short_urls (
    id UUID         PRIMARY KEY
  , destination_url TEXT NOT NULL 
  , slug            TEXT NOT NULL UNIQUE-- slug is the part after the slash, https://shurl.invalid/<this part here>
  , created_at      TIMESTAMPTZ NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE short_urls;
-- +goose StatementEnd
