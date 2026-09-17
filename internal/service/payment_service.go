package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/rs/zerolog/log"
)

type PaymentService interface {
	Create(ctx context.Context, orderID uint64, req dto.CreatePaymentRequest, actorID uint64, ipAddress string) (*dto.PaymentResponse, error)
	ListByOrder(ctx context.Context, orderID uint64) ([]dto.PaymentResponse, error)
}

type paymentService struct {
	paymentRepo  repository.PaymentRepository
	orderRepo    repository.OrderRepository
	auditService AuditService
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	auditService AuditService,
) PaymentService {
	return &paymentService{
		paymentRepo:  paymentRepo,
		orderRepo:    orderRepo,
		auditService: auditService,
	}
}

func toPaymentResponse(p *model.OrderPayment) *dto.PaymentResponse {
	return &dto.PaymentResponse{
		ID:            p.ID,
		OrderID:       p.OrderID,
		Amount:        p.Amount,
		PaymentMethod: p.PaymentMethod,
		Notes:         p.Notes,
		PaidAt:        p.PaidAt,
		CreatedAt:     p.CreatedAt,
	}
}

func (s *paymentService) Create(ctx context.Context, orderID uint64, req dto.CreatePaymentRequest, actorID uint64, ipAddress string) (*dto.PaymentResponse, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		log.Error().Err(err).Uint64("order_id", orderID).Str("component", "paymentService.Create").Msg("failed to find order")
		return nil, fmt.Errorf("paymentService.Create: %w", err)
	}
	if order == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "Order not found")
	}

	if order.Status == model.OrderStatusCancelled || order.Status == model.OrderStatusCompleted {
		return nil, response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("cannot make payment on %s order", order.Status))
	}

	remainingBill := order.TotalAmount - order.TotalPaid
	if req.Amount > remainingBill {
		return nil, response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("payment amount exceeds remaining bill (remaining: %.2f)", remainingBill))
	}

	newTotalPaid := order.TotalPaid + req.Amount
	var newStatus string

	if newTotalPaid >= order.TotalAmount {
		if order.Status == model.OrderStatusDraft || order.Status == model.OrderStatusDPPaid {
			newStatus = model.OrderStatusPaid
		}
	} else if order.Status == model.OrderStatusDraft {
		newStatus = model.OrderStatusDPPaid
	}

	payment := &model.OrderPayment{
		OrderID:       orderID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Notes:         req.Notes,
		PaidAt:        time.Now(),
	}

	if err := s.paymentRepo.CreatePaymentTx(ctx, payment, newTotalPaid, newStatus); err != nil {
		log.Error().Err(err).Uint64("order_id", orderID).Str("component", "paymentService.Create").Msg("failed to record payment transaction")
		return nil, fmt.Errorf("paymentService.Create: %w", err)
	}

	res := toPaymentResponse(payment)

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "CREATE",
			EntityType: "order_payment",
			EntityID:   &payment.ID,
			NewValue:   payment,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "paymentService.Create").Msg("audit log: failed to record entry")
		}
	}

	return res, nil
}

func (s *paymentService) ListByOrder(ctx context.Context, orderID uint64) ([]dto.PaymentResponse, error) {
	payments, err := s.paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		log.Error().Err(err).Uint64("order_id", orderID).Str("component", "paymentService.ListByOrder").Msg("failed to list payments")
		return nil, fmt.Errorf("paymentService.ListByOrder: %w", err)
	}

	responses := make([]dto.PaymentResponse, len(payments))
	for i := range payments {
		responses[i] = *toPaymentResponse(&payments[i])
	}

	return responses, nil
}
