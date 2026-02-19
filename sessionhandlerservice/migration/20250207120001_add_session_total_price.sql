-- +goose Up
-- +goose StatementBegin
ALTER TABLE session
    ADD COLUMN IF NOT EXISTS total_price DECIMAL(10, 5) DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE session
    DROP COLUMN IF EXISTS total_price;
-- +goose StatementEnd
