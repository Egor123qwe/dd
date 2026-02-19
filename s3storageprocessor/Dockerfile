FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/s3storageprocessor/ -o /s3storageprocessor

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /s3storageprocessor /app/s3storageprocessor
COPY --from=builder /app/config.yaml /app/config.yaml

EXPOSE 9090

ENTRYPOINT [ "/app/s3storageprocessor" ]