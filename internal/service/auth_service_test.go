package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	jwtpkg "github.com/ikkromm18/tokomakanan/internal/pkg/jwt"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) FindAll(ctx context.Context, page, limit int, role, search string) ([]model.User, int64, error) {
	args := m.Called(ctx, page, limit, role, search)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepo) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockAuditServiceForAuth struct {
	mock.Mock
}

func (m *mockAuditServiceForAuth) Log(ctx context.Context, entry service.AuditEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockAuditServiceForAuth) List(ctx context.Context, page, limit int, entityType, action string) ([]dto.AuditLogResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, entityType, action)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.AuditLogResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:      "12345678901234567890123456789012",
		JWTExpiryHours: 24,
		BcryptCost:     bcrypt.MinCost, // use MinCost for fast tests
	}
}

// -----------------------------------------------------------------------------
// Login Tests
// -----------------------------------------------------------------------------

func TestAuthService_Login_Success(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), cfg.BcryptCost)
	user := &model.User{
		ID:           10,
		Name:         "Super Admin",
		Email:        "superadmin@example.com",
		PasswordHash: string(hashedPassword),
		Role:         model.RoleSuperadmin,
		IsActive:     true,
	}

	repo.On("FindByEmail", mock.Anything, "superadmin@example.com").Return(user, nil).Once()
	auditSvc.On("Log", mock.Anything, mock.MatchedBy(func(e service.AuditEntry) bool {
		return e.Action == "LOGIN" &&
			e.EntityType == "user" &&
			*e.EntityID == 10 &&
			*e.UserID == 10 &&
			e.IPAddress == "127.0.0.1"
	})).Return(nil).Once()

	resp, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "superadmin@example.com",
		Password: "password123",
	}, "127.0.0.1")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, uint64(10), resp.User.ID)
	assert.Equal(t, "Super Admin", resp.User.Name)
	assert.Equal(t, "superadmin@example.com", resp.User.Email)
	assert.Equal(t, model.RoleSuperadmin, resp.User.Role)

	// Validate the returned JWT token
	claims, err := jwtpkg.ValidateToken(resp.Token, cfg.JWTSecret)
	require.NoError(t, err)
	assert.Equal(t, uint64(10), claims.UserID)
	assert.Equal(t, model.RoleSuperadmin, claims.Role)

	repo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	repo.On("FindByEmail", mock.Anything, "notfound@example.com").Return(nil, nil).Once()

	resp, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}, "127.0.0.1")

	assert.Error(t, err)
	assert.Nil(t, resp)

	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrUnauthorized))
	assert.Equal(t, "Invalid credentials", svcErr.Message)

	auditSvc.AssertNotCalled(t, "Log", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_DeactivatedAccount(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), cfg.BcryptCost)
	user := &model.User{
		ID:           5,
		Name:         "Inactive User",
		Email:        "inactive@example.com",
		PasswordHash: string(hashedPassword),
		Role:         model.RoleAdmin,
		IsActive:     false,
	}

	repo.On("FindByEmail", mock.Anything, "inactive@example.com").Return(user, nil).Once()

	resp, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "inactive@example.com",
		Password: "password123",
	}, "127.0.0.1")

	assert.Error(t, err)
	assert.Nil(t, resp)

	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrUnauthorized))
	assert.Equal(t, "Account is deactivated", svcErr.Message)

	auditSvc.AssertNotCalled(t, "Log", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), cfg.BcryptCost)
	user := &model.User{
		ID:           7,
		Name:         "User",
		Email:        "user@example.com",
		PasswordHash: string(hashedPassword),
		Role:         model.RoleAdmin,
		IsActive:     true,
	}

	repo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil).Once()

	resp, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "user@example.com",
		Password: "wrongpassword",
	}, "127.0.0.1")

	assert.Error(t, err)
	assert.Nil(t, resp)

	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrUnauthorized))
	assert.Equal(t, "Invalid credentials", svcErr.Message)

	auditSvc.AssertNotCalled(t, "Log", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_RepoError(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	repo.On("FindByEmail", mock.Anything, "error@example.com").
		Return(nil, errors.New("db query error")).Once()

	resp, err := svc.Login(context.Background(), dto.LoginRequest{
		Email:    "error@example.com",
		Password: "password123",
	}, "127.0.0.1")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "db query error", err.Error())

	auditSvc.AssertNotCalled(t, "Log", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

// -----------------------------------------------------------------------------
// GetMe Tests
// -----------------------------------------------------------------------------

func TestAuthService_GetMe_Success(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	user := &model.User{
		ID:       10,
		Name:     "Super Admin",
		Email:    "superadmin@example.com",
		Role:     model.RoleSuperadmin,
		IsActive: true,
	}

	repo.On("FindByID", mock.Anything, uint64(10)).Return(user, nil).Once()

	userInfo, err := svc.GetMe(context.Background(), 10)
	require.NoError(t, err)
	require.NotNil(t, userInfo)
	assert.Equal(t, uint64(10), userInfo.ID)
	assert.Equal(t, "Super Admin", userInfo.Name)
	assert.Equal(t, "superadmin@example.com", userInfo.Email)
	assert.Equal(t, model.RoleSuperadmin, userInfo.Role)

	repo.AssertExpectations(t)
}

func TestAuthService_GetMe_NotFound(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	repo.On("FindByID", mock.Anything, uint64(99)).Return(nil, nil).Once()

	userInfo, err := svc.GetMe(context.Background(), 99)
	assert.Error(t, err)
	assert.Nil(t, userInfo)

	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrNotFound))
	assert.Equal(t, "User not found", svcErr.Message)

	repo.AssertExpectations(t)
}

