#!/usr/bin/env sh
# Проверка: регистрация -> логин -> декодирование JWT (payload) -> запрос с токеном на защищённый маршрут.
set -e
BASE="${1:-http://localhost:8080}"

echo "=== 1. Register ==="
REG=$(curl -s -X POST "$BASE/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"verify@test.com","username":"verifyuser","password":"password123"}')
echo "$REG" | head -c 200
echo "..."

ACCESS=$(echo "$REG" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$ACCESS" ]; then
  echo "No access_token in response"
  exit 1
fi
echo "Got access_token (length ${#ACCESS})"

echo ""
echo "=== 2. Request without token (expect 401 or 502) ==="
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/users/me")
echo "HTTP $CODE"

echo ""
echo "=== 3. Request with Bearer token ==="
CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $ACCESS" "$BASE/api/v1/users/me")
echo "HTTP $CODE (502 = user service is gRPC only, no HTTP yet)"

echo ""
echo "=== 4. JWT payload (base64 middle part) ==="
MIDDLE=$(echo "$ACCESS" | cut -d'.' -f2)
# add padding if needed
PAD=$((4 - ${#MIDDLE} % 4))
[ "$PAD" -ne 4 ] && MIDDLE="${MIDDLE}$(printf '=%.0s' $(seq 1 $PAD))"
echo "$MIDDLE" | base64 -d 2>/dev/null || echo "$MIDDLE" | base64 -d 2>/dev/null
echo ""

echo "Done. Check payload for user_id, email, username, kind, exp."
