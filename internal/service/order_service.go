package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/rs/zerolog/log"
)

type OrderService interface {
	List(ctx context.Context, page, limit int, status, orderType, startDate, endDate string, userRole string) ([]dto.OrderResponse, dto.PaginationMeta, error)
	Create(ctx context.Context, userID uint64, userRole string, req dto.CreateOrderRequest, ipAddress string) (*dto.OrderResponse, error)
	GetByID(ctx context.Context, id uint64, userRole string) (*dto.OrderResponse, error)
	UpdateStatus(ctx context.Context, id uint64, newStatus string, actorID uint64, ipAddress string) error
}

type orderService struct {
	orderRepo    repository.OrderRepository
	productRepo  repository.ProductRepository
	packageRepo  repository.PackageRepository
	auditService AuditService
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	packageRepo repository.PackageRepository,
	auditService AuditService,
) OrderService {
	return &orderService{
		orderRepo:    orderRepo,
		productRepo:  productRepo,
		packageRepo:  packageRepo,
		auditService: auditService,
	}
}

func toOrderResponse(o *model.Order, userRole string) *dto.OrderResponse {
	remainingPaid := o.TotalAmount - o.TotalPaid
	if remainingPaid < 0 {
		remainingPaid = 0
	}

	res := &dto.OrderResponse{
		ID:             o.ID,
		InvoiceNo:      o.InvoiceNo,
		InvoiceToken:   o.InvoiceToken,
		CustomerID:     o.CustomerID,
		UserID:         o.UserID,
		OrderType:      o.OrderType,
		Status:         o.Status,
		Notes:          o.Notes,
		Subtotal:       o.Subtotal,
		DiscountAmount: o.DiscountAmount,
		TotalAmount:    o.TotalAmount,
		TotalPaid:      o.TotalPaid,
		RemainingPaid:  remainingPaid,
		PaymentMethod:  o.PaymentMethod,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}

	if o.PickupDate != nil {
		pDate := o.PickupDate.Format("2006-01-02")
		res.PickupDate = &pDate
	}

	if o.Customer != nil {
		res.CustomerName = &o.Customer.Name
	}
	if o.User != nil {
		res.UserName = o.User.Name
	}

	// Admins cannot see HPP / profit
	if userRole != model.RoleAdmin {
		totalHPPVal := o.TotalHPP
		res.TotalHPP = &totalHPPVal
	}

	if len(o.Items) > 0 {
		res.Items = make([]dto.OrderItemResponse, len(o.Items))
		for i, it := range o.Items {
			itResp := dto.OrderItemResponse{
				ID:        it.ID,
				ItemType:  it.ItemType,
				ItemID:    it.ItemID,
				ItemName:  it.ItemName,
				Quantity:  it.Quantity,
				UnitPrice: it.UnitPrice,
				Subtotal:  it.Subtotal,
			}
			if userRole != model.RoleAdmin {
				hppVal := it.UnitHPP
				itResp.UnitHPP = &hppVal
			}
			res.Items[i] = itResp
		}
	}

	if len(o.Payments) > 0 {
		res.Payments = make([]dto.OrderPaymentResponse, len(o.Payments))
		for i, p := range o.Payments {
			res.Payments[i] = dto.OrderPaymentResponse{
				ID:            p.ID,
				Amount:        p.Amount,
				PaymentMethod: p.PaymentMethod,
				Notes:         p.Notes,
				PaidAt:        p.PaidAt,
			}
		}
	}

	return res
}

func (s *orderService) List(ctx context.Context, page, limit int, status, orderType, startDate, endDate string, userRole string) ([]dto.OrderResponse, dto.PaginationMeta, error) {
	orders, totalRows, err := s.orderRepo.FindAll(ctx, page, limit, status, orderType, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Str("component", "orderService.List").Msg("failed to list orders")
		return nil, dto.PaginationMeta{}, fmt.Errorf("orderService.List: %w", err)
	}

	responses := make([]dto.OrderResponse, len(orders))
	for i := range orders {
		responses[i] = *toOrderResponse(&orders[i], userRole)
	}

	meta := pagination.FormatPagination(page, limit, totalRows)
	return responses, meta, nil
}

