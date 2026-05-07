# План итераций — GopherMart

Подход к балансу: ledger-паттерн (таблица transactions). Каждое изменение баланса — INSERT, баланс считается агрегатно. Таблица withdrawals не нужна.

---

## PR #1 — Фундамент и структура БД (День 1)

**Цель:** проект компилируется, запускается, подключается к БД, маршруты объявлены.

### Что сделать:

**1. go.mod / go.sum**
- Инициализировать модуль (`go mod init`)
- Зависимости: роутер (chi), драйвер БД (pgx), JWT-библиотека, bcrypt, логгер (slog или zap)

**2. internal/entity/**

- `user.go` — структура User (ID, Login, PasswordHash, CreatedAt)

- `order.go` — структура Order (ID, UserID, Number, Status, Accrual *int64, CreatedAt, UpdatedAt), тип OrderStatus string (NEW, PROCESSING, INVALID, PROCESSED), функция валидации по алгоритму Луна

- `balance.go` — структура Balance (Current int64, Withdrawn int64) — view-модель для API. Суммы в копейках, при отдаче в JSON конвертировать в рубли (делить на 100)

- `transaction.go` — структура Transaction (ID, UserID, OrderNumber, Type, Amount int64, CreatedAt), тип TransactionType int (0 = ACCRUAL, 1 = WITHDRAWAL). Amount в копейках. Для API-ответа GET /api/user/withdrawals — отдельная response-структура в хендлере

**3. internal/config/config.go**
- Парсинг флагов `-a`, `-d`, `-r`
- Чтение env-переменных `RUN_ADDRESS`, `DATABASE_URI`, `ACCRUAL_SYSTEM_ADDRESS`
- Env-переменные имеют приоритет над флагами

**4. migrations/001_init.sql**

```sql
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    login         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE orders (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id),
    number      VARCHAR(255) NOT NULL UNIQUE,
    status      VARCHAR(20)  NOT NULL DEFAULT 'NEW',
    accrual     BIGINT,          -- в копейках, NULL пока не рассчитано
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE transactions (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT       NOT NULL REFERENCES users(id),
    order_number VARCHAR(255) NOT NULL,
    type         SMALLINT     NOT NULL, -- 0 = ACCRUAL, 1 = WITHDRAWAL
    amount       BIGINT       NOT NULL, -- в копейках
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_type ON transactions(user_id, type);
```

**5. cmd/gophermart/main.go**
- Парсинг конфига
- Подключение к PostgreSQL
- Применение миграций
- Запуск HTTP-сервера

**6. internal/adapter/handler/router.go**
- Все маршруты:
  - POST /api/user/register
  - POST /api/user/login
  - POST /api/user/orders (auth)
  - GET  /api/user/orders (auth)
  - GET  /api/user/balance (auth)
  - POST /api/user/balance/withdraw (auth)
  - GET  /api/user/withdrawals (auth)
- Хендлеры пока возвращают `501 Not Implemented`

**7. internal/adapter/handler/middleware/**
- `logging.go` — логирование запросов (метод, путь, статус, время)
- `recovery.go` — перехват паник, ответ 500

### Результат PR #1:
Сервер стартует, БД с таблицами создаётся, маршруты отвечают 501.

---

## PR #2 — Аутентификация + Заказы (День 2)

**Цель:** регистрация → логин → загрузка заказа → просмотр заказов.

### Что сделать:

**1. internal/adapter/repo/user_postgres.go**
- `Create(ctx, user)` — INSERT, обработка конфликта по login (409)
- `FindByLogin(ctx, login)` — SELECT по логину

**2. internal/usecase/auth.go**
- Интерфейс `UserRepository` (Create, FindByLogin)
- `Register(ctx, login, password)` — bcrypt хеш, сохранение, генерация JWT
- `Login(ctx, login, password)` — поиск, проверка пароля, генерация JWT

**3. internal/adapter/handler/auth.go**
- `POST /api/user/register` — парсинг JSON {login, password}, вызов usecase, токен в ответ
- `POST /api/user/login` — аналогично

**4. internal/adapter/handler/middleware/auth.go**
- Извлечение JWT из запроса
- Валидация токена
- Запись userID в контекст

**5. internal/adapter/repo/order_postgres.go**
- `Save(ctx, userID, order)` — INSERT, обработка конфликта по number
- `FindByNumber(ctx, number)` — SELECT по номеру (для проверки: свой или чужой)
- `FindByUser(ctx, userID)` — SELECT заказов пользователя, ORDER BY uploaded_at DESC

**6. internal/usecase/order.go**
- Интерфейс `OrderRepository` (Save, FindByNumber, FindByUser)
- `SubmitOrder(ctx, userID, number)` — валидация Луна, проверка дубликатов (200 свой / 409 чужой), сохранение NEW
- `GetUserOrders(ctx, userID)` — список заказов

**7. internal/adapter/handler/order.go**
- `POST /api/user/orders` — чтение body как text/plain, вызов usecase
- `GET /api/user/orders` — JSON-ответ, 204 если пусто

### Результат PR #2:
Полный цикл: регистрация, логин, загрузка заказов, просмотр заказов.

---

## PR #3 — Баланс, списания, Accrual-воркер (День 3)

**Цель:** финальная функциональность — баланс, списания, фоновая синхронизация.

### Что сделать:

**1. internal/adapter/repo/balance_postgres.go**

Работает с таблицей `transactions`.

- `GetBalance(ctx, userID)`:
```sql
SELECT
    COALESCE(SUM(CASE WHEN type = 0 THEN amount ELSE 0 END), 0)
      - COALESCE(SUM(CASE WHEN type = 1 THEN amount ELSE 0 END), 0) AS current,
    COALESCE(SUM(CASE WHEN type = 1 THEN amount ELSE 0 END), 0) AS withdrawn
FROM transactions WHERE user_id = $1;
```
Результат в копейках — конвертировать в рубли при отдаче в JSON.

- `Withdraw(ctx, userID, orderNumber, sum)` — sum в копейках, в транзакции:
```sql
BEGIN;
SELECT pg_advisory_xact_lock($1);
-- вычислить текущий баланс (в копейках)
-- если current < sum → ROLLBACK, вернуть ошибку (402)
INSERT INTO transactions (user_id, order_number, type, amount)
    VALUES ($1, $2, 1, $3);
COMMIT;
```

- `GetWithdrawals(ctx, userID)`:
```sql
SELECT order_number, amount, created_at
FROM transactions
WHERE user_id = $1 AND type = 1
ORDER BY created_at DESC;
```
amount в копейках — конвертировать при отдаче.

- `AddAccrual(ctx, userID, orderNumber, amount)` — INSERT транзакции типа ACCRUAL (вызывается из accrual-воркера в одной транзакции с обновлением статуса заказа)

**2. internal/usecase/balance.go**
- Интерфейс `BalanceRepository` (GetBalance, Withdraw, GetWithdrawals, AddAccrual)
- `GetBalance(ctx, userID)`
- `Withdraw(ctx, userID, order, sum)` — валидация номера (Луна), вызов репозитория
- `GetWithdrawals(ctx, userID)`

**3. internal/adapter/handler/balance.go**
- `GET /api/user/balance` — JSON {current, withdrawn}
- `POST /api/user/balance/withdraw` — парсинг JSON {order, sum}, вызов usecase, 402 при нехватке
- `GET /api/user/withdrawals` — JSON-массив, 204 если пусто

**4. internal/adapter/accrual/client.go**
- HTTP-клиент к accrual-сервису
- `GetOrderAccrual(ctx, orderNumber)` — GET /api/orders/{number}
- Обработка: 200 (парсинг), 204 (не зарегистрирован), 429 (пауза Retry-After), 500

**5. internal/usecase/accrual.go**
- Фоновый воркер (горутина с тикером)
- Выбирает заказы со статусами NEW и PROCESSING
- Опрашивает accrual-сервис
- При PROCESSED — в одной транзакции (accrual от внешнего сервиса приходит в рублях — конвертировать в копейки перед сохранением):
```sql
BEGIN;
UPDATE orders SET status = 'PROCESSED', accrual = $1 WHERE number = $2;
INSERT INTO transactions (user_id, order_number, type, amount)
    VALUES ($3, $2, 0, $1);
COMMIT;
```
- При INVALID — UPDATE orders SET status = 'INVALID'
- При PROCESSING — UPDATE orders SET status = 'PROCESSING'
- При 429 — пауза на Retry-After

**6. cmd/gophermart/main.go — финальная сборка**
- Инициализация всех репозиториев, юзкейсов, хендлеров
- Запуск accrual-воркера в горутине
- Graceful shutdown (context + os.Signal)

### Результат PR #3:
Всё работает. 7 эндпоинтов, фоновый воркер, ledger-баланс.

---

## Советы

- Интерфейсы репозиториев определяй в usecase-слое, реализации — в adapter/repo
- Для JWT достаточно userID в claims, секрет можно вынести в конфиг
- `pg_advisory_xact_lock` — простой способ избежать гонки при списании, блокировка снимается автоматически при COMMIT/ROLLBACK
- Ledger-паттерн: баланс = SUM(type=0) - SUM(type=1). Никогда не UPDATE баланс напрямую, только INSERT в transactions. Все суммы в копейках, конвертация в рубли только на границе API
- Если в будущем агрегация станет медленной — добавить кеш-таблицу balances, обновляемую в той же транзакции
