FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /userservice ./cmd

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /userservice /app/userservice
COPY --from=builder /app/migration /app/migration

EXPOSE 8052

ENTRYPOINT ["/app/userservice"]
