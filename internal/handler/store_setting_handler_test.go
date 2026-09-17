package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/ikkromm18/tokomakanan/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockStoreSettingService struct {
	mock.Mock
}

func (m *mockStoreSettingService) Get(ctx context.Context) (*dto.StoreSettingResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StoreSettingResponse), args.Error(1)
}

func (m *mockStoreSettingService) Update(ctx context.Context, actorID uint64, req dto.UpdateStoreSettingRequest, ipAddress string) (*dto.StoreSettingResponse, error) {
	args := m.Called(ctx, actorID, req, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StoreSettingResponse), args.Error(1)
}

func strPtr(s string) *string {
	return &s
}

func setupStoreSettingRouter(h *handler.StoreSettingHandler, userRole string, userID uint64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1")
	if userRole != "" || userID > 0 {
		api.Use(func(c *gin.Context) {
			if userRole != "" {
				c.Set("user_role", userRole)
			}
			if userID > 0 {
				c.Set("user_id", userID)
			}
			c.Next()
		})
	}

	settings := api.Group("/store-settings")
	{
		// GET: all authenticated roles
		settings.GET("", middleware.RequireRoles("superadmin", "owner", "admin"), h.Get)
		// PUT: superadmin and owner only
		settings.PUT("", middleware.RequireRoles("superadmin", "owner"), h.Update)
	}
	return r
}

// ---------------- RBAC Tests ----------------

func TestStoreSettingHandler_RBAC_Get_SuperadminAllowed(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "superadmin", 1)

	svc.On("Get", mock.Anything).Return(&dto.StoreSettingResponse{ID: 1, Name: "Toko Makanan"}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/store-settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStoreSettingHandler_RBAC_Get_OwnerAllowed(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "owner", 2)

	svc.On("Get", mock.Anything).Return(&dto.StoreSettingResponse{ID: 1, Name: "Toko Makanan"}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/store-settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStoreSettingHandler_RBAC_Get_AdminAllowed(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "admin", 3)

	svc.On("Get", mock.Anything).Return(&dto.StoreSettingResponse{ID: 1, Name: "Toko Makanan"}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/store-settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStoreSettingHandler_RBAC_Get_UnauthenticatedForbidden(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "", 0)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/store-settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStoreSettingHandler_RBAC_Put_SuperadminAllowed(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "superadmin", 1)

	svc.On("Update", mock.Anything, uint64(1), mock.Anything, mock.Anything).Return(&dto.StoreSettingResponse{ID: 1, Name: "New Bakery"}, nil)

	body := []byte(`{"name":"New Bakery"}`)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/store-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStoreSettingHandler_RBAC_Put_OwnerAllowed(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "owner", 2)

	svc.On("Update", mock.Anything, uint64(2), mock.Anything, mock.Anything).Return(&dto.StoreSettingResponse{ID: 1, Name: "New Bakery"}, nil)

	body := []byte(`{"name":"New Bakery"}`)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/store-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStoreSettingHandler_RBAC_Put_AdminForbidden(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "admin", 3)

	body := []byte(`{"name":"New Bakery"}`)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/store-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStoreSettingHandler_RBAC_Put_UnauthenticatedForbidden(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "", 0)

	body := []byte(`{"name":"New Bakery"}`)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/store-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---------------- Functional / Endpoint Tests ----------------

func TestStoreSettingHandler_Get_Success(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "superadmin", 1)

	expected := &dto.StoreSettingResponse{
		ID:        1,
		Name:      "Toko Makanan",
		Address:   strPtr("Jl. Sudirman 10"),
		UpdatedAt: time.Now(),
	}
	svc.On("Get", mock.Anything).Return(expected, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/store-settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Store settings retrieved successfully", resp.Message)
}

func TestStoreSettingHandler_Get_ServiceError(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "superadmin", 1)

	svc.On("Get", mock.Anything).Return(nil, errors.New("db error"))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/store-settings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStoreSettingHandler_Update_Success(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "superadmin", 1)

	expected := &dto.StoreSettingResponse{
		ID:      1,
		Name:    "Toko Baru",
		Address: strPtr("Jl. Baru 20"),
	}
	svc.On("Update", mock.Anything, uint64(1), mock.MatchedBy(func(r dto.UpdateStoreSettingRequest) bool {
		return r.Name == "Toko Baru" && r.Address != nil && *r.Address == "Jl. Baru 20"
	}), mock.Anything).Return(expected, nil)

	body := []byte(`{"name":"Toko Baru","address":"Jl. Baru 20","logo_url":"https://example.com/logo.png"}`)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/store-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Store settings updated successfully", resp.Message)
}

func TestStoreSettingHandler_Update_ValidationError_NameTooShort(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "superadmin", 1)

	body := []byte(`{"name":"A"}`)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/store-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStoreSettingHandler_Update_ValidationError_NameMissing(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "superadmin", 1)

	body := []byte(`{"address":"Jl. Baru"}`)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/store-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStoreSettingHandler_Update_ValidationError_InvalidURL(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "superadmin", 1)

	body := []byte(`{"name":"Valid Name","logo_url":"not-a-valid-url"}`)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/store-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStoreSettingHandler_Update_ServiceError(t *testing.T) {
	svc := new(mockStoreSettingService)
	h := handler.NewStoreSettingHandler(svc)
	router := setupStoreSettingRouter(h, "superadmin", 1)

	svc.On("Update", mock.Anything, uint64(1), mock.Anything, mock.Anything).Return(nil, errors.New("update error"))

	body := []byte(`{"name":"Toko Baru"}`)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/store-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
