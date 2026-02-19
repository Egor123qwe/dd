FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/feedback/ -o /feedback

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /feedback /app/feedback
COPY --from=builder /app/config.yaml /app/config.yaml

EXPOSE 8095

ENTRYPOINT [ "/app/feedback" ]
