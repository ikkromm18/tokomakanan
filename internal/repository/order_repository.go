package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/model"
	invoicepkg "github.com/ikkromm18/tokomakanan/internal/pkg/invoice"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"gorm.io/gorm"
)

type OrderRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Order, error)
	FindByToken(ctx context.Context, token string) (*model.Order, error)
	CreateOrderTx(ctx context.Context, order *model.Order, initialPayment *model.OrderPayment) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
	FindAll(ctx context.Context, page, limit int, status, orderType, startDate, endDate string) ([]model.Order, int64, error)
	UpdateTotalPaidAndStatus(ctx context.Context, orderID uint64, newTotalPaid float64, newStatus string) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) FindByID(ctx context.Context, id uint64) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("User").
		Preload("Items").
		Preload("Payments").
		First(&order, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("orderRepository.FindByID: %w", err)
	}
	return &order, nil
}

func (r *orderRepository) FindByToken(ctx context.Context, token string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("User").
		Preload("Items").
		Preload("Payments").
		Where("invoice_token = ?", token).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("orderRepository.FindByToken: %w", err)
	}
	return &order, nil
}

func (r *orderRepository) CreateOrderTx(ctx context.Context, order *model.Order, initialPayment *model.OrderPayment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Generate invoice number within transaction to prevent race conditions
		if order.InvoiceNo == "" {
			invNo, err := invoicepkg.GenerateInvoiceNumber(ctx, tx, time.Now())
			if err != nil {
				return fmt.Errorf("orderRepository.CreateOrderTx generate invoice: %w", err)
			}
			order.InvoiceNo = invNo
		}

		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("orderRepository.CreateOrderTx create order: %w", err)
		}

		if initialPayment != nil {
			initialPayment.OrderID = order.ID
			if err := tx.Create(initialPayment).Error; err != nil {
				return fmt.Errorf("orderRepository.CreateOrderTx create initial payment: %w", err)
			}
		}

		return nil
	})
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	err := r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
	if err != nil {
		return fmt.Errorf("orderRepository.UpdateStatus: %w", err)
	}
	return nil
}

func (r *orderRepository) UpdateTotalPaidAndStatus(ctx context.Context, orderID uint64, newTotalPaid float64, newStatus string) error {
	updates := map[string]any{
		"total_paid": newTotalPaid,
	}
	if newStatus != "" {
		updates["status"] = newStatus
	}
	err := r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", orderID).
		Updates(updates).Error
	if err != nil {
		return fmt.Errorf("orderRepository.UpdateTotalPaidAndStatus: %w", err)
	}
	return nil
}

func (r *orderRepository) FindAll(ctx context.Context, page, limit int, status, orderType, startDate, endDate string) ([]model.Order, int64, error) {
	var orders []model.Order
	var totalRows int64

	query := r.db.WithContext(ctx).Model(&model.Order{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if orderType != "" {
		query = query.Where("order_type = ?", orderType)
	}
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, 0, fmt.Errorf("orderRepository.FindAll count: %w", err)
	}

	offset := pagination.GetOffset(page, limit)
	if err := query.
		Preload("Customer").
		Preload("User").
		Preload("Items").
		Preload("Payments").
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("orderRepository.FindAll fetch: %w", err)
	}

	return orders, totalRows, nil
}
