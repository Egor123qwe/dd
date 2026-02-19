FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/resourcepool/ -o /resourcepool

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /resourcepool /app/resourcepool
COPY --from=builder /app/config.yaml /app/config.yaml

EXPOSE 8091

ENTRYPOINT [ "/app/resourcepool" ]