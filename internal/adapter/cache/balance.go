package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"go.uber.org/zap"
)

type BalanceRepo interface {
	LockByUserId(ctx context.Context, userID int64) error
	GetBalance(ctx context.Context, userID int64) (entity.Balance, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, amount int64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]entity.Transaction, error)
	AddAccrual(ctx context.Context, userID int64, orderNumber string, amount int64) error
}

type CachedBalanceRepo struct {
	db    BalanceRepo
	cache Cache
	ttl   time.Duration
}

func NewCacheBalanceRepo(db BalanceRepo, cache Cache, ttl time.Duration) *CachedBalanceRepo {
	return &CachedBalanceRepo{db: db, cache: cache, ttl: ttl}
}

func balanceKey(userID int64) string {
	return fmt.Sprintf("balance:%d", userID)
}

func (r *CachedBalanceRepo) GetBalance(ctx context.Context, userID int64) (entity.Balance, error) {
	data, err := r.cache.Get(ctx, balanceKey(userID))
	if err == nil {
		var cachedBalance entity.Balance
		if json.Unmarshal(data, &cachedBalance) == nil {
			return cachedBalance, nil
		}
	}

	balance, err := r.db.GetBalance(ctx, userID)
	if err != nil {
		return balance, err
	}

	if data, err := json.Marshal(balance); err == nil {
		if err := r.cache.Set(ctx, balanceKey(userID), data, r.ttl); err != nil {
			zap.S().Warnw("failed to set balance cache", "error", err, "userID", userID)
		}
	}
	return balance, nil
}

func (r *CachedBalanceRepo) invalidate(ctx context.Context, userID int64) {
	if err := r.cache.Del(ctx, balanceKey(userID)); err != nil {
		zap.S().Warnw("failed to invalidate balance cache", "error", err, "userID", userID)
	}
}

func (r *CachedBalanceRepo) LockByUserId(ctx context.Context, userID int64) error {
	r.invalidate(ctx, userID)
	return r.db.LockByUserId(ctx, userID)
}

func (r *CachedBalanceRepo) Withdraw(ctx context.Context, userID int64, orderNumber string, amount int64) error {
	err := r.db.Withdraw(ctx, userID, orderNumber, amount)
	if err == nil {
		r.invalidate(ctx, userID)
	}
	return err
}

func (r *CachedBalanceRepo) GetWithdrawals(ctx context.Context, userID int64) ([]entity.Transaction, error) {
	return r.db.GetWithdrawals(ctx, userID)
}

func (r *CachedBalanceRepo) AddAccrual(ctx context.Context, userID int64, orderNumber string, amount int64) error {
	err := r.db.AddAccrual(ctx, userID, orderNumber, amount)
	if err == nil {
		r.invalidate(ctx, userID)
	}
	return err
}
