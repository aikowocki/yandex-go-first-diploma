package postgres

import (
	"context"
	"testing"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceRepo_WithdrawAndGetBalance(t *testing.T) {
	txm := setupTestDB(t)
	userRepo := NewUserRepo(txm)
	balanceRepo := NewBalanceRepo(txm)
	ctx := context.Background()

	// Создаем юзера
	user := &entity.User{Login: "test", PasswordHash: "hash"}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Начисляем
	err = balanceRepo.AddAccrual(ctx, user.ID, "7992739713", 500)
	require.NoError(t, err)

	// Проверяем баланс
	balance, err := balanceRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(500), balance.Current)
	assert.Equal(t, int64(0), balance.Withdrawn)

	// Списываем
	err = txm.Do(ctx, func(ctx context.Context) error {
		if err := balanceRepo.LockByUserId(ctx, user.ID); err != nil {
			return err
		}
		return balanceRepo.Withdraw(ctx, user.ID, "4992398716", 200)
	})
	require.NoError(t, err)

	// Проверка

	balance, err = balanceRepo.GetBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(300), balance.Current)
	assert.Equal(t, int64(200), balance.Withdrawn)
}

func TestBalanceRepo_DuplicateAccrual(t *testing.T) {
	txm := setupTestDB(t)
	userRepo := NewUserRepo(txm)
	balanceRepo := NewBalanceRepo(txm)
	ctx := context.Background()

	user := &entity.User{Login: "test", PasswordHash: "hash"}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	err = balanceRepo.AddAccrual(ctx, user.ID, "79927398713", 500)
	require.NoError(t, err)

	err = balanceRepo.AddAccrual(ctx, user.ID, "79927398713", 500)
	assert.ErrorIs(t, err, entity.ErrAccrualAlreadyExists)
}

func TestBalanceRepo_DuplicateWithdrawal(t *testing.T) {
	txm := setupTestDB(t)
	userRepo := NewUserRepo(txm)
	balanceRepo := NewBalanceRepo(txm)
	ctx := context.Background()

	user := &entity.User{Login: "test", PasswordHash: "hash"}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	err = txm.Do(ctx, func(ctx context.Context) error {
		return balanceRepo.Withdraw(ctx, user.ID, "4992398716", 200)
	})
	require.NoError(t, err)

	err = txm.Do(ctx, func(ctx context.Context) error {
		return balanceRepo.Withdraw(ctx, user.ID, "4992398716", 200)
	})
	assert.ErrorIs(t, err, entity.ErrWithdrawalAlreadyExists)
}
