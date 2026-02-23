-- +goose Up
-- +goose StatementBegin
ALTER TABLE template_settings
    ALTER COLUMN short_description TYPE TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE template_settings
    ALTER COLUMN short_description TYPE VARCHAR(255);
-- +goose StatementEnd
