.PHONY: build test run migrate docker-up docker-down

build:
	go build -o bin/auth_server ./cmd/auth_server
	go build -o bin/migrate ./cmd/migrate

test:
	go test ./...

run: build
	./bin/auth_server

migrate: build
	./bin/migrate

docker-up:
	docker compose up -d

docker-down:
	docker compose down
