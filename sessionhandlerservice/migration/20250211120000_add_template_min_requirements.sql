-- +goose Up
-- +goose StatementBegin
ALTER TABLE templates_template_info
  ADD COLUMN IF NOT EXISTS min_cpu INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS min_ram_bytes BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS min_storage_bytes BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS min_volume_storage_bytes BIGINT[] NOT NULL DEFAULT '{}';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE templates_template_info
  DROP COLUMN IF EXISTS min_cpu,
  DROP COLUMN IF EXISTS min_ram_bytes,
  DROP COLUMN IF EXISTS min_storage_bytes,
  DROP COLUMN IF EXISTS min_volume_storage_bytes;
-- +goose StatementEnd
