-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS templates_template_info (
    template_id VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    short_description TEXT NOT NULL DEFAULT '',
    version VARCHAR(255) NOT NULL DEFAULT '',
    container_image_name VARCHAR(512) NOT NULL DEFAULT '',
    container_image_tag VARCHAR(255) NOT NULL DEFAULT '',
    ports BYTEA[] NOT NULL DEFAULT '{}',
    envs BYTEA[] NOT NULL DEFAULT '{}',
    volumes BYTEA[] NOT NULL DEFAULT '{}',
    use_gpu BOOLEAN NOT NULL DEFAULT FALSE
);

-- Default templates (port JSON: Auth, Port, Type, Title)
INSERT INTO templates_template_info (
    template_id, title, type, description, short_description,
    version, container_image_name, container_image_tag,
    ports, envs, volumes, use_gpu
) VALUES
(
    'hello-world-api',
    'Hello World API',
    'proxy',
    'Simple API example (nmatsui/hello-world-api)',
    'HTTP API on port 3000',
    '1',
    'nmatsui/hello-world-api',
    'latest',
    ARRAY[convert_to('{"Auth":true,"Port":3000,"Type":"http","Title":"API"}', 'UTF8')],
    '{}', '{}', FALSE
),
(
    'hello-world-api-tcp',
    'Hello World API (TCP)',
    'proxy',
    'Same image, port exposed as TCP',
    'TCP 3000',
    '1',
    'nmatsui/hello-world-api',
    'latest',
    ARRAY[convert_to('{"Auth":false,"Port":3000,"Type":"tcp","Title":"API"}', 'UTF8')],
    '{}', '{}', FALSE
),
(
    'test',
    'Test template',
    'proxy',
    'Generic test template',
    'For start-session tests',
    '1',
    'nmatsui/hello-world-api',
    'latest',
    ARRAY[convert_to('{"Auth":true,"Port":3000,"Type":"http","Title":"API"}', 'UTF8')],
    '{}', '{}', FALSE
)
ON CONFLICT (template_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS templates_template_info;
-- +goose StatementEnd
