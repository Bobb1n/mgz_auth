FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /auth_server ./cmd/auth_server
RUN CGO_ENABLED=0 go build -o /migrate ./cmd/migrate

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /auth_server .
COPY --from=builder /migrate .

EXPOSE 8081

CMD ["sh", "-c", "./migrate && ./auth_server"]
