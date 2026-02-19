-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS
    rent (
        id VARCHAR(255) PRIMARY KEY, --request_id
        status VARCHAR(255) NOT NULL,
        client_id VARCHAR(255) NOT NULL,
        merchant_id VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'Europe/Moscow'),
        deleted_at TIMESTAMP
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rent;
-- +goose StatementEnd
