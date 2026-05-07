package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"github.com/aikowocki/yandex-go-first-diploma/internal/port"
)

func benchCreateOrder(b *testing.B, storage port.Storage) {
	ctx := context.Background()
	user := &entity.User{Login: b.Name(), PasswordHash: "hash"}
	if err := storage.UserRepo().Create(ctx, user); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.OrderRepo().Create(ctx, &entity.Order{
			UserID: user.ID,
			Number: fmt.Sprintf("%012d", i),
		})
	}
}

func benchFindByNumber(b *testing.B, storage port.Storage) {
	ctx := context.Background()
	user := &entity.User{Login: b.Name(), PasswordHash: "hash"}
	if err := storage.UserRepo().Create(ctx, user); err != nil {
		b.Fatal(err)
	}
	_ = storage.OrderRepo().Create(ctx, &entity.Order{
		UserID: user.ID,
		Number: "9999999999",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.OrderRepo().FindByNumber(ctx, "9999999999")
	}
}

func benchGetBalance(b *testing.B, storage port.Storage) {
	ctx := context.Background()
	user := &entity.User{Login: b.Name(), PasswordHash: "hash"}
	if err := storage.UserRepo().Create(ctx, user); err != nil {
		b.Fatal(err)
	}
	// Создаём несколько транзакций для реалистичности
	for i := 0; i < 10; i++ {
		_ = storage.BalanceRepo().AddAccrual(ctx, user.ID, fmt.Sprintf("%012d", i), 100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.BalanceRepo().GetBalance(ctx, user.ID)
	}
}

func BenchmarkCreateOrder_PGX(b *testing.B) {
	benchCreateOrder(b, setupPGXStorage(b))
}

func BenchmarkCreateOrder_GORM(b *testing.B) {
	benchCreateOrder(b, setupGORMStorage(b))
}

func BenchmarkFindByNumber_PGX(b *testing.B) {
	benchFindByNumber(b, setupPGXStorage(b))
}

func BenchmarkFindByNumber_GORM(b *testing.B) {
	benchFindByNumber(b, setupGORMStorage(b))
}

func BenchmarkGetBalance_PGX(b *testing.B) {
	benchGetBalance(b, setupPGXStorage(b))
}

func BenchmarkGetBalance_GORM(b *testing.B) {
	benchGetBalance(b, setupGORMStorage(b))
}

func benchWithdraw(b *testing.B, storage port.Storage) {
	ctx := context.Background()
	user := &entity.User{Login: b.Name(), PasswordHash: "hash"}
	if err := storage.UserRepo().Create(ctx, user); err != nil {
		b.Fatal(err)
	}
	// Начисляем достаточно для всех итераций
	_ = storage.BalanceRepo().AddAccrual(ctx, user.ID, "0000000001", int64(b.N*100+10000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.TxManager().Do(ctx, func(ctx context.Context) error {
			if err := storage.BalanceRepo().LockByUserId(ctx, user.ID); err != nil {
				return err
			}
			return storage.BalanceRepo().Withdraw(ctx, user.ID, fmt.Sprintf("%012d", i), 1)
		})
	}
}

func BenchmarkWithdraw_PGX(b *testing.B) {
	benchWithdraw(b, setupPGXStorage(b))
}

func BenchmarkWithdraw_GORM(b *testing.B) {
	benchWithdraw(b, setupGORMStorage(b))
}
