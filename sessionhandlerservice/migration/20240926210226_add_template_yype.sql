-- +goose Up
-- +goose StatementBegin
ALTER TABLE template_settings
    ADD COLUMN IF NOT EXISTS type VARCHAR(255);

UPDATE template_settings SET type = '';

ALTER TABLE template_settings ALTER COLUMN type SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE template_settings
    DROP COLUMN IF EXISTS type;
-- +goose StatementEnd
