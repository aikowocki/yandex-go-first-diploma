# Варианты структуры проекта «Гофермарт» под разные архитектуры

Ниже представлены варианты организации кодовой базы для системы лояльности «Гофермарт» (Go, PostgreSQL, HTTP API) в рамках разных архитектурных подходов.

---

## 1. Flat / Стандартная Go-структура (без явной архитектуры)

Минимальный подход — всё в одном пакете или с минимальным разделением. Подходит для MVP и небольших сервисов.

```
gophermart/
├── main.go                  // точка входа, конфигурация, запуск сервера
├── config.go                // парсинг флагов и env-переменных
├── server.go                // настройка роутера, middleware
├── handlers.go              // все HTTP-хендлеры
├── storage.go               // интерфейс хранилища
├── postgres.go              // реализация хранилища на PostgreSQL
├── models.go                // User, Order, Balance, Withdrawal
├── auth.go                  // JWT/cookie логика
├── luhn.go                  // валидация номера заказа
├── accrual.go               // клиент к системе начислений
├── worker.go                // фоновый опрос системы начислений
├── migrations/
│   └── 001_init.sql
├── go.mod
└── go.sum
```

Плюсы: просто, быстро, нет лишних абстракций.
Минусы: при росте проекта превращается в кашу, тяжело тестировать изолированно.

---

## 2. Clean Architecture (Чистая архитектура)

Классическое разделение на слои: entity → usecase → adapter → infrastructure. Зависимости направлены внутрь — внешние слои зависят от внутренних, но не наоборот.

```
gophermart/
├── cmd/
│   └── gophermart/
│       └── main.go                    // точка входа, сборка зависимостей (wire)
│
├── internal/
│   ├── entity/                        // бизнес-сущности (чистые, без зависимостей)
│   │   ├── user.go                    // User, UserID
│   │   ├── order.go                   // Order, OrderStatus, валидация Луна
│   │   ├── balance.go                 // Balance
│   │   └── withdrawal.go             // Withdrawal
│   │
│   ├── usecase/                       // бизнес-логика (зависит только от entity)
│   │   ├── auth.go                    // RegisterUser, LoginUser
│   │   ├── order.go                   // SubmitOrder, GetUserOrders
│   │   ├── balance.go                 // GetBalance, Withdraw
│   │   └── accrual.go                 // ProcessAccruals (фоновая обработка)
│   │
│   ├── adapter/
│   │   ├── handler/                   // HTTP-хендлеры (входящий адаптер)
│   │   │   ├── router.go
│   │   │   ├── auth.go
│   │   │   ├── order.go
│   │   │   ├── balance.go
│   │   │   └── middleware.go
│   │   │
│   │   ├── repo/                      // репозитории (исходящий адаптер)
│   │   │   ├── user_postgres.go
│   │   │   ├── order_postgres.go
│   │   │   └── balance_postgres.go
│   │   │
│   │   └── accrualclient/             // клиент к внешнему сервису начислений
│   │       └── client.go
│   │
│   └── config/
│       └── config.go                  // парсинг RUN_ADDRESS, DATABASE_URI, ACCRUAL_SYSTEM_ADDRESS
│
├── migrations/
│   └── 001_init.sql
├── go.mod
└── go.sum
```

Ключевые интерфейсы (определяются в usecase, реализуются в adapter):

```go
// internal/usecase/order.go
type OrderRepository interface {
    Save(ctx context.Context, userID int64, order entity.Order) error
    FindByNumber(ctx context.Context, number string) (*entity.Order, error)
    FindByUser(ctx context.Context, userID int64) ([]entity.Order, error)
    UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual float64) error
}

type AccrualClient interface {
    GetOrderAccrual(ctx context.Context, orderNumber string) (*entity.AccrualResponse, error)
}
```

Плюсы: чёткое разделение ответственности, легко тестировать через моки, бизнес-логика не зависит от фреймворков.
Минусы: больше файлов и boilerplate, может быть overkill для маленького сервиса.

---

## 3. Hexagonal Architecture (Ports & Adapters)

Похожа на Clean Architecture, но акцент на «портах» (интерфейсы) и «адаптерах» (реализации). Ядро приложения определяет порты, внешний мир подключается через адаптеры.

