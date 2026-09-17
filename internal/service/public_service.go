package service

import (
	"context"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/rs/zerolog/log"
)

type PublicService interface {
	GetInvoiceByToken(ctx context.Context, token string) (*dto.PublicInvoiceResponse, error)
}

type publicService struct {
	orderRepo   repository.OrderRepository
	settingRepo repository.StoreSettingRepository
}

func NewPublicService(
	orderRepo repository.OrderRepository,
	settingRepo repository.StoreSettingRepository,
) PublicService {
	return &publicService{
		orderRepo:   orderRepo,
		settingRepo: settingRepo,
	}
}

func (s *publicService) GetInvoiceByToken(ctx context.Context, token string) (*dto.PublicInvoiceResponse, error) {
	order, err := s.orderRepo.FindByToken(ctx, token)
	if err != nil {
		log.Error().Err(err).Str("token", token).Str("component", "publicService.GetInvoiceByToken").Msg("failed to find order by token")
		return nil, fmt.Errorf("publicService.GetInvoiceByToken: %w", err)
	}
	if order == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "invoice not found")
	}

	setting, err := s.settingRepo.Get(ctx)
	if err != nil {
		log.Warn().Err(err).Str("component", "publicService.GetInvoiceByToken").Msg("failed to fetch store settings for invoice")
	}

	var storeInfo dto.PublicInvoiceStoreInfo
	if setting != nil {
		storeInfo = dto.PublicInvoiceStoreInfo{
			Name:          setting.Name,
			Address:       setting.Address,
			Phone:         setting.Phone,
			LogoURL:       setting.LogoURL,
			ReceiptFooter: setting.ReceiptFooter,
		}
	} else {
		storeInfo = dto.PublicInvoiceStoreInfo{
			Name: "Bakery POS",
		}
	}

	var customerName *string
	if order.Customer != nil {
		customerName = &order.Customer.Name
	}

	cashierName := "System"
	if order.User != nil {
		cashierName = order.User.Name
	}

	items := make([]dto.PublicInvoiceItem, len(order.Items))
	for i, it := range order.Items {
		items[i] = dto.PublicInvoiceItem{
			ItemName:  it.ItemName,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
			Subtotal:  it.Subtotal,
		}
	}

	payments := make([]dto.PublicInvoicePayment, len(order.Payments))
	for i, p := range order.Payments {
		payments[i] = dto.PublicInvoicePayment{
			Amount:        p.Amount,
			PaymentMethod: p.PaymentMethod,
			PaidAt:        p.PaidAt,
		}
	}

	remaining := order.TotalAmount - order.TotalPaid
	if remaining < 0 {
		remaining = 0
	}

	var pickupDate *string
	if order.PickupDate != nil {
		formatted := order.PickupDate.Format("2006-01-02")
		pickupDate = &formatted
	}

	res := &dto.PublicInvoiceResponse{
		Store:          storeInfo,
		InvoiceNo:      order.InvoiceNo,
		CustomerName:   customerName,
		CashierName:    cashierName,
		OrderType:      order.OrderType,
		Status:         order.Status,
		PickupDate:     pickupDate,
		Notes:          order.Notes,
		Subtotal:       order.Subtotal,
		DiscountAmount: order.DiscountAmount,
		TotalAmount:    order.TotalAmount,
		TotalPaid:      order.TotalPaid,
		RemainingPaid:  remaining,
		Items:          items,
		Payments:       payments,
		CreatedAt:      order.CreatedAt,
	}

	return res, nil
}
