// Package realtime serves the collaborative editing WebSocket (`/ws/:project`)
// from inside the API process, so one server handles both REST and realtime.
package realtime

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"my-go-app/internal/realtime/config"
	"my-go-app/internal/realtime/delivery/http/middleware"
	"my-go-app/internal/realtime/delivery/ws/handlers"
	"my-go-app/internal/realtime/pkg/database"
	"my-go-app/internal/realtime/pkg/redis"
	"my-go-app/internal/realtime/repository"
	"my-go-app/internal/realtime/usecase"
	ulogger "my-go-app/internal/realtime/utils/logger"
)

// Mount connects the realtime engine's own PostgreSQL and Redis clients, starts
// its persistence worker and registers the /ws routes on app. It returns a
// function that stops the worker, flushes pending pages and closes Redis.
//
// The realtime engine keeps its own *gorm.DB: it shares tables with the API but
// maps columns with a different naming strategy, so the two must not share one.
func Mount(ctx context.Context, app *fiber.App) (shutdown func(), err error) {
	lg := ulogger.Default()

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("realtime config: %w", err)
	}

	db, err := database.NewPostgresConnection(database.Config{
		URL:      cfg.Database.URL,
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		return nil, fmt.Errorf("realtime database: %w", err)
	}

	rdb, err := redis.NewRedisClient(redis.Config{
		URL:      cfg.Redis.URL,
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		closeDB(db)
		return nil, fmt.Errorf("realtime redis: %w", err)
	}

	useCases := usecase.NewUseCase(repository.NewRepository(db, rdb), lg)

	workerCtx, stopWorker := context.WithCancel(ctx)
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		useCases.Persist.Start(workerCtx)
	}()

	wsHandler := handlers.NewHandler(useCases, rdb, lg)
	// Dev mode skips JWT checks (every client is DEV_USER_ID), so project access
	// is skipped too, unless REALTIME_ENFORCE_ACCESS=true.
	if !cfg.DevMode || os.Getenv("REALTIME_ENFORCE_ACCESS") == "true" {
		wsHandler.RequireAccess(useCases.Logic)
	}
	app.Use("/ws", middleware.AuthMiddleware(cfg))
	wsHandler.SetupRoutes(app)

	return func() {
		stopWorker()
		<-workerDone
		flushCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := useCases.Persist.Flush(flushCtx); err != nil {
			lg.Errorf("realtime: final flush failed: %v", err)
		}
		closeRedis(rdb)
		closeDB(db)
	}, nil
}

func closeRedis(rdb *goredis.Client) { _ = rdb.Close() }

func closeDB(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
