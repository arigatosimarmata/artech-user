package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/arigatosimarmata/artech-user/domain"
	"github.com/jmoiron/sqlx"
)

// AuditTrailRepository defines the interface for audit trail repository operations
type AuditTrailRepository interface {
	Create(ctx context.Context, audit *domain.AuditTrail) error
	FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.AuditTrail, error)
}

// auditTrailRepository is the implementation of AuditTrailRepository
type auditTrailRepository struct {
	db *sqlx.DB
}

// NewAuditTrailRepository creates a new audit trail repository
func NewAuditTrailRepository(db *sqlx.DB) AuditTrailRepository {
	return &auditTrailRepository{db: db}
}

// Create creates a new audit trail record
func (r *auditTrailRepository) Create(ctx context.Context, audit *domain.AuditTrail) error {
	query := `
		INSERT INTO audit_trails (user_id, action, entity_type, entity_id, old_value, new_value, 
		                          ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	audit.CreatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		audit.UserID,
		audit.Action,
		audit.EntityType,
		audit.EntityID,
		audit.OldValue,
		audit.NewValue,
		audit.IPAddress,
		audit.UserAgent,
		audit.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit trail: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	
	audit.ID = id
	return nil
}

// FindByUserID finds audit trails by user ID
func (r *auditTrailRepository) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.AuditTrail, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, old_value, new_value, 
		       ip_address, user_agent, created_at
		FROM audit_trails
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	var audits []domain.AuditTrail
	err := r.db.SelectContext(ctx, &audits, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find audit trails by user id: %w", err)
	}
	
	return audits, nil
}
