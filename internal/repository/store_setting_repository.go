package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"gorm.io/gorm"
)

type StoreSettingRepository interface {
	Get(ctx context.Context) (*model.StoreSetting, error)
	Update(ctx context.Context, setting *model.StoreSetting) error
}

type storeSettingRepository struct {
	db *gorm.DB
}

func NewStoreSettingRepository(db *gorm.DB) StoreSettingRepository {
	return &storeSettingRepository{db: db}
}

func (r *storeSettingRepository) Get(ctx context.Context) (*model.StoreSetting, error) {
	var setting model.StoreSetting
	err := r.db.WithContext(ctx).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("storeSettingRepository.Get: %w", err)
	}
	return &setting, nil
}

func (r *storeSettingRepository) Update(ctx context.Context, setting *model.StoreSetting) error {
	var count int64
	if setting.ID != 0 {
		if err := r.db.WithContext(ctx).Model(&model.StoreSetting{}).Where("id = ?", setting.ID).Count(&count).Error; err != nil {
			return fmt.Errorf("storeSettingRepository.Update count: %w", err)
		}
	}
	if count == 0 {
		if err := r.db.WithContext(ctx).Create(setting).Error; err != nil {
			return fmt.Errorf("storeSettingRepository.Update create: %w", err)
		}
		return nil
	}
	if err := r.db.WithContext(ctx).Save(setting).Error; err != nil {
		return fmt.Errorf("storeSettingRepository.Update save: %w", err)
	}
	return nil
}
