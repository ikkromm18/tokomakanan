package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockStoreSettingRepo struct {
	mock.Mock
}

func (m *mockStoreSettingRepo) Get(ctx context.Context) (*model.StoreSetting, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StoreSetting), args.Error(1)
}

func (m *mockStoreSettingRepo) Update(ctx context.Context, setting *model.StoreSetting) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

type mockAuditServiceForStoreSetting struct {
	mock.Mock
}

func (m *mockAuditServiceForStoreSetting) Log(ctx context.Context, entry service.AuditEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockAuditServiceForStoreSetting) List(ctx context.Context, page, limit int, entityType, action string) ([]dto.AuditLogResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, entityType, action)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.AuditLogResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func strPtr(s string) *string {
	return &s
}

func TestStoreSettingService_Get_Existing(t *testing.T) {
	repo := new(mockStoreSettingRepo)
	auditSvc := new(mockAuditServiceForStoreSetting)
	svc := service.NewStoreSettingService(repo, auditSvc)

	existing := &model.StoreSetting{
		ID:            1,
		Name:          "Toko Roti Sedap",
		Address:       strPtr("Jl. Dahlia No. 5"),
		Phone:         strPtr("08123456789"),
		LogoURL:       strPtr("https://example.com/logo.png"),
		ReceiptFooter: strPtr("Terima kasih"),
		UpdatedAt:     time.Now(),
	}

	repo.On("Get", mock.Anything).Return(existing, nil)

	res, err := svc.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(1), res.ID)
	assert.Equal(t, "Toko Roti Sedap", res.Name)
	assert.Equal(t, "Jl. Dahlia No. 5", *res.Address)
	assert.Equal(t, "08123456789", *res.Phone)
	assert.Equal(t, "https://example.com/logo.png", *res.LogoURL)
	assert.Equal(t, "Terima kasih", *res.ReceiptFooter)

	repo.AssertExpectations(t)
}

func TestStoreSettingService_Get_EmptyInitializesDefault(t *testing.T) {
	repo := new(mockStoreSettingRepo)
	auditSvc := new(mockAuditServiceForStoreSetting)
	svc := service.NewStoreSettingService(repo, auditSvc)

	repo.On("Get", mock.Anything).Return(nil, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(s *model.StoreSetting) bool {
		return s.Name == "Toko Makanan"
	})).Run(func(args mock.Arguments) {
		s := args.Get(1).(*model.StoreSetting)
		s.ID = 1
	}).Return(nil)

	res, err := svc.Get(context.Background())
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "Toko Makanan", res.Name)

	repo.AssertExpectations(t)
}