func (s *orderService) Create(ctx context.Context, userID uint64, userRole string, req dto.CreateOrderRequest, ipAddress string) (*dto.OrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, response.NewServiceError(response.ErrBusinessRule, "order must have at least one item")
	}

	var pickupDate *time.Time
	if req.OrderType == model.OrderTypePreOrder {
		if req.PickupDate == nil || *req.PickupDate == "" {
			return nil, response.NewServiceError(response.ErrBusinessRule, "pickup date is required for pre-order")
		}
		parsed, err := time.Parse("2006-01-02", *req.PickupDate)
		if err != nil {
			return nil, response.NewServiceError(response.ErrBusinessRule, "invalid pickup date format (must be YYYY-MM-DD)")
		}

		today := time.Now().Truncate(24 * time.Hour)
		if parsed.Before(today) {
			return nil, response.NewServiceError(response.ErrBusinessRule, "pickup date cannot be in the past")
		}
		pickupDate = &parsed
	}

	var subtotal float64
	var totalHPP float64
	orderItems := make([]model.OrderItem, len(req.Items))

	for i, itemReq := range req.Items {
		var itemName string
		var unitPrice float64
		var unitHPP float64

		switch itemReq.ItemType {
		case model.ItemTypeProduct:
			prod, err := s.productRepo.FindByID(ctx, itemReq.ItemID)
			if err != nil {
				return nil, fmt.Errorf("orderService.Create lookup product %d: %w", itemReq.ItemID, err)
			}
			if prod == nil {
				return nil, response.NewServiceError(response.ErrNotFound, fmt.Sprintf("product with ID %d not found", itemReq.ItemID))
			}
			if !prod.IsActive {
				return nil, response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("product %s is not active", prod.Name))
			}
			itemName = prod.Name
			unitPrice = prod.SellPrice
			unitHPP = prod.HPP

		case model.ItemTypePackage:
			pkg, err := s.packageRepo.FindByID(ctx, itemReq.ItemID)
			if err != nil {
				return nil, fmt.Errorf("orderService.Create lookup package %d: %w", itemReq.ItemID, err)
			}
			if pkg == nil {
				return nil, response.NewServiceError(response.ErrNotFound, fmt.Sprintf("package with ID %d not found", itemReq.ItemID))
			}
			if !pkg.IsActive {
				return nil, response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("package %s is not active", pkg.Name))
			}
			itemName = pkg.Name
			unitPrice = pkg.SellPrice
			unitHPP = pkg.TotalHPP

		default:
			return nil, response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("unsupported item type %s", itemReq.ItemType))
		}

		itemSubtotal := unitPrice * float64(itemReq.Quantity)
		subtotal += itemSubtotal
		totalHPP += unitHPP * float64(itemReq.Quantity)

		orderItems[i] = model.OrderItem{
			ItemType:  itemReq.ItemType,
			ItemID:    itemReq.ItemID,
			ItemName:  itemName,
			Quantity:  itemReq.Quantity,
			UnitPrice: unitPrice,
			UnitHPP:   unitHPP,
			Subtotal:  itemSubtotal,
		}
	}

	if req.DiscountAmount > subtotal {
		return nil, response.NewServiceError(response.ErrBusinessRule, "discount amount cannot exceed subtotal")
	}

	totalAmount := subtotal - req.DiscountAmount

	// Status determination and payment validation
	var status string
	var initialPayment *model.OrderPayment

	if req.OrderType == model.OrderTypeDirectSale {
		if req.InitialPaid < totalAmount {
			return nil, response.NewServiceError(response.ErrBusinessRule, "direct sale requires full payment")
		}
		if req.InitialPaid > totalAmount {
			return nil, response.NewServiceError(response.ErrBusinessRule, "payment amount cannot exceed total amount")
		}
		status = model.OrderStatusPaid
		initialPayment = &model.OrderPayment{
			Amount:        req.InitialPaid,
			PaymentMethod: req.PaymentMethod,
			PaidAt:        time.Now(),
		}
	} else { // PRE_ORDER
		if req.InitialPaid > totalAmount {
			return nil, response.NewServiceError(response.ErrBusinessRule, "payment amount cannot exceed total amount")
		}
		if req.InitialPaid == 0 {
			status = model.OrderStatusDraft
		} else if req.InitialPaid < totalAmount {
			status = model.OrderStatusDPPaid
			initialPayment = &model.OrderPayment{
				Amount:        req.InitialPaid,
				PaymentMethod: req.PaymentMethod,
				PaidAt:        time.Now(),
			}
		} else {
			status = model.OrderStatusPaid
			initialPayment = &model.OrderPayment{
				Amount:        req.InitialPaid,
				PaymentMethod: req.PaymentMethod,
				PaidAt:        time.Now(),
			}
		}
	}

	invoiceToken := uuid.New().String()

	order := &model.Order{
		InvoiceToken:   invoiceToken,
		CustomerID:     req.CustomerID,
		UserID:         userID,
		OrderType:      req.OrderType,
		Status:         status,
		PickupDate:     pickupDate,
		Notes:          req.Notes,
		Subtotal:       subtotal,
		DiscountAmount: req.DiscountAmount,
		TotalAmount:    totalAmount,
		TotalHPP:       totalHPP,
		TotalPaid:      req.InitialPaid,
		PaymentMethod:  req.PaymentMethod,
		Items:          orderItems,
	}

	if err := s.orderRepo.CreateOrderTx(ctx, order, initialPayment); err != nil {
		log.Error().Err(err).Str("component", "orderService.Create").Msg("failed to execute create order transaction")
		return nil, fmt.Errorf("orderService.Create: %w", err)
	}

	// Attach created payment into response if present
	if initialPayment != nil {
		order.Payments = []model.OrderPayment{*initialPayment}
	}

	res := toOrderResponse(order, userRole)

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(userID),
			Action:     "CREATE",
			EntityType: "order",
			EntityID:   &order.ID,
			NewValue:   order,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "orderService.Create").Msg("audit log: failed to record entry")
		}
	}

	return res, nil
}

