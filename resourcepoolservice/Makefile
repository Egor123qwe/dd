MIGRATION_DIR = "./migration"
DB_DRIVER = "postgres"
DB_STRING = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

PROTO_DIR = "./internal/proto"
PROTO_BUILD_DIR = "./internal/proto/gen"

.PHONY: run docker-build docker-run migrate-create migrate-up migrate-down compile-proto

run:
	go run ./cmd/resourcepool/main.go

docker-build:
	docker build -t resource-pool-service:0.0.1 -f Dockerfile .

docker-run:
	docker run --rm --name resource-pool-service -p 8081:8081 resource-pool-service:0.0.1

#make migrate-create name=your_migration_name
migrate-create:
	goose -dir ${MIGRATION_DIR} create $(name) sql

migrate-up:
	goose -dir ${MIGRATION_DIR} ${DB_DRIVER} ${DB_STRING} up

migrate-down:
	goose -dir ${MIGRATION_DIR} ${DB_DRIVER} ${DB_STRING} down

compile-proto:
	protoc -I$(PROTO_DIR) \
	--proto_path=$(PROTO_DIR) \
	--go_out=$(PROTO_BUILD_DIR) \
	--go-grpc_out=$(PROTO_BUILD_DIR) \
	$(PROTO_DIR)/*.proto
