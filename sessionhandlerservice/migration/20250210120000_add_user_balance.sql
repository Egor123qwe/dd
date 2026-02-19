-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_balance (
    user_id BIGINT PRIMARY KEY,
    balance DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (balance >= 0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_balance;
-- +goose StatementEnd