func (s *orderService) GetByID(ctx context.Context, id uint64, userRole string) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "orderService.GetByID").Msg("failed to find order")
		return nil, fmt.Errorf("orderService.GetByID: %w", err)
	}
	if order == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "Order not found")
	}

	return toOrderResponse(order, userRole), nil
}

func (s *orderService) UpdateStatus(ctx context.Context, id uint64, newStatus string, actorID uint64, ipAddress string) error {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "orderService.UpdateStatus").Msg("failed to find order")
		return fmt.Errorf("orderService.UpdateStatus: %w", err)
	}
	if order == nil {
		return response.NewServiceError(response.ErrNotFound, "Order not found")
	}

	// Terminal states cannot be changed
	if order.Status == model.OrderStatusCompleted || order.Status == model.OrderStatusCancelled {
		return response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("cannot change status of %s order", order.Status))
	}

	// Validate status transition
	if !isValidStatusTransition(order.Status, newStatus) {
		return response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("invalid status transition from %s to %s", order.Status, newStatus))
	}

	oldVal := *order

	if err := s.orderRepo.UpdateStatus(ctx, id, newStatus); err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "orderService.UpdateStatus").Msg("failed to update status")
		return fmt.Errorf("orderService.UpdateStatus: %w", err)
	}

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "UPDATE_STATUS",
			EntityType: "order",
			EntityID:   &id,
			OldValue:   oldVal.Status,
			NewValue:   newStatus,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "orderService.UpdateStatus").Msg("audit log: failed to record entry")
		}
	}

	return nil
}

func isValidStatusTransition(current, next string) bool {
	if current == next {
		return true
	}

	switch current {
	case model.OrderStatusDraft:
		return next == model.OrderStatusDPPaid || next == model.OrderStatusPaid || next == model.OrderStatusCancelled
	case model.OrderStatusDPPaid:
		return next == model.OrderStatusPaid || next == model.OrderStatusCancelled
	case model.OrderStatusPaid:
		return next == model.OrderStatusReady || next == model.OrderStatusCompleted || next == model.OrderStatusCancelled
	case model.OrderStatusReady:
		return next == model.OrderStatusCompleted || next == model.OrderStatusCancelled
	default:
		return false
	}
}
