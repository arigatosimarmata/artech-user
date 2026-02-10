package usecase

import (
	"context"

	"github.com/arigatosimarmata/artech-user/domain"
	"github.com/arigatosimarmata/artech-user/pkg/logger"
	"github.com/arigatosimarmata/artech-user/repository"
	"go.uber.org/zap"
)

// AuditTrailUsecase defines the interface for audit trail business logic
type AuditTrailUsecase interface {
	CreateAuditTrail(ctx context.Context, audit *domain.AuditTrail) error
	GetUserAuditTrails(ctx context.Context, userID int64, limit, offset int) ([]domain.AuditTrail, error)
}

// auditTrailUsecase is the implementation of AuditTrailUsecase
type auditTrailUsecase struct {
	auditTrailRepo repository.AuditTrailRepository
	logger         logger.Logger
}

// NewAuditTrailUsecase creates a new audit trail usecase
func NewAuditTrailUsecase(
	auditTrailRepo repository.AuditTrailRepository,
	logger logger.Logger,
) AuditTrailUsecase {
	return &auditTrailUsecase{
		auditTrailRepo: auditTrailRepo,
		logger:         logger,
	}
}

// CreateAuditTrail creates a new audit trail record
func (u *auditTrailUsecase) CreateAuditTrail(ctx context.Context, audit *domain.AuditTrail) error {
	if err := u.auditTrailRepo.Create(ctx, audit); err != nil {
		u.logger.Error("failed to create audit trail", zap.Error(err))
		return domain.ErrDatabaseError
	}
	return nil
}

// GetUserAuditTrails gets audit trails for a user
func (u *auditTrailUsecase) GetUserAuditTrails(ctx context.Context, userID int64, limit, offset int) ([]domain.AuditTrail, error) {
	audits, err := u.auditTrailRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		u.logger.Error("failed to get user audit trails", zap.Error(err))
		return nil, domain.ErrDatabaseError
	}
	return audits, nil
}
