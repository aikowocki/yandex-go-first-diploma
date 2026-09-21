package integration

import (
	"context"
	"testing"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testOrderRepo(t *testing.T, storage port.Storage) {
	userRepo := storage.UserRepo()
	orderRepo := storage.OrderRepo()
	ctx := context.Background()

	user := &entity.User{Login: "test", PasswordHash: "hash"}
	require.NoError(t, userRepo.Create(ctx, user))

	t.Run("CreateAndFind", func(t *testing.T) {
		order := &entity.Order{UserID: user.ID, Number: "7992739713"}
		err := orderRepo.Create(ctx, order)
		require.NoError(t, err)

		found, err := orderRepo.FindByNumber(ctx, order.Number)
		require.NoError(t, err)
		assert.Equal(t, order.Number, found.Number)
		assert.Equal(t, order.UserID, found.UserID)

		userOrders, err := orderRepo.FindByUser(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, userOrders, 1)
		assert.Equal(t, order.Number, userOrders[0].Number)
	})

	t.Run("Duplicate", func(t *testing.T) {
		order := &entity.Order{UserID: user.ID, Number: "7992739713"}
		err := orderRepo.Create(ctx, order)
		assert.ErrorIs(t, err, entity.ErrOrderExists)
	})

	t.Run("FindByNumber_NotFound", func(t *testing.T) {
		_, err := orderRepo.FindByNumber(ctx, "0000000000")
		assert.ErrorIs(t, err, entity.ErrOrderNotFound)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		accrual := int64(500)
		err := orderRepo.UpdateStatus(ctx, "7992739713", entity.OrderStatusProcessed, &accrual)
		require.NoError(t, err)

		found, err := orderRepo.FindByNumber(ctx, "7992739713")
		require.NoError(t, err)
		assert.Equal(t, entity.OrderStatusProcessed, found.Status)
		assert.Equal(t, &accrual, found.Accrual)
	})

	t.Run("FindPending", func(t *testing.T) {
		// Создаём ещё два заказа (будут NEW)
		_ = orderRepo.Create(ctx, &entity.Order{UserID: user.ID, Number: "49927398716"})
		_ = orderRepo.Create(ctx, &entity.Order{UserID: user.ID, Number: "1234567812345670"})

		pending, err := orderRepo.FindPending(ctx)
		require.NoError(t, err)
		// 7992739713 = PROCESSED, два новых = NEW
		assert.Len(t, pending, 2)
	})
}

func TestOrderRepo(t *testing.T) {
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
			testOrderRepo(t, storage)
		})
	}
}
