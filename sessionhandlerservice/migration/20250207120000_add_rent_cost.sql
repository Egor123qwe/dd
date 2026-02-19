-- +goose Up
-- +goose StatementBegin
ALTER TABLE rent
    ADD COLUMN IF NOT EXISTS cost DECIMAL(10, 2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rent
    DROP COLUMN IF EXISTS cost;
-- +goose StatementEnd
