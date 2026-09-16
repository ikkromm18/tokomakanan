package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ikkromm18/tokomakanan/internal/middleware"
)

func TestLoggerMiddleware_SuccessLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var logBuf bytes.Buffer
	r.Use(middleware.LoggerWithWriter(&logBuf))

	r.GET("/api/v1/test", func(c *gin.Context) {
		c.Set("user_id", uint64(123))
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/test", strings.NewReader(`{"sensitive":"password123"}`))
	req.RemoteAddr = "192.168.1.50:54321"
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Response should have X-Request-ID header
	reqID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, reqID)

	// Parse JSON log line
	logOutput := logBuf.String()
	require.NotEmpty(t, logOutput)

	var logEntry map[string]any
	err := json.Unmarshal([]byte(logOutput), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/api/v1/test", logEntry["path"])
	assert.Equal(t, float64(200), logEntry["status"])
	assert.Contains(t, logEntry["client_ip"], "192.168.1.50")
	assert.Equal(t, reqID, logEntry["request_id"])
	assert.Equal(t, float64(123), logEntry["user_id"])
	assert.NotNil(t, logEntry["latency_ms"])

	// Assert sensitive request body is NEVER logged
	assert.NotContains(t, logOutput, "password123")
	// Assert normal response body is NOT logged
	assert.Nil(t, logEntry["error_response"])
}

func TestLoggerMiddleware_PreservesExistingRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var logBuf bytes.Buffer
	r.Use(middleware.LoggerWithWriter(&logBuf))

	r.GET("/api/v1/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	customReqID := "custom-req-id-uuid-12345"
	req.Header.Set("X-Request-ID", customReqID)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, customReqID, w.Header().Get("X-Request-ID"))

	var logEntry map[string]any
	err := json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, customReqID, logEntry["request_id"])
}

func TestLoggerMiddleware_LogsResponseBodyOnError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	var logBuf bytes.Buffer
	r.Use(middleware.LoggerWithWriter(&logBuf))

	r.GET("/api/v1/error", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Validation failed"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/error", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var logEntry map[string]any
	err := json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)
	assert.Equal(t, float64(400), logEntry["status"])
	assert.NotNil(t, logEntry["error_response"])
	assert.Contains(t, logEntry["error_response"], "Validation failed")
}
