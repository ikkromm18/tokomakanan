package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Customer, error)
	FindByPhone(ctx context.Context, phone string) (*model.Customer, error)
	Create(ctx context.Context, c *model.Customer) error
	Update(ctx context.Context, c *model.Customer) error
	FindAll(ctx context.Context, page, limit int, search string) ([]model.Customer, int64, error)
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) FindByID(ctx context.Context, id uint64) (*model.Customer, error) {
	var customer model.Customer
	err := r.db.WithContext(ctx).First(&customer, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerRepository.FindByID: %w", err)
	}
	return &customer, nil
}

func (r *customerRepository) FindByPhone(ctx context.Context, phone string) (*model.Customer, error) {
	var customer model.Customer
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("customerRepository.FindByPhone: %w", err)
	}
	return &customer, nil
}

func (r *customerRepository) Create(ctx context.Context, c *model.Customer) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("customerRepository.Create: %w", err)
	}
	return nil
}

func (r *customerRepository) Update(ctx context.Context, c *model.Customer) error {
	if err := r.db.WithContext(ctx).Save(c).Error; err != nil {
		return fmt.Errorf("customerRepository.Update: %w", err)
	}
	return nil
}

func (r *customerRepository) FindAll(ctx context.Context, page, limit int, search string) ([]model.Customer, int64, error) {
	var customers []model.Customer
	var totalRows int64

	query := r.db.WithContext(ctx).Model(&model.Customer{})

	if search != "" {
		query = query.Where("name LIKE ? OR phone LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, fmt.Errorf("customerRepository.FindAll count: %w", err)
	}

	offset := pagination.GetOffset(page, limit)
	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&customers).Error; err != nil {
		return nil, 0, fmt.Errorf("customerRepository.FindAll fetch: %w", err)
	}

	return customers, totalRows, nil
}
