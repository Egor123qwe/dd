-- +goose Up
-- +goose StatementBegin
ALTER TABLE rent
    ADD COLUMN IF NOT EXISTS session_id VARCHAR(255) REFERENCES session (id) NOT NULL,
    DROP COLUMN IF EXISTS merchant_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rent
    ADD COLUMN IF NOT EXISTS merchant_id VARCHAR(255) NOT NULL,
    DROP COLUMN IF EXISTS session_id;
-- +goose StatementEnd
