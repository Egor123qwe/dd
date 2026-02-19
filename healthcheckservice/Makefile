export CONFIG_PATH=./config/local.yaml

build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/billing-serv cmd/main.go

run:
	go run cmd/main.go

docs-gen:
	swag init -g ./cmd/main.go -o ./cmd/docs
