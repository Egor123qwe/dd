-- +goose Up
-- +goose StatementBegin
ALTER TABLE session
ADD COLUMN status VARCHAR(32);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE session
DROP COLUMN status;
-- +goose StatementEnd
