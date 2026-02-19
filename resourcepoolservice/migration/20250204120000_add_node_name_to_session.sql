-- +goose Up
-- +goose StatementBegin
ALTER TABLE session ADD COLUMN IF NOT EXISTS node_name TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE session DROP COLUMN IF EXISTS node_name;
-- +goose StatementEnd
