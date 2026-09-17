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
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPaymentService struct {
	mock.Mock
}

func (m *mockPaymentService) Create(ctx context.Context, orderID uint64, req dto.CreatePaymentRequest, actorID uint64, ipAddress string) (*dto.PaymentResponse, error) {
	args := m.Called(ctx, orderID, req, actorID, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *mockPaymentService) ListByOrder(ctx context.Context, orderID uint64) ([]dto.PaymentResponse, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.PaymentResponse), args.Error(1)
}

func setupPaymentRouter(h *handler.PaymentHandler, userID uint64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		if userID > 0 {
			c.Set("user_id", userID)
		}
		c.Next()
	})

	api := r.Group("/api/v1/orders/:id/payments")
	api.POST("", h.Create)
	api.GET("", h.ListByOrder)

	return r
}

func TestPaymentHandler_Create_Success(t *testing.T) {
	svc := new(mockPaymentService)
	h := handler.NewPaymentHandler(svc)
	router := setupPaymentRouter(h, 1)

	req := dto.CreatePaymentRequest{
		Amount:        50000,
		PaymentMethod: "CASH",
	}

	expectedResp := &dto.PaymentResponse{
		ID:            10,
		OrderID:       1,
		Amount:        50000,
		PaymentMethod: "CASH",
		PaidAt:        time.Now(),
	}

	svc.On("Create", mock.Anything, uint64(1), req, uint64(1), mock.Anything).Return(expectedResp, nil)

	bodyBytes, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/orders/1/payments", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestPaymentHandler_Create_BusinessRuleError(t *testing.T) {
	svc := new(mockPaymentService)
	h := handler.NewPaymentHandler(svc)
	router := setupPaymentRouter(h, 1)

	req := dto.CreatePaymentRequest{
		Amount:        150000,
		PaymentMethod: "CASH",
	}

	svc.On("Create", mock.Anything, uint64(1), req, uint64(1), mock.Anything).
		Return(nil, response.NewServiceError(response.ErrBusinessRule, "payment amount exceeds remaining bill"))

	bodyBytes, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/orders/1/payments", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	svc.AssertExpectations(t)
}
