-- +goose Up
-- +goose StatementBegin
ALTER TABLE session
ADD COLUMN total_price DECIMAL(10, 5);

UPDATE session
SET
  total_price = 0;

ALTER TABLE session
ALTER COLUMN total_price
SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE session
ALTER COLUMN total_price
DROP NOT NULL;

UPDATE session
SET
  total_price = NULL;

ALTER TABLE session
DROP COLUMN total_price;
-- +goose StatementEnd
