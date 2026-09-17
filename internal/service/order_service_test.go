package service_test

import (
	"context"
	"testing"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockOrderRepo struct {
	mock.Mock
}

func (m *mockOrderRepo) FindByID(ctx context.Context, id uint64) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockOrderRepo) FindByToken(ctx context.Context, token string) (*model.Order, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockOrderRepo) CreateOrderTx(ctx context.Context, order *model.Order, initialPayment *model.OrderPayment) error {
	args := m.Called(ctx, order, initialPayment)
	return args.Error(0)
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id uint64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockOrderRepo) UpdateTotalPaidAndStatus(ctx context.Context, orderID uint64, newTotalPaid float64, newStatus string) error {
	args := m.Called(ctx, orderID, newTotalPaid, newStatus)
	return args.Error(0)
}

func (m *mockOrderRepo) FindAll(ctx context.Context, page, limit int, status, orderType, startDate, endDate string) ([]model.Order, int64, error) {
	args := m.Called(ctx, page, limit, status, orderType, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]model.Order), args.Get(1).(int64), args.Error(2)
}

func TestOrderService_DirectSale_InitialPaidLessThanTotal(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	prodRepo := new(mockProductRepo)
	pkgRepo := new(mockPackageRepo)
	audit := new(mockAuditServiceForProductSvc)

	svc := service.NewOrderService(orderRepo, prodRepo, pkgRepo, audit)

	req := dto.CreateOrderRequest{
		OrderType:     model.OrderTypeDirectSale,
		PaymentMethod: model.PaymentMethodCash,
		InitialPaid:   10000, // Product sell price is 15000, underpayment for direct sale
		Items: []dto.CreateOrderItemRequest{
			{ItemType: model.ItemTypeProduct, ItemID: 1, Quantity: 1},
		},
	}

	prod := &model.Product{
		ID:        1,
		Name:      "Roti Tawar",
		HPP:       10000,
		SellPrice: 15000,
		IsActive:  true,
	}
	prodRepo.On("FindByID", mock.Anything, uint64(1)).Return(prod, nil)

	res, err := svc.Create(context.Background(), 1, "superadmin", req, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrBusinessRule)
	assert.Contains(t, sErr.Message, "direct sale requires full payment")
}

func TestOrderService_PreOrder_RequiresPickupDate(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	prodRepo := new(mockProductRepo)
	pkgRepo := new(mockPackageRepo)
	audit := new(mockAuditServiceForProductSvc)

	svc := service.NewOrderService(orderRepo, prodRepo, pkgRepo, audit)

	req := dto.CreateOrderRequest{
		OrderType:     model.OrderTypePreOrder,
		PickupDate:    nil, // Missing pickup date
		PaymentMethod: model.PaymentMethodCash,
		InitialPaid:   0,
		Items: []dto.CreateOrderItemRequest{
			{ItemType: model.ItemTypeProduct, ItemID: 1, Quantity: 1},
		},
	}

	res, err := svc.Create(context.Background(), 1, "superadmin", req, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrBusinessRule)
	assert.Contains(t, sErr.Message, "pickup date is required for pre-order")
}

func TestOrderService_StatusTransition_Enforced(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	prodRepo := new(mockProductRepo)
	pkgRepo := new(mockPackageRepo)
	audit := new(mockAuditServiceForProductSvc)

	svc := service.NewOrderService(orderRepo, prodRepo, pkgRepo, audit)

	existingOrder := &model.Order{
		ID:        1,
		Status:    model.OrderStatusCompleted, // Terminal state
		OrderType: model.OrderTypeDirectSale,
	}
	orderRepo.On("FindByID", mock.Anything, uint64(1)).Return(existingOrder, nil)

	err := svc.UpdateStatus(context.Background(), 1, model.OrderStatusCancelled, 1, "127.0.0.1")
	assert.Error(t, err)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrBusinessRule)
	assert.Contains(t, sErr.Message, "cannot change status of COMPLETED order")
}
