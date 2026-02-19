-- +goose Up
-- +goose StatementBegin
ALTER TABLE piko_settings
    DROP COLUMN IF EXISTS endpoint,
    ADD COLUMN IF NOT EXISTS endpoints character varying[];

UPDATE piko_settings SET endpoints = '{}';

ALTER TABLE piko_settings ALTER COLUMN endpoints SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE piko_settings
    DROP COLUMN IF EXISTS endpoints,
    ADD COLUMN IF NOT EXISTS endpoint VARCHAR(255);

UPDATE piko_settings SET endpoint = 'undefined';

ALTER TABLE piko_settings ALTER COLUMN endpoint SET NOT NULL;
-- +goose StatementEnd
