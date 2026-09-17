package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Create(ctx context.Context, user *model.User) error
	FindAll(ctx context.Context, page, limit int, role, search string) ([]model.User, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindByEmail: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("userRepository.FindByID: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("userRepository.Update: %w", err)
	}
	return nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("userRepository.Create: %w", err)
	}
	return nil
}

func (r *userRepository) FindAll(ctx context.Context, page, limit int, role, search string) ([]model.User, int64, error) {
	offset := pagination.GetOffset(page, limit)
	limit = pagination.GetLimit(limit)

	var totalRows int64
	query := r.db.WithContext(ctx).Model(&model.User{})

	if role != "" {
		query = query.Where("role = ?", role)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, fmt.Errorf("userRepository.FindAll count: %w", err)
	}

	users := make([]model.User, 0)
	if err := query.Order("id DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("userRepository.FindAll find: %w", err)
	}

	return users, totalRows, nil
}

func (r *userRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&model.User{}, id).Error; err != nil {
		return fmt.Errorf("userRepository.Delete: %w", err)
	}
	return nil
}

