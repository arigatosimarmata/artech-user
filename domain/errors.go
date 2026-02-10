package domain

import (
	"errors"
	"fmt"
)

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrTokenRevoked       = errors.New("token has been revoked")

	// User errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserInactive      = errors.New("user is inactive")
	ErrUserSuspended     = errors.New("user is suspended")

	// Validation errors
	ErrInvalidInput     = errors.New("invalid input")
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")

	// Password reset errors
	ErrResetTokenNotFound = errors.New("password reset token not found")
	ErrResetTokenExpired  = errors.New("password reset token has expired")
	ErrResetTokenUsed     = errors.New("password reset token has already been used")

	// General errors
	ErrInternalServer = errors.New("internal server error")
	ErrDatabaseError  = errors.New("database error")
)

// AppError represents an application error with additional context
type AppError struct {
	Err        error
	Message    string
	StatusCode int
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Err.Error()
}

// NewAppError creates a new AppError
func NewAppError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}