```
gophermart/
├── cmd/
│   └── gophermart/
│       └── main.go
│
├── internal/
│   ├── domain/                        // ядро: сущности + бизнес-правила
│   │   ├── user.go
│   │   ├── order.go
│   │   ├── balance.go
│   │   └── luhn.go
│   │
│   ├── port/                          // порты (интерфейсы)
│   │   ├── incoming/                  // входящие порты (что приложение умеет делать)
│   │   │   ├── auth_service.go        // AuthService interface
│   │   │   ├── order_service.go       // OrderService interface
│   │   │   └── balance_service.go     // BalanceService interface
│   │   │
│   │   └── outgoing/                  // исходящие порты (что приложению нужно извне)
│   │       ├── user_repo.go           // UserRepository interface
│   │       ├── order_repo.go          // OrderRepository interface
│   │       ├── balance_repo.go        // BalanceRepository interface
│   │       └── accrual_provider.go    // AccrualProvider interface
│   │
│   ├── service/                       // реализация входящих портов (application services)
│   │   ├── auth.go
│   │   ├── order.go
│   │   └── balance.go
│   │
│   └── adapter/
│       ├── in/                        // входящие адаптеры
│       │   └── http/
│       │       ├── router.go
│       │       ├── auth_handler.go
│       │       ├── order_handler.go
│       │       ├── balance_handler.go
│       │       └── middleware.go
│       │
│       └── out/                       // исходящие адаптеры
│           ├── postgres/
│           │   ├── user_repo.go
│           │   ├── order_repo.go
│           │   └── balance_repo.go
│           └── accrual/
│               └── http_client.go
│
├── migrations/
│   └── 001_init.sql
├── go.mod
└── go.sum
```

Плюсы: максимальная заменяемость адаптеров (можно подменить PostgreSQL на SQLite, HTTP на gRPC), очень тестируемо.
Минусы: много мелких файлов, порог входа для новых разработчиков.

---

## 4. DDD (Domain-Driven Design)

Организация вокруг bounded contexts и агрегатов. Для Гофермарта можно выделить контексты: Identity (пользователи), Ordering (заказы), Loyalty (баланс и списания).

```
gophermart/
├── cmd/
│   └── gophermart/
│       └── main.go
│
├── internal/
│   ├── identity/                      // Bounded Context: Identity
│   │   ├── domain/
│   │   │   ├── user.go                // агрегат User
│   │   │   └── credentials.go        // value object Credentials
│   │   ├── app/
│   │   │   └── auth_service.go        // application service
│   │   ├── infra/
│   │   │   └── user_repo_pg.go        // инфраструктурная реализация
│   │   └── api/
│   │       ├── register_handler.go
│   │       └── login_handler.go
│   │
│   ├── ordering/                      // Bounded Context: Ordering
│   │   ├── domain/
│   │   │   ├── order.go               // агрегат Order
│   │   │   ├── order_status.go        // value object OrderStatus
│   │   │   └── luhn.go               // domain service — валидация
│   │   ├── app/
│   │   │   ├── submit_order.go        // команда
│   │   │   ├── get_orders.go          // запрос
│   │   │   └── accrual_worker.go      // процесс синхронизации с accrual
│   │   ├── infra/
│   │   │   ├── order_repo_pg.go
│   │   │   └── accrual_client.go
│   │   └── api/
│   │       └── order_handler.go
│   │
│   ├── loyalty/                       // Bounded Context: Loyalty
│   │   ├── domain/
│   │   │   ├── account.go             // агрегат LoyaltyAccount
│   │   │   └── withdrawal.go         // entity Withdrawal
│   │   ├── app/
│   │   │   ├── get_balance.go
│   │   │   └── withdraw.go
│   │   ├── infra/
│   │   │   └── account_repo_pg.go
│   │   └── api/
│   │       ├── balance_handler.go
│   │       └── withdrawal_handler.go
│   │
│   └── shared/                        // общие вещи между контекстами
│       ├── config/
│       │   └── config.go
│       ├── middleware/
│       │   └── auth.go
│       └── postgres/
│           └── connection.go
│
├── migrations/
│   ├── 001_users.sql
│   ├── 002_orders.sql
│   └── 003_loyalty.sql
├── go.mod
└── go.sum
```

Пример агрегата:

```go
// internal/loyalty/domain/account.go
type LoyaltyAccount struct {
    UserID    int64
    Current   float64
    Withdrawn float64
}

func (a *LoyaltyAccount) Withdraw(order string, sum float64) (Withdrawal, error) {
    if sum > a.Current {
        return Withdrawal{}, ErrInsufficientFunds
    }
    a.Current -= sum
    a.Withdrawn += sum
    return Withdrawal{Order: order, Sum: sum, ProcessedAt: time.Now()}, nil
}
```

