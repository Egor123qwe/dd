FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

COPY go.mod go.sum ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/ -o /healthcheck

FROM alpine:3.19

WORKDIR /app

ENV CONFIG_PATH=./app/config/dev.yaml

COPY --from=builder /healthcheck /app/healthcheck
COPY config/ app/config/

EXPOSE 9000

ENTRYPOINT [ "/app/healthcheck" ]
