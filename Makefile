.PHONY: build test run run-gateway migrate docker-up docker-down

build:
	cd auth_service && go build -o ../bin/auth_server ./cmd/auth_server
	cd auth_service && go build -o ../bin/migrate ./cmd/migrate
	cd api_gateway && go build -o ../bin/gateway ./cmd/service

test:
	cd auth_service && go test ./...
	cd api_gateway && go test ./...

run: build
	./bin/auth_server

run-gateway: build
	./bin/gateway

migrate: build
	./bin/migrate

docker-up:
	docker compose up -d

# Поднять всё, включая чат (ожидает репозиторий chat-message-mgz рядом: ../chat-message-mgz или CHAT_REPO_PATH)
docker-up-chat:
	CHAT_REPO_PATH=$${CHAT_REPO_PATH:-../chat-message-mgz} CHAT_MIGRATIONS_PATH=$${CHAT_MIGRATIONS_PATH:-../chat-message-mgz} docker compose up -d

# Поднять весь стек: auth, gateway, chat, user, postgres, redis, minio
docker-up-all:
	CHAT_REPO_PATH=$${CHAT_REPO_PATH:-../chat-message-mgz} CHAT_MIGRATIONS_PATH=$${CHAT_MIGRATIONS_PATH:-../chat-message-mgz} \
	USER_REPO_PATH=$${USER_REPO_PATH:-../user-mgz} USER_MIGRATIONS_PATH=$${USER_MIGRATIONS_PATH:-../user-mgz} \
	docker compose up -d

docker-down:
	docker compose down
