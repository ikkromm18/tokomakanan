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

func TestProductService_List_And_Update_And_Delete_Success(t *testing.T) {
	repo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewProductService(repo, audit)

	// Test List
	prods := []model.Product{
		{ID: 1, Name: "Roti", Category: "Bakery", HPP: 5000, SellPrice: 10000, IsActive: true},
	}
	repo.On("FindAll", mock.Anything, 1, 20, "Bakery", mock.Anything, "").Return(prods, int64(1), nil)

	list, meta, err := svc.List(context.Background(), 1, 20, "Bakery", nil, "", "owner")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, int64(1), meta.TotalRows)

	// Test Update
	existing := &model.Product{ID: 1, Name: "Roti Lama", Category: "Bakery", HPP: 5000, SellPrice: 10000}
	repo.On("FindByID", mock.Anything, uint64(1)).Return(existing, nil)
	repo.On("FindByNameAndCategory", mock.Anything, "Roti Baru", "Bakery").Return(nil, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	isActive := true
	updateReq := dto.UpdateProductRequest{
		Name:      "Roti Baru",
		Category:  "Bakery",
		HPP:       6000,
		SellPrice: 12000,
		IsActive:  &isActive,
	}
	updated, err := svc.Update(context.Background(), 1, updateReq, 1, "owner", "127.0.0.1")
	assert.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Roti Baru", updated.Name)

	// Test Delete Success
	repo.On("ExistsInActivePackages", mock.Anything, uint64(1)).Return(false, nil)
	repo.On("Delete", mock.Anything, uint64(1)).Return(nil)

	err = svc.Delete(context.Background(), 1, 1, "127.0.0.1")
	assert.NoError(t, err)
}

func TestCustomerService_List_And_Update(t *testing.T) {
	repo := new(mockCustomerRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewCustomerService(repo, audit)

	// List
	custs := []model.Customer{
		{ID: 1, Name: "Ani", Phone: "0811111111"},
	}
	repo.On("FindAll", mock.Anything, 1, 20, "").Return(custs, int64(1), nil)

	list, meta, err := svc.List(context.Background(), 1, 20, "")
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), meta.TotalRows)

	// Update
	existing := &model.Customer{ID: 1, Name: "Ani Lama", Phone: "0811111111"}
	repo.On("FindByID", mock.Anything, uint64(1)).Return(existing, nil)
	repo.On("FindByPhone", mock.Anything, "0822222222").Return(nil, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	updated, err := svc.Update(context.Background(), 1, dto.UpdateCustomerRequest{
		Name:  "Ani Baru",
		Phone: "0822222222",
	}, 1, "127.0.0.1")
	assert.NoError(t, err)
	assert.Equal(t, "Ani Baru", updated.Name)
	assert.Equal(t, "0822222222", updated.Phone)
}

func TestPackageService_List_Update_Delete(t *testing.T) {
	pkgRepo := new(mockPackageRepo)
	prodRepo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewPackageService(pkgRepo, prodRepo, audit)

	// List
	pkgs := []model.ProductPackage{
		{ID: 1, Name: "Paket 1", TotalHPP: 10000, SellPrice: 15000, IsActive: true},
	}
	pkgRepo.On("FindAll", mock.Anything, 1, 20, mock.Anything, "").Return(pkgs, int64(1), nil)

	list, meta, err := svc.List(context.Background(), 1, 20, nil, "", "owner")
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), meta.TotalRows)

	// Update
	existingPkg := &model.ProductPackage{ID: 1, Name: "Paket Lama", TotalHPP: 10000, SellPrice: 15000}
	pkgRepo.On("FindByID", mock.Anything, uint64(1)).Return(existingPkg, nil)
	pkgRepo.On("FindByName", mock.Anything, "Paket Baru").Return(nil, nil)

	prod := &model.Product{ID: 1, Name: "Roti", HPP: 5000, SellPrice: 10000, IsActive: true}
	prodRepo.On("FindByID", mock.Anything, uint64(1)).Return(prod, nil)
	pkgRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	isActive := true
	upReq := dto.UpdatePackageRequest{
		Name:      "Paket Baru",
		SellPrice: 20000,
		IsActive:  &isActive,
		Items: []dto.PackageItemRequest{
			{ProductID: 1, Quantity: 2},
		},
	}
	updated, err := svc.Update(context.Background(), 1, upReq, 1, "owner", "127.0.0.1")
	assert.NoError(t, err)
	assert.Equal(t, "Paket Baru", updated.Name)

	// Delete
	pkgRepo.On("Delete", mock.Anything, uint64(1)).Return(nil)
	err = svc.Delete(context.Background(), 1, 1, "127.0.0.1")
	assert.NoError(t, err)
}

