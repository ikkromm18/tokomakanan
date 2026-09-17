package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Login(ctx context.Context, req dto.LoginRequest, ipAddress string) (*dto.LoginResponse, error) {
	args := m.Called(ctx, req, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

func (m *mockAuthService) GetMe(ctx context.Context, userID uint64) (*dto.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserInfo), args.Error(1)
}

func (m *mockAuthService) ChangePassword(ctx context.Context, userID uint64, req dto.ChangePasswordRequest, ipAddress string) error {
	args := m.Called(ctx, userID, req, ipAddress)
	return args.Error(0)
}

func setupAuthRouter(h *handler.AuthHandler, currentUserID *uint64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1/auth")
	api.POST("/login", h.Login)

	protected := api.Group("")
	if currentUserID != nil {
		protected.Use(func(c *gin.Context) {
			c.Set("user_id", *currentUserID)
			c.Next()
		})
	}
	protected.GET("/me", h.Me)
	protected.PUT("/change-password", h.ChangePassword)

	return r
}

// -----------------------------------------------------------------------------
// Login Handler Tests
// -----------------------------------------------------------------------------

func TestAuthHandler_Login_Success(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	router := setupAuthRouter(h, nil)

	reqBody := dto.LoginRequest{
		Email:    "superadmin@tokomakanan.com",
		Password: "password123",
	}
	expectedResp := &dto.LoginResponse{
		Token: "jwt.token.here",
		User: dto.UserInfo{
			ID:    1,
			Name:  "Super Admin",
			Email: "superadmin@tokomakanan.com",
			Role:  model.RoleSuperadmin,
		},
	}

	svc.On("Login", mock.Anything, reqBody, mock.Anything).Return(expectedResp, nil).Once()

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Login successful", res.Message)

	dataBytes, _ := json.Marshal(res.Data)
	var loginData dto.LoginResponse
	_ = json.Unmarshal(dataBytes, &loginData)
	assert.Equal(t, "jwt.token.here", loginData.Token)
	assert.Equal(t, uint64(1), loginData.User.ID)
	assert.Equal(t, "superadmin@tokomakanan.com", loginData.User.Email)

	svc.AssertExpectations(t)
}

func TestAuthHandler_Login_ValidationError_InvalidEmail(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	router := setupAuthRouter(h, nil)

	invalidJSON := `{"email": "not-an-email", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Validation failed", res.Message)

	svc.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_Login_ValidationError_ShortPassword(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	router := setupAuthRouter(h, nil)

	invalidJSON := `{"email": "test@example.com", "password": "short"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)

	svc.AssertNotCalled(t, "Login", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	router := setupAuthRouter(h, nil)

	reqBody := dto.LoginRequest{
		Email:    "wrong@tokomakanan.com",
		Password: "wrongpassword",
	}

	svc.On("Login", mock.Anything, reqBody, mock.Anything).
		Return(nil, response.NewServiceError(response.ErrUnauthorized, "Invalid credentials")).Once()

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid credentials", res.Message)

	svc.AssertExpectations(t)
}

func TestAuthHandler_Login_DeactivatedAccount(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	router := setupAuthRouter(h, nil)

	reqBody := dto.LoginRequest{
		Email:    "deactivated@tokomakanan.com",
		Password: "password123",
	}

	svc.On("Login", mock.Anything, reqBody, mock.Anything).
		Return(nil, response.NewServiceError(response.ErrUnauthorized, "Account is deactivated")).Once()

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Account is deactivated", res.Message)

	svc.AssertExpectations(t)
}

func TestAuthHandler_Login_InternalServerError(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	router := setupAuthRouter(h, nil)

	reqBody := dto.LoginRequest{
		Email:    "admin@tokomakanan.com",
		Password: "password123",
	}

	svc.On("Login", mock.Anything, reqBody, mock.Anything).
		Return(nil, errors.New("unexpected database error")).Once()

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Internal server error", res.Message)

	svc.AssertExpectations(t)
}

// -----------------------------------------------------------------------------
// Me Handler Tests
// -----------------------------------------------------------------------------

func TestAuthHandler_Me_Success(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	uid := uint64(1)
	router := setupAuthRouter(h, &uid)

	expectedUser := &dto.UserInfo{
		ID:    1,
		Name:  "Super Admin",
		Email: "superadmin@tokomakanan.com",
		Role:  model.RoleSuperadmin,
	}

	svc.On("GetMe", mock.Anything, uint64(1)).Return(expectedUser, nil).Once()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "User profile retrieved successfully", res.Message)

	dataBytes, _ := json.Marshal(res.Data)
	var userInfo dto.UserInfo
	_ = json.Unmarshal(dataBytes, &userInfo)
	assert.Equal(t, uint64(1), userInfo.ID)
	assert.Equal(t, "superadmin@tokomakanan.com", userInfo.Email)

	svc.AssertExpectations(t)
}

func TestAuthHandler_Me_Unauthorized_NoUserInContext(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	router := setupAuthRouter(h, nil) // no user_id in context

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)

	svc.AssertNotCalled(t, "GetMe", mock.Anything, mock.Anything)
}

