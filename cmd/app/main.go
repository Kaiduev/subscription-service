package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	_ "subscription/docs"
	"subscription/internal/httpapi"
	"subscription/internal/repository"
	"subscription/pkg/db"
	"subscription/pkg/logger"
)

// @title Subscription API
// @version 1.0
// @description REST-сервис для агрегации онлайн-подписок
// @BasePath /
func main() {
	_ = godotenv.Load()

	log := logger.New(getEnv("LOG_LEVEL", "info"))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	log.Info("db connected")

	health := httpapi.NewHealth()
	repo := repository.NewSubscriptionRepository(pool)
	subHandler := httpapi.NewSubscriptionHandler(log, repo)

	router := httpapi.NewRouter(health, subHandler)

	srv := &http.Server{
		Addr:    ":" + getEnv("PORT", "8080"),
		Handler: router,
	}

	go func() {
		log.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", "err", err)
	}
	log.Info("server exited")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
