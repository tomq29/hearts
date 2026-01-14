FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o notification-service ./cmd/server/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/notification-service .

EXPOSE 8081
CMD ["./notification-service"]