func TestAuthHandler_Me_NotFound(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	uid := uint64(99)
	router := setupAuthRouter(h, &uid)

	svc.On("GetMe", mock.Anything, uint64(99)).
		Return(nil, response.NewServiceError(response.ErrNotFound, "User not found")).Once()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "User not found", res.Message)

	svc.AssertExpectations(t)
}

func TestAuthHandler_Me_InternalServerError(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	uid := uint64(1)
	router := setupAuthRouter(h, &uid)

	svc.On("GetMe", mock.Anything, uint64(1)).
		Return(nil, errors.New("db timeout")).Once()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)

	svc.AssertExpectations(t)
}

// -----------------------------------------------------------------------------
// ChangePassword Handler Tests
// -----------------------------------------------------------------------------

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	uid := uint64(2)
	router := setupAuthRouter(h, &uid)

	reqBody := dto.ChangePasswordRequest{
		OldPassword:     "oldpassword123",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	svc.On("ChangePassword", mock.Anything, uint64(2), reqBody, mock.Anything).Return(nil).Once()

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/auth/change-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Password changed successfully", res.Message)

	svc.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_Unauthorized_NoUserInContext(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	router := setupAuthRouter(h, nil)

	reqBody := dto.ChangePasswordRequest{
		OldPassword:     "oldpassword123",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/auth/change-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	svc.AssertNotCalled(t, "ChangePassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_ChangePassword_ValidationError_Mismatch(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	uid := uint64(2)
	router := setupAuthRouter(h, &uid)

	invalidJSON := `{"old_password": "oldpassword123", "new_password": "newpassword123", "confirm_password": "different123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/auth/change-password", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Validation failed", res.Message)

	svc.AssertNotCalled(t, "ChangePassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_ChangePassword_WrongOldPassword(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	uid := uint64(2)
	router := setupAuthRouter(h, &uid)

	reqBody := dto.ChangePasswordRequest{
		OldPassword:     "wrongoldpassword",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	svc.On("ChangePassword", mock.Anything, uint64(2), reqBody, mock.Anything).
		Return(response.NewServiceError(response.ErrUnauthorized, "Invalid old password")).Once()

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/auth/change-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid old password", res.Message)

	svc.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_BusinessRuleViolation(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	uid := uint64(2)
	router := setupAuthRouter(h, &uid)

	reqBody := dto.ChangePasswordRequest{
		OldPassword:     "oldpassword123",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	svc.On("ChangePassword", mock.Anything, uint64(2), reqBody, mock.Anything).
		Return(response.NewServiceError(response.ErrBusinessRule, "New password cannot be same as old password")).Once()

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/auth/change-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "New password cannot be same as old password", res.Message)

	svc.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_InternalServerError(t *testing.T) {
	svc := new(mockAuthService)
	h := handler.NewAuthHandler(svc)
	uid := uint64(2)
	router := setupAuthRouter(h, &uid)

	reqBody := dto.ChangePasswordRequest{
		OldPassword:     "oldpassword123",
		NewPassword:     "newpassword123",
		ConfirmPassword: "newpassword123",
	}

	svc.On("ChangePassword", mock.Anything, uint64(2), reqBody, mock.Anything).
		Return(errors.New("fatal db error")).Once()

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/auth/change-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var res dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.False(t, res.Success)

	svc.AssertExpectations(t)
}
