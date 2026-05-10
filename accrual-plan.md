# Accrual: план реализации

## Часть 1: HTTP-клиент (`internal/adapter/accrual/client.go`)

### API accrual-сервиса

Один эндпоинт: `GET /api/orders/{number}`

Ответ 200:
```json
{
    "order": "2377225624",
    "status": "PROCESSED",
    "accrual": 500
}
```

Статусы: `REGISTERED`, `PROCESSING`, `INVALID`, `PROCESSED`.
Accrual приходит только при `PROCESSED`.

Другие ответы:
- 204 — заказ не зарегистрирован
- 429 — слишком много запросов, заголовок `Retry-After` (секунды)

### Структура клиента

```go
package accrual

type Client struct {
    baseURL    string
    httpClient *http.Client
}

type AccrualResponse struct {
    Order   string  `json:"order"`
    Status  string  `json:"status"`
    Accrual float64 `json:"accrual"`
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL:    baseURL,
        httpClient: &http.Client{Timeout: 5 * time.Second},
    }
}

func (c *Client) GetOrderAccrual(ctx context.Context, number string) (*AccrualResponse, error) {
    // GET baseURL/api/orders/{number}
    // 200 → распарсить JSON
    // 204 → return nil, nil (заказ не найден)
    // 429 → вернуть специальную ошибку с Retry-After
    // остальное → return nil, err
}
```

### Ошибка для 429

```go
type ErrTooManyRequests struct {
    RetryAfter time.Duration
}

func (e *ErrTooManyRequests) Error() string {
    return fmt.Sprintf("too many requests, retry after %s", e.RetryAfter)
}
```

---

## Часть 2: Воркер (`internal/usecase/accrual.go`)

Фоновая горутина:
1. Раз в 5 секунд берёт pending-заказы из базы
2. Для каждого опрашивает accrual-сервис
3. Обновляет статус заказа
4. Если `PROCESSED` — начисляет баллы

### FindPending в `repo_order.go`

```go
func (r *OrderRepo) FindPending(ctx context.Context) ([]entity.Order, error) {
    q := `SELECT id, user_id, number, status FROM orders WHERE status IN ('NEW', 'PROCESSING')`
    // ...
}
```

### Структура воркера

```go
type AccrualWorker struct {
    orderRepo     OrderRepository   // FindPending + UpdateStatus
    balanceRepo   BalanceRepository // AddAccrual
    accrualClient *accrual.Client
    txManager     TxManager
}

func (w *AccrualWorker) Run(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            w.processOrders(ctx)
        }
    }
}
```

### processOrders — для каждого pending-заказа:
- Запрос к accrual
- Если 429 — подождать `RetryAfter`
- Если статус изменился — `txManager.Do`: `UpdateStatus` + `AddAccrual` (если PROCESSED) в одной транзакции

---

## Порядок действий
1. Написать `client.go` — HTTP-клиент
2. Добавить `FindPending` в `repo_order.go`
3. Написать `accrual.go` — воркер
4. Подключить в `main.go` — создать клиент, воркер, запустить горутину
