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

type mockCustomerService struct {
	mock.Mock
}

func (m *mockCustomerService) List(ctx context.Context, page, limit int, search string) ([]dto.CustomerResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, search)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.CustomerResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func (m *mockCustomerService) Create(ctx context.Context, req dto.CreateCustomerRequest, actorID uint64, ipAddress string) (*dto.CustomerResponse, error) {
	args := m.Called(ctx, req, actorID, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CustomerResponse), args.Error(1)
}

func (m *mockCustomerService) GetByID(ctx context.Context, id uint64) (*dto.CustomerResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CustomerResponse), args.Error(1)
}

func (m *mockCustomerService) Update(ctx context.Context, id uint64, req dto.UpdateCustomerRequest, actorID uint64, ipAddress string) (*dto.CustomerResponse, error) {
	args := m.Called(ctx, id, req, actorID, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CustomerResponse), args.Error(1)
}

func setupCustomerRouter(h *handler.CustomerHandler, userID uint64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		if userID > 0 {
			c.Set("user_id", userID)
		}
		c.Next()
	})

	api := r.Group("/api/v1/customers")
	api.GET("", h.List)
	api.POST("", h.Create)
	api.GET("/:id", h.GetByID)
	api.PUT("/:id", h.Update)

	return r
}

func TestCustomerHandler_Create_Success(t *testing.T) {
	svc := new(mockCustomerService)
	h := handler.NewCustomerHandler(svc)
	router := setupCustomerRouter(h, 1)

	req := dto.CreateCustomerRequest{
		Name:  "Jane Doe",
		Phone: "08987654321",
	}

	expectedResp := &dto.CustomerResponse{
		ID:        1,
		Name:      "Jane Doe",
		Phone:     "08987654321",
		CreatedAt: time.Now(),
	}

	svc.On("Create", mock.Anything, req, uint64(1), mock.Anything).Return(expectedResp, nil)

	bodyBytes, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestCustomerHandler_Create_DuplicatePhone(t *testing.T) {
	svc := new(mockCustomerService)
	h := handler.NewCustomerHandler(svc)
	router := setupCustomerRouter(h, 1)

	req := dto.CreateCustomerRequest{
		Name:  "Jane Doe",
		Phone: "08987654321",
	}

	svc.On("Create", mock.Anything, req, uint64(1), mock.Anything).
		Return(nil, response.NewServiceError(response.ErrDuplicate, "phone number already registered"))

	bodyBytes, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/customers", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusConflict, w.Code)
	svc.AssertExpectations(t)
}
