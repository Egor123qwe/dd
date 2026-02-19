-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS feedback_local  (
  id BIGSERIAL PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  type VARCHAR(255) NOT NULL,
  text TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS feedback_local;
-- +goose StatementEnd
