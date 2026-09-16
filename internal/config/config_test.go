package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Success(t *testing.T) {
	os.Setenv("APP_NAME", "tokomakanan")
	os.Setenv("APP_PORT", "8080")
	os.Setenv("APP_ENV", "development")
	os.Setenv("GIN_MODE", "debug")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_PASSWORD", "ikrom1214")
	os.Setenv("DB_NAME", "tokomakanan")
	os.Setenv("JWT_SECRET", "super-secret-jwt-key-with-minimum-32-chars")
	os.Setenv("JWT_EXPIRY_HOURS", "24")
	os.Setenv("RATE_LIMIT_RPS", "10")
	os.Setenv("RATE_LIMIT_BURST", "20")
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	os.Setenv("BCRYPT_COST", "12")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "tokomakanan", cfg.AppName)
	assert.Equal(t, "8080", cfg.AppPort)
	assert.Equal(t, "root", cfg.DBUser)
	assert.Equal(t, 24, cfg.JWTExpiryHours)
	assert.Contains(t, cfg.CORSAllowedOrigins, "http://localhost:3000")
}

func TestLoadConfig_MissingRequired(t *testing.T) {
	os.Clearenv()
	_, err := LoadConfig()
	assert.Error(t, err)
}
