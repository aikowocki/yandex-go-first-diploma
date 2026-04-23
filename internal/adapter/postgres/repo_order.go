package postgres

import (
	"context"
	"errors"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrderRepo struct {
	baseRepo
}

func NewOrderRepo(txManger *TxManager) *OrderRepo {
	return &OrderRepo{baseRepo: baseRepo{txManager: txManger}}
}

func (r *OrderRepo) Create(ctx context.Context, order *entity.Order) error {
	q := "INSERT INTO orders (user_id, number) VALUES ($1, $2) RETURNING id, status, created_at, updated_at"
	err := r.db(ctx).QueryRow(ctx, q, order.UserID, order.Number).
		Scan(&order.ID, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {

		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrOrderExists
		}

		return err
	}

	return nil
}

func (r *OrderRepo) FindByNumber(ctx context.Context, number string) (*entity.Order, error) {
	order := &entity.Order{}
	q := "SELECT id, user_id, number, status, accrual, created_at, updated_at FROM orders WHERE number = $1"

	err := r.db(ctx).QueryRow(ctx, q, number).
		Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrOrderNotFound
		}
		return nil, err
	}

	return order, nil
}

func (r *OrderRepo) FindByUser(ctx context.Context, userID int64) ([]entity.Order, error) {
	q := `
		SELECT id, user_id, number, status, accrual, created_at, updated_at 
		FROM orders 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db(ctx).Query(ctx, q, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)

	}
	return orders, rows.Err()
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *int64) error {
	q := `
		UPDATE orders
		SET status = $2, accrual = $3 
		WHERE number = $1
	`
	_, err := r.db(ctx).Exec(ctx, q, number, status, accrual)
	return err
}

func (r *OrderRepo) FindPending(ctx context.Context) ([]entity.Order, error) {
	q := `
		SELECT	id, user_id, number, status
		FROM orders
		WHERE status = ANY($1)
	`
	rows, err := r.db(ctx).Query(ctx, q, entity.OrderPendingStatuses)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var o entity.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Number, &o.Status); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
