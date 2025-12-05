package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/yourusername/dmn-engine-go/internal/api"
	"github.com/yourusername/dmn-engine-go/internal/config"
	"github.com/yourusername/dmn-engine-go/internal/storage"
)

func main() {
	// Config
	cfg := config.Load()

	// Logger
	logLevel := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	logger.Info("starting DMN Engine",
		"port", cfg.Server.Port,
		"db_host", cfg.Database.Host,
		"db_name", cfg.Database.Name,
		"db_max_conns", cfg.Database.MaxConns,
	)

	// Database
	ctx := context.Background()
	pool, err := storage.NewPool(ctx, &cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("connected to database")

	// Run migrations
	if err := storage.RunMigrations(ctx, pool); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations completed")

	// Repository
	repo := storage.NewPostgresRepository(pool)

	// HTTP Server
	app := fiber.New(fiber.Config{
		AppName:               "DMN Engine Go",
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		DisableStartupMessage: false,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, X-Tenant-ID, X-Request-ID",
	}))
	app.Use(api.RequestIDMiddleware())
	app.Use(api.LoggerMiddleware(logger))

	// Routes
	handler := api.NewHandler(repo, logger)
	api.SetupRoutes(app, handler)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.Server.Port); err != nil {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("server started", "port", cfg.Server.Port, "url", "http://localhost:"+cfg.Server.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	logger.Info("server stopped")
}
