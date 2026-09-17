package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	jwtpkg "github.com/ikkromm18/tokomakanan/internal/pkg/jwt"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepoForUserSvc struct {
	mock.Mock
}

func (m *mockUserRepoForUserSvc) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepoForUserSvc) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepoForUserSvc) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepoForUserSvc) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepoForUserSvc) FindAll(ctx context.Context, page, limit int, role, search string) ([]model.User, int64, error) {
	args := m.Called(ctx, page, limit, role, search)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepoForUserSvc) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockAuditServiceForUserSvc struct {
	mock.Mock
}

func (m *mockAuditServiceForUserSvc) Log(ctx context.Context, entry service.AuditEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockAuditServiceForUserSvc) List(ctx context.Context, page, limit int, entityType, action string) ([]dto.AuditLogResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, entityType, action)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.AuditLogResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func TestUserService_Create_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	auditSvc := new(mockAuditServiceForUserSvc)
	cfg := &config.Config{BcryptCost: 4}
	svc := service.NewUserService(userRepo, auditSvc, cfg)

	req := dto.CreateUserRequest{
		Name:     "Owner Bob",
		Email:    "bob@example.com",
		Password: "password123",
		Role:     model.RoleOwner,
	}

	userRepo.On("FindByEmail", mock.Anything, "bob@example.com").Return(nil, nil)
	userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		u.ID = 42
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
		return u.Name == "Owner Bob" && u.Email == "bob@example.com" && u.Role == model.RoleOwner && u.IsActive == true && jwtpkg.ComparePassword(u.PasswordHash, "password123") == nil
	})).Return(nil)

	auditSvc.On("Log", mock.Anything, mock.MatchedBy(func(e service.AuditEntry) bool {
		return e.Action == "CREATE" && e.EntityType == "user" && e.EntityID != nil && *e.EntityID == 42 && e.IPAddress == "127.0.0.1" && e.UserID != nil && *e.UserID == 1
	})).Return(nil)

	res, err := svc.Create(context.Background(), req, 1, "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(42), res.ID)
	assert.Equal(t, "Owner Bob", res.Name)
	assert.Equal(t, "bob@example.com", res.Email)
	assert.Equal(t, model.RoleOwner, res.Role)
	assert.True(t, res.IsActive)

	userRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	req := dto.CreateUserRequest{
		Name:     "Duplicate Bob",
		Email:    "existing@example.com",
		Password: "password123",
		Role:     model.RoleAdmin,
	}

	existingUser := &model.User{ID: 10, Email: "existing@example.com"}
	userRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

	res, err := svc.Create(context.Background(), req, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrDuplicate))

	userRepo.AssertExpectations(t)
}

func TestUserService_Create_RepoFindByEmailError(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	req := dto.CreateUserRequest{
		Name:     "Bob",
		Email:    "bob@example.com",
		Password: "password123",
		Role:     model.RoleAdmin,
	}

	userRepo.On("FindByEmail", mock.Anything, "bob@example.com").Return(nil, errors.New("db error"))

	res, err := svc.Create(context.Background(), req, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "db error", err.Error())
}

func TestUserService_Create_RepoCreateError(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, &config.Config{BcryptCost: 4})

	req := dto.CreateUserRequest{
		Name:     "Bob",
		Email:    "bob@example.com",
		Password: "password123",
		Role:     model.RoleAdmin,
	}

	userRepo.On("FindByEmail", mock.Anything, "bob@example.com").Return(nil, nil)
	userRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("insert failed"))

	res, err := svc.Create(context.Background(), req, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "insert failed", err.Error())
}

func TestUserService_GetByID_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	user := &model.User{
		ID:        5,
		Name:      "Jane Doe",
		Email:     "jane@example.com",
		Role:      model.RoleOwner,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userRepo.On("FindByID", mock.Anything, uint64(5)).Return(user, nil)

	res, err := svc.GetByID(context.Background(), 5)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(5), res.ID)
	assert.Equal(t, "Jane Doe", res.Name)
	assert.Equal(t, "jane@example.com", res.Email)
	assert.Equal(t, model.RoleOwner, res.Role)
	assert.True(t, res.IsActive)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	userRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, nil)

	res, err := svc.GetByID(context.Background(), 99)
	assert.Error(t, err)
	assert.Nil(t, res)

	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrNotFound))
}

func TestUserService_GetByID_RepoError(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	userRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, errors.New("db error"))

	res, err := svc.GetByID(context.Background(), 99)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "db error", err.Error())
}

