package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/handler"
	"github.com/aikowocki/yandex-go-first-diploma/internal/adapter/postgres"
	"github.com/aikowocki/yandex-go-first-diploma/internal/config"
	"github.com/aikowocki/yandex-go-first-diploma/internal/logger"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/auth"
	"github.com/aikowocki/yandex-go-first-diploma/internal/usecase"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	if err := godotenv.Load(); err != nil {
		log.Println("failed to load .env", "error", err)
	}

	loggerCleanup, err := logger.New()

	if err != nil {
		log.Fatal(err)
	}

	defer loggerCleanup()

	cfg, err := config.New()

	if err != nil {
		zap.S().Fatalw("failed to load config", "error", err)
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseDSN)
	if err != nil {
		zap.S().Fatalw("failed to connect to database", "error", err)
	}
	defer pool.Close()

	if err := postgres.RunMigrations(cfg.DatabaseDSN); err != nil {
		zap.S().Fatalw("failed to run migrations", "error", err)
	}

	zap.S().Infow(
		"Server starting",
		"address", cfg.ServerAddress,
	)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	userRepo := postgres.NewUserRepo(pool)
	orderRepo := postgres.NewOrderRepo(pool)
	authUC := usecase.NewAuthUseCase(userRepo, jwtManager)
	orderUC := usecase.NewOrderUseCase(orderRepo)
	authHandler := handler.NewAuthHandler(authUC)
	orderHandler := handler.NewOrderHandler(orderUC)

	r := handler.NewRouter(authHandler, orderHandler, jwtManager)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.S().Fatalw(err.Error(), "event", "start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	zap.S().Infow("shutting down...")
	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err = srv.Shutdown(shutdownContext); err != nil {
		zap.S().Fatalw(err.Error(), "event", "Shutdown server error")
	}
	cancel()
	zap.S().Infow("server stopped")
}
