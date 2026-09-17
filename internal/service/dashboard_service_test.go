package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDashboardRepo struct {
	mock.Mock
}

func (m *mockDashboardRepo) GetTodaySummary(ctx context.Context, today string) (float64, int64, int64, float64, error) {
	args := m.Called(ctx, today)
	return args.Get(0).(float64), args.Get(1).(int64), args.Get(2).(int64), args.Get(3).(float64), args.Error(4)
}

func (m *mockDashboardRepo) GetSalesChart(ctx context.Context, startDate, endDate string) ([]dto.SalesChartPoint, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.SalesChartPoint), args.Error(1)
}

func (m *mockDashboardRepo) GetPOReminders(ctx context.Context, date string) ([]model.Order, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Order), args.Error(1)
}

func TestDashboardService_GetSummary_AdminRole_HidesGrossProfit(t *testing.T) {
	dashRepo := new(mockDashboardRepo)
	svc := service.NewDashboardService(dashRepo)

	today := time.Now().Format("2006-01-02")
	dashRepo.On("GetTodaySummary", mock.Anything, today).
		Return(float64(500000), int64(15), int64(3), float64(150000), nil)

	res, err := svc.GetSummary(context.Background(), "admin")
	assert.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, float64(500000), res.TodayRevenue)
	assert.Equal(t, int64(15), res.TodayTransactions)
	assert.Equal(t, int64(3), res.ActivePreOrders)
	assert.Nil(t, res.TodayGrossProfit, "Gross profit MUST be hidden for admin role")

	dashRepo.AssertExpectations(t)
}

func TestDashboardService_GetSummary_SuperadminRole_ShowsGrossProfit(t *testing.T) {
	dashRepo := new(mockDashboardRepo)
	svc := service.NewDashboardService(dashRepo)

	today := time.Now().Format("2006-01-02")
	dashRepo.On("GetTodaySummary", mock.Anything, today).
		Return(float64(500000), int64(15), int64(3), float64(150000), nil)

	res, err := svc.GetSummary(context.Background(), "superadmin")
	assert.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, float64(500000), res.TodayRevenue)
	assert.Equal(t, int64(15), res.TodayTransactions)
	assert.Equal(t, int64(3), res.ActivePreOrders)
	require.NotNil(t, res.TodayGrossProfit)
	assert.Equal(t, float64(150000), *res.TodayGrossProfit)

	dashRepo.AssertExpectations(t)
}
