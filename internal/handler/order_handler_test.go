package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOrderService struct {
	mock.Mock
}

func (m *mockOrderService) List(ctx context.Context, page, limit int, status, orderType, startDate, endDate string, userRole string) ([]dto.OrderResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, status, orderType, startDate, endDate, userRole)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.OrderResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func (m *mockOrderService) Create(ctx context.Context, userID uint64, userRole string, req dto.CreateOrderRequest, ipAddress string) (*dto.OrderResponse, error) {
	args := m.Called(ctx, userID, userRole, req, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *mockOrderService) GetByID(ctx context.Context, id uint64, userRole string) (*dto.OrderResponse, error) {
	args := m.Called(ctx, id, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *mockOrderService) UpdateStatus(ctx context.Context, id uint64, newStatus string, actorID uint64, ipAddress string) error {
	args := m.Called(ctx, id, newStatus, actorID, ipAddress)
	return args.Error(0)
}

func setupOrderRouter(h *handler.OrderHandler, userID uint64, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		if userID > 0 {
			c.Set("user_id", userID)
		}
		if role != "" {
			c.Set("role", role)
		}
		c.Next()
	})

	api := r.Group("/api/v1/orders")
	api.GET("", h.List)
	api.POST("", h.Create)
	api.GET("/:id", h.GetByID)
	api.PATCH("/:id/status", h.UpdateStatus)

	return r
}

func TestOrderHandler_Create_Success(t *testing.T) {
	svc := new(mockOrderService)
	h := handler.NewOrderHandler(svc)
	router := setupOrderRouter(h, 1, "superadmin")

	req := dto.CreateOrderRequest{
		OrderType:     model.OrderTypeDirectSale,
		PaymentMethod: model.PaymentMethodCash,
		InitialPaid:   20000,
		Items: []dto.CreateOrderItemRequest{
			{ItemType: model.ItemTypeProduct, ItemID: 1, Quantity: 1},
		},
	}

	expectedResp := &dto.OrderResponse{
		ID:           1,
		InvoiceNo:    "INV/20260917/00001",
		InvoiceToken: "uuid-1234",
		Status:       model.OrderStatusPaid,
		TotalAmount:  20000,
		TotalPaid:    20000,
		CreatedAt:    time.Now(),
	}

	svc.On("Create", mock.Anything, uint64(1), "superadmin", req, mock.Anything).Return(expectedResp, nil)

	bodyBytes, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestOrderHandler_UpdateStatus_BusinessRule(t *testing.T) {
	svc := new(mockOrderService)
	h := handler.NewOrderHandler(svc)
	router := setupOrderRouter(h, 1, "superadmin")

	req := dto.UpdateOrderStatusRequest{
		Status: model.OrderStatusCancelled,
	}

	svc.On("UpdateStatus", mock.Anything, uint64(5), model.OrderStatusCancelled, uint64(1), mock.Anything).
		Return(response.NewServiceError(response.ErrBusinessRule, "cannot change status of COMPLETED order"))

	bodyBytes, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPatch, "/api/v1/orders/5/status", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	svc.AssertExpectations(t)
}
