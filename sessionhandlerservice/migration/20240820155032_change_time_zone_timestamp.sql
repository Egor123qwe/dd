-- +goose Up
-- +goose StatementBegin
ALTER TABLE session
ALTER COLUMN created_at SET DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE session
ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
