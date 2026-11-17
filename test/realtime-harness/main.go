// Command realtime-harness is a loopback-only test server. It uses the API's
// realtime repositories and handlers, with synthetic identities for tests.
// It must never be used as an authenticated production server.
package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"my-go-app/internal/realtime/delivery/ws/handlers"
	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/pkg/database"
	rtredis "my-go-app/internal/realtime/pkg/redis"
	"my-go-app/internal/realtime/repository"
	"my-go-app/internal/realtime/usecase"
	"my-go-app/internal/realtime/utils/logger"
)

func main() {
	dsn, addr := os.Getenv("WEBBUILDER_TEST_DATABASE_URL"), os.Getenv("WEBBUILDER_TEST_REDIS_ADDR")
	if dsn == "" || addr == "" {
		log.Fatal("test database and Redis addresses are required")
	}
	db, err := database.NewPostgresConnection(database.Config{URL: dsn})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}()
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal(err)
	}
	rdb, err := rtredis.NewRedisClient(rtredis.Config{Host: host, Port: port})
	if err != nil {
		log.Fatal(err)
	}
	defer rdb.Close()
	lg := logger.Default()
	uc := usecase.NewUseCase(repository.NewRepository(db, rdb), lg)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); uc.Persist.Start(ctx) }()
	defer func() {
		cancel()
		<-done
		flush, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_ = uc.Persist.Flush(flush)
	}()
	h := handlers.NewHandler(uc, rdb, lg)
	defer h.Shutdown()
	app := fiber.New()
	app.Use("/ws", func(c fiber.Ctx) error {
		id := c.Query("token")
		if id == "" {
			return fiber.ErrUnauthorized
		}
		c.Locals("user", &entity.User{ID: id, Email: id + "@test.invalid"})
		return c.Next()
	})
	h.SetupRoutes(app)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("WEBBUILDER_TEST_LISTEN_FILE"), []byte(listener.Addr().String()), 0600); err != nil {
		log.Fatal(err)
	}
	log.Fatal(app.Listener(listener, fiber.ListenConfig{DisableStartupMessage: true}))
}
