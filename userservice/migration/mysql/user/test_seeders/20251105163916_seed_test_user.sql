-- +goose Up
-- +goose StatementBegin
INSERT IGNORE INTO users (
    username,
    email,
    hash_password,
    name,
    is_password_updated,
    created_at,
    updated_at
) VALUES (
    'test',
    'test@example.com',
    '$2a$10$4pphahvk3JEn47ZpJuPuQ.CQjqjucLI01aywfSZldheBv6a0nrJgy', -- admisafaWWWn123
    'test',
    1,
    UTC_TIMESTAMP(),
    UTC_TIMESTAMP()
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT IGNORE INTO user_roles (user_id, role_id) SELECT u.id, r.id FROM users u
CROSS JOIN roles r
WHERE u.username = 'test' AND r.name = 'SUI';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE username = 'test';
-- +goose StatementEnd