func TestAuthService_GetMe_RepoError(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	repo.On("FindByID", mock.Anything, uint64(99)).Return(nil, errors.New("db error")).Once()

	userInfo, err := svc.GetMe(context.Background(), 99)
	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Equal(t, "db error", err.Error())

	repo.AssertExpectations(t)
}

// -----------------------------------------------------------------------------
// ChangePassword Tests
// -----------------------------------------------------------------------------

func TestAuthService_ChangePassword_Success(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword123"), cfg.BcryptCost)
	user := &model.User{
		ID:           3,
		Name:         "Admin User",
		Email:        "admin@example.com",
		PasswordHash: string(oldHash),
		Role:         model.RoleAdmin,
		IsActive:     true,
	}

	repo.On("FindByID", mock.Anything, uint64(3)).Return(user, nil).Once()
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		// verify that the new password hash matches "newpassword123"
		return u.ID == 3 &&
			jwtpkg.CheckPassword("newpassword123", u.PasswordHash)
	})).Return(nil).Once()

	auditSvc.On("Log", mock.Anything, mock.MatchedBy(func(e service.AuditEntry) bool {
		return e.Action == "UPDATE" &&
			e.EntityType == "user" &&
			*e.EntityID == 3 &&
			*e.UserID == 3 &&
			e.IPAddress == "127.0.0.1"
	})).Return(nil).Once()

	err := svc.ChangePassword(context.Background(), 3, dto.ChangePasswordRequest{
		OldPassword:     "oldpassword123",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}, "127.0.0.1")

	require.NoError(t, err)
	repo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestAuthService_ChangePassword_ConfirmPasswordMismatch(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	err := svc.ChangePassword(context.Background(), 3, dto.ChangePasswordRequest{
		OldPassword:     "oldpassword123",
		NewPassword:     "newpassword123",
		ConfirmPassword: "differentpassword",
	}, "127.0.0.1")

	assert.Error(t, err)
	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrBusinessRule))
	assert.Equal(t, "New password and confirm password do not match", svcErr.Message)

	repo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	auditSvc.AssertNotCalled(t, "Log", mock.Anything, mock.Anything)
}

func TestAuthService_ChangePassword_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	repo.On("FindByID", mock.Anything, uint64(99)).Return(nil, nil).Once()

	err := svc.ChangePassword(context.Background(), 99, dto.ChangePasswordRequest{
		OldPassword:     "oldpassword123",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}, "127.0.0.1")

	assert.Error(t, err)
	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrNotFound))
	assert.Equal(t, "User not found", svcErr.Message)

	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	auditSvc.AssertNotCalled(t, "Log", mock.Anything, mock.Anything)
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	oldHash, _ := bcrypt.GenerateFromPassword([]byte("correctoldpassword"), cfg.BcryptCost)
	user := &model.User{
		ID:           3,
		Name:         "Admin User",
		Email:        "admin@example.com",
		PasswordHash: string(oldHash),
		Role:         model.RoleAdmin,
		IsActive:     true,
	}

	repo.On("FindByID", mock.Anything, uint64(3)).Return(user, nil).Once()

	err := svc.ChangePassword(context.Background(), 3, dto.ChangePasswordRequest{
		OldPassword:     "wrongoldpassword",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}, "127.0.0.1")

	assert.Error(t, err)
	var svcErr *response.ServiceError
	require.True(t, errors.As(err, &svcErr))
	assert.True(t, errors.Is(svcErr.Type, response.ErrUnauthorized))
	assert.Equal(t, "Invalid old password", svcErr.Message)

	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	auditSvc.AssertNotCalled(t, "Log", mock.Anything, mock.Anything)
}

func TestAuthService_ChangePassword_UpdateRepoError(t *testing.T) {
	repo := new(mockUserRepo)
	auditSvc := new(mockAuditServiceForAuth)
	cfg := testConfig()
	svc := service.NewAuthService(repo, auditSvc, cfg)

	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword123"), cfg.BcryptCost)
	user := &model.User{
		ID:           3,
		Name:         "Admin User",
		Email:        "admin@example.com",
		PasswordHash: string(oldHash),
		Role:         model.RoleAdmin,
		IsActive:     true,
	}

	repo.On("FindByID", mock.Anything, uint64(3)).Return(user, nil).Once()
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db write failure")).Once()

	err := svc.ChangePassword(context.Background(), 3, dto.ChangePasswordRequest{
		OldPassword:     "oldpassword123",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}, "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, "db write failure", err.Error())

	repo.AssertExpectations(t)
	auditSvc.AssertNotCalled(t, "Log", mock.Anything, mock.Anything)
}
