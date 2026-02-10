package domain

import "time"

// User represents a user in the system
type User struct {
	ID             int64      `db:"id"`
	Email          string     `db:"email"`
	PasswordHash   string     `db:"password_hash"`
	FullName       string     `db:"full_name"`
	Phone          string     `db:"phone"`
	ProfilePicture *string    `db:"profile_picture"`
	Status         string     `db:"status"` // active, inactive, suspended
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"`
}

// UserToken represents a user token (access/refresh token)
type UserToken struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Token     string    `db:"token"`
	TokenType string    `db:"token_type"` // access, refresh
	ExpiresAt time.Time `db:"expires_at"`
	Revoked   bool      `db:"revoked"`
	CreatedAt time.Time `db:"created_at"`
}

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	ID             int64     `db:"id"`
	UserID         int64     `db:"user_id"`
	Provider       string    `db:"provider"` // google, facebook, github
	ProviderUserID string    `db:"provider_user_id"`
	CreatedAt      time.Time `db:"created_at"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        int64      `db:"id"`
	UserID    int64      `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}
