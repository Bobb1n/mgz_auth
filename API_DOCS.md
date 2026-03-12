## API через gateway (единственная точка входа)

Все запросы идут только на шлюз: `http://localhost:8080`.

Шлюз:
- проверяет JWT (для всех путей, кроме `/health` и `/api/v1/auth*`);
- при валидном токене добавляет заголовки `X-User-Id`, `X-User-Email`, `X-User-Username`;
- проксирует запросы:
  - `/api/v1/auth*` → auth_service (HTTP),
  - `/v1/*` → chat-message (HTTP/gRPC-фасад – когда появится),
  - `/api/v1/users*` → user-mgz (HTTP/gRPC-фасад – когда появится).

---

### 1. Health шлюза

```bash
curl -i http://localhost:8080/health
```

Ожидаемый ответ: `200 OK` и тело:

```json
{"status":"ok","service":"api-gateway"}
```

---

### 2. Auth: регистрация пользователя

`POST /api/v1/auth/register`

```bash
curl -i -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@example.com",
    "username": "user1",
    "password": "password123"
  }'
```

Ожидаемый результат:
- код `201` или `200`;
- JSON с полями: `id`, `email`, `username`, `access_token`, `refresh_token`.

---

### 3. Auth: логин (email / username)

`POST /api/v1/auth/login`

```bash
# логин по email
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "user1@example.com",
    "password": "password123"
  }'

# логин по username
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "user1",
    "password": "password123"
  }'
```

Из ответа забираем `access_token` и `refresh_token`.

---

### 4. Auth: обновление access по refresh

`POST /api/v1/auth/refresh`

```bash
REFRESH="...refresh_token из логина/регистрации..."

curl -i -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"${REFRESH}\"
  }"
```

Ожидаемый результат:
- код `200`;
- новый `access_token` (и при необходимости новый `refresh_token`).

---

### 5. Auth: logout (инвалидация refresh-токена)

`POST /api/v1/auth/logout`

```bash
REFRESH="...refresh_token..."

curl -i -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"${REFRESH}\"
  }"
```

После этого refresh-токен добавляется в blacklist (Redis), повторное использование должно приводить к ошибке.

---

### 6. Пример защищённого эндпоинта (через gateway)

Все пути, кроме `/health` и `/api/v1/auth*`, требуют заголовок:

`Authorization: Bearer <access_token>`

Пример запроса к защищённому ресурсу auth-сервиса (предположим, что есть `GET /api/v1/auth/me`):

```bash
ACCESS_TOKEN="...access_token..."

curl -i http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

Ожидания:
- без заголовка или с битым токеном → **401** от шлюза;
- с валидным токеном:
  - шлюз декодирует JWT;
  - добавляет заголовки `X-User-Id`, `X-User-Email`, `X-User-Username`;
  - проксирует запрос в auth_service.

---

### 7. Проверка полей JWT (payload access-токена)

Проще всего посмотреть payload токена через `jwt.io`:

1. Получить `access_token` (из логина/регистрации).
2. Открыть в браузере сайт `https://jwt.io`.
3. Вставить токен целиком в левое поле — в правой части появится раскодированный JSON.

В payload должны быть:
- `user_id` (int64);
- `email`;
- `username`;
- `kind`: `"access"` для access-токена;
- стандартные поля `exp`, `iat`.

---

### 8. Chat-service HTTP фасад через gateway

Шлюз поднимает простой HTTP‑фасад над gRPC‑сервисом `chat-message-mgz` (`ChatMessageService`).

#### 8.1. Создать личный чат

`POST /v1/chats/direct`

Текущий пользователь берётся из access‑токена (`X-User-Id`), второго участника передаём в `other_user_id`.

```bash
ACCESS_TOKEN="...access_token..."

curl -i -X POST http://localhost:8080/v1/chats/direct \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "other_user_id": "42"
  }'
```

Ответ: JSON с объектом `Chat` (id, user1_id, user2_id, created_at).

#### 8.2. Отправить сообщение в чат

`POST /v1/messages`

```bash
ACCESS_TOKEN="...access_token..."

curl -i -X POST http://localhost:8080/v1/messages \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "chat_id": "chat-id-here",
    "text": "hello world"
  }'
```

Ответ: JSON с объектом `Message` (id, chat_id, sender_user_id, text, status, created_at, updated_at).

#### 8.3. Получить список чатов пользователя

`GET /v1/chats?limit=&offset=`

```bash
ACCESS_TOKEN="...access_token..."

curl -i "http://localhost:8080/v1/chats?limit=10&offset=0" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

Ответ:

```json
{
  "chats": [ /* массив ChatPreview */ ],
  "limit": 10,
  "offset": 0
}
```

---

### 9. User-service (через gateway)

User‑сервис (`/Users/bobb1n/GolandProjects/user-mgz`) тоже сейчас gRPC‑only (порт `50052` внутри Docker).  
Шлюз проксирует:

- `/api/v1/users` и `/api/v1/users/*` → `USER_SERVICE_URL` (по умолчанию `http://user:8083` внутри сети Docker).

После появления HTTP/REST или HTTP→gRPC‑фасада в user‑сервисе, доступ будет через gateway:

```bash
ACCESS_TOKEN="...access_token..."

curl -i http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

Сейчас, пока в user‑сервисе нет HTTP‑обработчиков, такие запросы будут возвращать 502.

---

### 10. Единственный внешний порт

В `docker-compose.yml` наружу мапится только порт шлюза:

- `gateway`: `8080:8080`

Остальные сервисы (auth_service, chat-message, user-mgz, postgres, redis, minio) доступны только внутри Docker‑сети и ходят друг к другу по внутренним адресам:

- `app:8081` – auth_service
- `chat:50051` – chat-message (gRPC)
- `user:50052` – user-mgz (gRPC)
- `postgres:5432`, `redis:6379`, `minio:9000`

Все внешние клиенты используют только `http://localhost:8080/...`.

