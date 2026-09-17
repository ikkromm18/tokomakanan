package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockStoreSettingRepoForPublic struct {
	mock.Mock
}

func (m *mockStoreSettingRepoForPublic) Get(ctx context.Context) (*model.StoreSetting, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StoreSetting), args.Error(1)
}

func (m *mockStoreSettingRepoForPublic) Update(ctx context.Context, setting *model.StoreSetting) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

func TestPublicService_GetInvoiceByToken_Success(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	settingRepo := new(mockStoreSettingRepoForPublic)

	svc := service.NewPublicService(orderRepo, settingRepo)

	token := "4c13a2d0-9999-4e78-bc45-989f66e01a8f"

	store := &model.StoreSetting{
		ID:   1,
		Name: "Bakery Delight",
	}

	order := &model.Order{
		ID:           100,
		InvoiceNo:    "INV/20260917/00001",
		InvoiceToken: token,
		OrderType:    model.OrderTypeDirectSale,
		Status:       model.OrderStatusPaid,
		TotalAmount:  50000,
		TotalHPP:     30000, // Must NOT be in the public response struct
		TotalPaid:    50000,
		User: &model.User{
			Name: "Kasir Sarah",
		},
		Items: []model.OrderItem{
			{
				ItemName:  "Roti Coklat",
				Quantity:  2,
				UnitPrice: 25000,
				UnitHPP:   15000, // Must NOT be in public response
				Subtotal:  50000,
			},
		},
		CreatedAt: time.Now(),
	}

	orderRepo.On("FindByToken", mock.Anything, token).Return(order, nil)
	settingRepo.On("Get", mock.Anything).Return(store, nil)

	res, err := svc.GetInvoiceByToken(context.Background(), token)
	assert.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "Bakery Delight", res.Store.Name)
	assert.Equal(t, "INV/20260917/00001", res.InvoiceNo)
	assert.Equal(t, "Kasir Sarah", res.CashierName)
	assert.Equal(t, float64(50000), res.TotalAmount)
	assert.Len(t, res.Items, 1)
	assert.Equal(t, "Roti Coklat", res.Items[0].ItemName)

	orderRepo.AssertExpectations(t)
	settingRepo.AssertExpectations(t)
}

func TestPublicService_GetInvoiceByToken_NotFound(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	settingRepo := new(mockStoreSettingRepoForPublic)

	svc := service.NewPublicService(orderRepo, settingRepo)

	orderRepo.On("FindByToken", mock.Anything, "non-existent-token").Return(nil, nil)

	res, err := svc.GetInvoiceByToken(context.Background(), "non-existent-token")
	assert.Error(t, err)
	assert.Nil(t, res)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrNotFound)
	assert.Contains(t, sErr.Message, "invoice not found")
}
