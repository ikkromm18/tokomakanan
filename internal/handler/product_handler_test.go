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

type mockProductService struct {
	mock.Mock
}

func (m *mockProductService) List(ctx context.Context, page, limit int, category string, isActive *bool, search string, userRole string) ([]dto.ProductResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, category, isActive, search, userRole)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.ProductResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func (m *mockProductService) Create(ctx context.Context, req dto.CreateProductRequest, actorID uint64, userRole string, ipAddress string) (*dto.ProductResponse, error) {
	args := m.Called(ctx, req, actorID, userRole, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponse), args.Error(1)
}

func (m *mockProductService) GetByID(ctx context.Context, id uint64, userRole string) (*dto.ProductResponse, error) {
	args := m.Called(ctx, id, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponse), args.Error(1)
}

func (m *mockProductService) Update(ctx context.Context, id uint64, req dto.UpdateProductRequest, actorID uint64, userRole string, ipAddress string) (*dto.ProductResponse, error) {
	args := m.Called(ctx, id, req, actorID, userRole, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponse), args.Error(1)
}

func (m *mockProductService) Delete(ctx context.Context, id uint64, actorID uint64, ipAddress string) error {
	args := m.Called(ctx, id, actorID, ipAddress)
	return args.Error(0)
}

func setupProductRouter(h *handler.ProductHandler, userID uint64, role string) *gin.Engine {
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

	api := r.Group("/api/v1/products")
	api.GET("", h.List)
	api.POST("", h.Create)
	api.GET("/:id", h.GetByID)
	api.PUT("/:id", h.Update)
	api.DELETE("/:id", h.Delete)

	return r
}

func TestProductHandler_Create_Success(t *testing.T) {
	svc := new(mockProductService)
	h := handler.NewProductHandler(svc)
	router := setupProductRouter(h, 1, "superadmin")

	req := dto.CreateProductRequest{
		Name:      "Baguette",
		Category:  "Bakery",
		HPP:       12000,
		SellPrice: 20000,
	}

	expectedResp := &dto.ProductResponse{
		ID:        1,
		Name:      "Baguette",
		Category:  "Bakery",
		SellPrice: 20000,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	svc.On("Create", mock.Anything, req, uint64(1), "superadmin", mock.Anything).Return(expectedResp, nil)

	bodyBytes, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestProductHandler_Create_ValidationError(t *testing.T) {
	svc := new(mockProductService)
	h := handler.NewProductHandler(svc)
	router := setupProductRouter(h, 1, "superadmin")

	// Missing Name
	body := []byte(`{"category":"Bakery","hpp":1000,"sell_price":2000}`)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductHandler_Create_BusinessRuleError(t *testing.T) {
	svc := new(mockProductService)
	h := handler.NewProductHandler(svc)
	router := setupProductRouter(h, 1, "superadmin")

	req := dto.CreateProductRequest{
		Name:      "Baguette",
		Category:  "Bakery",
		HPP:       20000,
		SellPrice: 15000,
	}

	svc.On("Create", mock.Anything, req, uint64(1), "superadmin", mock.Anything).
		Return(nil, response.NewServiceError(response.ErrBusinessRule, "sell price cannot be less than HPP"))

	bodyBytes, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	svc.AssertExpectations(t)
}

func TestProductHandler_GetByID_NotFound(t *testing.T) {
	svc := new(mockProductService)
	h := handler.NewProductHandler(svc)
	router := setupProductRouter(h, 1, "admin")

	svc.On("GetByID", mock.Anything, uint64(999), "admin").
		Return(nil, response.NewServiceError(response.ErrNotFound, "Product not found"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/products/999", nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestProductHandler_Delete_BusinessRule(t *testing.T) {
	svc := new(mockProductService)
	h := handler.NewProductHandler(svc)
	router := setupProductRouter(h, 1, "superadmin")

	svc.On("Delete", mock.Anything, uint64(5), uint64(1), mock.Anything).
		Return(response.NewServiceError(response.ErrBusinessRule, "cannot delete product: product is used in active package"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodDelete, "/api/v1/products/5", nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	svc.AssertExpectations(t)
}
