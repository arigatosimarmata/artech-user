package handler

import (
	"strings"

	"github.com/arigatosimarmata/artech-user/domain"
	"github.com/arigatosimarmata/artech-user/dto/request"
	"github.com/arigatosimarmata/artech-user/dto/response"
	"github.com/arigatosimarmata/artech-user/pkg/logger"
	"github.com/arigatosimarmata/artech-user/usecase"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userUsecase usecase.UserUsecase
	logger      logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUsecase usecase.UserUsecase, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
		logger:      logger,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RegisterRequest true "Registration request"
// @Success 201 {object} response.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req request.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	authResp, err := h.userUsecase.Register(c.Context(), &req, ipAddress, userAgent)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    authResp,
	})
}

// Login handles user login
// @Summary Login
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "Login request"
// @Success 200 {object} response.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req request.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	authResp, err := h.userUsecase.Login(c.Context(), &req, ipAddress, userAgent)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Login successful",
		Data:    authResp,
	})
}

// Logout handles user logout
// @Summary Logout
// @Description Logout user and revoke tokens
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *UserHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	
	// Extract access token from header
	authHeader := c.Get("Authorization")
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	if err := h.userUsecase.Logout(c.Context(), userID, accessToken, ipAddress, userAgent); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Logout successful",
	})
}

// RefreshToken handles token refresh
// @Summary Refresh Token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} response.AuthResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *UserHandler) RefreshToken(c *fiber.Ctx) error {
	var req request.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
	}

	authResp, err := h.userUsecase.RefreshToken(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Data:    authResp,
	})
}

// ForgotPassword handles forgot password request
// @Summary Forgot Password
// @Description Request password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var req request.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
	}

	if err := h.userUsecase.ForgotPassword(c.Context(), &req); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset
// @Summary Reset Password
// @Description Reset password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/reset-password [post]
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	var req request.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	if err := h.userUsecase.ResetPassword(c.Context(), &req, ipAddress, userAgent); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}

// GetProfile handles get user profile
// @Summary Get Profile
// @Description Get current user profile
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	userResp, err := h.userUsecase.GetProfile(c.Context(), userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Data:    userResp,
	})
}

// handleError handles errors and returns appropriate HTTP responses
func (h *UserHandler) handleError(c *fiber.Ctx, err error) error {
	h.logger.Error("handler error", zap.Error(err))

	statusCode := fiber.StatusInternalServerError
	errorCode := "internal_server_error"
	message := err.Error()

	switch err {
	case domain.ErrInvalidCredentials:
		statusCode = fiber.StatusUnauthorized
		errorCode = "invalid_credentials"
	case domain.ErrUnauthorized:
		statusCode = fiber.StatusUnauthorized
		errorCode = "unauthorized"
	case domain.ErrTokenExpired:
		statusCode = fiber.StatusUnauthorized
		errorCode = "token_expired"
	case domain.ErrTokenInvalid:
		statusCode = fiber.StatusUnauthorized
		errorCode = "token_invalid"
	case domain.ErrTokenRevoked:
		statusCode = fiber.StatusUnauthorized
		errorCode = "token_revoked"
	case domain.ErrUserNotFound:
		statusCode = fiber.StatusNotFound
		errorCode = "user_not_found"
	case domain.ErrUserAlreadyExists:
		statusCode = fiber.StatusConflict
		errorCode = "user_already_exists"
	case domain.ErrUserInactive:
		statusCode = fiber.StatusForbidden
		errorCode = "user_inactive"
	case domain.ErrUserSuspended:
		statusCode = fiber.StatusForbidden
		errorCode = "user_suspended"
	case domain.ErrInvalidInput:
		statusCode = fiber.StatusBadRequest
		errorCode = "invalid_input"
	case domain.ErrResetTokenNotFound:
		statusCode = fiber.StatusNotFound
		errorCode = "reset_token_not_found"
	case domain.ErrResetTokenExpired:
		statusCode = fiber.StatusBadRequest
		errorCode = "reset_token_expired"
	case domain.ErrResetTokenUsed:
		statusCode = fiber.StatusBadRequest
		errorCode = "reset_token_used"
	}

	return c.Status(statusCode).JSON(response.ErrorResponse{
		Success: false,
		Error:   errorCode,
		Message: message,
	})
}
