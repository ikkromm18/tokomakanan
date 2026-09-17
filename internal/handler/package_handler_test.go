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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPackageService struct {
	mock.Mock
}

func (m *mockPackageService) List(ctx context.Context, page, limit int, isActive *bool, search string, userRole string) ([]dto.PackageResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, isActive, search, userRole)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.PackageResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func (m *mockPackageService) Create(ctx context.Context, req dto.CreatePackageRequest, actorID uint64, userRole string, ipAddress string) (*dto.PackageResponse, error) {
	args := m.Called(ctx, req, actorID, userRole, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PackageResponse), args.Error(1)
}

func (m *mockPackageService) GetByID(ctx context.Context, id uint64, userRole string) (*dto.PackageResponse, error) {
	args := m.Called(ctx, id, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PackageResponse), args.Error(1)
}

func (m *mockPackageService) Update(ctx context.Context, id uint64, req dto.UpdatePackageRequest, actorID uint64, userRole string, ipAddress string) (*dto.PackageResponse, error) {
	args := m.Called(ctx, id, req, actorID, userRole, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PackageResponse), args.Error(1)
}

func (m *mockPackageService) Delete(ctx context.Context, id uint64, actorID uint64, ipAddress string) error {
	args := m.Called(ctx, id, actorID, ipAddress)
	return args.Error(0)
}

func setupPackageRouter(h *handler.PackageHandler, userID uint64, role string) *gin.Engine {
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

	api := r.Group("/api/v1/packages")
	api.GET("", h.List)
	api.POST("", h.Create)
	api.GET("/:id", h.GetByID)
	api.PUT("/:id", h.Update)
	api.DELETE("/:id", h.Delete)

	return r
}

func TestPackageHandler_Create_Success(t *testing.T) {
	svc := new(mockPackageService)
	h := handler.NewPackageHandler(svc)
	router := setupPackageRouter(h, 1, "superadmin")

	req := dto.CreatePackageRequest{
		Name:      "Paket Sarapan Hemat",
		SellPrice: 25000,
		Items: []dto.PackageItemRequest{
			{ProductID: 1, Quantity: 2},
		},
	}

	expectedResp := &dto.PackageResponse{
		ID:        1,
		Name:      "Paket Sarapan Hemat",
		SellPrice: 25000,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	svc.On("Create", mock.Anything, req, uint64(1), "superadmin", mock.Anything).Return(expectedResp, nil)

	bodyBytes, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/packages", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestPackageHandler_Create_ValidationError(t *testing.T) {
	svc := new(mockPackageService)
	h := handler.NewPackageHandler(svc)
	router := setupPackageRouter(h, 1, "superadmin")

	// Missing items (must have at least 1 item)
	body := []byte(`{"name":"Paket Sarapan","sell_price":25000,"items":[]}`)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodPost, "/api/v1/packages", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
