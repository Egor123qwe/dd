-- +goose Up
-- +goose StatementBegin
ALTER TABLE template_settings
    ADD COLUMN IF NOT EXISTS volumes character varying[];

UPDATE template_settings SET volumes = '{}';

ALTER TABLE template_settings ALTER COLUMN volumes SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE template_settings
    DROP COLUMN IF EXISTS volumes;
-- +goose StatementEnd
