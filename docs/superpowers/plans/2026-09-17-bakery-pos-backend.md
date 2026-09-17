# Backend Bakery POS & Management System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a production-ready Backend API for a Bakery POS and Management System using Go 1.22+, Gin, GORM, MySQL, JWT, and Zerolog following 3-layer Clean Architecture.

**Architecture:** Clean 3-Layer Architecture (`handler/` -> `service/` -> `repository/`), communicating via Go interfaces. Handlers manage HTTP binding and response formatting, Services encapsulate business rules, transactions, and audit logging, and Repositories handle database operations using GORM.

**Tech Stack:** Go 1.22+, Gin Gonic v1.9+, GORM v1.25+, MySQL Driver v1.5+, golang-migrate v4.17+, golang-jwt/jwt/v5, zerolog v1.32+, testify v1.9+.

## Global Constraints

- Go 1.22+ clean 3-layer architecture: Handler -> Service -> Repository. Handlers do not call repositories directly; services do not import `gin` or `net/http`.
- Database: MySQL 8.0 with database name `tokomakanan`, user `root`, password `ikrom1214` (configurable via `.env`).
- Git remote: `git@github.com:ikkromm18/tokomakanan.git`. Every medium-scale milestone (end of each phase) must be committed and pushed to `origin main`.
- Business rules: `sell_price >= hpp`, `pickup_date` required for `PRE_ORDER`, overpayment prevention, snapshot of prices on order items, role-based visibility (admins cannot see HPP / profit).
- Public invoice endpoint `/api/v1/public/invoice/:token` requires NO authentication and NEVER exposes HPP or gross profit.

---

## Phase 1: Foundation, Environment, DB Migrations & Scaffolding

### Task 1: Go Module Initialization, Config & Env Setup

**Files:**
- Create: `go.mod`
- Create: `.env.example`
- Create: `.env`
- Create: `.gitignore`
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Interfaces:**
- Produces: `config.LoadConfig() (*Config, error)`, `config.Config` struct

- [ ] **Step 1: Write the failing test for Config loading**

Create `internal/config/config_test.go`:
```go
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
```

- [ ] **Step 2: Initialize module and run test to verify failure**

Run:
```bash
go mod init github.com/ikkromm18/tokomakanan
go get -u github.com/stretchr/testify
go test ./internal/config/...
```
Expected: FAIL compilation error (package config undefined)

- [ ] **Step 3: Implement Config loader and environment templates**

Create `.gitignore`:
```gitignore
bin/
*.exe
*.out
.env
tmp/
vendor/
.idea/
.vscode/
```

Create `.env.example` and `.env`:
```env
APP_NAME=tokomakanan
APP_PORT=8080
APP_ENV=development
GIN_MODE=debug

DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=ikrom1214
DB_NAME=tokomakanan

JWT_SECRET=super-secret-jwt-key-with-minimum-32-chars
JWT_EXPIRY_HOURS=24

RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20

CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

BCRYPT_COST=12
```

Create `internal/config/config.go`:
```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run:
```bash
go get -u github.com/joho/godotenv
go test -v ./internal/config/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum .gitignore .env.example internal/config/
git commit -m "feat: setup go module and environment config loader"
```

---

### Task 2: Database Connection, Response & Pagination Helpers, and Health Check

**Files:**
- Create: `internal/database/mysql.go`
- Create: `internal/pkg/response/response.go`
- Create: `internal/pkg/pagination/pagination.go`
- Create: `internal/dto/common_dto.go`
- Create: `cmd/api/main.go`
- Test: `internal/pkg/response/response_test.go`
- Test: `internal/pkg/pagination/pagination_test.go`

**Interfaces:**
- Produces: `database.NewMySQLConnection(cfg *config.Config) (*gorm.DB, error)`
- Produces: `response.Success(c, status, message, data)`, `response.Paginated(c, message, data, meta)`, `response.Error(c, status, message)`
- Produces: `pagination.FormatPagination(page, limit, totalRows)`

- [ ] **Step 1: Write failing tests for response & pagination helpers**

Create `internal/pkg/pagination/pagination_test.go`:
```go
package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatPagination(t *testing.T) {
	meta := FormatPagination(1, 20, 45)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 20, meta.Limit)
	assert.Equal(t, int64(45), meta.TotalRows)
	assert.Equal(t, 3, meta.TotalPages)
}
```

Create `internal/pkg/response/response_test.go`:
```go
package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, http.StatusOK, "Operation succeeded", map[string]string{"foo": "bar"})

	assert.Equal(t, http.StatusOK, w.Code)
	var resp StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Operation succeeded", resp.Message)
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test ./internal/pkg/pagination/... ./internal/pkg/response/...`
Expected: FAIL (packages not found)

- [ ] **Step 3: Implement Database connection, DTOs, Response, Pagination, and Health Check**

Create `internal/dto/common_dto.go`:
```go
package dto

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalRows  int64 `json:"total_rows"`
	TotalPages int   `json:"total_pages"`
}

type StandardResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type PaginatedResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    any            `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors,omitempty"`
}
```

Create `internal/pkg/pagination/pagination.go`:
```go
package pagination

import (
	"math"

	"github.com/ikkromm18/tokomakanan/internal/dto"
)

func FormatPagination(page, limit int, totalRows int64) dto.PaginationMeta {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))
	if totalPages == 0 && totalRows > 0 {
		totalPages = 1
	}

	return dto.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}
}

func GetOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return (page - 1) * limit
}
```

Create `internal/pkg/response/response.go`:
```go
package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ikkromm18/tokomakanan/internal/dto"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrDuplicate    = errors.New("resource already exists")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBusinessRule = errors.New("business rule violation")
)

type ServiceError struct {
	Type    error
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewServiceError(errType error, message string) error {
	return &ServiceError{Type: errType, Message: message}
}

func Success(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, dto.StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data any) {
	Success(c, http.StatusCreated, message, data)
}

func Paginated(c *gin.Context, message string, data any, meta dto.PaginationMeta) {
	c.JSON(http.StatusOK, dto.PaginatedResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, dto.ErrorResponse{
		Success: false,
		Message: message,
	})
}

func ValidationError(c *gin.Context, err error) {
	var valErrors []dto.FieldError
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			valErrors = append(valErrors, dto.FieldError{
				Field:   fe.Field(),
				Message: fe.Tag() + " validation failed",
			})
		}
	}
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  valErrors,
	})
}

func HandleServiceError(c *gin.Context, err error) {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		switch {
		case errors.Is(serviceErr.Type, ErrNotFound):
			Error(c, http.StatusNotFound, serviceErr.Message)
		case errors.Is(serviceErr.Type, ErrDuplicate):
			Error(c, http.StatusConflict, serviceErr.Message)
		case errors.Is(serviceErr.Type, ErrUnauthorized):
			Error(c, http.StatusUnauthorized, serviceErr.Message)
		case errors.Is(serviceErr.Type, ErrForbidden):
			Error(c, http.StatusForbidden, serviceErr.Message)
		case errors.Is(serviceErr.Type, ErrBusinessRule):
			Error(c, http.StatusUnprocessableEntity, serviceErr.Message)
		default:
			Error(c, http.StatusBadRequest, serviceErr.Message)
		}
		return
	}
	Error(c, http.StatusInternalServerError, "Internal server error")
}
```

Create `internal/database/mysql.go`:
```go
package database