func TestStoreSettingService_Get_RepoGetError(t *testing.T) {
	repo := new(mockStoreSettingRepo)
	auditSvc := new(mockAuditServiceForStoreSetting)
	svc := service.NewStoreSettingService(repo, auditSvc)

	repo.On("Get", mock.Anything).Return(nil, errors.New("database connection failed"))

	res, err := svc.Get(context.Background())
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestStoreSettingService_Get_RepoUpdateErrorOnDefault(t *testing.T) {
	repo := new(mockStoreSettingRepo)
	auditSvc := new(mockAuditServiceForStoreSetting)
	svc := service.NewStoreSettingService(repo, auditSvc)

	repo.On("Get", mock.Anything).Return(nil, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db write error"))

	res, err := svc.Get(context.Background())
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "db write error")

	repo.AssertExpectations(t)
}

func TestStoreSettingService_Update_SuccessExisting(t *testing.T) {
	repo := new(mockStoreSettingRepo)
	auditSvc := new(mockAuditServiceForStoreSetting)
	svc := service.NewStoreSettingService(repo, auditSvc)

	existing := &model.StoreSetting{
		ID:            1,
		Name:          "Old Name",
		Address:       strPtr("Old Address"),
		Phone:         strPtr("0811111111"),
		LogoURL:       strPtr("https://old.com/logo.png"),
		ReceiptFooter: strPtr("Old Footer"),
		UpdatedAt:     time.Now().Add(-time.Hour),
	}

	req := dto.UpdateStoreSettingRequest{
		Name:          "New Bakery",
		Address:       strPtr("New Address"),
		Phone:         strPtr("0822222222"),
		LogoURL:       strPtr("https://new.com/logo.png"),
		ReceiptFooter: strPtr("New Footer"),
	}

	repo.On("Get", mock.Anything).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(s *model.StoreSetting) bool {
		return s.Name == "New Bakery" && *s.Address == "New Address"
	})).Return(nil)

	auditSvc.On("Log", mock.Anything, mock.MatchedBy(func(entry service.AuditEntry) bool {
		return entry.Action == "UPDATE" &&
			entry.EntityType == "store_setting" &&
			entry.EntityID != nil && *entry.EntityID == 1 &&
			entry.UserID != nil && *entry.UserID == 99 &&
			entry.IPAddress == "127.0.0.1" &&
			entry.OldValue != nil &&
			entry.NewValue != nil
	})).Return(nil)

	res, err := svc.Update(context.Background(), 99, req, "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(1), res.ID)
	assert.Equal(t, "New Bakery", res.Name)
	assert.Equal(t, "New Address", *res.Address)
	assert.Equal(t, "0822222222", *res.Phone)
	assert.Equal(t, "https://new.com/logo.png", *res.LogoURL)
	assert.Equal(t, "New Footer", *res.ReceiptFooter)

	repo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestStoreSettingService_Update_SuccessWhenEmptyInitially(t *testing.T) {
	repo := new(mockStoreSettingRepo)
	auditSvc := new(mockAuditServiceForStoreSetting)
	svc := service.NewStoreSettingService(repo, auditSvc)

	req := dto.UpdateStoreSettingRequest{
		Name:    "Fresh Bakery",
		Address: strPtr("Jl. Baru No. 1"),
	}

	repo.On("Get", mock.Anything).Return(nil, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(s *model.StoreSetting) bool {
		return s.Name == "Fresh Bakery"
	})).Run(func(args mock.Arguments) {
		s := args.Get(1).(*model.StoreSetting)
		s.ID = 1
	}).Return(nil)

	auditSvc.On("Log", mock.Anything, mock.MatchedBy(func(entry service.AuditEntry) bool {
		return entry.Action == "UPDATE" &&
			entry.EntityType == "store_setting" &&
			entry.EntityID != nil && *entry.EntityID == 1 &&
			entry.OldValue == nil
	})).Return(nil)

	res, err := svc.Update(context.Background(), 42, req, "192.168.1.1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "Fresh Bakery", res.Name)
	assert.Equal(t, "Jl. Baru No. 1", *res.Address)

	repo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestStoreSettingService_Update_RepoGetError(t *testing.T) {
	repo := new(mockStoreSettingRepo)
	auditSvc := new(mockAuditServiceForStoreSetting)
	svc := service.NewStoreSettingService(repo, auditSvc)

	repo.On("Get", mock.Anything).Return(nil, errors.New("db read failure"))

	req := dto.UpdateStoreSettingRequest{Name: "Store Name"}
	res, err := svc.Update(context.Background(), 1, req, "127.0.0.1")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "db read failure")

	repo.AssertExpectations(t)
}

func TestStoreSettingService_Update_RepoUpdateError(t *testing.T) {
	repo := new(mockStoreSettingRepo)
	auditSvc := new(mockAuditServiceForStoreSetting)
	svc := service.NewStoreSettingService(repo, auditSvc)

	existing := &model.StoreSetting{ID: 1, Name: "Current"}
	repo.On("Get", mock.Anything).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db write failure"))

	req := dto.UpdateStoreSettingRequest{Name: "New Name"}
	res, err := svc.Update(context.Background(), 1, req, "127.0.0.1")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "db write failure")

	repo.AssertExpectations(t)
}
