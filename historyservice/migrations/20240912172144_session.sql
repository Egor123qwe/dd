-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS
  session (
    id VARCHAR(255) PRIMARY KEY, --session_id
    total_ram BIGINT NOT NULL,
    available_ram BIGINT NOT NULL,
    used_ram BIGINT NOT NULL,
    price_ram FLOAT NOT NULL,
    load_speed FLOAT NOT NULL,
    upload_speed FLOAT NOT NULL,
    ping INT NOT NULL,
    price_internet FLOAT NOT NULL,
    price_total FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table session;
-- +goose StatementEnd
