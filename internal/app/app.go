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
	"github.com/aikowocki/yandex-go-first-diploma/internal/config"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/auth"
	"github.com/aikowocki/yandex-go-first-diploma/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	server *http.Server
	worker *accrual.Worker
	pool   *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	pool, err := postgres.NewPool(ctx, cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := postgres.RunMigrations(cfg.DatabaseDSN); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to run migration: %w", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	txManager := postgres.NewTxManager(pool)

	userRepo := postgres.NewUserRepo(txManager)
	authUC := usecase.NewAuthUseCase(userRepo, jwtManager)
	authHandler := handler.NewAuthHandler(authUC)

	orderRepo := postgres.NewOrderRepo(txManager)
	orderUC := usecase.NewOrderUseCase(orderRepo)
	orderHandler := handler.NewOrderHandler(orderUC)

	balanceRepo := postgres.NewBalanceRepo(txManager)
	balanceUC := usecase.NewBalanceUseCase(balanceRepo, txManager)

	accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
	accrualWorker := accrual.NewWorker(accrualClient, orderRepo, balanceRepo, txManager)

	if cfg.CacheAddress != "" {
		redisCache := cache.NewRedisCache(cfg.CacheAddress)
		cachedBalanceRepo := cache.NewCacheBalanceRepo(balanceRepo, redisCache, 5*time.Minute)
		balanceUC = usecase.NewBalanceUseCase(cachedBalanceRepo, txManager)
		accrualWorker = accrual.NewWorker(accrualClient, orderRepo, cachedBalanceRepo, txManager)
	}

	balanceHandler := handler.NewBalanceHandler(balanceUC)
	healthHandler := handler.NewHealthHandler(pool)

	r := handler.NewRouter(authHandler, orderHandler, balanceHandler, healthHandler, jwtManager)

	zap.S().Infow(
		"Server starting",
		"address", cfg.ServerAddress,
	)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	return &App{server: srv, worker: accrualWorker, pool: pool}, nil
}

func (a *App) Run(ctx context.Context) {
	go a.worker.Run(ctx)

	zap.S().Infow("server starting", "address", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zap.S().Fatalw("server failed", "error", err)
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func (a *App) Close() {
	a.pool.Close()
}
