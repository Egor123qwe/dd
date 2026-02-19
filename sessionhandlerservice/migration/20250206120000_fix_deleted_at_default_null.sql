-- +goose Up
-- +goose StatementBegin
-- deleted_at must be NULL for active rows; default was wrongly set to CURRENT_TIMESTAMP in 20241115133122
ALTER TABLE rent
    ALTER COLUMN deleted_at SET DEFAULT NULL;

ALTER TABLE session
    ALTER COLUMN deleted_at SET DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rent
    ALTER COLUMN deleted_at SET DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow');

ALTER TABLE session
    ALTER COLUMN deleted_at SET DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow');
-- +goose StatementEnd
