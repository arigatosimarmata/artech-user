package domain

import "time"

// AuditTrail represents an audit trail record for tracking user activities
type AuditTrail struct {
	ID         int64      `db:"id"`
	UserID     int64      `db:"user_id"`
	Action     string     `db:"action"` // login, logout, register, update_profile, etc.
	EntityType string     `db:"entity_type"` // user, user_token, etc.
	EntityID   string     `db:"entity_id"`
	OldValue   *string    `db:"old_value"`
	NewValue   *string    `db:"new_value"`
	IPAddress  string     `db:"ip_address"`
	UserAgent  string     `db:"user_agent"`
	CreatedAt  time.Time  `db:"created_at"`
}
