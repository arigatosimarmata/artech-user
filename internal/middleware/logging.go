package middleware

import (
	"time"

	"github.com/arigatosimarmata/artech-user/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// LoggingMiddleware creates a request logging middleware
func LoggingMiddleware(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request
		log.Info("HTTP Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.String("user_agent", c.Get("User-Agent")),
		)

		return err
	}
}
