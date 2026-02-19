.PHONY: run docker-build docker-run

run:
	go run ./cmd/expirationnotification/main.go

docker-build:
	docker build -t expiration-notification-service:0.0.1 -f Dockerfile .

docker-run: docker-build
	docker run --rm --network test-network --name expiration-notification-service -p 8092:8092 expiration-notification-service:0.0.1
