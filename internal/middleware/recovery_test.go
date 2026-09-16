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

func TestRecoveryMiddleware_CatchesPanicAndHidesStack(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(middleware.Recovery())

	r.GET("/panic", func(c *gin.Context) {
		panic("database connection blew up unexpectedly!")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "Internal server error", resp.Message)

	// Verify panic message and stack trace are NEVER exposed to client
	assert.NotContains(t, w.Body.String(), "blew up unexpectedly")
	assert.NotContains(t, w.Body.String(), "runtime/debug.Stack")
	assert.NotContains(t, w.Body.String(), "goroutine")
}

func TestRecoveryMiddleware_SetsSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(middleware.Recovery())

	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
}

func TestRecoveryMiddleware_SecurityHeadersOnPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(middleware.Recovery())

	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
}
