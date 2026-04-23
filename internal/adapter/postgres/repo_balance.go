package postgres

import (
	"context"
	"errors"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type BalanceRepo struct {
	baseRepo
}

func NewBalanceRepo(txManager *TxManager) *BalanceRepo {
	return &BalanceRepo{baseRepo: baseRepo{txManager: txManager}}
}

func (r *BalanceRepo) LockByUserId(ctx context.Context, userID int64) error {
	_, err := r.db(ctx).Exec(ctx, "SELECT pg_advisory_xact_lock(1, $1)", userID)
	return err
}

func (r *BalanceRepo) GetBalance(ctx context.Context, userID int64) (entity.Balance, error) {
	q := `
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE type = $2), 0) as accrued,
			COALESCE(SUM(amount) FILTER (WHERE type = $3), 0) as withdrawn
		FROM transactions WHERE user_id = $1
	`
	var accrued, withdrawn int64
	if err := r.db(ctx).QueryRow(ctx, q, userID, entity.TransactionTypeAccrual, entity.TransactionTypeWithdrawal).Scan(&accrued, &withdrawn); err != nil {
		return entity.Balance{}, err
	}
	return entity.Balance{
		Current:   accrued - withdrawn,
		Withdrawn: withdrawn,
	}, nil
}

func (r *BalanceRepo) Withdraw(ctx context.Context, userID int64, orderNumber string, amount int64) error {
	q := `
		INSERT INTO transactions (user_id, order_number, type, amount)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db(ctx).Exec(ctx, q, userID, orderNumber, entity.TransactionTypeWithdrawal, amount)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrWithdrawalAlreadyExists
		}
		return err
	}
	return nil
}

func (r *BalanceRepo) GetWithdrawals(ctx context.Context, userID int64) ([]entity.Transaction, error) {
	q := `
		SELECT id, user_id, order_number, type, amount, created_at 
		FROM transactions
		WHERE user_id = $1 AND type = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db(ctx).Query(ctx, q, userID, entity.TransactionTypeWithdrawal)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var withdrawals []entity.Transaction
	for rows.Next() {
		var transaction entity.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.OrderNumber,
			&transaction.Type,
			&transaction.Amount,
			&transaction.CreatedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, transaction)

	}
	return withdrawals, rows.Err()
}

func (r *BalanceRepo) AddAccrual(ctx context.Context, userID int64, orderNumber string, amount int64) error {
	q := `
		INSERT INTO transactions (user_id, order_number, type, amount)
		values ($1, $2, $3, $4)
	`
	_, err := r.db(ctx).Exec(ctx, q, userID, orderNumber, entity.TransactionTypeAccrual, amount)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrAccrualAlreadyExists
		}
		return err
	}
	return nil
}
