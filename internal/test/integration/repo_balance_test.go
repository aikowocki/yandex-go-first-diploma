package integration

import (
	"context"
	"testing"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBalanceRepo_WithdrawAndGetBalance(t *testing.T, storage port.Storage) {
	userRepo := storage.UserRepo()
	balanceRepo := storage.BalanceRepo()
	ctx := context.Background()

	// Создаем юзера
	user := &entity.User{Login: t.Name(), PasswordHash: "hash"}
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
	err = storage.TxManager().Do(ctx, func(ctx context.Context) error {
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

func testBalanceRepo_DuplicateAccrual(t *testing.T, storage port.Storage) {
	userRepo := storage.UserRepo()
	balanceRepo := storage.BalanceRepo()
	ctx := context.Background()

	user := &entity.User{Login: t.Name(), PasswordHash: "hash"}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	err = balanceRepo.AddAccrual(ctx, user.ID, "79927398713", 500)
	require.NoError(t, err)

	err = balanceRepo.AddAccrual(ctx, user.ID, "79927398713", 500)
	assert.ErrorIs(t, err, entity.ErrAccrualAlreadyExists)
}

func testBalanceRepo_DuplicateWithdrawal(t *testing.T, storage port.Storage) {
	userRepo := storage.UserRepo()
	balanceRepo := storage.BalanceRepo()
	ctx := context.Background()

	user := &entity.User{Login: t.Name(), PasswordHash: "hash"}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	err = storage.TxManager().Do(ctx, func(ctx context.Context) error {
		return balanceRepo.Withdraw(ctx, user.ID, "4992398716", 200)
	})
	require.NoError(t, err)

	err = storage.TxManager().Do(ctx, func(ctx context.Context) error {
		return balanceRepo.Withdraw(ctx, user.ID, "4992398716", 200)
	})
	assert.ErrorIs(t, err, entity.ErrWithdrawalAlreadyExists)
}

func TestBalanceRepo(t *testing.T) {
	drivers := []struct {
		name    string
		storage func(tb testing.TB) port.Storage
	}{
		{"pgx", setupPGXStorage},
		{"gorm", setupGORMStorage},
	}
	for _, d := range drivers {
		t.Run(d.name, func(t *testing.T) {
			storage := d.storage(t)
			t.Run("WithdrawAndGetBalance", func(t *testing.T) {
				testBalanceRepo_WithdrawAndGetBalance(t, storage)
			})
			t.Run("DuplicateAccrual", func(t *testing.T) {
				testBalanceRepo_DuplicateAccrual(t, storage)
			})
			t.Run("DuplicateWithdrawal", func(t *testing.T) {
				testBalanceRepo_DuplicateWithdrawal(t, storage)
			})
		})
	}
}
