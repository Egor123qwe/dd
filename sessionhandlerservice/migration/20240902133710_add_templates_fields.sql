-- +goose Up
-- +goose StatementBegin
ALTER TABLE template_settings
    ADD COLUMN IF NOT EXISTS title VARCHAR(100) NOT NULL,
    ADD COLUMN IF NOT EXISTS description text NOT NULL,
    ADD COLUMN IF NOT EXISTS short_description VARCHAR(255),
    ADD COLUMN IF NOT EXISTS ports character varying[] NOT NULL,
    ADD COLUMN IF NOT EXISTS envs character varying[] NOT NULL,
    ADD COLUMN IF NOT EXISTS use_gpu boolean NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE template_settings
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS short_description,
    DROP COLUMN IF EXISTS ports,
    DROP COLUMN IF EXISTS envs,
    DROP COLUMN IF EXISTS use_gpu;
-- +goose StatementEnd
