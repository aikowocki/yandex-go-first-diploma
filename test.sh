#!/bin/bash

BASE_URL="http://localhost:8081"
ACCRUAL_URL="http://localhost:8080"
COOKIES="cookies.txt"
USER="user_$(date +%s)"
# Номера заказов, проходящие Luhn
ORDER1="12345678903"
ORDER2="1234567812345670"
WITHDRAW_ORDER="2377225624"

echo "=== 0. Регистрация механики вознаграждения в accrual ==="
curl -s -X POST "$ACCRUAL_URL/api/goods" \
  -H "Content-Type: application/json" \
  -d '{"match":"Bork","reward":10,"reward_type":"%"}'
echo

echo "=== 1. Регистрация заказа в accrual ($ORDER1) ==="
curl -s -X POST "$ACCRUAL_URL/api/orders" \
  -H "Content-Type: application/json" \
  -d "{\"order\":\"$ORDER1\",\"goods\":[{\"description\":\"Чайник Bork\",\"price\":7000}]}"
echo

echo "=== 2. Регистрация пользователя ($USER) ==="
curl -s -c "$COOKIES" -X POST "$BASE_URL/api/user/register" \
  -H "Content-Type: application/json" \
  -d "{\"login\":\"$USER\",\"password\":\"test123\"}"
echo

echo "=== 3. Загрузка заказа ($ORDER1) ==="
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -b "$COOKIES" -X POST "$BASE_URL/api/user/orders" \
  -H "Content-Type: text/plain" \
  -d "$ORDER1")
echo "HTTP $STATUS"

echo "=== 4. Проверка заказов ==="
curl -s -b "$COOKIES" "$BASE_URL/api/user/orders" | python3 -m json.tool 2>/dev/null || echo "(пусто)"

echo "=== 5. Ждём 15 сек (воркер опросит accrual) ==="
sleep 15

echo "=== 6. Проверка заказов (после воркера) ==="
curl -s -b "$COOKIES" "$BASE_URL/api/user/orders" | python3 -m json.tool 2>/dev/null || echo "(пусто)"

echo "=== 7. Проверка баланса ==="
curl -s -b "$COOKIES" "$BASE_URL/api/user/balance" | python3 -m json.tool

echo "=== 8. Списание (1.5 балла) ==="
curl -s -b "$COOKIES" -X POST "$BASE_URL/api/user/balance/withdraw" \
  -H "Content-Type: application/json" \
  -d "{\"order\":\"$WITHDRAW_ORDER\",\"sum\":1.5}"
echo

echo "=== 9. Баланс после списания ==="
curl -s -b "$COOKIES" "$BASE_URL/api/user/balance" | python3 -m json.tool

echo "=== 10. Список списаний ==="
curl -s -b "$COOKIES" "$BASE_URL/api/user/withdrawals" | python3 -m json.tool 2>/dev/null || echo "(пусто)"

echo "=== 11. Health check ==="
curl -s -o /dev/null -w "HTTP %{http_code}" "$BASE_URL/ping"
echo

rm -f "$COOKIES"
echo
echo "=== Done ==="
