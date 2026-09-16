package handler_test

import (
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
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAuditService struct {
	mock.Mock
}

func (m *mockAuditService) Log(ctx context.Context, entry service.AuditEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockAuditService) List(ctx context.Context, page, limit int, entityType, action string) ([]dto.AuditLogResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, entityType, action)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.AuditLogResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func setupAuditRouter(h *handler.AuditHandler, userRole string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1")
	if userRole != "" {
		api.Use(func(c *gin.Context) {
			c.Set("user_role", userRole)
			c.Next()
		})
	}

	api.GET("/audit-logs", middleware.RequireRoles("superadmin"), h.List)
	return r
}

func TestAuditHandler_List_Success(t *testing.T) {
	svc := new(mockAuditService)
	h := handler.NewAuditHandler(svc)
	router := setupAuditRouter(h, "superadmin")

	expectedLogs := []dto.AuditLogResponse{
		{
			ID:         1,
			Action:     "CREATE",
			EntityType: "product",
			CreatedAt:  time.Now(),
		},
	}
	expectedMeta := dto.PaginationMeta{
		Page:       1,
		Limit:      20,
		TotalRows:  1,
		TotalPages: 1,
	}

	svc.On("List", mock.Anything, 1, 20, "product", "CREATE").
		Return(expectedLogs, expectedMeta, nil).Once()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/audit-logs?page=1&limit=20&entity_type=product&action=CREATE", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, int64(1), resp.Meta.TotalRows)

	svc.AssertExpectations(t)

	// Test page=0 and limit=0 fallback to defaults (1 and 20)
	svc.On("List", mock.Anything, 1, 20, "", "").
		Return(expectedLogs, expectedMeta, nil).Once()

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/audit-logs?page=0&limit=0", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	svc.AssertExpectations(t)
}

func TestAuditHandler_List_ServiceError(t *testing.T) {
	svc := new(mockAuditService)
	h := handler.NewAuditHandler(svc)
	router := setupAuditRouter(h, "superadmin")

	svc.On("List", mock.Anything, 1, 20, "", "").
		Return(nil, dto.PaginationMeta{}, errors.New("database failure")).Once()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/audit-logs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)

	svc.AssertExpectations(t)
}

func TestAuditHandler_List_Forbidden_CashierRole(t *testing.T) {
	svc := new(mockAuditService)
	h := handler.NewAuditHandler(svc)
	router := setupAuditRouter(h, "cashier")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/audit-logs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "Forbidden")

	svc.AssertNotCalled(t, "List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuditHandler_List_Forbidden_NoRole(t *testing.T) {
	svc := new(mockAuditService)
	h := handler.NewAuditHandler(svc)
	router := setupAuditRouter(h, "")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/audit-logs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertNotCalled(t, "List", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
