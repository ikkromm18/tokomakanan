package repository

import (
	"context"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"gorm.io/gorm"
)

type ReportRepository interface {
	GetSalesData(ctx context.Context, startDate, endDate, orderType string) ([]model.Order, error)
	GetProfitDailyData(ctx context.Context, startDate, endDate string) ([]dto.ProfitReportItem, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetSalesData(ctx context.Context, startDate, endDate, orderType string) ([]model.Order, error) {
	var orders []model.Order
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("User").
		Where("DATE(created_at) >= ? AND DATE(created_at) <= ? AND status != ?", startDate, endDate, model.OrderStatusCancelled)

	if orderType != "" {
		query = query.Where("order_type = ?", orderType)
	}

	err := query.Order("created_at ASC").Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("reportRepository.GetSalesData: %w", err)
	}

	return orders, nil
}

func (r *reportRepository) GetProfitDailyData(ctx context.Context, startDate, endDate string) ([]dto.ProfitReportItem, error) {
	type rowResult struct {
		Date        string
		Revenue     float64
		TotalHPP    float64
		GrossProfit float64
	}

	var rows []rowResult
	err := r.db.WithContext(ctx).
		Table("orders").
		Select("DATE(created_at) as date, COALESCE(SUM(total_paid), 0) as revenue, COALESCE(SUM(total_hpp), 0) as total_hpp, (COALESCE(SUM(total_paid), 0) - COALESCE(SUM(total_hpp), 0)) as gross_profit").
		Where("DATE(created_at) >= ? AND DATE(created_at) <= ? AND status != ?", startDate, endDate, model.OrderStatusCancelled).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("reportRepository.GetProfitDailyData: %w", err)
	}

	items := make([]dto.ProfitReportItem, len(rows))
	for i, row := range rows {
		var marginPct float64
		if row.Revenue > 0 {
			marginPct = (row.GrossProfit / row.Revenue) * 100
		}
		items[i] = dto.ProfitReportItem{
			Date:        row.Date,
			Revenue:     row.Revenue,
			TotalHPP:    row.TotalHPP,
			GrossProfit: row.GrossProfit,
			MarginPct:   marginPct,
		}
	}

	return items, nil
}