func TestUserService_Update_SuccessAllFields(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	auditSvc := new(mockAuditServiceForUserSvc)
	cfg := &config.Config{BcryptCost: 4}
	svc := service.NewUserService(userRepo, auditSvc, cfg)

	existingUser := &model.User{
		ID:           7,
		Name:         "Old Name",
		Email:        "old@example.com",
		PasswordHash: "oldhash",
		Role:         model.RoleAdmin,
		IsActive:     true,
		CreatedAt:    time.Now().Add(-time.Hour),
		UpdatedAt:    time.Now().Add(-time.Hour),
	}
	userRepo.On("FindByID", mock.Anything, uint64(7)).Return(existingUser, nil)

	newName := "New Name"
	newEmail := "new@example.com"
	newPassword := "newpassword123"
	newRole := model.RoleOwner
	newIsActive := false

	req := dto.UpdateUserRequest{
		Name:     &newName,
		Email:    &newEmail,
		Password: &newPassword,
		Role:     &newRole,
		IsActive: &newIsActive,
	}

	userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, nil)
	userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.ID == 7 && u.Name == "New Name" && u.Email == "new@example.com" && u.Role == model.RoleOwner && u.IsActive == false && jwtpkg.ComparePassword(u.PasswordHash, "newpassword123") == nil
	})).Return(nil)

	auditSvc.On("Log", mock.Anything, mock.MatchedBy(func(e service.AuditEntry) bool {
		return e.Action == "UPDATE" && e.EntityType == "user" && e.EntityID != nil && *e.EntityID == 7 && e.IPAddress == "10.0.0.1" && e.UserID != nil && *e.UserID == 1
	})).Return(nil)

	res, err := svc.Update(context.Background(), 7, req, 1, "10.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, uint64(7), res.ID)
	assert.Equal(t, "New Name", res.Name)
	assert.Equal(t, "new@example.com", res.Email)
	assert.Equal(t, model.RoleOwner, res.Role)
	assert.False(t, res.IsActive)

	userRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestUserService_Update_DuplicateEmail(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	existingUser := &model.User{
		ID:    7,
		Name:  "User Seven",
		Email: "user7@example.com",
	}
	userRepo.On("FindByID", mock.Anything, uint64(7)).Return(existingUser, nil)

	takenEmail := "taken@example.com"
	req := dto.UpdateUserRequest{
		Email: &takenEmail,
	}

	conflictUser := &model.User{ID: 8, Email: "taken@example.com"}
	userRepo.On("FindByEmail", mock.Anything, "taken@example.com").Return(conflictUser, nil)

	res, err := svc.Update(context.Background(), 7, req, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrDuplicate))
}

func TestUserService_Update_SameEmailNoConflict(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	existingUser := &model.User{
		ID:           7,
		Name:         "User Seven",
		Email:        "same@example.com",
		PasswordHash: "keepthis",
		Role:         model.RoleAdmin,
		IsActive:     true,
	}
	userRepo.On("FindByID", mock.Anything, uint64(7)).Return(existingUser, nil)

	sameEmail := "same@example.com"
	newName := "Updated Name"
	req := dto.UpdateUserRequest{
		Name:  &newName,
		Email: &sameEmail,
	}

	userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.Name == "Updated Name" && u.PasswordHash == "keepthis"
	})).Return(nil)

	res, err := svc.Update(context.Background(), 7, req, 1, "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", res.Name)
}

func TestUserService_Update_NotFound(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	userRepo.On("FindByID", mock.Anything, uint64(88)).Return(nil, nil)

	name := "Any"
	req := dto.UpdateUserRequest{Name: &name}

	res, err := svc.Update(context.Background(), 88, req, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, res)

	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrNotFound))
}

func TestUserService_Delete_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	auditSvc := new(mockAuditServiceForUserSvc)
	svc := service.NewUserService(userRepo, auditSvc, nil)

	existingUser := &model.User{
		ID:       15,
		Name:     "To Delete",
		Email:    "todelete@example.com",
		Role:     model.RoleAdmin,
		IsActive: true,
	}
	userRepo.On("FindByID", mock.Anything, uint64(15)).Return(existingUser, nil)
	userRepo.On("Delete", mock.Anything, uint64(15)).Return(nil)

	auditSvc.On("Log", mock.Anything, mock.MatchedBy(func(e service.AuditEntry) bool {
		return e.Action == "DELETE" && e.EntityType == "user" && e.EntityID != nil && *e.EntityID == 15 && e.IPAddress == "127.0.0.1" && e.UserID != nil && *e.UserID == 1
	})).Return(nil)

	err := svc.Delete(context.Background(), 15, 1, "127.0.0.1")
	require.NoError(t, err)

	userRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	userRepo.On("FindByID", mock.Anything, uint64(99)).Return(nil, nil)

	err := svc.Delete(context.Background(), 99, 1, "127.0.0.1")
	assert.Error(t, err)

	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrNotFound))
}

func TestUserService_Delete_RepoDeleteError(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	existingUser := &model.User{ID: 15}
	userRepo.On("FindByID", mock.Anything, uint64(15)).Return(existingUser, nil)
	userRepo.On("Delete", mock.Anything, uint64(15)).Return(errors.New("db delete failure"))

	err := svc.Delete(context.Background(), 15, 1, "127.0.0.1")
	assert.Error(t, err)
	assert.Equal(t, "db delete failure", err.Error())
}

func TestUserService_List_Success(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	users := []model.User{
		{ID: 1, Name: "Admin User", Email: "admin@test.com", Role: model.RoleAdmin, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Name: "Owner User", Email: "owner@test.com", Role: model.RoleOwner, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	userRepo.On("FindAll", mock.Anything, 1, 20, "admin", "User").Return(users, int64(2), nil)

	res, meta, err := svc.List(context.Background(), 1, 20, "admin", "User")
	require.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, uint64(1), res[0].ID)
	assert.Equal(t, "admin@test.com", res[0].Email)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 20, meta.Limit)
	assert.Equal(t, int64(2), meta.TotalRows)
	assert.Equal(t, 1, meta.TotalPages)
}

func TestUserService_List_RepoError(t *testing.T) {
	userRepo := new(mockUserRepoForUserSvc)
	svc := service.NewUserService(userRepo, nil, nil)

	userRepo.On("FindAll", mock.Anything, 1, 20, "", "").Return(nil, int64(0), errors.New("db query error"))

	res, _, err := svc.List(context.Background(), 1, 20, "", "")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "db query error", err.Error())
}
