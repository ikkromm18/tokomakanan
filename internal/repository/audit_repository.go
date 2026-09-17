package repository

import (
	"context"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	FindAll(ctx context.Context, page, limit int, entityType, action string) ([]model.AuditLog, int64, error)
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("auditRepository.Create: %w", err)
	}
	return nil
}

func (r *auditRepository) FindAll(ctx context.Context, page, limit int, entityType, action string) ([]model.AuditLog, int64, error) {
	offset := pagination.GetOffset(page, limit)
	limit = pagination.GetLimit(limit)

	var totalRows int64
	query := r.db.WithContext(ctx).Model(&model.AuditLog{})

	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, fmt.Errorf("auditRepository.FindAll count: %w", err)
	}

	logs := make([]model.AuditLog, 0)
	if err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("auditRepository.FindAll find: %w", err)
	}

	return logs, totalRows, nil
}
