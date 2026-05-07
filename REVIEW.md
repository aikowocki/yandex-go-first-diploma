# Code Review: Clean Architecture & Идиоматичность

## Архитектура

- [x] **1. Хендлеры зависят от интерфейсов, а не конкретных структур**
  AuthHandler, OrderHandler, BalanceHandler, HealthHandler — все переведены на интерфейсы.

- [~] **2. Доменные ошибки не должны знать про HTTP** — пропущено, не критично
  В `entity/order.go` комментарии `// -> 200`, `// -> 409` — косметика, на код не влияет.

- [x] **3. TxManager вынесен в отдельный пакет**
  Создан `internal/port/tx.go`, usecase и worker используют `port.TxManager`.

- [x] **4. Worker содержит бизнес-логику, а живёт в adapter**
  Исправлено — бизнес-логика вынесена в `usecase/accrual.go`, worker остался планировщиком.

## Идиоматичность

- [x] **5. `switch true` → `switch`**
  Исправлено в order.go и balance.go.

- [~] **6. Глобальный логгер `zap.S()`** — пропущено, оставлено как есть
  Работает, рефакторинг на DI-логгер — большой объём изменений ради небольшого выигрыша.

- [x] **7. Баг в `client.go`: `error` вместо `"error"`**
  Исправлено — теперь `"error"`, а не `error`.

- [x] **8. Retry-клиент не уважает context**
  Исправлено — `time.Sleep` заменён на `select` с `time.After` и `r.Context().Done()`.

- [x] **9. `JWTManager` — конкретный тип в usecase**
  Исправлено — добавлен интерфейс `TokenGenerator`, usecase больше не импортирует `pkg/auth`.

## Мелочи

- [x] **10. `updated_at` не обновляется при UpdateStatus**
  Исправлено — добавлено `updated_at = now()` в запрос.

- [~] **11. Нет валидации длины логина/пароля в usecase** — пропущено, нет в ТЗ

- [x] **12. Параметр конструктора `NewHealthHandler(pool Pinger)` → `NewHealthHandler(pinger Pinger)`**
