package usecase

import (
	"context"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
)

type BalanceRepository interface {
	LockByUserId(ctx context.Context, userID int64) error
	GetBalance(ctx context.Context, userID int64) (entity.Balance, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, amount int64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]entity.Transaction, error)
}

type BalanceUseCase struct {
	repo      BalanceRepository
	txManager TxManager
}

func NewBalanceUseCase(repo BalanceRepository, txManager TxManager) *BalanceUseCase {
	return &BalanceUseCase{repo: repo, txManager: txManager}
}

func (uc *BalanceUseCase) Withdraw(ctx context.Context, userID int64, order string, amount int64) error {
	if !entity.ValidateLuhn(order) {
		return entity.ErrOrderNumberNotValid
	}

	return uc.txManager.Do(ctx, func(ctx context.Context) error {
		// Lock - что бы два паралельных списания не проходили одновременно
		if err := uc.repo.LockByUserId(ctx, userID); err != nil {
			return err
		}

		balance, err := uc.repo.GetBalance(ctx, userID)
		if err != nil {
			return err
		}

		if balance.Current < amount {
			return entity.ErrBalanceInsufficientFunds
		}

		return uc.repo.Withdraw(ctx, userID, order, amount)
	})
}

func (uc *BalanceUseCase) GetBalance(ctx context.Context, userID int64) (entity.Balance, error) {
	return uc.repo.GetBalance(ctx, userID)
}

func (uc *BalanceUseCase) GetWithdrawals(ctx context.Context, userID int64) ([]entity.Transaction, error) {
	return uc.repo.GetWithdrawals(ctx, userID)
}
