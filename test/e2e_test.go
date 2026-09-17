package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/handler"
	"github.com/ikkromm18/tokomakanan/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckE2E(t *testing.T) {
	cfg := &config.Config{
		AppName:            "tokomakanan",
		AppEnv:             "test",
		GinMode:            "test",
		JWTSecret:          "12345678901234567890123456789012",
		RateLimitRPS:       100,
		RateLimitBurst:     200,
		CORSAllowedOrigins: []string{"*"},
	}

	r := router.SetupRouter(cfg, &router.Handlers{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Service is healthy", resp.Message)
}

func TestUnauthenticatedAccessRejected(t *testing.T) {
	cfg := &config.Config{
		AppName:            "tokomakanan",
		AppEnv:             "test",
		GinMode:            "test",
		JWTSecret:          "12345678901234567890123456789012",
		RateLimitRPS:       100,
		RateLimitBurst:     200,
		CORSAllowedOrigins: []string{"*"},
	}

	r := router.SetupRouter(cfg, &router.Handlers{
		Order: &handler.OrderHandler{},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
