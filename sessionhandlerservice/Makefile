MIGRATION_DIR = "./migration"
DB_DRIVER = "postgres"
DB_STRING = "postgres://postgres:roy9@192.168.20.14:5432/roy9?sslmode=disable"

.PHONY: run docker-build docker-run migrate-create migrate-up migrate-down migrate-down

run:
	go run ./cmd/sessionHandler/main.go

docker-build:
	docker build -t session-handler-service:0.0.1 -f Dockerfile .

docker-run:
	docker run --rm --name session-handler-service -p 8081:8081 session-handler-service:0.0.1

#make migrate-create name=your_migration_name
migrate-create:
	goose -dir ${MIGRATION_DIR} create $(name) sql

migrate-up:
	goose -dir ${MIGRATION_DIR} ${DB_DRIVER} ${DB_STRING} up

migrate-down:
	goose -dir ${MIGRATION_DIR} ${DB_DRIVER} ${DB_STRING} down
