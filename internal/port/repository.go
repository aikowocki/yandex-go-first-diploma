package port

import (
	"context"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	FindByNumber(ctx context.Context, number string) (*entity.Order, error)
	FindByUser(ctx context.Context, userID int64) ([]entity.Order, error)
	UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *int64) error
	FindPending(ctx context.Context) ([]entity.Order, error)
}

type BalanceRepository interface {
	LockByUserId(ctx context.Context, userID int64) error
	GetBalance(ctx context.Context, userID int64) (entity.Balance, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, amount int64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]entity.Transaction, error)
	AddAccrual(ctx context.Context, userID int64, orderNumber string, amount int64) error
}
