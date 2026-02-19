-- +goose Up
-- +goose StatementBegin
ALTER TABLE rent
    ALTER COLUMN deleted_at SET DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow');

ALTER TABLE session
    ALTER COLUMN deleted_at SET DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rent
    ALTER COLUMN deleted_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE session
    ALTER COLUMN deleted_at SET DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
