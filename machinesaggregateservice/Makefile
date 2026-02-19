export CONFIG_PATH=./config/config.yaml

.PHONY: run docs-gen

run:
	go run cmd/main.go

docs-gen:
	swag init -g ./cmd/main.go -o ./docs
