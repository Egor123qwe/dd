package config

import (
	_ "embed"
)

const (
	CheatModeKey = "cheat_mode"

	RdPortKey          = "runtimedaemon.port"
	SmConnectionUrlKey = "state_machine.connection_url"

	AuthTokenKey       = "auth.token"
	AuthServiceURLKey  = "auth.service_url"

	LogLevelKey = "logger.level"

	HttpPortKey = "http.port"

	// StatusCheckURLKey — базовый URL сервиса для HTTP-проверки активности сессии (например healthcheckservice). Если пусто — проверка по HTTP не выполняется.
	StatusCheckURLKey = "status_check.url"
)

// EnvBackendHost — переменная окружения и .env: IP или хост бэкенда. Если задана, подставляется вместо localhost в auth.service_url, status_check.url и state_machine.connection_url.
const EnvBackendHost = "BACKEND_HOST"

//go:embed config.yaml
var Data []byte
