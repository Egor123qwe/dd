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
| **Сайт (rental-portal)**  | **3000** |

Если порт занят, в `docker-compose.yml` можно изменить маппинг, например: `"18090:8090"` вместо `"8090:8090"`.

## Тестирование с другого компьютера в локальной сети

Если порты в основном compose привязаны к localhost (`127.0.0.1:PORT:PORT`), с другого ПК в сети к ним не подключиться. Чтобы открыть доступ по LAN:

**1. Запуск с привязкой к всем интерфейсам**

```bash
docker compose -f docker-compose.yml -f docker-compose.lan.yml up -d
```

В `docker-compose.lan.yml` порты userservice (8052) и connectioncoordinatorservice (8090) переопределены на `0.0.0.0`, чтобы к ним можно было обращаться по IP ноутбука из сети.

**2. Вариант А — веб-интерфейс с другого ПК**

- На ноутбуке: поднять compose (с `docker-compose.lan.yml` при необходимости), запустить **merchantclient** (он слушает `:8080` на всех интерфейсах).
- Узнать IP ноутбука в LAN (macOS: `ipconfig getifaddr en0`, Linux: `hostname -I`).
- На другом ПК в браузере открыть: `http://<IP_ноутбука>:8080`.
- Если не открывается — проверить фаервол на ноутбуке (разрешить входящие на порт 8080).

**3. Вариант Б — merchantclient на другом ПК, бэкенд на ноутбуке**

- На ноутбуке: `docker compose -f docker-compose.yml -f docker-compose.lan.yml up -d`.
- На другом ПК: собрать/скопировать merchantclient и runtimedaemon, запустить с указанием адреса ноутбука:

```bash
export ROY9_AUTH_SERVICE_URL="http://<IP_ноутбука>:8052"
export ROY9_CONNECTION_URL="ws://<IP_ноутбука>:8090/api/v1/ws"
./merchantclient
```

В браузере на этом же ПК открыть `http://localhost:8080`. Runtimedaemon и Docker должны быть на этом ПК.

## Переменные окружения

Брокер Kafka задаётся через:

- `BROKER_URLS` или `KAFKA_BROKERS`: `redpanda:9092` (внутри сети Docker)

Базы и кэш:

- `DATABASE_HOST` / `DB_HOST`: `postgres`
- `REDIS_HOST`: `redis`
- Подключение к MinIO для s3storageprocessor: `S3_HOST`: `minio`

Сервисы читают конфиги из образов; при необходимости можно переопределить через `environment` в `docker-compose.yml` или смонтировать свой конфиг через `volumes`.
