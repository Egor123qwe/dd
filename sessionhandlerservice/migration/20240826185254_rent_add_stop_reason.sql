-- +goose Up
-- +goose StatementBegin
ALTER TABLE rent
    ADD COLUMN IF NOT EXISTS stop_reason VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rent
    DROP COLUMN IF EXISTS stop_reason;
-- +goose StatementEnd
