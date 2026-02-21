# proto compile config
proto_dir         = internal/handler/grpc/proto
proto_build_dir   = internal/handler/grpc/generate

compile-proto:
	protoc -I$(proto_dir) \
	--proto_path=$(proto_dir) \
	--go_out=$(proto_build_dir) \
	--go-grpc_out=$(proto_build_dir) \
	$(proto_dir)/*.proto

build:
	GOTOOLCHAIN=local go build -o ./runtimedaemon cmd/runtimedaemon/main.go