func TestOrderService_List_Create_UpdateStatus_Success(t *testing.T) {
	orderRepo := new(mockOrderRepo)
	prodRepo := new(mockProductRepo)
	pkgRepo := new(mockPackageRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewOrderService(orderRepo, prodRepo, pkgRepo, audit)

	// List
	orders := []model.Order{
		{ID: 1, InvoiceNo: "INV/001", Status: model.OrderStatusPaid, TotalAmount: 50000, TotalPaid: 50000, CreatedAt: time.Now()},
	}
	orderRepo.On("FindAll", mock.Anything, 1, 20, "", "", "", "").Return(orders, int64(1), nil)

	list, _, err := svc.List(context.Background(), 1, 20, "", "", "", "", "owner")
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	// Create PreOrder Success
	pickupDate := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	prod := &model.Product{ID: 1, Name: "Kue", HPP: 20000, SellPrice: 35000, IsActive: true}
	prodRepo.On("FindByID", mock.Anything, uint64(1)).Return(prod, nil)

	orderRepo.On("CreateOrderTx", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	audit.On("Log", mock.Anything, mock.Anything).Return(nil)

	req := dto.CreateOrderRequest{
		OrderType:     model.OrderTypePreOrder,
		PickupDate:    &pickupDate,
		PaymentMethod: model.PaymentMethodCash,
		InitialPaid:   10000, // DP
		Items: []dto.CreateOrderItemRequest{
			{ItemType: model.ItemTypeProduct, ItemID: 1, Quantity: 1},
		},
	}
	created, err := svc.Create(context.Background(), 1, "superadmin", req, "127.0.0.1")
	assert.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, model.OrderStatusDPPaid, created.Status)

	// GetByID
	orderRepo.On("FindByID", mock.Anything, uint64(10)).Return(&orders[0], nil)
	byId, err := svc.GetByID(context.Background(), 10, "admin")
	assert.NoError(t, err)
	assert.Equal(t, "INV/001", byId.InvoiceNo)

	// Update status valid transition: PAID -> READY
	paidOrder := &model.Order{ID: 2, Status: model.OrderStatusPaid}
	orderRepo.On("FindByID", mock.Anything, uint64(2)).Return(paidOrder, nil)
	orderRepo.On("UpdateStatus", mock.Anything, uint64(2), model.OrderStatusReady).Return(nil)

	err = svc.UpdateStatus(context.Background(), 2, model.OrderStatusReady, 1, "127.0.0.1")
	assert.NoError(t, err)
}

func TestPaymentService_ListByOrder(t *testing.T) {
	payRepo := new(mockPaymentRepo)
	orderRepo := new(mockOrderRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewPaymentService(payRepo, orderRepo, audit)

	payments := []model.OrderPayment{
		{ID: 1, OrderID: 1, Amount: 50000, PaymentMethod: "CASH", PaidAt: time.Now()},
	}
	payRepo.On("FindByOrderID", mock.Anything, uint64(1)).Return(payments, nil)

	res, err := svc.ListByOrder(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestDashboardService_GetChart_And_POReminders(t *testing.T) {
	dashRepo := new(mockDashboardRepo)
	svc := service.NewDashboardService(dashRepo)

	points := []dto.SalesChartPoint{
		{Date: "2026-09-17", TotalSales: 100000, TotalOrders: 2},
	}
	dashRepo.On("GetSalesChart", mock.Anything, mock.Anything, mock.Anything).Return(points, nil)

	chart, err := svc.GetChart(context.Background(), "", "")
	assert.NoError(t, err)
	assert.Len(t, chart, 1)

	orders := []model.Order{
		{ID: 1, InvoiceNo: "INV/PO/01", Status: model.OrderStatusPaid},
	}
	dashRepo.On("GetPOReminders", mock.Anything, mock.Anything).Return(orders, nil)

	reminders, err := svc.GetPOReminders(context.Background(), "admin")
	assert.NoError(t, err)
	assert.Len(t, reminders, 1)
}

func TestReportService_GetProfitReport_And_ExportProfit(t *testing.T) {
	reportRepo := new(mockReportRepo)
	svc := service.NewReportService(reportRepo)

	daily := []dto.ProfitReportItem{
		{Date: "2026-09-17", Revenue: 100000, TotalHPP: 60000, GrossProfit: 40000, MarginPct: 40.0},
	}
	reportRepo.On("GetProfitDailyData", mock.Anything, "2026-09-01", "2026-09-17").Return(daily, nil)

	res, err := svc.GetProfitReport(context.Background(), dto.ReportFilterRequest{
		StartDate: "2026-09-01",
		EndDate:   "2026-09-17",
	})
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, float64(100000), res.TotalRevenue)
	assert.Equal(t, float64(40000), res.TotalGrossProfit)

	data, filename, err := svc.ExportCSV(context.Background(), dto.ReportFilterRequest{
		StartDate: "2026-09-01",
		EndDate:   "2026-09-17",
		Type:      "profit",
	})
	assert.NoError(t, err)
	assert.Contains(t, filename, "laporan_laba_")
	assert.True(t, len(data) > 0)
}
