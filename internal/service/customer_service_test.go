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

type mockCustomerRepo struct {
	mock.Mock
}

func (m *mockCustomerRepo) FindByID(ctx context.Context, id uint64) (*model.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Customer), args.Error(1)
}

func (m *mockCustomerRepo) FindByPhone(ctx context.Context, phone string) (*model.Customer, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Customer), args.Error(1)
}

func (m *mockCustomerRepo) Create(ctx context.Context, c *model.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *mockCustomerRepo) Update(ctx context.Context, c *model.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *mockCustomerRepo) FindAll(ctx context.Context, page, limit int, search string) ([]model.Customer, int64, error) {
	args := m.Called(ctx, page, limit, search)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]model.Customer), args.Get(1).(int64), args.Error(2)
}

func TestCustomerService_Create_Success(t *testing.T) {
	repo := new(mockCustomerRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewCustomerService(repo, audit)

	addr := "Jl. Sudirman No. 45"
	req := dto.CreateCustomerRequest{
		Name:    "Budi Santoso",
		Phone:   "081234567890",
		Address: &addr,
	}

	repo.On("FindByPhone", mock.Anything, "081234567890").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(c *model.Customer) bool {
		c.ID = 10
		c.CreatedAt = time.Now()
		c.UpdatedAt = time.Now()
		return c.Name == "Budi Santoso" && c.Phone == "081234567890"
	})).Return(nil)

	audit.On("Log", mock.Anything, mock.MatchedBy(func(e service.AuditEntry) bool {
		return e.Action == "CREATE" && e.EntityType == "customer" && *e.EntityID == 10
	})).Return(nil)

	res, err := svc.Create(context.Background(), req, 1, "127.0.0.1")
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(10), res.ID)
	assert.Equal(t, "Budi Santoso", res.Name)
	assert.Equal(t, "081234567890", res.Phone)

	repo.AssertExpectations(t)
	audit.AssertExpectations(t)
}

func TestCustomerService_Create_DuplicatePhone(t *testing.T) {
	repo := new(mockCustomerRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewCustomerService(repo, audit)

	req := dto.CreateCustomerRequest{
		Name:  "Budi Santoso",
		Phone: "081234567890",
	}

	existing := &model.Customer{
		ID:    10,
		Name:  "Budi Lama",
		Phone: "081234567890",
	}

	repo.On("FindByPhone", mock.Anything, "081234567890").Return(existing, nil)

	res, err := svc.Create(context.Background(), req, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrDuplicate)
	assert.Contains(t, sErr.Message, "phone number already registered")

	repo.AssertExpectations(t)
}

func TestCustomerService_GetByID_NotFound(t *testing.T) {
	repo := new(mockCustomerRepo)
	audit := new(mockAuditServiceForProductSvc)
	svc := service.NewCustomerService(repo, audit)

	repo.On("FindByID", mock.Anything, uint64(99)).Return(nil, nil)

	res, err := svc.GetByID(context.Background(), 99)
	assert.Error(t, err)
	assert.Nil(t, res)

	var sErr *response.ServiceError
	require.ErrorAs(t, err, &sErr)
	assert.ErrorIs(t, sErr.Type, response.ErrNotFound)

	repo.AssertExpectations(t)
}
