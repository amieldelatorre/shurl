-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS idempotency_keys (
    id              UUID PRIMARY KEY
  , i_key           UUID UNIQUE NOT NULL
  , reference_id    UUID NOT NULL
  , created_at      TIMESTAMPTZ NOT NULL
  , expires_at      TIMESTAMPTZ NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE idempotency_keys;
-- +goose StatementEnd
