package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/middleware"
)

func TestRateLimiter_AllowWithinBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 1 rps, burst 3
	r.Use(middleware.RateLimiter(1, 3))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed within burst", i+1)
	}
}

func TestRateLimiter_ExceedBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 1 rps, burst 2
	r.Use(middleware.RateLimiter(1, 2))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	clientIP := "192.168.1.2:12345"

	// 2 requests pass
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = clientIP
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 3rd request should fail with 429
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = clientIP
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var resp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Message)
}

func TestRateLimiter_DifferentIPsAreIsolated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 1 rps, burst 1
	r.Use(middleware.RateLimiter(1, 1))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// IP 1 uses its burst
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// IP 1 next request throttled
	w1b := httptest.NewRecorder()
	req1b, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req1b.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w1b, req1b)
	assert.Equal(t, http.StatusTooManyRequests, w1b.Code)

	// IP 2 is unaffected and still succeeds
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req2.RemoteAddr = "10.0.0.2:1234"
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestIPRateLimiter_Cleanup(t *testing.T) {
	limiter := middleware.NewIPRateLimiter(1, 2, 10*time.Millisecond, 20*time.Millisecond)
	defer limiter.Stop()

	_ = limiter.GetLimiter("1.1.1.1")
	assert.Equal(t, 1, limiter.Count())

	// Wait for cleanup of idle IP
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, limiter.Count())
}
