package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/rs/zerolog/log"
)

type DashboardService interface {
	GetSummary(ctx context.Context, userRole string) (*dto.DashboardSummaryResponse, error)
	GetChart(ctx context.Context, startDate, endDate string) ([]dto.SalesChartPoint, error)
	GetPOReminders(ctx context.Context, userRole string) ([]dto.OrderResponse, error)
}

type dashboardService struct {
	dashboardRepo repository.DashboardRepository
}

func NewDashboardService(dashboardRepo repository.DashboardRepository) DashboardService {
	return &dashboardService{dashboardRepo: dashboardRepo}
}

func (s *dashboardService) GetSummary(ctx context.Context, userRole string) (*dto.DashboardSummaryResponse, error) {
	today := time.Now().Format("2006-01-02")
	revenue, orders, activePreOrders, grossProfit, err := s.dashboardRepo.GetTodaySummary(ctx, today)
	if err != nil {
		log.Error().Err(err).Str("component", "dashboardService.GetSummary").Msg("failed to get dashboard summary")
		return nil, fmt.Errorf("dashboardService.GetSummary: %w", err)
	}

	res := &dto.DashboardSummaryResponse{
		TodayRevenue:      revenue,
		TodayTransactions: orders,
		ActivePreOrders:   activePreOrders,
	}

	// Admins cannot see gross profit
	if userRole != model.RoleAdmin {
		res.TodayGrossProfit = &grossProfit
	}

	return res, nil
}

func (s *dashboardService) GetChart(ctx context.Context, startDate, endDate string) ([]dto.SalesChartPoint, error) {
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -6).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	chartPoints, err := s.dashboardRepo.GetSalesChart(ctx, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Str("component", "dashboardService.GetChart").Msg("failed to get sales chart")
		return nil, fmt.Errorf("dashboardService.GetChart: %w", err)
	}

	return chartPoints, nil
}

func (s *dashboardService) GetPOReminders(ctx context.Context, userRole string) ([]dto.OrderResponse, error) {
	today := time.Now().Format("2006-01-02")
	orders, err := s.dashboardRepo.GetPOReminders(ctx, today)
	if err != nil {
		log.Error().Err(err).Str("component", "dashboardService.GetPOReminders").Msg("failed to get PO reminders")
		return nil, fmt.Errorf("dashboardService.GetPOReminders: %w", err)
	}

	responses := make([]dto.OrderResponse, len(orders))
	for i := range orders {
		responses[i] = *toOrderResponse(&orders[i], userRole)
	}

	return responses, nil
}
