package postgres_gorm

import (
	"context"
	"errors"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type OrderRepo struct {
	baseRepo
}

func NewOrderRepo(txManger *TxManager) *OrderRepo {
	return &OrderRepo{baseRepo: baseRepo{txManager: txManger}}
}

func (r *OrderRepo) Create(ctx context.Context, order *entity.Order) error {
	if err := r.db(ctx).Create(order).Error; err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrOrderExists
		}
		return err
	}
	return nil
}

func (r *OrderRepo) FindByNumber(ctx context.Context, number string) (*entity.Order, error) {
	order := &entity.Order{}

	if err := r.db(ctx).Where("number = ?", number).First(order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrOrderNotFound
		}
		return nil, err
	}

	return order, nil
}

func (r *OrderRepo) FindByUser(ctx context.Context, userID int64) ([]entity.Order, error) {
	var orders []entity.Order

	if err := r.db(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *int64) error {
	return r.db(ctx).
		Model(&entity.Order{}).
		Where("number = ?", number).
		Updates(map[string]interface{}{
			"status":  status,
			"accrual": accrual,
		}).
		Error
}

func (r *OrderRepo) FindPending(ctx context.Context) ([]entity.Order, error) {
	var orders []entity.Order

	if err := r.db(ctx).
		Where("status IN ?", entity.OrderPendingStatuses).
		Limit(100).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
