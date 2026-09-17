package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductHandler_List_And_Update(t *testing.T) {
	svc := new(mockProductService)
	h := handler.NewProductHandler(svc)
	router := setupProductRouter(h, 1, "owner")

	// List
	svc.On("List", mock.Anything, 1, 20, "Bakery", mock.Anything, "", "owner").
		Return([]dto.ProductResponse{{ID: 1, Name: "Roti"}}, dto.PaginationMeta{Page: 1, Limit: 20, TotalRows: 1}, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/products?category=Bakery", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update
	svc.On("Update", mock.Anything, uint64(1), mock.Anything, uint64(1), "owner", mock.Anything).
		Return(&dto.ProductResponse{ID: 1, Name: "Roti Baru"}, nil)

	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest(http.MethodPut, "/api/v1/products/1", bytes.NewBufferString(`{"name":"Roti Baru","category":"Bakery","hpp":6000,"sell_price":12000,"is_active":true}`))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPackageHandler_List_GetByID_Update_Delete(t *testing.T) {
	svc := new(mockPackageService)
	h := handler.NewPackageHandler(svc)
	router := setupPackageRouter(h, 1, "owner")

	// List
	svc.On("List", mock.Anything, 1, 20, mock.Anything, "", "owner").
		Return([]dto.PackageResponse{{ID: 1, Name: "Paket"}}, dto.PaginationMeta{TotalRows: 1}, nil)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/packages", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// GetByID
	svc.On("GetByID", mock.Anything, uint64(1), "owner").
		Return(&dto.PackageResponse{ID: 1, Name: "Paket"}, nil)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest(http.MethodGet, "/api/v1/packages/1", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update
	svc.On("Update", mock.Anything, uint64(1), mock.Anything, uint64(1), "owner", mock.Anything).
		Return(&dto.PackageResponse{ID: 1, Name: "Paket"}, nil)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest(http.MethodPut, "/api/v1/packages/1", bytes.NewBufferString(`{"name":"Paket","sell_price":25000,"is_active":true,"items":[{"product_id":1,"quantity":2}]}`))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete
	svc.On("Delete", mock.Anything, uint64(1), uint64(1), mock.Anything).Return(nil)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest(http.MethodDelete, "/api/v1/packages/1", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomerHandler_List_GetByID_Update(t *testing.T) {
	svc := new(mockCustomerService)
	h := handler.NewCustomerHandler(svc)
	router := setupCustomerRouter(h, 1)

	// List
	svc.On("List", mock.Anything, 1, 20, "").
		Return([]dto.CustomerResponse{{ID: 1, Name: "Budi"}}, dto.PaginationMeta{TotalRows: 1}, nil)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// GetByID
	svc.On("GetByID", mock.Anything, uint64(1)).Return(&dto.CustomerResponse{ID: 1, Name: "Budi"}, nil)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest(http.MethodGet, "/api/v1/customers/1", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update
	svc.On("Update", mock.Anything, uint64(1), mock.Anything, uint64(1), mock.Anything).
		Return(&dto.CustomerResponse{ID: 1, Name: "Budi Baru"}, nil)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest(http.MethodPut, "/api/v1/customers/1", bytes.NewBufferString(`{"name":"Budi Baru","phone":"08111111"}`))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrderHandler_List_GetByID(t *testing.T) {
	svc := new(mockOrderService)
	h := handler.NewOrderHandler(svc)
	router := setupOrderRouter(h, 1, "admin")

	svc.On("List", mock.Anything, 1, 20, "", "", "", "", "admin").
		Return([]dto.OrderResponse{{ID: 1, InvoiceNo: "INV/01"}}, dto.PaginationMeta{TotalRows: 1}, nil)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	svc.On("GetByID", mock.Anything, uint64(1), "admin").
		Return(&dto.OrderResponse{ID: 1, InvoiceNo: "INV/01"}, nil)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest(http.MethodGet, "/api/v1/orders/1", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPaymentHandler_ListByOrder(t *testing.T) {
	svc := new(mockPaymentService)
	h := handler.NewPaymentHandler(svc)
	router := setupPaymentRouter(h, 1)

	svc.On("ListByOrder", mock.Anything, uint64(1)).
		Return([]dto.PaymentResponse{{ID: 1, Amount: 10000}}, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/orders/1/payments", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDashboardHandler_Chart_And_Reminders(t *testing.T) {
	svc := new(mockDashboardService)
	h := handler.NewDashboardHandler(svc)
	router := setupDashboardRouter(h, "admin")

	svc.On("GetChart", mock.Anything, "", "").Return([]dto.SalesChartPoint{{Date: "2026-09-17"}}, nil)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/chart", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	svc.On("GetPOReminders", mock.Anything, "admin").Return([]dto.OrderResponse{{ID: 1}}, nil)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest(http.MethodGet, "/api/v1/dashboard/po-reminders", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_Profit_And_Export(t *testing.T) {
	svc := new(mockReportService)
	h := handler.NewReportHandler(svc)
	router := setupReportRouter(h)

	svc.On("GetProfitReport", mock.Anything, mock.Anything).
		Return(&dto.ProfitReportResponse{TotalRevenue: 100000}, nil)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest(http.MethodGet, "/api/v1/reports/profit?start_date=2026-09-01&end_date=2026-09-17", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	svc.On("ExportCSV", mock.Anything, mock.Anything).
		Return([]byte("CSV DATA"), "report.csv", nil)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest(http.MethodGet, "/api/v1/reports/export?start_date=2026-09-01&end_date=2026-09-17", nil)
	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)
}
