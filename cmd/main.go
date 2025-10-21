package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edwaldo/test_blue_horn_tech/backend/internal/app"
	"github.com/edwaldo/test_blue_horn_tech/backend/internal/config"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	ctx := context.Background()
	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: application.Router,
	}

	go func() {
		application.Logger.Info("server online",
			zap.String("address", addr),
			zap.String("mode", cfg.App.Env),
			zap.String("app", cfg.App.Name),
			zap.String("db", fmt.Sprintf("postgres://%s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)),
		)
		fmt.Printf("🚀 %s (%s) ready at http://%s\n", cfg.App.Name, cfg.App.Env, addr)
		fmt.Printf("🗄️  connected to PostgreSQL %s@%s:%d/%s\n", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
		fmt.Println("📘 API docs available at /docs")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			application.Logger.Fatal("http server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.Logger.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		application.Logger.Error("server shutdown failed", zap.Error(err))
	}

	if err := application.Shutdown(shutdownCtx); err != nil {
		application.Logger.Error("failed to shutdown application", zap.Error(err))
	}
}
