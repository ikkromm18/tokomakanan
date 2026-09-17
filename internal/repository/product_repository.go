package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"gorm.io/gorm"
)

type ProductRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Product, error)
	FindByNameAndCategory(ctx context.Context, name, category string) (*model.Product, error)
	Create(ctx context.Context, p *model.Product) error
	Update(ctx context.Context, p *model.Product) error
	Delete(ctx context.Context, id uint64) error
	FindAll(ctx context.Context, page, limit int, category string, isActive *bool, search string) ([]model.Product, int64, error)
	ExistsInActivePackages(ctx context.Context, productID uint64) (bool, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) FindByID(ctx context.Context, id uint64) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("productRepository.FindByID: %w", err)
	}
	return &product, nil
}

func (r *productRepository) FindByNameAndCategory(ctx context.Context, name, category string) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).
		Where("name = ? AND category = ?", name, category).
		First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("productRepository.FindByNameAndCategory: %w", err)
	}
	return &product, nil
}

func (r *productRepository) Create(ctx context.Context, p *model.Product) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return fmt.Errorf("productRepository.Create: %w", err)
	}
	return nil
}

func (r *productRepository) Update(ctx context.Context, p *model.Product) error {
	if err := r.db.WithContext(ctx).Save(p).Error; err != nil {
		return fmt.Errorf("productRepository.Update: %w", err)
	}
	return nil
}

func (r *productRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&model.Product{}, id).Error; err != nil {
		return fmt.Errorf("productRepository.Delete: %w", err)
	}
	return nil
}

func (r *productRepository) FindAll(ctx context.Context, page, limit int, category string, isActive *bool, search string) ([]model.Product, int64, error) {
	var products []model.Product
	var totalRows int64

	query := r.db.WithContext(ctx).Model(&model.Product{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, fmt.Errorf("productRepository.FindAll count: %w", err)
	}

	offset := pagination.GetOffset(page, limit)
	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("productRepository.FindAll fetch: %w", err)
	}

	return products, totalRows, nil
}

func (r *productRepository) ExistsInActivePackages(ctx context.Context, productID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("package_items").
		Joins("JOIN product_packages ON product_packages.id = package_items.package_id").
		Where("package_items.product_id = ? AND product_packages.deleted_at IS NULL AND product_packages.is_active = 1", productID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("productRepository.ExistsInActivePackages: %w", err)
	}
	return count > 0, nil
}
