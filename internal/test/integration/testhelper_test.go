package integration

import (
	"context"
	"testing"

	pgadapter "github.com/aikowocki/yandex-go-first-diploma/internal/adapter/postgres"
	postgres_gorm "github.com/aikowocki/yandex-go-first-diploma/internal/adapter/postgres/gorm"
	"github.com/aikowocki/yandex-go-first-diploma/internal/port"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(tb testing.TB) string {
	tb.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		tb.Fatal(err)
	}

	if err := pgadapter.RunMigrations(dsn); err != nil {
		tb.Fatal(err)
	}
	return dsn
}

func setupPGXStorage(tb testing.TB) port.Storage {
	dsn := setupTestDB(tb)
	storage, err := pgadapter.NewStorage(context.Background(), dsn)

	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		storage.DB().Close()
	})
	return storage
}

func setupGORMStorage(tb testing.TB) port.Storage {
	dsn := setupTestDB(tb)
	storage, err := postgres_gorm.NewStorage(dsn)
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		storage.DB().Close()
	})
	return storage
}
