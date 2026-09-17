package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"gorm.io/gorm"
)

type PackageRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.ProductPackage, error)
	FindByName(ctx context.Context, name string) (*model.ProductPackage, error)
	Create(ctx context.Context, pkg *model.ProductPackage) error
	Update(ctx context.Context, pkg *model.ProductPackage, items []model.PackageItem) error
	Delete(ctx context.Context, id uint64) error
	FindAll(ctx context.Context, page, limit int, isActive *bool, search string) ([]model.ProductPackage, int64, error)
}

type packageRepository struct {
	db *gorm.DB
}

func NewPackageRepository(db *gorm.DB) PackageRepository {
	return &packageRepository{db: db}
}

func (r *packageRepository) FindByID(ctx context.Context, id uint64) (*model.ProductPackage, error) {
	var pkg model.ProductPackage
	err := r.db.WithContext(ctx).
		Preload("Items.Product").
		First(&pkg, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("packageRepository.FindByID: %w", err)
	}
	return &pkg, nil
}

func (r *packageRepository) FindByName(ctx context.Context, name string) (*model.ProductPackage, error) {
	var pkg model.ProductPackage
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&pkg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("packageRepository.FindByName: %w", err)
	}
	return &pkg, nil
}

func (r *packageRepository) Create(ctx context.Context, pkg *model.ProductPackage) error {
	// GORM will automatically insert associated Items within transaction
	if err := r.db.WithContext(ctx).Create(pkg).Error; err != nil {
		return fmt.Errorf("packageRepository.Create: %w", err)
	}
	return nil
}

func (r *packageRepository) Update(ctx context.Context, pkg *model.ProductPackage, items []model.PackageItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update parent package fields
		if err := tx.Save(pkg).Error; err != nil {
			return fmt.Errorf("packageRepository.Update save package: %w", err)
		}

		// Delete existing package items
		if err := tx.Where("package_id = ?", pkg.ID).Delete(&model.PackageItem{}).Error; err != nil {
			return fmt.Errorf("packageRepository.Update delete items: %w", err)
		}

		// Re-insert new package items
		if len(items) > 0 {
			for i := range items {
				items[i].PackageID = pkg.ID
				items[i].ID = 0 // ensure new IDs
			}
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("packageRepository.Update recreate items: %w", err)
			}
		}

		return nil
	})
}

func (r *packageRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&model.ProductPackage{}, id).Error; err != nil {
		return fmt.Errorf("packageRepository.Delete: %w", err)
	}
	return nil
}

func (r *packageRepository) FindAll(ctx context.Context, page, limit int, isActive *bool, search string) ([]model.ProductPackage, int64, error) {
	var packages []model.ProductPackage
	var totalRows int64

	query := r.db.WithContext(ctx).Model(&model.ProductPackage{})

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, fmt.Errorf("packageRepository.FindAll count: %w", err)
	}

	offset := pagination.GetOffset(page, limit)
	if err := query.Preload("Items.Product").
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&packages).Error; err != nil {
		return nil, 0, fmt.Errorf("packageRepository.FindAll fetch: %w", err)
	}

	return packages, totalRows, nil
}
