# Валидация реквестов в Go

## Варианты

### 1. go-playground/validator — стандарт де-факто
Теги на структурах, встроен в gin из коробки.

```go
type WithdrawRequest struct {
    Order string  `json:"order" validate:"required"`
    Sum   float64 `json:"sum" validate:"required,gt=0"`
}

validate := validator.New()
err := validate.Struct(req)
```

### 2. ozzo-validation — валидация через код, без тегов
```go
err := validation.ValidateStruct(&req,
    validation.Field(&req.Order, validation.Required),
    validation.Field(&req.Sum, validation.Required, validation.Min(0.01)),
)
```

### 3. Ручная валидация
```go
if req.Sum <= 0 {
    return error
}
```

Для маленьких структур (2-3 поля) — ок. Для больших — go-playground/validator.

---

## Сложные примеры go-playground/validator

```go
type CreateOrderRequest struct {
    // Базовые
    Email    string `validate:"required,email"`
    Age      int    `validate:"required,gte=18,lte=120"`
    Password string `validate:"required,min=8,max=64"`

    // Условная валидация — если PaymentType == "card", CardNumber обязателен
    PaymentType string `validate:"required,oneof=card cash crypto"`
    CardNumber  string `validate:"required_if=PaymentType card"`

    // Вложенные структуры
    Address Address `validate:"required,dive"`

    // Слайсы — каждый элемент валидируется
    Tags []string `validate:"max=5,dive,min=1,max=20"`

    // Кастомный формат
    Phone string `validate:"required,e164"` // +79991234567
    URL   string `validate:"omitempty,url"` // необязательно, но если есть — валидный URL
}

type Address struct {
    City   string `validate:"required,min=2"`
    Street string `validate:"required"`
    Zip    string `validate:"required,numeric,len=6"`
}
```

## Кастомный валидатор (например Luhn)

```go
validate := validator.New()

validate.RegisterValidation("luhn", func(fl validator.FieldLevel) bool {
    return entity.ValidateLuhn(fl.Field().String())
})

// теперь можно использовать в тегах
type WithdrawRequest struct {
    Order string  `validate:"required,luhn"`
    Sum   float64 `validate:"required,gt=0"`
}
```

## Красивые ошибки для API

```go
err := validate.Struct(req)
if err != nil {
    var ve validator.ValidationErrors
    if errors.As(err, &ve) {
        out := make(map[string]string, len(ve))
        for _, fe := range ve {
            // fe.Field() = "Email", fe.Tag() = "required"
            out[fe.Field()] = fmt.Sprintf("failed on '%s'", fe.Tag())
        }
        // {"Email": "failed on 'required'", "Sum": "failed on 'gt'"}
        c.JSON(400, out)
    }
}
```
