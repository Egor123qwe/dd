# Makefile для запуска облака в Docker Compose
# Использование: make [цель] или make help

COMPOSE = docker compose
COMPOSE_FILE = docker-compose.yml

.PHONY: help up down build build-% infra logs logs-% ps restart clean topics migrate

# Цель по умолчанию
.DEFAULT_GOAL := help

help:
	@echo "Cloud Docker Compose — доступные цели:"
	@echo ""
	@echo "  make up          — поднять все сервисы в фоне"
	@echo "  make down        — остановить и удалить контейнеры"
	@echo "  make build       — собрать образы всех сервисов"
	@echo "  make build-SVC   — собрать образ одного сервиса (например make build-connectioncoordinatorservice)"
	@echo "  make infra       — поднять только инфраструктуру (postgres, redis, minio, redpanda + топики)"
	@echo "  make logs        — показать логи всех сервисов (Ctrl+C для выхода)"
	@echo "  make logs-SVC    — логи одного сервиса (например make logs-redpanda)"
	@echo "  make ps          — список контейнеров проекта"
	@echo "  make restart     — down + up"
	@echo "  make clean       — down и удалить тома (volumes)"
	@echo "  make topics      — создать топики Kafka (запустить redpanda-init)"
	@echo "  make migrate     — прогнать миграции БД (postgres должен быть запущен)"
	@echo ""

up:
	$(COMPOSE) -f $(COMPOSE_FILE) up -d

down:
	$(COMPOSE) -f $(COMPOSE_FILE) down

build:
	$(COMPOSE) -f $(COMPOSE_FILE) build

build-%:
	$(COMPOSE) -f $(COMPOSE_FILE) build $*

infra:
	$(COMPOSE) -f $(COMPOSE_FILE) up -d postgres redis minio redpanda
	@echo "Ожидание готовности Red Panda..."
	@sleep 5
	$(COMPOSE) -f $(COMPOSE_FILE) up redpanda-init
	@echo "Инфраструктура запущена (postgres, redis, minio, redpanda + топики)."

logs:
	$(COMPOSE) -f $(COMPOSE_FILE) logs -f

logs-%:
	$(COMPOSE) -f $(COMPOSE_FILE) logs -f $*

ps:
	$(COMPOSE) -f $(COMPOSE_FILE) ps

restart: down up

clean: down
	$(COMPOSE) -f $(COMPOSE_FILE) down -v
	@echo "Контейнеры и тома удалены."

topics:
	$(COMPOSE) -f $(COMPOSE_FILE) up redpanda-init

migrate:
	$(COMPOSE) -f $(COMPOSE_FILE) up migrate-sessionhandler migrate-resourcepool migrate-feedback migrate-history
	@echo "Миграции выполнены (порядок: sessionhandler → resourcepool → feedback → history)."