Плюсы: отлично масштабируется, каждый контекст можно развивать независимо, бизнес-логика максимально выразительна.
Минусы: для маленького проекта — перебор, требует глубокого понимания домена.

---

## 5. TDD-ориентированная структура

Не столько архитектура, сколько подход к организации, где тесты — первоклассные граждане. Структура может быть любой (здесь — на базе Clean Architecture), но с акцентом на тестируемость.

```
gophermart/
├── cmd/
│   └── gophermart/
│       └── main.go
│
├── internal/
│   ├── entity/
│   │   ├── user.go
│   │   ├── user_test.go               // unit-тесты сущностей
│   │   ├── order.go
│   │   ├── order_test.go
│   │   └── luhn_test.go
│   │
│   ├── usecase/
│   │   ├── auth.go
│   │   ├── auth_test.go               // тесты с моками репозиториев
│   │   ├── order.go
│   │   ├── order_test.go
│   │   ├── balance.go
│   │   └── balance_test.go
│   │
│   ├── mock/                          // моки для тестирования
│   │   ├── user_repo.go               // сгенерированные или ручные моки
│   │   ├── order_repo.go
│   │   └── accrual_client.go
│   │
│   ├── handler/
│   │   ├── auth.go
│   │   ├── auth_test.go               // тесты хендлеров через httptest
│   │   ├── order.go
│   │   ├── order_test.go
│   │   ├── balance.go
│   │   └── balance_test.go
│   │
│   ├── repo/
│   │   ├── user_postgres.go
│   │   ├── user_postgres_test.go      // интеграционные тесты с testcontainers
│   │   ├── order_postgres.go
│   │   └── order_postgres_test.go
│   │
│   └── config/
│       └── config.go
│
├── integration/                       // e2e / интеграционные тесты
│   ├── auth_test.go
│   ├── order_flow_test.go
│   ├── balance_test.go
│   └── testhelpers.go                 // хелперы: поднятие БД, создание клиента
│
├── testdata/                          // фикстуры
│   ├── valid_orders.json
│   └── invalid_orders.json
│
├── migrations/
│   └── 001_init.sql
├── go.mod
└── go.sum
```

Пирамида тестирования:

```
         ┌─────────┐
         │  E2E    │  integration/ — полный цикл через HTTP
         ├─────────┤
         │ Integr. │  repo/*_test.go — с реальной БД (testcontainers)
         ├─────────┤
         │  Unit   │  entity/*_test.go, usecase/*_test.go, handler/*_test.go
         └─────────┘
```

Плюсы: высокое покрытие, уверенность в рефакторинге, моки позволяют тестировать слои изолированно.
Минусы: написание и поддержка тестов требует времени, моки могут расходиться с реальностью.

---

## 6. Onion Architecture (Луковая архитектура)

Вариация Clean Architecture с более явным разделением на концентрические слои. Каждый слой может зависеть только от более внутренних слоёв.

```
gophermart/
├── cmd/
│   └── gophermart/
│       └── main.go
│
├── internal/
│   ├── core/                          // самый внутренний слой
│   │   ├── model/                     // доменные модели
│   │   │   ├── user.go
│   │   │   ├── order.go
│   │   │   ├── balance.go
│   │   │   └── withdrawal.go
│   │   │
│   │   └── port/                      // интерфейсы репозиториев
│   │       ├── user_repository.go
│   │       ├── order_repository.go
│   │       └── balance_repository.go
│   │
│   ├── application/                   // слой приложения (use cases)
│   │   ├── auth_service.go
│   │   ├── order_service.go
│   │   ├── balance_service.go
│   │   └── accrual_service.go
│   │
│   └── infrastructure/                // самый внешний слой
│       ├── http/
│       │   ├── router.go
│       │   ├── auth_handler.go
│       │   ├── order_handler.go
│       │   ├── balance_handler.go
│       │   └── middleware.go
│       │
│       ├── persistence/
│       │   ├── postgres.go            // подключение к БД
│       │   ├── user_repo.go
│       │   ├── order_repo.go
│       │   └── balance_repo.go
│       │
│       ├── external/
│       │   └── accrual_client.go
│       │
│       └── config/
│           └── config.go
│
├── migrations/
│   └── 001_init.sql
├── go.mod
└── go.sum
```

