FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

COPY go.mod go.sum ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /history-service cmd/main.go

FROM alpine:3.19

WORKDIR /app

ENV CONFIG_PATH=./app/config/dev.yaml

COPY --from=builder /history-service /app/history-service
COPY config/ app/config/

EXPOSE 8899

ENTRYPOINT [ "/app/history-service" ]
