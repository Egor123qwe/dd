FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -C cmd/coordinator/ -o /connectioncoordinator

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /connectioncoordinator /app/connectioncoordinator
COPY config/        /app/config/

EXPOSE 8090

ENTRYPOINT [ "/app/connectioncoordinator" ]