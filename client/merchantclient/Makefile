# proto compile config
proto_dir         = pkg/proto/runtimedaemon/proto
proto_build_dir   = pkg/proto/runtimedaemon/generate

compile-proto:
	protoc -I$(proto_dir) \
	--proto_path=$(proto_dir) \
	--go_out=$(proto_build_dir) \
	--go-grpc_out=$(proto_build_dir) \
	$(proto_dir)/*.proto

build:
	go build -o ./merchantclient cmd/merchantclient/main.go
