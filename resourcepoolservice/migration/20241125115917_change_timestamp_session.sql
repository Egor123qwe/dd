-- +goose Up
-- +goose StatementBegin
ALTER TABLE session
ALTER COLUMN created_at TYPE TIMESTAMP,
ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE session
ALTER COLUMN created_at SET DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow'::text);;
-- +goose StatementEnd
