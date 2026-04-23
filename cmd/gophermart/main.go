package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aikowocki/yandex-go-first-diploma/internal/app"
	"github.com/aikowocki/yandex-go-first-diploma/internal/config"
	"github.com/aikowocki/yandex-go-first-diploma/internal/logger"
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

	application, err := app.New(ctx, cfg)
	if err != nil {
		zap.S().Fatalw("failed to init app", "error", err)
	}
	defer application.Close()
	go application.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	zap.S().Infow("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err = application.Shutdown(shutdownCtx); err != nil {
		zap.S().Fatalw("shutdown server error", "error", err)
	}
	cancel()
	zap.S().Infow("server stopped")
}
