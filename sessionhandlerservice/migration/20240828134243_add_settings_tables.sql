-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS
    template_settings (
        id BIGSERIAL PRIMARY KEY,

        -- template:
        template_id VARCHAR(255) NOT NULL,
        image_name VARCHAR(255) NOT NULL,
        image_tag VARCHAR(255) NOT NULL,
        version VARCHAR(255) NOT NULL,

        -- model authentication:
        login VARCHAR(255) NOT NULL,
        password VARCHAR(255) NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    piko_settings (
        id BIGSERIAL PRIMARY KEY,

        auth_key VARCHAR(255) NOT NULL,
        endpoint VARCHAR(255) NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    tailscale_settings (
        id BIGSERIAL PRIMARY KEY,

        merchant_key VARCHAR(255) NOT NULL,
        client_key VARCHAR(255) NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    network_settings (
        id BIGSERIAL PRIMARY KEY,

        -- references to piko_settings or tailscale_settings:
        -- reference depends on mode in rent_settings
        network_way_id BIGINT
    );

CREATE TABLE IF NOT EXISTS
    rent_settings (
        id BIGSERIAL PRIMARY KEY,

        -- general settings:
        mode VARCHAR(255) NOT NULL,

        -- template:
        template_id BIGINT REFERENCES template_settings (id),

        -- network:
        network_id BIGINT REFERENCES network_settings (id)
    );

ALTER TABLE rent
    ADD COLUMN IF NOT EXISTS settings_id BIGINT REFERENCES rent_settings (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rent
    DROP COLUMN IF EXISTS settings_id;

DROP TABLE rent_settings;
DROP TABLE network_settings;
DROP TABLE template_settings;
DROP TABLE piko_settings;
DROP TABLE tailscale_settings;
-- +goose StatementEnd
