MIGRATION_DIR = "./migration"
DB_DRIVER = "postgres"
DB_STRING = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

.PHONY: run docker-build docker-run migrate-create migrate-up migrate-down docs-gen

run:
	go run ./cmd/feedback/main.go

docker-build:
	docker build -t feedback-service:0.0.1 -f Dockerfile .

docker-run: docker-build
	docker run --rm --rm --network test-network --name feedback-service -p 8094:8094 feedback-service:0.0.1

#make migrate-create name=your_migration_name
migrate-create:
	goose -dir ${MIGRATION_DIR} create $(name) sql

migrate-up:
	goose -dir ${MIGRATION_DIR} ${DB_DRIVER} ${DB_STRING} up

migrate-down:
	goose -dir ${MIGRATION_DIR} ${DB_DRIVER} ${DB_STRING} down

docs-gen:
	swag init -g ./cmd/feedback/main.go -o ./docs


env: "dev"
db_config:
  db_host: "192.168.20.14"
  db_port: "5432"
  db_user: "postgres"
  db_name: "roy9"
  db_password: "5NLLbYo3Mc4HzQ"
http_server:
  address: "0.0.0.0:8000"
  timeout: 4s
  idle_timeout: 30s
  user: "user"
