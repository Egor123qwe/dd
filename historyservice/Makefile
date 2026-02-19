export CONFIG_PATH=./config/local.yaml

build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/history-service cmd/main.go

run:
	go run cmd/main.go

migrate: 
	GOOSE_DRIVER=postgres GOOSE_DBSTRING=postgres://myuser:mypassword@localhost:35432/mydatabase?sslmode=disable GOOSE_MIGRATION_DIR=migrations/ goose up
