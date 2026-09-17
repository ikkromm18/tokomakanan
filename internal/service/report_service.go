package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/rs/zerolog/log"
)

type ReportService interface {
	GetSalesReport(ctx context.Context, req dto.ReportFilterRequest) (*dto.SalesReportResponse, error)
	GetProfitReport(ctx context.Context, req dto.ReportFilterRequest) (*dto.ProfitReportResponse, error)
	ExportCSV(ctx context.Context, req dto.ReportFilterRequest) ([]byte, string, error)
}

type reportService struct {
	reportRepo repository.ReportRepository
}

func NewReportService(reportRepo repository.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

func (s *reportService) GetSalesReport(ctx context.Context, req dto.ReportFilterRequest) (*dto.SalesReportResponse, error) {
	orders, err := s.reportRepo.GetSalesData(ctx, req.StartDate, req.EndDate, req.OrderType)
	if err != nil {
		log.Error().Err(err).Str("component", "reportService.GetSalesReport").Msg("failed to get sales data")
		return nil, fmt.Errorf("reportService.GetSalesReport: %w", err)
	}

	var totalRevenue float64
	items := make([]dto.SalesReportItem, len(orders))
	for i, o := range orders {
		var customerName *string
		if o.Customer != nil {
			customerName = &o.Customer.Name
		}
		totalRevenue += o.TotalPaid
		items[i] = dto.SalesReportItem{
			InvoiceNo:    o.InvoiceNo,
			Date:         o.CreatedAt,
			CustomerName: customerName,
			OrderType:    o.OrderType,
			TotalAmount:  o.TotalAmount,
			TotalPaid:    o.TotalPaid,
			Status:       o.Status,
		}
	}

	return &dto.SalesReportResponse{
		TotalTransactions: int64(len(orders)),
		TotalRevenue:      totalRevenue,
		Items:             items,
	}, nil
}

func (s *reportService) GetProfitReport(ctx context.Context, req dto.ReportFilterRequest) (*dto.ProfitReportResponse, error) {
	dailyBreakdown, err := s.reportRepo.GetProfitDailyData(ctx, req.StartDate, req.EndDate)
	if err != nil {
		log.Error().Err(err).Str("component", "reportService.GetProfitReport").Msg("failed to get profit data")
		return nil, fmt.Errorf("reportService.GetProfitReport: %w", err)
	}

	var totalRevenue float64
	var totalHPP float64
	var totalGrossProfit float64

	for _, d := range dailyBreakdown {
		totalRevenue += d.Revenue
		totalHPP += d.TotalHPP
		totalGrossProfit += d.GrossProfit
	}

	var avgMargin float64
	if totalRevenue > 0 {
		avgMargin = (totalGrossProfit / totalRevenue) * 100
	}

	return &dto.ProfitReportResponse{
		TotalRevenue:     totalRevenue,
		TotalHPP:         totalHPP,
		TotalGrossProfit: totalGrossProfit,
		AverageMarginPct: avgMargin,
		DailyBreakdown:   dailyBreakdown,
	}, nil
}

func (s *reportService) ExportCSV(ctx context.Context, req dto.ReportFilterRequest) ([]byte, string, error) {
	var buf bytes.Buffer
	// Write UTF-8 BOM for Microsoft Excel compatibility
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)
	cleanStart := strings.ReplaceAll(req.StartDate, "-", "")
	cleanEnd := strings.ReplaceAll(req.EndDate, "-", "")

	if req.Type == "profit" {
		profitReport, err := s.GetProfitReport(ctx, req)
		if err != nil {
			return nil, "", err
		}

		_ = w.Write([]string{"Tanggal", "Pendapatan", "Total HPP", "Laba Kotor", "Margin (%)"})
		for _, d := range profitReport.DailyBreakdown {
			_ = w.Write([]string{
				d.Date,
				fmt.Sprintf("%.2f", d.Revenue),
				fmt.Sprintf("%.2f", d.TotalHPP),
				fmt.Sprintf("%.2f", d.GrossProfit),
				fmt.Sprintf("%.2f%%", d.MarginPct),
			})
		}
		_ = w.Write([]string{
			"TOTAL",
			fmt.Sprintf("%.2f", profitReport.TotalRevenue),
			fmt.Sprintf("%.2f", profitReport.TotalHPP),
			fmt.Sprintf("%.2f", profitReport.TotalGrossProfit),
			fmt.Sprintf("%.2f%%", profitReport.AverageMarginPct),
		})
		w.Flush()

		filename := fmt.Sprintf("laporan_laba_%s_%s.csv", cleanStart, cleanEnd)
		return buf.Bytes(), filename, nil
	}

	// Default: sales report
	salesReport, err := s.GetSalesReport(ctx, req)
	if err != nil {
		return nil, "", err
	}

	_ = w.Write([]string{"Nomor Invoice", "Tanggal", "Pelanggan", "Tipe Order", "Total Tagihan", "Total Bayar", "Status"})
	for _, it := range salesReport.Items {
		cust := "-"
		if it.CustomerName != nil {
			cust = *it.CustomerName
		}
		_ = w.Write([]string{
			it.InvoiceNo,
			it.Date.Format("2006-01-02 15:04:05"),
			cust,
			it.OrderType,
			fmt.Sprintf("%.2f", it.TotalAmount),
			fmt.Sprintf("%.2f", it.TotalPaid),
			it.Status,
		})
	}
	_ = w.Write([]string{
		"TOTAL",
		"",
		"",
		"",
		"",
		fmt.Sprintf("%.2f", salesReport.TotalRevenue),
		fmt.Sprintf("%d transaksi", salesReport.TotalTransactions),
	})
	w.Flush()

	filename := fmt.Sprintf("laporan_penjualan_%s_%s.csv", cleanStart, cleanEnd)
	return buf.Bytes(), filename, nil
}
