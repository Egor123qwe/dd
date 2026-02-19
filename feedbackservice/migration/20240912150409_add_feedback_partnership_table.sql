-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS feedback_partnership (
  id BIGSERIAL PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  contact_name VARCHAR(255) NOT NULL,
  company_name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  phone_num VARCHAR(32) NOT NULL,
  comment TEXT
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS feedback_partnership;
-- +goose StatementEnd
