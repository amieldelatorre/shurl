-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS idempotency_keys (
    id              UUID PRIMARY KEY
  , reference_id    UUID NOT NULL
  , created_at      TIMESTAMPTZ NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE short_urls;
-- +goose StatementEnd
