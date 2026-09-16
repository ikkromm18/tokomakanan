package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/middleware"
)

func setupRBACRouter(userRole string, allowedRoles ...string) (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Inject role into context
	r.Use(func(c *gin.Context) {
		if userRole != "" {
			c.Set("user_role", userRole)
		}
		c.Next()
	})

	r.GET("/admin-only", middleware.RequireRoles(allowedRoles...), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	return r, w
}

func TestRequireRoles_Allowed(t *testing.T) {
	r, w := setupRBACRouter("superadmin", "superadmin", "owner")
	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRoles_Forbidden(t *testing.T) {
	r, w := setupRBACRouter("admin", "superadmin", "owner")
	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Message)
}

func TestRequireRoles_MissingRole(t *testing.T) {
	r, w := setupRBACRouter("", "superadmin")
	req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
