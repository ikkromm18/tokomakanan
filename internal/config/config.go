package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName            string
	AppPort            string
	AppEnv             string
	GinMode            string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	JWTSecret          string
	JWTExpiryHours     int
	RateLimitRPS       int
	RateLimitBurst     int
	CORSAllowedOrigins []string
	BcryptCost         int
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	appName := getEnv("APP_NAME", "tokomakanan")
	appPort := getEnv("APP_PORT", "8080")
	appEnv := getEnv("APP_ENV", "development")
	ginMode := getEnv("GIN_MODE", "debug")

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		return nil, errors.New("DB_HOST is required")
	}
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		return nil, errors.New("DB_USER is required")
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return nil, errors.New("DB_NAME is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return nil, errors.New("JWT_SECRET is required and must be at least 32 characters")
	}

	jwtExpiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	rateLimitRPS, _ := strconv.Atoi(getEnv("RATE_LIMIT_RPS", "10"))
	rateLimitBurst, _ := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "20"))
	bcryptCost, _ := strconv.Atoi(getEnv("BCRYPT_COST", "12"))

	corsRaw := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	var origins []string
	for _, o := range strings.Split(corsRaw, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return &Config{
		AppName:            appName,
		AppPort:            appPort,
		AppEnv:             appEnv,
		GinMode:            ginMode,
		DBHost:             dbHost,
		DBPort:             dbPort,
		DBUser:             dbUser,
		DBPassword:         dbPassword,
		DBName:             dbName,
		JWTSecret:          jwtSecret,
		JWTExpiryHours:     jwtExpiryHours,
		RateLimitRPS:       rateLimitRPS,
		RateLimitBurst:     rateLimitBurst,
		CORSAllowedOrigins: origins,
		BcryptCost:         bcryptCost,
	}, nil
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