```
┌──────────────────────────────────────────┐
│           Infrastructure                 │  HTTP, PostgreSQL, Accrual Client
│  ┌────────────────────────────────────┐  │
│  │         Application                │  │  Use Cases / Services
│  │  ┌──────────────────────────────┐  │  │
│  │  │          Core                │  │  │  Models + Port Interfaces
│  │  └──────────────────────────────┘  │  │
│  └────────────────────────────────────┘  │
└──────────────────────────────────────────┘
```

Плюсы: простая ментальная модель «слои как луковица», хорошо подходит для средних проектов.
Минусы: на практике мало отличается от Clean Architecture, может путать новичков.

---

## 7. CQRS (Command Query Responsibility Segregation)

Разделение операций записи (commands) и чтения (queries). Для Гофермарта: команды — регистрация, загрузка заказа, списание; запросы — получение заказов, баланса, списаний.

```
gophermart/
├── cmd/
│   └── gophermart/
│       └── main.go
│
├── internal/
│   ├── domain/
│   │   ├── user.go
│   │   ├── order.go
│   │   ├── balance.go
│   │   └── withdrawal.go
│   │
│   ├── command/                       // операции записи
│   │   ├── register_user.go           // RegisterUserCommand + Handler
│   │   ├── login_user.go
│   │   ├── submit_order.go
│   │   ├── withdraw.go
│   │   └── process_accrual.go
│   │
│   ├── query/                         // операции чтения
│   │   ├── get_orders.go              // GetOrdersQuery + Handler
│   │   ├── get_balance.go
│   │   └── get_withdrawals.go
│   │
│   ├── bus/                           // шина команд/запросов (опционально)
│   │   ├── command_bus.go
│   │   └── query_bus.go
│   │
│   ├── infra/
│   │   ├── http/
│   │   │   ├── router.go
│   │   │   ├── handlers.go           // маппинг HTTP → command/query
│   │   │   └── middleware.go
│   │   ├── persistence/
│   │   │   ├── write_repo.go         // репозиторий для записи
│   │   │   └── read_repo.go          // репозиторий для чтения (может быть оптимизирован)
│   │   └── accrual/
│   │       └── client.go
│   │
│   └── config/
│       └── config.go
│
├── migrations/
│   └── 001_init.sql
├── go.mod
└── go.sum
```

Пример команды:

```go
// internal/command/submit_order.go
type SubmitOrderCommand struct {
    UserID      int64
    OrderNumber string
}

type SubmitOrderHandler struct {
    repo   WriteOrderRepository
    luhn   func(string) bool
}

func (h *SubmitOrderHandler) Handle(ctx context.Context, cmd SubmitOrderCommand) error {
    if !h.luhn(cmd.OrderNumber) {
        return ErrInvalidOrderNumber
    }
    return h.repo.SaveOrder(ctx, cmd.UserID, cmd.OrderNumber)
}
```

Плюсы: чёткое разделение read/write, можно оптимизировать чтение отдельно от записи, хорошо ложится на event-driven.
Минусы: для простого CRUD — overkill, добавляет сложность маршрутизации команд.

---

## Сравнительная таблица

| Критерий               | Flat  | Clean | Hexagonal | DDD   | TDD   | Onion | CQRS  |
|------------------------|-------|-------|-----------|-------|-------|-------|-------|
| Простота старта        | ★★★★★ | ★★★   | ★★☆       | ★★☆   | ★★★   | ★★★   | ★★☆   |
| Тестируемость          | ★★☆   | ★★★★★ | ★★★★★     | ★★★★  | ★★★★★ | ★★★★  | ★★★★  |
| Масштабируемость       | ★☆☆   | ★★★★  | ★★★★      | ★★★★★ | ★★★   | ★★★★  | ★★★★★ |
| Кол-во файлов          | мало  | средне| много     | много | средне| средне| средне|
| Порог входа            | низкий| средний| высокий  | высокий| средний| средний| средний|
| Подходит для Гофермарта| MVP   | да    | да        | скорее нет | да | да   | скорее нет |

---

## Рекомендация для «Гофермарт»

Для учебного проекта такого масштаба оптимальный выбор — Clean Architecture или Onion Architecture. Они дают достаточную структуру для демонстрации навыков, при этом не перегружают проект лишними абстракциями. DDD и CQRS имеют смысл, если проект будет расти в сторону микросервисов или сложной доменной логики.
