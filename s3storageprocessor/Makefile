PROTO_DIR = "./internal/proto"
PROTO_BUILD_DIR = "./internal/proto/gen"

.PHONY: run lint compile-proto

run:
	go run ./cmd/s3storageprocessor/main.go

lint:
	golangci-lint run

compile-proto:
	protoc -I$(PROTO_DIR) \
	--proto_path=$(PROTO_DIR) \
	--go_out=$(PROTO_BUILD_DIR) \
	--go-grpc_out=$(PROTO_BUILD_DIR) \
	$(PROTO_DIR)/*.proto
