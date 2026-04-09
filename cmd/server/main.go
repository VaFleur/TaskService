package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "taskService/docs"
	"taskService/internal/config"
	"taskService/internal/handler"
	"taskService/internal/repository/postgres"
	"taskService/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Task Service API
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// PostgreSQL Pool
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to create db pool: %v", err)
	}
	defer pool.Close()
	pool.Config().MaxConns = int32(cfg.Database.MaxOpenConns)
	pool.Config().MinConns = int32(cfg.Database.MaxIdleConns)

	// Redis Client
	opts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		log.Fatalf("Invalid REDIS_URL: %v", err)
	}
	rdb := redis.NewClient(opts)
	defer rdb.Close()

	// Redis checkup
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Redis ping failed: %v", err)
	} else {
		log.Println("✅ Connected to Redis")
	}

	// Layer init
	repo := postgres.NewTaskRepo(pool)
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30)) // chi timeout в секундах

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Mount("/tasks", h.Routes())
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("failed to write health check response: %v", err)
		}
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Printf("🚀 Server starting on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⏳ Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("✅ Server exited properly")
}
