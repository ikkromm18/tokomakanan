package dto

type DashboardSummaryResponse struct {
	TodayRevenue      float64  `json:"today_revenue"`
	TodayTransactions int64    `json:"today_transactions"`
	ActivePreOrders   int64    `json:"active_pre_orders"`
	TodayGrossProfit  *float64 `json:"today_gross_profit,omitempty"`
}

type SalesChartPoint struct {
	Date        string  `json:"date"`
	TotalSales  float64 `json:"total_sales"`
	TotalOrders int64   `json:"total_orders"`
}
