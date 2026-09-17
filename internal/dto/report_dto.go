package dto

import "time"

type ReportFilterRequest struct {
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
	OrderType string `form:"order_type"`
	Type      string `form:"type"` // "sales" or "profit" for export
}

type SalesReportItem struct {
	InvoiceNo    string    `json:"invoice_no"`
	Date         time.Time `json:"date"`
	CustomerName *string   `json:"customer_name,omitempty"`
	OrderType    string    `json:"order_type"`
	TotalAmount  float64   `json:"total_amount"`
	TotalPaid    float64   `json:"total_paid"`
	Status       string    `json:"status"`
}

type SalesReportResponse struct {
	TotalTransactions int64             `json:"total_transactions"`
	TotalRevenue      float64           `json:"total_revenue"`
	Items             []SalesReportItem `json:"items"`
}

type ProfitReportItem struct {
	Date        string  `json:"date"`
	Revenue     float64 `json:"revenue"`
	TotalHPP    float64 `json:"total_hpp"`
	GrossProfit float64 `json:"gross_profit"`
	MarginPct   float64 `json:"margin_pct"`
}

type ProfitReportResponse struct {
	TotalRevenue     float64            `json:"total_revenue"`
	TotalHPP         float64            `json:"total_hpp"`
	TotalGrossProfit float64            `json:"total_gross_profit"`
	AverageMarginPct float64            `json:"average_margin_pct"`
	DailyBreakdown   []ProfitReportItem `json:"daily_breakdown"`
}
