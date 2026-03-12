FROM golang:1.25-alpine AS builder
WORKDIR /app

# Скопировать исходники gateway (с go.mod внутри) и chat-message-mgz внутрь билда
COPY auth_service/api_gateway ./api_gateway
COPY chat-message-mgz /chat-message-mgz

WORKDIR /app/api_gateway
RUN go mod download
RUN CGO_ENABLED=0 go build -o /service ./cmd/service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /service .
EXPOSE 8080
CMD ["./service"]
