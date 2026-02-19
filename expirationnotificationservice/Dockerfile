FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/expirationnotification/ -o /expirationnotification

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /expirationnotification /app/expirationnotification
COPY --from=builder /app/config.yaml /app/config.yaml

EXPOSE 8092

ENTRYPOINT [ "/app/expirationnotification" ]