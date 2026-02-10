package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/arigatosimarmata/artech-user/domain"
	"github.com/arigatosimarmata/artech-user/dto/request"
	"github.com/arigatosimarmata/artech-user/dto/response"
	"github.com/arigatosimarmata/artech-user/pkg/audit"
	"github.com/arigatosimarmata/artech-user/pkg/email"
	"github.com/arigatosimarmata/artech-user/pkg/logger"
	"github.com/arigatosimarmata/artech-user/pkg/password"
	"github.com/arigatosimarmata/artech-user/pkg/token"
	"github.com/arigatosimarmata/artech-user/repository"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// UserUsecase defines the interface for user business logic
type UserUsecase interface {
	Register(ctx context.Context, req *request.RegisterRequest, ipAddress, userAgent string) (*response.AuthResponse, error)
	Login(ctx context.Context, req *request.LoginRequest, ipAddress, userAgent string) (*response.AuthResponse, error)
	Logout(ctx context.Context, userID int64, accessToken string, ipAddress, userAgent string) error
	RefreshToken(ctx context.Context, req *request.RefreshTokenRequest) (*response.AuthResponse, error)
	ForgotPassword(ctx context.Context, req *request.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *request.ResetPasswordRequest, ipAddress, userAgent string) error
	GetProfile(ctx context.Context, userID int64) (*response.UserResponse, error)
}

// userUsecase is the implementation of UserUsecase
type userUsecase struct {
	userRepo       repository.UserRepository
	auditTrailRepo repository.AuditTrailRepository
	jwtManager     *token.JWTManager
	emailSender    *email.EmailSender
	logger         logger.Logger
	validator      *validator.Validate
}

// NewUserUsecase creates a new user usecase
func NewUserUsecase(
	userRepo repository.UserRepository,
	auditTrailRepo repository.AuditTrailRepository,
	jwtManager *token.JWTManager,
	emailSender *email.EmailSender,
	logger logger.Logger,
) UserUsecase {
	return &userUsecase{
		userRepo:       userRepo,
		auditTrailRepo: auditTrailRepo,
		jwtManager:     jwtManager,
		emailSender:    emailSender,
		logger:         logger,
		validator:      validator.New(),
	}
}

