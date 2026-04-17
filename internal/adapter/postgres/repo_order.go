package postgres

import (
	"context"
	"errors"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	baseRepo
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{baseRepo: baseRepo{pool: pool}}
}

func (r *OrderRepo) Create(ctx context.Context, order *entity.Order) error {
	q := "INSERT INTO orders (user_id, number) VALUES ($1, $2) RETURNING id, status, created_at, updated_at"
	err := r.pool.QueryRow(ctx, q, order.UserID, order.Number).
		Scan(&order.ID, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrOrderExists
		}

		return err
	}

	return nil
}

func (r *OrderRepo) FindByNumber(ctx context.Context, number string) (*entity.Order, error) {
	order := &entity.Order{}
	q := "SELECT id, user_id, number, status, accrual, created_at, updated_at FROM orders WHERE number = $1"

	err := r.pool.QueryRow(ctx, q, number).
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

	rows, err := r.pool.Query(ctx, q, userID)

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
