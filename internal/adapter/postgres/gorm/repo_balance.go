package postgres_gorm

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
	return r.db(ctx).Exec("SELECT pg_advisory_xact_lock(1, ?)", userID).Error
}

func (r *BalanceRepo) GetBalance(ctx context.Context, userID int64) (entity.Balance, error) {
	q := `
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE type = ?), 0) as accrued,
			COALESCE(SUM(amount) FILTER (WHERE type = ?), 0) as withdrawn
		FROM transactions WHERE user_id = ?
	`

	var result struct {
		Accrued   int64
		Withdrawn int64
	}
	if err := r.db(ctx).Raw(q, entity.TransactionTypeAccrual, entity.TransactionTypeWithdrawal, userID).
		Scan(&result).Error; err != nil {
		return entity.Balance{}, err
	}
	return entity.Balance{
		Current:   result.Accrued - result.Withdrawn,
		Withdrawn: result.Withdrawn,
	}, nil
}

func (r *BalanceRepo) Withdraw(ctx context.Context, userID int64, orderNumber string, amount int64) error {

	if err := r.db(ctx).Create(&entity.Transaction{
		UserID:      userID,
		OrderNumber: orderNumber,
		Type:        entity.TransactionTypeWithdrawal,
		Amount:      amount,
	}).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrWithdrawalAlreadyExists
		}
		return err
	}
	return nil
}

func (r *BalanceRepo) GetWithdrawals(ctx context.Context, userID int64) ([]entity.Transaction, error) {
	var withdrawals []entity.Transaction

	err := r.db(ctx).Where("user_id = ? AND type = ?", userID, entity.TransactionTypeWithdrawal).
		Order("created_at DESC").
		Find(&withdrawals).Error

	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

func (r *BalanceRepo) AddAccrual(ctx context.Context, userID int64, orderNumber string, amount int64) error {
	if err := r.db(ctx).Create(&entity.Transaction{
		UserID:      userID,
		OrderNumber: orderNumber,
		Type:        entity.TransactionTypeAccrual,
		Amount:      amount,
	}).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.ErrAccrualAlreadyExists
		}
		return err
	}
	return nil
}
