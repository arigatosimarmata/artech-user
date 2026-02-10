package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/arigatosimarmata/artech-user/domain"
	"github.com/jmoiron/sqlx"
)

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
	
	// Token operations
	CreateToken(ctx context.Context, token *domain.UserToken) error
	FindTokenByToken(ctx context.Context, tokenStr string) (*domain.UserToken, error)
	RevokeToken(ctx context.Context, tokenStr string) error
	RevokeAllUserTokens(ctx context.Context, userID int64) error
	
	// Password reset operations
	CreatePasswordResetToken(ctx context.Context, token *domain.PasswordResetToken) error
	FindPasswordResetToken(ctx context.Context, tokenStr string) (*domain.PasswordResetToken, error)
	MarkPasswordResetTokenAsUsed(ctx context.Context, id int64) error
}

// userRepository is the implementation of UserRepository
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, full_name, phone, profile_picture, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Phone,
		user.ProfilePicture,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	
	user.ID = id
	return nil
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, phone, profile_picture, status, 
		       created_at, updated_at, deleted_at
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`
	
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	
	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, phone, profile_picture, status, 
		       created_at, updated_at, deleted_at
		FROM users
		WHERE email = ? AND deleted_at IS NULL
	`
	
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	
	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = ?, password_hash = ?, full_name = ?, phone = ?, 
		    profile_picture = ?, status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	
	user.UpdatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Phone,
		user.ProfilePicture,
		user.Status,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	
	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE users
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	
	return nil
}

// CreateToken creates a new user token
func (r *userRepository) CreateToken(ctx context.Context, token *domain.UserToken) error {
	query := `
		INSERT INTO user_tokens (user_id, token, token_type, expires_at, revoked, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	token.CreatedAt = time.Now()
	token.Revoked = false
	
	result, err := r.db.ExecContext(ctx, query,
		token.UserID,
		token.Token,
		token.TokenType,
		token.ExpiresAt,
		token.Revoked,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	
	token.ID = id
	return nil
}

// FindTokenByToken finds a token by token string
func (r *userRepository) FindTokenByToken(ctx context.Context, tokenStr string) (*domain.UserToken, error) {
	query := `
		SELECT id, user_id, token, token_type, expires_at, revoked, created_at
		FROM user_tokens
		WHERE token = ?
	`
	
	var token domain.UserToken
	err := r.db.GetContext(ctx, &token, query, tokenStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTokenInvalid
		}
		return nil, fmt.Errorf("failed to find token: %w", err)
	}
	
	return &token, nil
}

// RevokeToken revokes a token
func (r *userRepository) RevokeToken(ctx context.Context, tokenStr string) error {
	query := `
		UPDATE user_tokens
		SET revoked = true
		WHERE token = ?
	`
	
	_, err := r.db.ExecContext(ctx, query, tokenStr)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	
	return nil
}

// RevokeAllUserTokens revokes all tokens for a user
func (r *userRepository) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	query := `
		UPDATE user_tokens
		SET revoked = true
		WHERE user_id = ? AND revoked = false
	`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}
	
	return nil
}

// CreatePasswordResetToken creates a new password reset token
func (r *userRepository) CreatePasswordResetToken(ctx context.Context, token *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`
	
	token.CreatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	
	token.ID = id
	return nil
}

// FindPasswordResetToken finds a password reset token
func (r *userRepository) FindPasswordResetToken(ctx context.Context, tokenStr string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = ?
	`
	
	var token domain.PasswordResetToken
	err := r.db.GetContext(ctx, &token, query, tokenStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrResetTokenNotFound
		}
		return nil, fmt.Errorf("failed to find password reset token: %w", err)
	}
	
	return &token, nil
}

// MarkPasswordResetTokenAsUsed marks a password reset token as used
func (r *userRepository) MarkPasswordResetTokenAsUsed(ctx context.Context, id int64) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = ?
		WHERE id = ?
	`
	
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark password reset token as used: %w", err)
	}
	
	return nil
}
