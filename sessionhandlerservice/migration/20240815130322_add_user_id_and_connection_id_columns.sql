-- +goose Up
-- +goose StatementBegin
ALTER TABLE session
  ADD COLUMN IF NOT EXISTS user_id VARCHAR(255) NOT NULL,
  ADD COLUMN IF NOT EXISTS connection_id VARCHAR(255) NOT NULL,
  ADD COLUMN IF NOT EXISTS deletion_reason VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE session
  DROP COLUMN IF EXISTS user_id,
  DROP COLUMN IF EXISTS connection_id,
  DROP COLUMN IF EXISTS deletion_reason;
-- +goose StatementEnd
