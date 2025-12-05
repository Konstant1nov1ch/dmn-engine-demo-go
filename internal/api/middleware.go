package api

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("X-Request-ID", requestID)
		c.Locals("requestId", requestID)
		return c.Next()
	}
}

// LoggerMiddleware logs request information
func LoggerMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Skip logging for health checks
		path := c.Path()
		if path == "/health" || path == "/ready" {
			return err
		}

		logger.Info("request",
			"method", c.Method(),
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"request_id", c.Locals("requestId"),
			"tenant_id", c.Get("X-Tenant-ID"),
		)

		return err
	}
}

// TenantMiddleware extracts tenant ID from header
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantID := c.Get("X-Tenant-ID")
		c.Locals("tenantId", tenantID)
		return c.Next()
	}
}
