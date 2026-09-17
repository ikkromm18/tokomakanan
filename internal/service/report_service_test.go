package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockReportRepo struct {
	mock.Mock
}

func (m *mockReportRepo) GetSalesData(ctx context.Context, startDate, endDate, orderType string) ([]model.Order, error) {
	args := m.Called(ctx, startDate, endDate, orderType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *mockReportRepo) GetProfitDailyData(ctx context.Context, startDate, endDate string) ([]dto.ProfitReportItem, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.ProfitReportItem), args.Error(1)
}

func TestReportService_GetSalesReport_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	svc := service.NewReportService(reportRepo)

	custName := "Budi"
	orders := []model.Order{
		{
			InvoiceNo:   "INV/001",
			CreatedAt:   time.Now(),
			OrderType:   model.OrderTypeDirectSale,
			TotalAmount: 50000,
			TotalPaid:   50000,
			Status:      model.OrderStatusPaid,
			Customer:    &model.Customer{Name: custName},
		},
		{
			InvoiceNo:   "INV/002",
			CreatedAt:   time.Now(),
			OrderType:   model.OrderTypePreOrder,
			TotalAmount: 100000,
			TotalPaid:   50000,
			Status:      model.OrderStatusDPPaid,
		},
	}

	reportRepo.On("GetSalesData", mock.Anything, "2026-09-01", "2026-09-17", "").Return(orders, nil)

	res, err := svc.GetSalesReport(context.Background(), dto.ReportFilterRequest{
		StartDate: "2026-09-01",
		EndDate:   "2026-09-17",
	})

	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int64(2), res.TotalTransactions)
	assert.Equal(t, float64(100000), res.TotalRevenue) // 50000 + 50000 total paid
	assert.Len(t, res.Items, 2)
	assert.Equal(t, "Budi", *res.Items[0].CustomerName)

	reportRepo.AssertExpectations(t)
}

func TestReportService_ExportCSV_Sales_HasUTF8BOM(t *testing.T) {
	reportRepo := new(mockReportRepo)
	svc := service.NewReportService(reportRepo)

	orders := []model.Order{
		{
			InvoiceNo:   "INV/001",
			CreatedAt:   time.Date(2026, 9, 10, 8, 30, 0, 0, time.UTC),
			OrderType:   model.OrderTypeDirectSale,
			TotalAmount: 50000,
			TotalPaid:   50000,
			Status:      model.OrderStatusPaid,
		},
	}

	reportRepo.On("GetSalesData", mock.Anything, "2026-09-01", "2026-09-17", "").Return(orders, nil)

	data, filename, err := svc.ExportCSV(context.Background(), dto.ReportFilterRequest{
		StartDate: "2026-09-01",
		EndDate:   "2026-09-17",
		Type:      "sales",
	})

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(filename, "laporan_penjualan_"))
	require.True(t, len(data) >= 3)
	// Check UTF-8 BOM: 0xEF, 0xBB, 0xBF
	assert.Equal(t, byte(0xEF), data[0])
	assert.Equal(t, byte(0xBB), data[1])
	assert.Equal(t, byte(0xBF), data[2])

	reportRepo.AssertExpectations(t)
}
