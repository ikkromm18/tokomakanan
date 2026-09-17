package dto

import "time"

type PublicInvoiceStoreInfo struct {
	Name          string  `json:"name"`
	Address       *string `json:"address,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	LogoURL       *string `json:"logo_url,omitempty"`
	ReceiptFooter *string `json:"receipt_footer,omitempty"`
}

type PublicInvoiceItem struct {
	ItemName  string  `json:"item_name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`
}

type PublicInvoicePayment struct {
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	PaidAt        time.Time `json:"paid_at"`
}

type PublicInvoiceResponse struct {
	Store          PublicInvoiceStoreInfo `json:"store"`
	InvoiceNo      string                 `json:"invoice_no"`
	CustomerName   *string                `json:"customer_name,omitempty"`
	CashierName    string                 `json:"cashier_name"`
	OrderType      string                 `json:"order_type"`
	Status         string                 `json:"status"`
	PickupDate     *string                `json:"pickup_date,omitempty"`
	Notes          *string                `json:"notes,omitempty"`
	Subtotal       float64                `json:"subtotal"`
	DiscountAmount float64                `json:"discount_amount"`
	TotalAmount    float64                `json:"total_amount"`
	TotalPaid      float64                `json:"total_paid"`
	RemainingPaid  float64                `json:"remaining_paid"`
	Items          []PublicInvoiceItem    `json:"items"`
	Payments       []PublicInvoicePayment `json:"payments"`
	CreatedAt      time.Time              `json:"created_at"`
}
