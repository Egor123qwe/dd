-- +goose Up
-- +goose StatementBegin
INSERT INTO users (
    username,
    email,
    hash_password,
    name,
    created_at,
    updated_at
) VALUES (
    'admin',
    'admin@example.com',
    '$2a$10$WgdPwzfAuVolxTPdjxJ33uNnHwGs6Wmok8PkjX/OS5oW/24o.A..O', -- AdminPassword@007
    'ServiceUser',
    UTC_TIMESTAMP(),
    UTC_TIMESTAMP()
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO user_roles (user_id, role_id) SELECT u.id, r.id FROM users u
CROSS JOIN roles r
WHERE u.username = 'admin' AND r.name = 'Service';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE username = 'admin';
-- +goose StatementEnd
