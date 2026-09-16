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
	jwtpkg "github.com/ikkromm18/tokomakanan/internal/pkg/jwt"
)

const testSecret = "jwt-secret-for-middleware-testing-32-chars!"

func setupAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(middleware.Auth(testSecret))
	r.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		userRole, _ := c.Get("user_role")
		c.JSON(http.StatusOK, gin.H{
			"user_id":   userID,
			"user_role": userRole,
		})
	})
	return r
}

func TestAuthMiddleware_Success(t *testing.T) {
	r := setupAuthRouter()

	token, err := jwtpkg.GenerateToken(99, "superadmin", testSecret, 1)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, float64(99), body["user_id"])
	assert.Equal(t, "superadmin", body["user_role"])
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	r := setupAuthRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Message)
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	r := setupAuthRouter()

	testCases := []string{
		"Basic 12345",
		"Bearer",
		"Bearer ",
		"InvalidFormat token",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tc)
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			var resp dto.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.False(t, resp.Success)
		})
	}
}

func TestAuthMiddleware_InvalidOrExpiredToken(t *testing.T) {
	r := setupAuthRouter()

	// Expired token
	expiredToken, err := jwtpkg.GenerateToken(1, "admin", testSecret, -1)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Tampered token
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "Bearer "+expiredToken+"tampered")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}