// Register registers a new user
func (u *userUsecase) Register(ctx context.Context, req *request.RegisterRequest, ipAddress, userAgent string) (*response.AuthResponse, error) {
	// Validate request
	if err := u.validator.Struct(req); err != nil {
		u.logger.Error("validation failed", zap.Error(err))
		return nil, domain.ErrInvalidInput
	}

	// Check if user already exists
	existingUser, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && err != domain.ErrUserNotFound {
		u.logger.Error("failed to check existing user", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		u.logger.Error("failed to hash password", zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	// Create user
	user := &domain.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Status:       "active",
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		u.logger.Error("failed to create user", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}

	// Generate tokens
	accessToken, err := u.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		u.logger.Error("failed to generate access token", zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	refreshToken, err := u.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		u.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	// Store refresh token
	userToken := &domain.UserToken{
		UserID:    user.ID,
		Token:     refreshToken,
		TokenType: "refresh",
		ExpiresAt: time.Now().Add(u.jwtManager.GetRefreshTokenDuration()),
	}

	if err := u.userRepo.CreateToken(ctx, userToken); err != nil {
		u.logger.Error("failed to store refresh token", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}

	// Create audit trail
	newValue, _ := audit.FormatValue(user)
	auditTrail := &domain.AuditTrail{
		UserID:     user.ID,
		Action:     "register",
		EntityType: "user",
		EntityID:   strconv.FormatInt(user.ID, 10),
		NewValue:   newValue,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	if err := u.auditTrailRepo.Create(ctx, auditTrail); err != nil {
		u.logger.Warn("failed to create audit trail", zap.Error(err))
	}

	// Send welcome email (async, don't block on failure)
	go func() {
		if err := u.emailSender.SendWelcomeEmail(user.Email, user.FullName); err != nil {
			u.logger.Warn("failed to send welcome email", zap.Error(err))
		}
	}()

	return &response.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(u.jwtManager.GetAccessTokenDuration().Seconds()),
		User:         u.mapUserToResponse(user),
	}, nil
}

// Login authenticates a user
func (u *userUsecase) Login(ctx context.Context, req *request.LoginRequest, ipAddress, userAgent string) (*response.AuthResponse, error) {
	// Validate request
	if err := u.validator.Struct(req); err != nil {
		u.logger.Error("validation failed", zap.Error(err))
		return nil, domain.ErrInvalidInput
	}

	// Find user by email
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		u.logger.Error("failed to find user", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}

	// Check user status
	if user.Status != "active" {
		if user.Status == "suspended" {
			return nil, domain.ErrUserSuspended
		}
		return nil, domain.ErrUserInactive
	}

	// Verify password
	if err := password.Verify(user.PasswordHash, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := u.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		u.logger.Error("failed to generate access token", zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	refreshToken, err := u.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		u.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	// Store refresh token
	userToken := &domain.UserToken{
		UserID:    user.ID,
		Token:     refreshToken,
		TokenType: "refresh",
		ExpiresAt: time.Now().Add(u.jwtManager.GetRefreshTokenDuration()),
	}

	if err := u.userRepo.CreateToken(ctx, userToken); err != nil {
		u.logger.Error("failed to store refresh token", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}

	// Create audit trail
	auditTrail := &domain.AuditTrail{
		UserID:     user.ID,
		Action:     "login",
		EntityType: "user",
		EntityID:   strconv.FormatInt(user.ID, 10),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	if err := u.auditTrailRepo.Create(ctx, auditTrail); err != nil {
		u.logger.Warn("failed to create audit trail", zap.Error(err))
	}

	return &response.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(u.jwtManager.GetAccessTokenDuration().Seconds()),
		User:         u.mapUserToResponse(user),
	}, nil
}

// Logout logs out a user
func (u *userUsecase) Logout(ctx context.Context, userID int64, accessToken string, ipAddress, userAgent string) error {
	// Revoke all user tokens
	if err := u.userRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		u.logger.Error("failed to revoke user tokens", zap.Error(err))
		return domain.ErrDatabaseError
	}

	// Create audit trail
	auditTrail := &domain.AuditTrail{
		UserID:     userID,
		Action:     "logout",
		EntityType: "user",
		EntityID:   strconv.FormatInt(userID, 10),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	if err := u.auditTrailRepo.Create(ctx, auditTrail); err != nil {
		u.logger.Warn("failed to create audit trail", zap.Error(err))
	}

	return nil
}

// RefreshToken refreshes an access token
func (u *userUsecase) RefreshToken(ctx context.Context, req *request.RefreshTokenRequest) (*response.AuthResponse, error) {
	// Validate request
	if err := u.validator.Struct(req); err != nil {
		u.logger.Error("validation failed", zap.Error(err))
		return nil, domain.ErrInvalidInput
	}

	// Validate refresh token
	claims, err := u.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Check if token exists in database
	storedToken, err := u.userRepo.FindTokenByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	// Check if token is revoked
	if storedToken.Revoked {
		return nil, domain.ErrTokenRevoked
	}

	// Check if token is expired
	if time.Now().After(storedToken.ExpiresAt) {
		return nil, domain.ErrTokenExpired
	}

	// Get user
	user, err := u.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrUserNotFound
		}
		u.logger.Error("failed to find user", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}

	// Check user status
	if user.Status != "active" {
		if user.Status == "suspended" {
			return nil, domain.ErrUserSuspended
		}
		return nil, domain.ErrUserInactive
	}

	// Generate new tokens
	newAccessToken, err := u.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		u.logger.Error("failed to generate access token", zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	newRefreshToken, err := u.jwtManager.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		u.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	// Revoke old refresh token
	if err := u.userRepo.RevokeToken(ctx, req.RefreshToken); err != nil {
		u.logger.Error("failed to revoke old token", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}

	// Store new refresh token
	userToken := &domain.UserToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		TokenType: "refresh",
		ExpiresAt: time.Now().Add(u.jwtManager.GetRefreshTokenDuration()),
	}

	if err := u.userRepo.CreateToken(ctx, userToken); err != nil {
		u.logger.Error("failed to store refresh token", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}

	return &response.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(u.jwtManager.GetAccessTokenDuration().Seconds()),
		User:         u.mapUserToResponse(user),
	}, nil
}

// ForgotPassword initiates password reset
func (u *userUsecase) ForgotPassword(ctx context.Context, req *request.ForgotPasswordRequest) error {
	// Validate request
	if err := u.validator.Struct(req); err != nil {
		u.logger.Error("validation failed", zap.Error(err))
		return domain.ErrInvalidInput
	}

	// Find user by email
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if user exists or not
		if err == domain.ErrUserNotFound {
			return nil
		}
		u.logger.Error("failed to find user", zap.Error(err))
		return domain.ErrDatabaseError
	}

	// Generate reset token (simple random string for now)
	resetToken := fmt.Sprintf("%d-%d", user.ID, time.Now().Unix())

	// Store reset token
	passwordResetToken := &domain.PasswordResetToken{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := u.userRepo.CreatePasswordResetToken(ctx, passwordResetToken); err != nil {
		u.logger.Error("failed to create password reset token", zap.Error(err))
		return domain.ErrDatabaseError
	}

	// Send password reset email (async, don't block on failure)
	go func() {
		if err := u.emailSender.SendForgotPasswordEmail(user.Email, resetToken); err != nil {
			u.logger.Warn("failed to send password reset email", zap.Error(err))
		}
	}()

	return nil
}

// ResetPassword resets user password
func (u *userUsecase) ResetPassword(ctx context.Context, req *request.ResetPasswordRequest, ipAddress, userAgent string) error {
	// Validate request
	if err := u.validator.Struct(req); err != nil {
		u.logger.Error("validation failed", zap.Error(err))
		return domain.ErrInvalidInput
	}

	// Find password reset token
	resetToken, err := u.userRepo.FindPasswordResetToken(ctx, req.Token)
	if err != nil {
		return err
	}

	// Check if token has been used
	if resetToken.UsedAt != nil {
		return domain.ErrResetTokenUsed
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return domain.ErrResetTokenExpired
	}

	// Get user
	user, err := u.userRepo.FindByID(ctx, resetToken.UserID)
	if err != nil {
		u.logger.Error("failed to find user", zap.Error(err))
		return domain.ErrDatabaseError
	}

	// Hash new password
	hashedPassword, err := password.Hash(req.NewPassword)
	if err != nil {
		u.logger.Error("failed to hash password", zap.Error(err))
		return domain.ErrInternalServer
	}

	// Update user password
	user.PasswordHash = hashedPassword
	if err := u.userRepo.Update(ctx, user); err != nil {
		u.logger.Error("failed to update user password", zap.Error(err))
		return domain.ErrDatabaseError
	}

	// Mark reset token as used
	if err := u.userRepo.MarkPasswordResetTokenAsUsed(ctx, resetToken.ID); err != nil {
		u.logger.Error("failed to mark reset token as used", zap.Error(err))
	}

	// Revoke all user tokens for security
	if err := u.userRepo.RevokeAllUserTokens(ctx, user.ID); err != nil {
		u.logger.Warn("failed to revoke user tokens", zap.Error(err))
	}

	// Create audit trail
	auditTrail := &domain.AuditTrail{
		UserID:     user.ID,
		Action:     "reset_password",
		EntityType: "user",
		EntityID:   strconv.FormatInt(user.ID, 10),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	if err := u.auditTrailRepo.Create(ctx, auditTrail); err != nil {
		u.logger.Warn("failed to create audit trail", zap.Error(err))
	}

	return nil
}

// GetProfile gets user profile
func (u *userUsecase) GetProfile(ctx context.Context, userID int64) (*response.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrUserNotFound
		}
		u.logger.Error("failed to find user", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}

	return &response.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		FullName:       user.FullName,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}, nil
}

// mapUserToResponse maps domain.User to response.UserResponse
func (u *userUsecase) mapUserToResponse(user *domain.User) response.UserResponse {
	return response.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		FullName:       user.FullName,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}
