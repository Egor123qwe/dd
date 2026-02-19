-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(255) NULL,
    is_default TINYINT(1) NOT NULL DEFAULT 0,

    permissions BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO roles (name, description, is_default, permissions, created_at, updated_at) VALUES
    ('Farmer', 'farmer role', 1, 13510801029590912, UTC_TIMESTAMP(), UTC_TIMESTAMP()),
    ('Manager', 'manager role', 1, 13581165478801406, UTC_TIMESTAMP(), UTC_TIMESTAMP()),
    ('Service', 'service role', 1, 18014398509477886, UTC_TIMESTAMP(), UTC_TIMESTAMP()),
    ('SUI', 'sui role', 1, 18014398509481983, UTC_TIMESTAMP(), UTC_TIMESTAMP()); -- All 54 permissions (bits 0-53)
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id INT NOT NULL,

    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_roles;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
