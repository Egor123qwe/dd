.PHONY: run docker-build docker-run

run:
	go run ./cmd/coordinator/main.go

docker-build:
	docker build -t coordinator-service:0.0.1 -f Dockerfile .

docker-run:
	docker run --rm --name coordinator-service -p 8085:8085 coordinator-service:0.0.1