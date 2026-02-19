-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS
  session (
    id VARCHAR(255) PRIMARY KEY, --session_id
    total_ram BIGINT NOT NULL,
    available_ram BIGINT NOT NULL,
    used_ram BIGINT NOT NULL,
    price_ram FLOAT,
    load_speed FLOAT NOT NULL,
    upload_speed FLOAT NOT NULL,
    ping INT NOT NULL,
    price_internet FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
  );

CREATE TABLE IF NOT EXISTS
  cpu (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) REFERENCES session (id) NOT NULL,
    name VARCHAR(255),
    total BIGINT NOT NULL,
    available BIGINT NOT NULL,
    price FLOAT
  );

CREATE TABLE IF NOT EXISTS
  storage (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) REFERENCES session (id) NOT NULL,
    total BIGINT NOT NULL,
    available BIGINT NOT NULL,
    used BIGINT NOT NULL,
    bandwidth FLOAT NOT NULL,
    price FLOAT
  );

CREATE TABLE IF NOT EXISTS
  prepull (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    template_id BIGINT NOT NULL -- in future, we need to add foreign key
  );

CREATE TABLE IF NOT EXISTS
  gpu_dict (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    total_vram BIGINT NOT NULL,
    price FLOAT,
    avg_dlperf FLOAT
  );

CREATE TABLE IF NOT EXISTS
  gpu (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) REFERENCES session (id) NOT NULL,
    gpu_dict_id BIGINT REFERENCES gpu_dict (id) NOT NULL, -- if not found, push existed gpu
    available_vram BIGINT NOT NULL,
    used_vram BIGINT NOT NULL,
    dlperf FLOAT NOT NULL,
    price FLOAT
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE gpu;

DROP TABLE storage;

DROP TABLE cpu;

DROP TABLE prepull;

DROP TABLE session;

DROP TABLE gpu_dict;
-- +goose StatementEnd
