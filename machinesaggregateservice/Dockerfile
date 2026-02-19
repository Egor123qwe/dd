FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/aggregator cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/bin/aggregator .
COPY config/config.yaml /root/config/config.yaml
COPY config/dev.yaml /root/config/dev.yaml

ENV CONFIG_PATH=/root/config/config.yaml

EXPOSE 8000 8001

CMD ["./aggregator"]
