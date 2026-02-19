-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_balance (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(12,2) NOT NULL DEFAULT 0 CHECK (balance >= 0)
);

-- Create row for existing users with 0 balance (optional: run only if you have existing users)
-- INSERT INTO user_balance (user_id, balance) SELECT id, 0 FROM users ON CONFLICT (user_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_balance;
-- +goose StatementEnd
