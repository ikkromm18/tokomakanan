package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPaymentRepo struct {
	mock.Mock
}

func (m *mockPaymentRepo) CreatePaymentTx(ctx context.Context, payment *model.OrderPayment, newTotalPaid float64, newStatus string) error {
	args := m.Called(ctx, payment, newTotalPaid, newStatus)
	return args.Error(0)
}

func (m *mockPaymentRepo) FindByOrderID(ctx context.Context, orderID uint64) ([]model.OrderPayment, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.OrderPayment), args.Error(1)
}

func TestPaymentService_Create_OverpaymentRejected(t *testing.T) {
	payRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	audit := new(mockAuditServiceForProductSvc)

	svc := service.NewPaymentService(payRepo, orderRepo, audit)

	order := &model.Order{
		ID:          1,
		Status:      model.OrderStatusDPPaid,
		TotalAmount: 100000,
		TotalPaid:   60000, // Remaining is 40000
	}
	orderRepo.On("FindByID", mock.Anything, uint64(1)).Return(order, nil)

	req := dto.CreatePaymentRequest{
		Amount:        50000, // Overpayment: 60000 + 50000 = 110000 > 100000
		PaymentMethod: model.PaymentMethodCash,
	}

	res, err := svc.Create(context.Background(), 1, req, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrBusinessRule)
	assert.Contains(t, sErr.Message, "payment amount exceeds remaining bill")
}

func TestPaymentService_Create_Success_TransitionsToPaid(t *testing.T) {
	payRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	audit := new(mockAuditServiceForProductSvc)

	svc := service.NewPaymentService(payRepo, orderRepo, audit)

	order := &model.Order{
		ID:          1,
		Status:      model.OrderStatusDPPaid,
		TotalAmount: 100000,
		TotalPaid:   60000,
	}
	orderRepo.On("FindByID", mock.Anything, uint64(1)).Return(order, nil)

	req := dto.CreatePaymentRequest{
		Amount:        40000, // Completes payment -> status should become PAID
		PaymentMethod: model.PaymentMethodTransfer,
	}

	payRepo.On("CreatePaymentTx", mock.Anything, mock.MatchedBy(func(p *model.OrderPayment) bool {
		p.ID = 12
		p.PaidAt = time.Now()
		p.CreatedAt = time.Now()
		return p.OrderID == 1 && p.Amount == 40000 && p.PaymentMethod == model.PaymentMethodTransfer
	}), float64(100000), model.OrderStatusPaid).Return(nil)

	audit.On("Log", mock.Anything, mock.MatchedBy(func(e service.AuditEntry) bool {
		return e.Action == "CREATE" && e.EntityType == "order_payment" && *e.EntityID == 12
	})).Return(nil)

	res, err := svc.Create(context.Background(), 1, req, 1, "127.0.0.1")
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(12), res.ID)
	assert.Equal(t, float64(40000), res.Amount)

	orderRepo.AssertExpectations(t)
	payRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}
