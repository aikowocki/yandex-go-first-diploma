package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/accrual"
	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/cache"
	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/handler"
	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/postgres"
	postgres_gorm "github.com/aikowocki/yandex-go-first-diploma/internal/adapter/postgres/gorm"
	"github.com/aikowocki/yandex-go-first-diploma/internal/config"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/auth"
	"github.com/aikowocki/yandex-go-first-diploma/internal/port"
	"github.com/aikowocki/yandex-go-first-diploma/internal/usecase"
	"go.uber.org/zap"
)

type App struct {
	server     *http.Server
	workerPool *accrual.WorkerPool
	db         port.DB
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	var pgStorage port.Storage
	var err error

	switch cfg.PostgresDriver {
	case "gorm":
		pgStorage, err = postgres_gorm.NewStorage(cfg.DatabaseDSN)
		zap.S().Infow("Used gorm driver")
	default:
		zap.S().Infow("Used pgx driver")
		pgStorage, err = postgres.NewStorage(ctx, cfg.DatabaseDSN)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := postgres.RunMigrations(cfg.DatabaseDSN); err != nil {
		pgStorage.DB().Close()
		return nil, fmt.Errorf("failed to run migration: %w", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	authUC := usecase.NewAuthUseCase(pgStorage.UserRepo(), jwtManager)
	authHandler := handler.NewAuthHandler(authUC)

	orderUC := usecase.NewOrderUseCase(pgStorage.OrderRepo())
	orderHandler := handler.NewOrderHandler(orderUC)

	balanceUC := usecase.NewBalanceUseCase(pgStorage.BalanceRepo(), pgStorage.TxManager())

	accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
	accrualUC := usecase.NewAccrualUseCase(accrualClient, pgStorage.OrderRepo(), pgStorage.BalanceRepo(), pgStorage.TxManager())

	if cfg.CacheAddress != "" {
		redisCache := cache.NewRedisCache(cfg.CacheAddress)
		cachedBalanceRepo := cache.NewCacheBalanceRepo(pgStorage.BalanceRepo(), redisCache, 5*time.Minute)
		balanceUC = usecase.NewBalanceUseCase(cachedBalanceRepo, pgStorage.TxManager())
		accrualUC = usecase.NewAccrualUseCase(accrualClient, pgStorage.OrderRepo(), cachedBalanceRepo, pgStorage.TxManager())
	}

	accrualWorkerPool := accrual.NewWorkerPool(accrualUC, 3)
	balanceHandler := handler.NewBalanceHandler(balanceUC)
	healthHandler := handler.NewHealthHandler(pgStorage.DB())

	r := handler.NewRouter(authHandler, orderHandler, balanceHandler, healthHandler, jwtManager)

	zap.S().Infow(
		"Server starting",
		"address", cfg.ServerAddress,
	)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	return &App{server: srv, workerPool: accrualWorkerPool, db: pgStorage.DB()}, nil
}

func (a *App) Run(ctx context.Context) {
	go a.workerPool.Run(ctx)

	zap.S().Infow("server starting", "address", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zap.S().Fatalw("server failed", "error", err)
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func (a *App) Close() {
	a.db.Close()
}
