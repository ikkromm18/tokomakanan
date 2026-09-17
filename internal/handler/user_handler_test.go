package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/ikkromm18/tokomakanan/internal/middleware"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) List(ctx context.Context, page, limit int, role, search string) ([]dto.UserResponse, dto.PaginationMeta, error) {
	args := m.Called(ctx, page, limit, role, search)
	if args.Get(0) == nil {
		return nil, args.Get(1).(dto.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]dto.UserResponse), args.Get(1).(dto.PaginationMeta), args.Error(2)
}

func (m *mockUserService) Create(ctx context.Context, req dto.CreateUserRequest, actorID uint64, ipAddress string) (*dto.UserResponse, error) {
	args := m.Called(ctx, req, actorID, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *mockUserService) GetByID(ctx context.Context, id uint64) (*dto.UserResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *mockUserService) Update(ctx context.Context, id uint64, req dto.UpdateUserRequest, actorID uint64, ipAddress string) (*dto.UserResponse, error) {
	args := m.Called(ctx, id, req, actorID, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *mockUserService) Delete(ctx context.Context, id uint64, actorID uint64, ipAddress string) error {
	args := m.Called(ctx, id, actorID, ipAddress)
	return args.Error(0)
}

func setupUserRouter(h *handler.UserHandler, userRole string, userID uint64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1")
	if userRole != "" || userID > 0 {
		api.Use(func(c *gin.Context) {
			if userRole != "" {
				c.Set("user_role", userRole)
			}
			if userID > 0 {
				c.Set("user_id", userID)
			}
			c.Next()
		})
	}

	users := api.Group("/users", middleware.RequireRoles("superadmin"))
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.GET("/:id", h.GetByID)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
	}
	return r
}

// ---------------- RBAC Tests ----------------

func TestUserHandler_RBAC_SuperadminAllowed(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	svc.On("List", mock.Anything, 1, 20, "", "").Return([]dto.UserResponse{}, dto.PaginationMeta{Page: 1, Limit: 20}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_RBAC_OwnerForbidden(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "owner", 2)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUserHandler_RBAC_AdminForbidden(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "admin", 3)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUserHandler_RBAC_UnauthenticatedForbidden(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "", 0)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---------------- List Tests ----------------

func TestUserHandler_List_Success(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	expectedUsers := []dto.UserResponse{
		{ID: 1, Name: "User 1", Email: "u1@test.com", Role: model.RoleSuperadmin, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Name: "User 2", Email: "u2@test.com", Role: model.RoleOwner, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	expectedMeta := dto.PaginationMeta{Page: 2, Limit: 10, TotalRows: 25, TotalPages: 3}

	svc.On("List", mock.Anything, 2, 10, "owner", "User").Return(expectedUsers, expectedMeta, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users?page=2&limit=10&role=owner&search=User", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Users retrieved successfully", resp.Message)
	assert.Equal(t, 2, resp.Meta.Page)
	assert.Equal(t, int64(25), resp.Meta.TotalRows)
}

func TestUserHandler_List_ServiceError(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	svc.On("List", mock.Anything, 1, 20, "", "").Return(nil, dto.PaginationMeta{}, errors.New("db error"))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------- Create Tests ----------------

func TestUserHandler_Create_Success(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	createReq := dto.CreateUserRequest{
		Name:     "New Cashier",
		Email:    "cashier@example.com",
		Password: "password123",
		Role:     model.RoleAdmin,
	}
	expectedResp := &dto.UserResponse{
		ID:        10,
		Name:      "New Cashier",
		Email:     "cashier@example.com",
		Role:      model.RoleAdmin,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	svc.On("Create", mock.Anything, createReq, uint64(1), mock.Anything).Return(expectedResp, nil)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "User created successfully", resp.Message)
}

func TestUserHandler_Create_ValidationError_InvalidEmail(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	invalidReq := map[string]any{
		"name":     "Bob",
		"email":    "not-an-email",
		"password": "password123",
		"role":     "admin",
	}
	body, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Create_ValidationError_ShortPassword(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	invalidReq := map[string]any{
		"name":     "Bob",
		"email":    "bob@example.com",
		"password": "short",
		"role":     "admin",
	}
	body, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Create_ValidationError_InvalidRole(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	invalidReq := map[string]any{
		"name":     "Bob",
		"email":    "bob@example.com",
		"password": "password123",
		"role":     "godmode",
	}
	body, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Create_DuplicateEmail(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	createReq := dto.CreateUserRequest{
		Name:     "Bob",
		Email:    "existing@example.com",
		Password: "password123",
		Role:     model.RoleAdmin,
	}
	svc.On("Create", mock.Anything, createReq, uint64(1), mock.Anything).Return(nil, response.NewServiceError(response.ErrDuplicate, "Email already registered"))

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// ---------------- GetByID Tests ----------------

func TestUserHandler_GetByID_Success(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	expectedUser := &dto.UserResponse{
		ID:        5,
		Name:      "Found User",
		Email:     "found@example.com",
		Role:      model.RoleOwner,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	svc.On("GetByID", mock.Anything, uint64(5)).Return(expectedUser, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "User retrieved successfully", resp.Message)
}

func TestUserHandler_GetByID_InvalidID(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/notanumber", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	svc.On("GetByID", mock.Anything, uint64(99)).Return(nil, response.NewServiceError(response.ErrNotFound, "User not found"))

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/99", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------- Update Tests ----------------

func TestUserHandler_Update_Success(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	newName := "Updated Name"
	updateReq := dto.UpdateUserRequest{Name: &newName}
	expectedResp := &dto.UserResponse{
		ID:        5,
		Name:      "Updated Name",
		Email:     "u5@example.com",
		Role:      model.RoleAdmin,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	svc.On("Update", mock.Anything, uint64(5), updateReq, uint64(1), mock.Anything).Return(expectedResp, nil)

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/5", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "User updated successfully", resp.Message)
}

func TestUserHandler_Update_InvalidID(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/abc", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Update_ValidationError(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	invalidReq := map[string]any{"role": "invalid_role"}
	body, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/5", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Update_DuplicateEmail(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	email := "conflict@example.com"
	updateReq := dto.UpdateUserRequest{Email: &email}
	svc.On("Update", mock.Anything, uint64(5), updateReq, uint64(1), mock.Anything).Return(nil, response.NewServiceError(response.ErrDuplicate, "Email already registered"))

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/5", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUserHandler_Update_NotFound(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	updateReq := dto.UpdateUserRequest{}
	svc.On("Update", mock.Anything, uint64(99), updateReq, uint64(1), mock.Anything).Return(nil, response.NewServiceError(response.ErrNotFound, "User not found"))

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/99", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------- Delete Tests ----------------

func TestUserHandler_Delete_Success(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	svc.On("Delete", mock.Anything, uint64(5), uint64(1), mock.Anything).Return(nil)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.Bytes())
}

func TestUserHandler_Delete_InvalidID(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/notanumber", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	svc.On("Delete", mock.Anything, uint64(99), uint64(1), mock.Anything).Return(response.NewServiceError(response.ErrNotFound, "User not found"))

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/99", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_Delete_InternalServerError(t *testing.T) {
	svc := new(mockUserService)
	h := handler.NewUserHandler(svc)
	router := setupUserRouter(h, "superadmin", 1)

	svc.On("Delete", mock.Anything, uint64(5), uint64(1), mock.Anything).Return(errors.New("db delete failed"))

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
