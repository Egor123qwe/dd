-- +goose Up
-- +goose StatementBegin
ALTER TABLE gpu ADD COLUMN IF NOT EXISTS name VARCHAR(512);
ALTER TABLE gpu ADD COLUMN IF NOT EXISTS total_vram BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE gpu DROP COLUMN IF EXISTS name;
ALTER TABLE gpu DROP COLUMN IF EXISTS total_vram;
-- +goose StatementEnd
