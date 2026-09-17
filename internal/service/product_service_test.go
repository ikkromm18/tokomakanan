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

type mockProductRepo struct {
	mock.Mock
}

func (m *mockProductRepo) FindByID(ctx context.Context, id uint64) (*model.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *mockProductRepo) FindByNameAndCategory(ctx context.Context, name, category string) (*model.Product, error) {
	args := m.Called(ctx, name, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *mockProductRepo) Create(ctx context.Context, p *model.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockProductRepo) Update(ctx context.Context, p *model.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockProductRepo) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockProductRepo) FindAll(ctx context.Context, page, limit int, category string, isActive *bool, search string) ([]model.Product, int64, error) {
	args := m.Called(ctx, page, limit, category, isActive, search)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]model.Product), args.Get(1).(int64), args.Error(2)
}

func (m *mockProductRepo) ExistsInActivePackages(ctx context.Context, productID uint64) (bool, error) {
	args := m.Called(ctx, productID)
	return args.Bool(0), args.Error(1)
}

type mockAuditServiceForProductSvc struct {
	mock.Mock
}

func (m *mockAuditServiceForProductSvc) Log(ctx context.Context, entry service.AuditEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockAuditServiceForProductSvc) List(ctx context.Context, page, limit int, entityType, action string) ([]dto.AuditLogResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, entityType, action)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.AuditLogResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func TestProductService_Create_SellPriceLessThanHPP(t *testing.T) {
	repo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewProductService(repo, audit)

	req := dto.CreateProductRequest{
		Name:      "Roti Tawar",
		Category:  "Bakery",
		HPP:       15000,
		SellPrice: 12000, // Invalid: sell_price < hpp
	}

	res, err := svc.Create(context.Background(), req, 1, "superadmin", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrBusinessRule)
	assert.Contains(t, sErr.Message, "sell price cannot be less than HPP")
}

func TestProductService_Create_Success(t *testing.T) {
	repo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewProductService(repo, audit)

	req := dto.CreateProductRequest{
		Name:      "Roti Tawar",
		Category:  "Bakery",
		HPP:       10000,
		SellPrice: 15000,
	}

	repo.On("FindByNameAndCategory", mock.Anything, "Roti Tawar", "Bakery").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(p *model.Product) bool {
		p.ID = 101
		p.CreatedAt = time.Now()
		p.UpdatedAt = time.Now()
		return p.Name == "Roti Tawar" && p.HPP == 10000 && p.SellPrice == 15000 && p.IsActive == true
	})).Return(nil)
	audit.On("Log", mock.Anything, mock.MatchedBy(func(e service.AuditEntry) bool {
		return e.Action == "CREATE" && e.EntityType == "product" && *e.EntityID == 101
	})).Return(nil)

	res, err := svc.Create(context.Background(), req, 1, "superadmin", "127.0.0.1")
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(101), res.ID)
	assert.Equal(t, "Roti Tawar", res.Name)
	assert.Equal(t, float64(15000), res.SellPrice)
	require.NotNil(t, res.HPP)
	assert.Equal(t, float64(10000), *res.HPP)

	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestProductService_Create_DuplicateNameAndCategory(t *testing.T) {
	repo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewProductService(repo, audit)

	req := dto.CreateProductRequest{
		Name:      "Roti Tawar",
		Category:  "Bakery",
		HPP:       10000,
		SellPrice: 15000,
	}

	existing := &model.Product{
		ID:       1,
		Name:     "Roti Tawar",
		Category: "Bakery",
	}
	repo.On("FindByNameAndCategory", mock.Anything, "Roti Tawar", "Bakery").Return(existing, nil)

	res, err := svc.Create(context.Background(), req, 1, "superadmin", "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrDuplicate)
}

func TestProductService_AdminRole_HidesHPP(t *testing.T) {
	repo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewProductService(repo, audit)

	p := &model.Product{
		ID:        5,
		Name:      "Croissant",
		Category:  "Pastry",
		HPP:       8000,
		SellPrice: 16000,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.On("FindByID", mock.Anything, uint64(5)).Return(p, nil)

	res, err := svc.GetByID(context.Background(), 5, "admin")
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Nil(t, res.HPP, "HPP must be hidden (nil) for admin role")
	assert.Equal(t, float64(16000), res.SellPrice)
}

func TestProductService_Delete_ReferencedInPackage(t *testing.T) {
	repo := new(mockProductRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewProductService(repo, audit)

	p := &model.Product{
		ID:       5,
		Name:     "Croissant",
		Category: "Pastry",
	}
	repo.On("FindByID", mock.Anything, uint64(5)).Return(p, nil)
	repo.On("ExistsInActivePackages", mock.Anything, uint64(5)).Return(true, nil)

	err := svc.Delete(context.Background(), 5, 1, "127.0.0.1")
	assert.Error(t, err)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrBusinessRule)
	assert.Contains(t, sErr.Message, "product is used in active package")
}
