-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS
  feedback (
    id BIGSERIAL PRIMARY KEY,
    score INT NOT NULL,
    text TEXT,
    rent_id VARCHAR(255) NOT NULL
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table feedback;
-- +goose StatementEnd
