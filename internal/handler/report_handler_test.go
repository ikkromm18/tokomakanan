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

type mockReportService struct {
	mock.Mock
}

func (m *mockReportService) GetSalesReport(ctx context.Context, req dto.ReportFilterRequest) (*dto.SalesReportResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SalesReportResponse), args.Error(1)
}

func (m *mockReportService) GetProfitReport(ctx context.Context, req dto.ReportFilterRequest) (*dto.ProfitReportResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfitReportResponse), args.Error(1)
}

func (m *mockReportService) ExportCSV(ctx context.Context, req dto.ReportFilterRequest) ([]byte, string, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func setupReportRouter(h *handler.ReportHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1/reports")
	api.GET("/sales", h.GetSalesReport)
	api.GET("/profit", h.GetProfitReport)
	api.GET("/export", h.ExportCSV)

	return r
}

func TestReportHandler_GetSalesReport_Success(t *testing.T) {
	svc := new(mockReportService)
	h := handler.NewReportHandler(svc)
	router := setupReportRouter(h)

	filter := dto.ReportFilterRequest{
		StartDate: "2026-09-01",
		EndDate:   "2026-09-17",
	}

	expectedResp := &dto.SalesReportResponse{
		TotalTransactions: 10,
		TotalRevenue:      1000000,
	}

	svc.On("GetSalesReport", mock.Anything, filter).Return(expectedResp, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/reports/sales?start_date=2026-09-01&end_date=2026-09-17", nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}
