package middleware

import (
	"github.com/arigatosimarmata/artech-user/dto/response"
	"github.com/arigatosimarmata/artech-user/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ErrorHandlerMiddleware creates an error handling middleware
func ErrorHandlerMiddleware(log logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default error code
		code := fiber.StatusInternalServerError

		// Retrieve error code if it's a Fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		// Log error
		log.Error("Request error",
			zap.Error(err),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status_code", code),
		)

		// Send JSON response
		return c.Status(code).JSON(response.ErrorResponse{
			Success: false,
			Error:   "error",
			Message: err.Error(),
		})
	}
}
