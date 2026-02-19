# Запуск всех сервисов в Docker Compose

В корне `cloud/` лежит общий `docker-compose.yml` и **Makefile** для удобного запуска.

**Быстрый старт:** `make up` — поднять всё; `make help` — список всех команд.

- **Red Panda** (Kafka-совместимый брокер) с топиками: `input`, `output`, `ttl-notification`, `keepalive`; **Redpanda Console** — веб-UI на http://localhost:18080
- **PostgreSQL**, **Redis**, **MinIO**
- Все микросервисы: connectioncoordinatorservice, sessionhandlerservice, resourcepoolservice, expirationnotificationservice, healthcheckservice, feedbackservice, userservice, historyservice, machinesaggregateservice, s3storageprocessor

## Запуск

```bash
cd /path/to/cloud
make up
# или
docker compose up -d
```

**Через Makefile:**
- `make up` — поднять все сервисы в фоне
- `make down` — остановить контейнеры
- `make build` — собрать образы
- `make infra` — только инфраструктура (postgres, redis, minio, redpanda + топики)
- `make logs` / `make logs-<сервис>` — логи
- `make ps` — список контейнеров
- `make clean` — остановить и удалить тома
- `make migrate` — прогнать миграции БД (нужен запущенный postgres)
- `make help` — полный список команд

**Миграции** выполняются автоматически при `make up` в порядке: sessionhandlerservice (создаёт таблицу `session` и др.) → resourcepoolservice (изменяет `session`) → feedbackservice → historyservice. Сервисы, использующие БД, стартуют только после завершения всех миграций.

Только инфраструктура и Kafka (без сервисов):

```bash
make infra
# или
docker compose up -d postgres redis minio redpanda
docker compose up redpanda-init
```

## Порты

| Сервис                    | Порт  |
|---------------------------|-------|
| Kafka (Red Panda)         | 19092 |
| Redpanda Console (веб-UI) | 18080 |
| Postgres                  | 5432  |
| Redis                     | 6379  |
| MinIO API / Console       | 9000, 9001 |
| Connection Coordinator    | 8090  |
| Session Handler           | 8094  |
| Resource Pool             | 8091  |
| Expiration Notification   | 8092  |
| Healthcheck               | 19000 (внутри контейнера 9000) |
| Feedback                  | 8095  |
| User                      | 8052  |
| History                   | 8899  |
| Machines Aggregate        | 8000  |
| S3 Storage Processor (gRPC) | 9090 |

Если порт занят, в `docker-compose.yml` можно изменить маппинг, например: `"18090:8090"` вместо `"8090:8090"`.

## Переменные окружения

Брокер Kafka задаётся через:

- `BROKER_URLS` или `KAFKA_BROKERS`: `redpanda:9092` (внутри сети Docker)

Базы и кэш:

- `DATABASE_HOST` / `DB_HOST`: `postgres`
- `REDIS_HOST`: `redis`
- Подключение к MinIO для s3storageprocessor: `S3_HOST`: `minio`

Сервисы читают конфиги из образов; при необходимости можно переопределить через `environment` в `docker-compose.yml` или смонтировать свой конфиг через `volumes`.
