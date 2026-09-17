package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPackageRepo struct {
	mock.Mock
}

func (m *mockPackageRepo) FindByID(ctx context.Context, id uint64) (*model.ProductPackage, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductPackage), args.Error(1)
}

func (m *mockPackageRepo) FindByName(ctx context.Context, name string) (*model.ProductPackage, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductPackage), args.Error(1)
}

func (m *mockPackageRepo) Create(ctx context.Context, pkg *model.ProductPackage) error {
	args := m.Called(ctx, pkg)
	return args.Error(0)
}

func (m *mockPackageRepo) Update(ctx context.Context, pkg *model.ProductPackage, items []model.PackageItem) error {
	args := m.Called(ctx, pkg, items)
	return args.Error(0)
}

func (m *mockPackageRepo) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPackageRepo) FindAll(ctx context.Context, page, limit int, isActive *bool, search string) ([]model.ProductPackage, int64, error) {
	args := m.Called(ctx, page, limit, isActive, search)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]model.ProductPackage), args.Get(1).(int64), args.Error(2)
}

func TestPackageService_Create_SellPriceLessThanCalculatedHPP(t *testing.T) {
	pkgRepo := new(mockPackageRepo)
	prodRepo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewPackageService(pkgRepo, prodRepo, audit)

	req := dto.CreatePackageRequest{
		Name:      "Paket Sarapan Hemat",
		SellPrice: 15000, // Violation: calculated total_hpp will be (10000*2) = 20000
		Items: []dto.PackageItemRequest{
			{ProductID: 1, Quantity: 2},
		},
	}

	prod := &model.Product{
		ID:        1,
		Name:      "Roti Tawar",
		HPP:       10000,
		SellPrice: 14000,
		IsActive:  true,
	}

	pkgRepo.On("FindByName", mock.Anything, "Paket Sarapan Hemat").Return(nil, nil)
	prodRepo.On("FindByID", mock.Anything, uint64(1)).Return(prod, nil)

	res, err := svc.Create(context.Background(), req, 1, "superadmin", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrBusinessRule)
	assert.Contains(t, sErr.Message, "sell price cannot be less than total HPP")
}

func TestPackageService_Create_Success(t *testing.T) {
	pkgRepo := new(mockPackageRepo)
	prodRepo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewPackageService(pkgRepo, prodRepo, audit)

	req := dto.CreatePackageRequest{
		Name:      "Paket Sarapan Hemat",
		SellPrice: 25000, // Total HPP = 10000*2 = 20000 <= 25000 (Valid)
		Items: []dto.PackageItemRequest{
			{ProductID: 1, Quantity: 2},
		},
	}

	prod := &model.Product{
		ID:        1,
		Name:      "Roti Tawar",
		HPP:       10000,
		SellPrice: 14000,
		IsActive:  true,
	}

	pkgRepo.On("FindByName", mock.Anything, "Paket Sarapan Hemat").Return(nil, nil)
	prodRepo.On("FindByID", mock.Anything, uint64(1)).Return(prod, nil)

	pkgRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *model.ProductPackage) bool {
		p.ID = 50
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
		return p.Name == "Paket Sarapan Hemat" && p.TotalHPP == 20000 && p.SellPrice == 25000 && len(p.Items) == 1
	})).Return(nil)

	audit.On("Log", mock.Anything, mock.MatchedBy(func(e service.AuditEntry) bool {
		return e.Action == "CREATE" && e.EntityType == "package" && *e.EntityID == 50
	})).Return(nil)

	res, err := svc.Create(context.Background(), req, 1, "superadmin", "127.0.0.1")
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(50), res.ID)
	assert.Equal(t, "Paket Sarapan Hemat", res.Name)
	assert.Equal(t, float64(25000), res.SellPrice)
	require.NotNil(t, res.TotalHPP)
	assert.Equal(t, float64(20000), *res.TotalHPP)

	pkgRepo.AssertExpectations(t)
	prodRepo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestPackageService_AdminRole_HidesHPP(t *testing.T) {
	pkgRepo := new(mockPackageRepo)
	prodRepo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewPackageService(pkgRepo, prodRepo, audit)

	pkg := &model.ProductPackage{
		ID:        50,
		Name:      "Paket Sarapan Hemat",
		TotalHPP:  20000,
		SellPrice: 25000,
		IsActive:  true,
		Items: []model.PackageItem{
			{
				ID:        1,
				PackageID: 50,
				ProductID: 1,
				Quantity:  2,
				Product: model.Product{
					ID:        1,
					Name:      "Roti Tawar",
					HPP:       10000,
					SellPrice: 14000,
				},
			},
		},
	}

	pkgRepo.On("FindByID", mock.Anything, uint64(50)).Return(pkg, nil)

	res, err := svc.GetByID(context.Background(), 50, "admin")
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Nil(t, res.TotalHPP, "TotalHPP should be masked for admin")
	require.Len(t, res.Items, 1)
	assert.Nil(t, res.Items[0].UnitHPP, "UnitHPP of items should be masked for admin")

	pkgRepo.AssertExpectations(t)
}
