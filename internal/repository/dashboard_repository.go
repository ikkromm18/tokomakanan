package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetTodaySummary(ctx context.Context, today string) (float64, int64, int64, float64, error)
	GetSalesChart(ctx context.Context, startDate, endDate string) ([]dto.SalesChartPoint, error)
	GetPOReminders(ctx context.Context, date string) ([]model.Order, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetTodaySummary(ctx context.Context, today string) (float64, int64, int64, float64, error) {
	// Revenue & Gross Profit from completed / paid orders today
	type summaryResult struct {
		TotalRevenue sql.NullFloat64
		TotalHPP     sql.NullFloat64
		TotalOrders  int64
	}

	var res summaryResult
	err := r.db.WithContext(ctx).
		Table("orders").
		Select("COALESCE(SUM(total_paid), 0) as total_revenue, COALESCE(SUM(total_hpp), 0) as total_hpp, COUNT(id) as total_orders").
		Where("DATE(created_at) = ? AND status != ?", today, model.OrderStatusCancelled).
		Scan(&res).Error
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("dashboardRepository.GetTodaySummary metrics: %w", err)
	}

	revenue := res.TotalRevenue.Float64
	grossProfit := revenue - res.TotalHPP.Float64

	// Active pre-orders (not COMPLETED and not CANCELLED)
	var activePreOrders int64
	err = r.db.WithContext(ctx).
		Table("orders").
		Where("order_type = ? AND status NOT IN (?, ?)", model.OrderTypePreOrder, model.OrderStatusCompleted, model.OrderStatusCancelled).
		Count(&activePreOrders).Error
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("dashboardRepository.GetTodaySummary preorders: %w", err)
	}

	return revenue, res.TotalOrders, activePreOrders, grossProfit, nil
}

func (r *dashboardRepository) GetSalesChart(ctx context.Context, startDate, endDate string) ([]dto.SalesChartPoint, error) {
	type rowResult struct {
		Date        string
		TotalSales  float64
		TotalOrders int64
	}

	var rows []rowResult
	err := r.db.WithContext(ctx).
		Table("orders").
		Select("DATE(created_at) as date, COALESCE(SUM(total_paid), 0) as total_sales, COUNT(id) as total_orders").
		Where("DATE(created_at) >= ? AND DATE(created_at) <= ? AND status != ?", startDate, endDate, model.OrderStatusCancelled).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetSalesChart: %w", err)
	}

	points := make([]dto.SalesChartPoint, len(rows))
	for i, row := range rows {
		points[i] = dto.SalesChartPoint{
			Date:        row.Date,
			TotalSales:  row.TotalSales,
			TotalOrders: row.TotalOrders,
		}
	}

	return points, nil
}

func (r *dashboardRepository) GetPOReminders(ctx context.Context, date string) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("User").
		Preload("Items").
		Where("order_type = ? AND pickup_date = ? AND status NOT IN (?, ?)", model.OrderTypePreOrder, date, model.OrderStatusCompleted, model.OrderStatusCancelled).
		Order("id ASC").
		Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("dashboardRepository.GetPOReminders: %w", err)
	}
	return orders, nil
}
