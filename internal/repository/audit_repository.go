package repository

import (
	"context"

	"github.com/ikkromm18/tokomakanan/internal/model"
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
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditRepository) FindAll(ctx context.Context, page, limit int, entityType, action string) ([]model.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	var totalRows int64
	query := r.db.WithContext(ctx).Model(&model.AuditLog{})

	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]model.AuditLog, 0)
	if err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, totalRows, nil
}
