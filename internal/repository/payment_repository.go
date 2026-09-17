package repository

import (
	"context"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreatePaymentTx(ctx context.Context, payment *model.OrderPayment, newTotalPaid float64, newStatus string) error
	FindByOrderID(ctx context.Context, orderID uint64) ([]model.OrderPayment, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) CreatePaymentTx(ctx context.Context, payment *model.OrderPayment, newTotalPaid float64, newStatus string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("paymentRepository.CreatePaymentTx create payment: %w", err)
		}

		updates := map[string]any{
			"total_paid": newTotalPaid,
		}
		if newStatus != "" {
			updates["status"] = newStatus
		}

		if err := tx.Model(&model.Order{}).Where("id = ?", payment.OrderID).Updates(updates).Error; err != nil {
			return fmt.Errorf("paymentRepository.CreatePaymentTx update order: %w", err)
		}

		return nil
	})
}

func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID uint64) ([]model.OrderPayment, error) {
	var payments []model.OrderPayment
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("id ASC").
		Find(&payments).Error
	if err != nil {
		return nil, fmt.Errorf("paymentRepository.FindByOrderID: %w", err)
	}
	return payments, nil
}
