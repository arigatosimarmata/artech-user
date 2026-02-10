package middleware

import (
	"strings"

	"github.com/arigatosimarmata/artech-user/domain"
	"github.com/arigatosimarmata/artech-user/dto/response"
	"github.com/arigatosimarmata/artech-user/pkg/token"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(jwtManager *token.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponse{
				Success: false,
				Error:   "unauthorized",
				Message: "Missing authorization header",
			})
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponse{
				Success: false,
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponse{
				Success: false,
				Error:   "unauthorized",
				Message: "Missing token",
			})
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			statusCode := fiber.StatusUnauthorized
			errorCode := "invalid_token"
			message := err.Error()

			switch err {
			case token.ErrExpiredToken:
				errorCode = "token_expired"
			case token.ErrInvalidToken:
				errorCode = "token_invalid"
			case domain.ErrTokenExpired:
				errorCode = "token_expired"
			case domain.ErrTokenInvalid:
				errorCode = "token_invalid"
			}

			return c.Status(statusCode).JSON(response.ErrorResponse{
				Success: false,
				Error:   errorCode,
				Message: message,
			})
		}

		// Set user information in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)

		return c.Next()
	}
}
