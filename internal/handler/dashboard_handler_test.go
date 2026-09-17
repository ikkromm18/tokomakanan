package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDashboardService struct {
	mock.Mock
}

func (m *mockDashboardService) GetSummary(ctx context.Context, userRole string) (*dto.DashboardSummaryResponse, error) {
	args := m.Called(ctx, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DashboardSummaryResponse), args.Error(1)
}

func (m *mockDashboardService) GetChart(ctx context.Context, startDate, endDate string) ([]dto.SalesChartPoint, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.SalesChartPoint), args.Error(1)
}

func (m *mockDashboardService) GetPOReminders(ctx context.Context, userRole string) ([]dto.OrderResponse, error) {
	args := m.Called(ctx, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.OrderResponse), args.Error(1)
}

func setupDashboardRouter(h *handler.DashboardHandler, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		if role != "" {
			c.Set("role", role)
		}
		c.Next()
	})

	api := r.Group("/api/v1/dashboard")
	api.GET("/summary", h.GetSummary)
	api.GET("/chart", h.GetChart)
	api.GET("/po-reminders", h.GetPOReminders)

	return r
}

func TestDashboardHandler_GetSummary_Success(t *testing.T) {
	svc := new(mockDashboardService)
	h := handler.NewDashboardHandler(svc)
	router := setupDashboardRouter(h, "admin")

	expectedResp := &dto.DashboardSummaryResponse{
		TodayRevenue:      250000,
		TodayTransactions: 10,
		ActivePreOrders:   2,
	}

	svc.On("GetSummary", mock.Anything, "admin").Return(expectedResp, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}
