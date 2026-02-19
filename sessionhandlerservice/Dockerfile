FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/sessionHandler/ -o /sessionhandler


FROM alpine:3.19

WORKDIR /app

COPY --from=builder /sessionhandler /app/sessionhandler
COPY config/        /app/config/

EXPOSE 8094

ENTRYPOINT [ "/app/sessionhandler" ]
