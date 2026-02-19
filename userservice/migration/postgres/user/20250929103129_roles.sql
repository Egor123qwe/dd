-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(255) NULL,
    is_default SMALLINT NOT NULL DEFAULT 0,

    permissions BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO roles (name, description, is_default, permissions, created_at, updated_at) VALUES
    ('Farmer', 'farmer role', 1, 13510801029590912, (NOW() AT TIME ZONE 'UTC'), (NOW() AT TIME ZONE 'UTC')),
    ('Manager', 'manager role', 1, 13581165478801406, (NOW() AT TIME ZONE 'UTC'), (NOW() AT TIME ZONE 'UTC')),
    ('Service', 'service role', 1, 18014398509477886, (NOW() AT TIME ZONE 'UTC'), (NOW() AT TIME ZONE 'UTC')),
    ('SUI', 'sui role', 1, 18014398509481983, (NOW() AT TIME ZONE 'UTC'), (NOW() AT TIME ZONE 'UTC'));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT NOT NULL,
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