import (
	"fmt"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQLConnection(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	gormLogger := logger.Default.LogMode(logger.Warn)
	if cfg.AppEnv == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
```

- [ ] **Step 4: Run tests and verify health check**

Run:
```bash
go get -u gorm.io/gorm gorm.io/driver/mysql github.com/gin-gonic/gin
go test -v ./internal/pkg/pagination/... ./internal/pkg/response/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pkg/ internal/dto/ internal/database/
git commit -m "feat: implement database connector, response and pagination packages"
```

---

### Task 3: SQL Migrations for All 10 Tables & Superadmin Seed

**Files:**
- Create: `migrations/000001_create_users_table.up.sql`
- Create: `migrations/000001_create_users_table.down.sql`
- Create: `migrations/000002_create_store_settings_table.up.sql`
- Create: `migrations/000002_create_store_settings_table.down.sql`
- Create: `migrations/000003_create_customers_table.up.sql`
- Create: `migrations/000003_create_customers_table.down.sql`
- Create: `migrations/000004_create_products_table.up.sql`
- Create: `migrations/000004_create_products_table.down.sql`
- Create: `migrations/000005_create_product_packages_table.up.sql`
- Create: `migrations/000005_create_product_packages_table.down.sql`
- Create: `migrations/000006_create_package_items_table.up.sql`
- Create: `migrations/000006_create_package_items_table.down.sql`
- Create: `migrations/000007_create_orders_table.up.sql`
- Create: `migrations/000007_create_orders_table.down.sql`
- Create: `migrations/000008_create_order_items_table.up.sql`
- Create: `migrations/000008_create_order_items_table.down.sql`
- Create: `migrations/000009_create_order_payments_table.up.sql`
- Create: `migrations/000009_create_order_payments_table.down.sql`
- Create: `migrations/000010_create_audit_logs_table.up.sql`
- Create: `migrations/000010_create_audit_logs_table.down.sql`
- Create: `migrations/000011_seed_superadmin.up.sql`
- Create: `migrations/000011_seed_superadmin.down.sql`
- Create: `internal/database/migrator.go`

- [ ] **Step 1: Write all migration up/down SQL files strictly according to PRD section 6**

Ensure tables match schema:
- `users`: id, name, email, password_hash, role ENUM('superadmin','owner','admin'), is_active, created_at, updated_at, deleted_at
- `store_settings`: id, name, address, phone, logo_url, receipt_footer, updated_at
- `customers`: id, name, phone UNIQUE, address, created_at, updated_at, deleted_at
- `products`: id, name, category, hpp DECIMAL(12,2), sell_price DECIMAL(12,2), is_active, created_at, updated_at, deleted_at
- `product_packages`: id, name, total_hpp DECIMAL(12,2), sell_price DECIMAL(12,2), is_active, created_at, updated_at, deleted_at
- `package_items`: id, package_id FK, product_id FK, quantity INT UNSIGNED, UNIQUE(package_id, product_id)
- `orders`: id, invoice_no UNIQUE, invoice_token CHAR(36) UNIQUE, customer_id FK NULLABLE, user_id FK, order_type ENUM, status ENUM, pickup_date DATE NULLABLE, notes TEXT, subtotal DECIMAL(14,2), discount_amount DECIMAL(14,2), total_amount DECIMAL(14,2), total_hpp DECIMAL(14,2), total_paid DECIMAL(14,2), payment_method ENUM, created_at, updated_at, deleted_at
- `order_items`: id, order_id FK, item_type ENUM('PRODUCT','PACKAGE'), item_id BIGINT UNSIGNED, item_name VARCHAR(150), quantity INT UNSIGNED, unit_price DECIMAL(12,2), unit_hpp DECIMAL(12,2), subtotal DECIMAL(14,2), created_at
- `order_payments`: id, order_id FK, amount DECIMAL(14,2), payment_method ENUM, notes VARCHAR(255), paid_at TIMESTAMP, created_at
- `audit_logs`: id, user_id BIGINT UNSIGNED NULLABLE, action VARCHAR(30), entity_type VARCHAR(50), entity_id BIGINT UNSIGNED NULLABLE, old_value JSON NULLABLE, new_value JSON NULLABLE, ip_address VARCHAR(45), created_at
- `seed_superadmin`: bcrypt hash of default password `SuperAdmin123!` with cost 12.

- [ ] **Step 2: Implement Golang runner or CLI runner for migrations**

Create `internal/database/migrator.go` using `github.com/golang-migrate/migrate/v4` with mysql driver.

- [ ] **Step 3: Execute migrations on local MySQL `tokomakanan`**

Run:
```bash
go run cmd/api/main.go --migrate
```
or run migrate script directly against `mysql -u root -pikrom1214 tokomakanan`.
Verify with `SHOW TABLES;` and `SELECT id, email, role FROM users;`.

- [ ] **Step 4: Commit & Git Push Phase 1**

```bash
git add migrations/ internal/database/
git commit -m "feat: add schema migrations and initial superadmin seed"
git push -u origin main
```

---

## Phase 2: Authentication, Security Middleware, User Management & Store Settings

### Task 4: JWT, Password Hash & Middleware Stack

**Files:**
- Create: `internal/pkg/jwt/jwt.go`
- Create: `internal/middleware/auth.go`
- Create: `internal/middleware/rbac.go`
- Create: `internal/middleware/cors.go`
- Create: `internal/middleware/rate_limiter.go`
- Create: `internal/middleware/logger.go`
- Create: `internal/middleware/recovery.go`
- Test: `internal/pkg/jwt/jwt_test.go`
- Test: `internal/middleware/auth_test.go`
- Test: `internal/middleware/rbac_test.go`

**Interfaces:**
- Produces: `jwt.GenerateToken(userID uint64, role string, secret string, expiryHours int) (string, error)`
- Produces: `jwt.ValidateToken(tokenString, secret string) (*JWTClaims, error)`
- Produces: `middleware.Auth(jwtSecret string) gin.HandlerFunc`
- Produces: `middleware.RequireRoles(roles ...string) gin.HandlerFunc`
- Produces: `middleware.CORS(origins []string) gin.HandlerFunc`
- Produces: `middleware.RateLimiter(rps, burst int) gin.HandlerFunc`
- Produces: `middleware.Logger() gin.HandlerFunc`
- Produces: `middleware.Recovery() gin.HandlerFunc`

- [ ] **Step 1: Write failing tests for JWT generation/validation and RBAC middleware**
- [ ] **Step 2: Run tests to verify failure**
- [ ] **Step 3: Implement JWT helper and Middleware components**
- [ ] **Step 4: Run tests to verify pass**
- [ ] **Step 5: Commit**

---

### Task 5: Audit Log Service & Repository

**Files:**
- Create: `internal/model/audit_log.go`
- Create: `internal/repository/audit_repository.go`
- Create: `internal/service/audit_service.go`
- Create: `internal/handler/audit_handler.go`
- Test: `internal/service/audit_service_test.go`

**Interfaces:**
- Consumes: GORM DB
- Produces: `AuditService.Log(ctx, entry AuditEntry) error` (non-blocking, never fails the calling flow)
- Produces: `AuditHandler.List(c *gin.Context)` (superadmin only)

- [ ] **Step 1: Write failing test for AuditService**
- [ ] **Step 2: Run test to verify failure**
- [ ] **Step 3: Implement AuditLog Model, Repository, Service and Handler**
- [ ] **Step 4: Run test to verify pass**
- [ ] **Step 5: Commit**

---

### Task 6: Auth Service & Handler (Login, Me, Change Password)

**Files:**
- Create: `internal/model/user.go`
- Create: `internal/dto/auth_dto.go`
- Create: `internal/repository/user_repository.go`
- Create: `internal/service/auth_service.go`
- Create: `internal/handler/auth_handler.go`
- Test: `internal/service/auth_service_test.go`
- Test: `internal/handler/auth_handler_test.go`

**Interfaces:**
- Produces: `AuthService.Login(ctx, req dto.LoginRequest) (*dto.LoginResponse, error)`
- Produces: `AuthService.GetMe(ctx, userID uint64) (*dto.UserInfo, error)`
- Produces: `AuthService.ChangePassword(ctx, userID uint64, req dto.ChangePasswordRequest) error`

- [ ] **Step 1: Write failing unit tests for AuthService** (valid credentials, invalid credentials, deactivated account, password mismatch)
- [ ] **Step 2: Run tests to verify failure**
- [ ] **Step 3: Implement UserRepository, AuthService, AuthHandler**
- [ ] **Step 4: Run tests to verify pass**
- [ ] **Step 5: Commit**

---

### Task 7: User Management (Superadmin only CRUD)

**Files:**
- Create: `internal/dto/user_dto.go`
- Create: `internal/service/user_service.go`
- Create: `internal/handler/user_handler.go`
- Test: `internal/service/user_service_test.go`

**Interfaces:**
- Produces: `UserService` interface with `List`, `Create`, `GetByID`, `Update`, `Delete` methods.
- Produces: `UserHandler` handling `/api/v1/users` endpoints guarded by `RequireRoles("superadmin")`.

- [ ] **Step 1: Write failing unit tests for UserService**
- [ ] **Step 2: Run tests to verify failure**
- [ ] **Step 3: Implement UserService and UserHandler**
- [ ] **Step 4: Run tests to verify pass**
- [ ] **Step 5: Commit**

---

### Task 8: Store Settings Module

**Files:**
- Create: `internal/model/store_setting.go`
- Create: `internal/dto/store_setting_dto.go`
- Create: `internal/repository/store_setting_repository.go`
- Create: `internal/service/store_setting_service.go`
- Create: `internal/handler/store_setting_handler.go`
- Test: `internal/service/store_setting_service_test.go`

**Interfaces:**
- Produces: `StoreSettingService.Get(ctx) (*dto.StoreSettingResponse, error)`
- Produces: `StoreSettingService.Update(ctx, req dto.UpdateStoreSettingRequest) (*dto.StoreSettingResponse, error)`

- [ ] **Step 1: Write failing unit test for StoreSettingService**
- [ ] **Step 2: Run test to verify failure**
- [ ] **Step 3: Implement StoreSettings repository, service, handler**
- [ ] **Step 4: Run test to verify pass**
- [ ] **Step 5: Commit & Git Push Phase 2**

```bash
git add .
git commit -m "feat: implement auth, rbac, audit log, user management and store settings"
git push origin main
```

---

## Phase 3: Master Data Catalog (Products, Packages & Customers)

### Task 9: Product Module (CRUD & HPP Validation)

**Files:**
- Create: `internal/model/product.go`
- Create: `internal/dto/product_dto.go`
- Create: `internal/repository/product_repository.go`
- Create: `internal/service/product_service.go`
- Create: `internal/handler/product_handler.go`
- Test: `internal/service/product_service_test.go`

**Interfaces:**
- Consumes: `AuditService`
- Produces: `ProductService.Create`, `Update`, `Delete`, `GetByID`, `List`
- Rule: `sell_price >= hpp`, name unique in category, soft delete.

- [x] **Step 1: Write failing unit test verifying `sell_price < hpp` returns ErrBusinessRule**
- [x] **Step 2: Run test to verify failure**
- [x] **Step 3: Implement Product repository, service and handler**
- [x] **Step 4: Run test to verify pass**
- [x] **Step 5: Commit**

---

### Task 10: Product Package (Bundling) Module

**Files:**
- Create: `internal/model/product_package.go`
- Create: `internal/model/package_item.go`
- Create: `internal/dto/package_dto.go`
- Create: `internal/repository/package_repository.go`
- Create: `internal/service/package_service.go`
- Create: `internal/handler/package_handler.go`
- Test: `internal/service/package_service_test.go`

**Interfaces:**
- Produces: `PackageService.Create`, `Update`, `Delete`, `GetByID`, `List`
- Rule: Auto-calculate `total_hpp = SUM(product.hpp * quantity)`. `sell_price >= total_hpp`. In update: delete-and-recreate items in single DB transaction.

- [x] **Step 1: Write failing test for package HPP calculation and validation**
- [x] **Step 2: Run test to verify failure**
- [x] **Step 3: Implement Package repository, service and handler**
- [x] **Step 4: Run test to verify pass**
- [x] **Step 5: Commit**

---

### Task 11: Customer Module

**Files:**
- Create: `internal/model/customer.go`
- Create: `internal/dto/customer_dto.go`
- Create: `internal/repository/customer_repository.go`
- Create: `internal/service/customer_service.go`
- Create: `internal/handler/customer_handler.go`
- Test: `internal/service/customer_service_test.go`

**Interfaces:**
- Produces: `CustomerService.Create`, `Update`, `GetByID`, `List`
- Rule: Phone is required and unique identifier.

- [x] **Step 1: Write failing test for customer creation and duplicate phone validation**
- [x] **Step 2: Run test to verify failure**
- [x] **Step 3: Implement Customer repository, service and handler**
- [x] **Step 4: Run test to verify pass**
- [x] **Step 5: Commit & Git Push Phase 3**

```bash
git add .
git commit -m "feat: implement products, bundling packages and customers modules"
git push origin main
```

---

## Phase 4: Transactions, Invoicing, Orders & Payments

### Task 12: Invoice Number Generator

**Files:**
- Create: `internal/pkg/invoice/invoice.go`
- Test: `internal/pkg/invoice/invoice_test.go`

**Interfaces:**
- Produces: `invoice.GenerateInvoiceNumber(db *gorm.DB, date time.Time) (string, error)` (Format: `INV/YYYYMMDD/XXXXX`)

- [x] **Step 1: Write failing test for invoice number generation**
- [x] **Step 2: Run test to verify failure**
- [x] **Step 3: Implement invoice counter query and formatting**
- [x] **Step 4: Run test to verify pass**
- [x] **Step 5: Commit**

---

### Task 13: Order Module (Direct Sale & Pre-Order)

**Files:**
- Create: `internal/model/order.go`
- Create: `internal/model/order_item.go`
- Create: `internal/dto/order_dto.go`
- Create: `internal/repository/order_repository.go`
- Create: `internal/service/order_service.go`
- Create: `internal/handler/order_handler.go`
- Test: `internal/service/order_service_test.go`

**Interfaces:**
- Produces: `OrderService.Create(ctx, userID uint64, req dto.CreateOrderRequest) (*dto.OrderResponse, error)`
- Produces: `OrderService.UpdateStatus(ctx, orderID uint64, newStatus string) error`
- Rules:
  - Snapshots: `item_name`, `unit_price`, `unit_hpp` saved at transaction time.
  - Pre-order validation: `pickup_date` required and >= today.
  - Status transition state machine strictly enforced.
  - Single DB transaction for Order + OrderItems + initial Payment.

- [x] **Step 1: Write failing tests for Order creation and status transition validation**
- [x] **Step 2: Run tests to verify failure**
- [x] **Step 3: Implement Order repository, service, handler**
- [x] **Step 4: Run tests to verify pass**
- [x] **Step 5: Commit**

---

### Task 14: Payment Module & Status Transition

**Files:**
- Create: `internal/model/order_payment.go`
- Create: `internal/dto/payment_dto.go`
- Create: `internal/repository/payment_repository.go`
- Create: `internal/service/payment_service.go`
- Create: `internal/handler/payment_handler.go`
- Test: `internal/service/payment_service_test.go`

**Interfaces:**
- Produces: `PaymentService.Create(ctx, orderID uint64, req dto.CreatePaymentRequest) (*dto.PaymentResponse, error)`
- Produces: `PaymentService.ListByOrder(ctx, orderID uint64) ([]dto.PaymentResponse, error)`
- Rules: Prevent overpayment (`total_paid + amount <= total_amount`). Trigger status transition (DRAFT -> DP_PAID / PAID).

- [x] **Step 1: Write failing test verifying overpayment is rejected**
- [x] **Step 2: Run test to verify failure**
- [x] **Step 3: Implement Payment repository, service, handler**
- [x] **Step 4: Run test to verify pass**
- [x] **Step 5: Commit**

---

### Task 15: Public Invoice Module

**Files:**
- Create: `internal/dto/public_invoice_dto.go`
- Create: `internal/service/public_service.go`
- Create: `internal/handler/public_handler.go`
- Test: `internal/service/public_service_test.go`

**Interfaces:**
- Produces: `PublicService.GetInvoiceByToken(ctx, token string) (*dto.PublicInvoiceResponse, error)`
- Rules: Unauthenticated, strictly excludes HPP and Gross Profit.

- [x] **Step 1: Write failing test verifying public invoice does not expose HPP**
- [x] **Step 2: Run test to verify failure**
- [x] **Step 3: Implement Public invoice service and handler**
- [x] **Step 4: Run test to verify pass**
- [x] **Step 5: Commit & Git Push Phase 4**

```bash
git add .
git commit -m "feat: implement orders, payments, invoicing and public receipt token endpoint"
git push origin main
```

---

## Phase 5: Reporting, Analytics Dashboard & Final Integration

### Task 16: Dashboard Module

**Files:**
- Create: `internal/dto/dashboard_dto.go`
- Create: `internal/repository/dashboard_repository.go`
- Create: `internal/service/dashboard_service.go`
- Create: `internal/handler/dashboard_handler.go`
- Test: `internal/service/dashboard_service_test.go`

**Interfaces:**
- Produces: `DashboardService.GetSummary(ctx, userRole string) (*dto.DashboardSummaryResponse, error)`
- Produces: `DashboardService.GetChart(ctx, startDate, endDate string) ([]dto.SalesChartPoint, error)`
- Produces: `DashboardService.GetPOReminders(ctx) ([]dto.OrderResponse, error)`
- Rules: Admin role receives summary with `today_gross_profit` omitted.

- [ ] **Step 1: Write failing test verifying admin dashboard hides `today_gross_profit`**
- [ ] **Step 2: Run test to verify failure**
- [ ] **Step 3: Implement Dashboard repository, service, handler**
- [ ] **Step 4: Run test to verify pass**
- [ ] **Step 5: Commit**

---

### Task 17: Reports Module & CSV Export

**Files:**
- Create: `internal/dto/report_dto.go`
- Create: `internal/repository/report_repository.go`
- Create: `internal/service/report_service.go`
- Create: `internal/handler/report_handler.go`
- Test: `internal/service/report_service_test.go`

**Interfaces:**
- Produces: `ReportService.GetSalesReport(ctx, req dto.ReportFilterRequest) ([]dto.SalesReportResponse, error)`
- Produces: `ReportService.GetProfitReport(ctx, req dto.ReportFilterRequest) ([]dto.ProfitReportResponse, error)`
- Produces: `ReportService.ExportCSV(ctx, req dto.ReportFilterRequest) ([]byte, string, error)`
- Rules: Superadmin & Owner only. Export with UTF-8 BOM.

- [ ] **Step 1: Write failing test for sales and profit aggregation calculation**
- [ ] **Step 2: Run test to verify failure**
- [ ] **Step 3: Implement Report repository, service, handler**
- [ ] **Step 4: Run test to verify pass**
- [ ] **Step 5: Commit**

---

### Task 18: Full Router Wiring, Dockerfile, Makefile & End-to-End Verification

**Files:**
- Create: `internal/router/router.go`
- Modify: `cmd/api/main.go`
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `Makefile`
- Test: `test/e2e_test.go`

**Interfaces:**
- Consumes: All handlers and middlewares
- Produces: Full Gin HTTP server with `/health` and `/api/v1/*` routes wired.

- [ ] **Step 1: Wire all routes according to PRD section 8 & middleware stack section 10**
- [ ] **Step 2: Implement Dockerfile, docker-compose.yml, and Makefile**
- [ ] **Step 3: Run full test suite with coverage**
Run: `go test -v -cover ./...`
Expected: ALL PASS with Service coverage >= 80%, Handler >= 70%.
- [ ] **Step 4: Verify application launch and health check**
Run: `go run cmd/api/main.go` and check `curl http://localhost:8080/health`.
- [ ] **Step 5: Final Commit & Git Push Phase 5 to GitHub**

```bash
git add .
git commit -m "feat: complete backend wiring, docker setup, makefile and end-to-end verification"
git push origin main
```
