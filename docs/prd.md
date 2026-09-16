# PRD — Backend Bakery POS & Management System

> **Versi:** 1.0.0
> **Tanggal:** 17 September 2026
> **Stack:** Go 1.22+ · Gin Gonic · GORM · MySQL · JWT · golang-migrate
> **Scope:** Backend API only (Phase 1 / MVP)

---

## Daftar Isi

1. [Executive Summary](#1-executive-summary)
2. [Tech Stack & Dependencies](#2-tech-stack--dependencies)
3. [Project Architecture](#3-project-architecture)
4. [Folder Structure](#4-folder-structure)
5. [Configuration & Environment](#5-configuration--environment)
6. [Database Design](#6-database-design)
7. [User Roles & RBAC](#7-user-roles--rbac)
8. [API Specification](#8-api-specification)
9. [Module Specifications](#9-module-specifications)
10. [Middleware Stack](#10-middleware-stack)
11. [Error Handling & Response Contract](#11-error-handling--response-contract)
12. [Security Requirements](#12-security-requirements)
13. [Observability & Logging](#13-observability--logging)
14. [Testing Strategy](#14-testing-strategy)
15. [Docker & Development Setup](#15-docker--development-setup)
16. [Coding Conventions](#16-coding-conventions)
17. [Out-of-Scope (Phase 2)](#17-out-of-scope-phase-2)

---

## 1. Executive Summary

### Problem

Operasional toko roti masih mengandalkan pencatatan manual untuk pesanan, transaksi kasir, dan nota fisik. Ini menyebabkan:

- Human error tinggi pada rekapitulasi
- Tidak ada visibilitas real-time terhadap laba rugi
- Selisih data stok dan keuangan yang lambat terdeteksi

### Solution

Backend API untuk sistem POS dan manajemen operasional toko roti yang menangani:

- **Direct Sales** — penjualan langsung di kasir
- **Pre-Order (PO)** — pesanan katering dengan alur DP dan pelunasan
- **Laporan keuangan** — kalkulasi HPP dan Laba Kotor otomatis
- **Nota digital** — struk thermal + link nota publik via WhatsApp

### Scope Batasan

Dokumen ini **hanya** mencakup backend Go Gin Gonic. Frontend React TypeScript tidak dibahas di sini.

---

## 2. Tech Stack & Dependencies

| Kategori | Library | Versi Min. | Fungsi |
|---|---|---|---|
| Framework | `github.com/gin-gonic/gin` | v1.9+ | HTTP router & middleware |
| ORM | `gorm.io/gorm` | v1.25+ | Database ORM |
| DB Driver | `gorm.io/driver/mysql` | v1.5+ | MySQL driver untuk GORM |
| Migration | `github.com/golang-migrate/migrate/v4` | v4.17+ | Versioned SQL migration |
| Auth | `github.com/golang-jwt/jwt/v5` | v5.2+ | JWT token generation & validation |
| Password | `golang.org/x/crypto/bcrypt` | latest | Password hashing |
| Validation | `github.com/go-playground/validator/v10` | v10.20+ | Struct tag validation (built-in Gin) |
| Logging | `github.com/rs/zerolog` | v1.32+ | Structured JSON logging |
| Config | `github.com/joho/godotenv` | v1.5+ | Load .env file |
| UUID | `github.com/google/uuid` | v1.6+ | UUID v4 generation |
| Rate Limit | `golang.org/x/time/rate` | latest | Token bucket rate limiter |
| CORS | `github.com/gin-contrib/cors` | v1.7+ | CORS middleware |
| Testing | `github.com/stretchr/testify` | v1.9+ | Test assertions & mocks |

---

## 3. Project Architecture

Arsitektur menggunakan **Clean Architecture 3-Layer**:

```
┌─────────────────────────────────────────────────┐
│                   HTTP Request                  │
└────────────────────┬────────────────────────────┘
                     │
         ┌───────────▼───────────┐
         │   Middleware Stack    │
         │  (CORS, RateLimiter, │
         │   Logger, JWT Auth,  │
         │   RBAC)              │
         └───────────┬──────────┘
                     │
         ┌───────────▼───────────┐
         │      Handler Layer    │  ← HTTP binding, validation, response formatting
         │   (internal/handler)  │
         └───────────┬──────────┘
                     │  interface
         ┌───────────▼───────────┐
         │      Service Layer    │  ← Business logic, orchestration, rules
         │   (internal/service)  │
         └───────────┬──────────┘
                     │  interface
         ┌───────────▼───────────┐
         │    Repository Layer   │  ← Database query, GORM operations
         │  (internal/repository)│
         └───────────┬──────────┘
                     │
         ┌───────────▼───────────┐
         │       MySQL DB        │
         └───────────────────────┘
```

### Prinsip Utama

1. **Dependency flows downward** — Handler depends on Service, Service depends on Repository. Tidak boleh sebaliknya.
2. **Communication via interface** — Setiap layer berkomunikasi melalui interface Go, bukan concrete struct. Ini memudahkan unit testing dengan mock.
3. **No business logic in Handler** — Handler hanya bertanggung jawab: bind request, panggil service, format response.
4. **No HTTP concern in Service** — Service tidak boleh import `gin.Context` atau package HTTP apapun.
5. **No business logic in Repository** — Repository hanya menjalankan query database. Tidak ada if-else bisnis di sini.

---

## 4. Folder Structure

```
tokomakanan/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point aplikasi
│
├── internal/
│   ├── config/
│   │   └── config.go               # Load env vars, struct Config
│   │
│   ├── database/
│   │   └── mysql.go                # GORM connection setup
│   │
│   ├── middleware/
│   │   ├── auth.go                 # JWT authentication middleware
│   │   ├── rbac.go                 # Role-based access control
│   │   ├── cors.go                 # CORS configuration
│   │   ├── rate_limiter.go         # Rate limiting
│   │   ├── logger.go               # Request/response logging
│   │   └── recovery.go            # Panic recovery
│   │
│   ├── model/
│   │   ├── user.go                 # User entity
│   │   ├── store_setting.go        # Store settings entity
│   │   ├── customer.go             # Customer entity
│   │   ├── product.go              # Product entity
│   │   ├── product_package.go      # Package/bundling entity
│   │   ├── package_item.go         # Package items entity
│   │   ├── order.go                # Order entity
│   │   ├── order_item.go           # Order items entity
│   │   ├── order_payment.go        # Order payments entity
│   │   └── audit_log.go            # Audit log entity
│   │
│   ├── dto/
│   │   ├── auth_dto.go             # Login request/response
│   │   ├── user_dto.go             # User CRUD DTOs
│   │   ├── product_dto.go          # Product CRUD DTOs
│   │   ├── package_dto.go          # Package CRUD DTOs
│   │   ├── customer_dto.go         # Customer CRUD DTOs
│   │   ├── order_dto.go            # Order & PO DTOs
│   │   ├── payment_dto.go          # Payment DTOs
│   │   ├── report_dto.go           # Report filter & response DTOs
│   │   └── common_dto.go           # Pagination, standard response
│   │
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── product_handler.go
│   │   ├── package_handler.go
│   │   ├── customer_handler.go
│   │   ├── order_handler.go
│   │   ├── payment_handler.go
│   │   ├── report_handler.go
│   │   ├── public_handler.go       # Public invoice endpoint
│   │   ├── store_setting_handler.go
│   │   └── dashboard_handler.go
│   │
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── product_service.go
│   │   ├── package_service.go
│   │   ├── customer_service.go
│   │   ├── order_service.go
│   │   ├── payment_service.go
│   │   ├── report_service.go
│   │   ├── public_service.go
│   │   ├── store_setting_service.go
│   │   ├── dashboard_service.go
│   │   └── audit_service.go
│   │
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── product_repository.go
│   │   ├── package_repository.go
│   │   ├── customer_repository.go
│   │   ├── order_repository.go
│   │   ├── payment_repository.go
│   │   ├── report_repository.go
│   │   ├── store_setting_repository.go
│   │   ├── dashboard_repository.go
│   │   └── audit_repository.go
│   │
│   ├── router/
│   │   └── router.go               # Semua route registration
│   │
│   └── pkg/
│       ├── response/
│       │   └── response.go         # Standard JSON response helper
│       ├── pagination/
│       │   └── pagination.go       # Pagination helper
│       ├── jwt/
│       │   └── jwt.go              # JWT generate & parse helper
│       └── invoice/
│           └── invoice.go          # Invoice number generator
│
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_store_settings_table.up.sql
│   ├── 000002_create_store_settings_table.down.sql
│   ├── ... (satu file up + down per tabel)
│   └── README.md
│
├── docs/
│   └── prd.md                       # Dokumen ini
│
├── .env.example
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

### Aturan Folder

| Folder | Boleh Import | Tidak Boleh Import |
|---|---|---|
| `handler/` | `service/`, `dto/`, `pkg/`, `middleware/` | `repository/`, `database/` |
| `service/` | `repository/`, `model/`, `dto/`, `pkg/` | `handler/`, `gin`, `net/http` |
| `repository/` | `model/`, `gorm` | `handler/`, `service/`, `dto/` |
| `dto/` | standard lib only | apapun dari `internal/` |
| `model/` | `gorm`, standard lib | apapun dari `internal/` |

---

## 5. Configuration & Environment

### .env.example

```env
# Application
APP_NAME=tokomakanan
APP_PORT=8080
APP_ENV=development          # development | staging | production
GIN_MODE=debug               # debug | release

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=secret
DB_NAME=tokomakanan

# JWT
JWT_SECRET=your-super-secret-key-min-32-chars
JWT_EXPIRY_HOURS=24

# Rate Limiter
RATE_LIMIT_RPS=10            # requests per second
RATE_LIMIT_BURST=20          # burst capacity

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Bcrypt
BCRYPT_COST=12
```

### Config Struct

```go
// internal/config/config.go

type Config struct {
    AppName           string
    AppPort           string
    AppEnv            string
    GinMode           string

    DBHost            string
    DBPort            string
    DBUser            string
    DBPassword        string
    DBName            string

    JWTSecret         string
    JWTExpiryHours    int

    RateLimitRPS      int
    RateLimitBurst    int

    CORSAllowedOrigins []string

    BcryptCost        int
}
```

**Implementasi:** Gunakan `godotenv.Load()` di `main.go`, lalu baca dengan `os.Getenv()`. Buat fungsi `LoadConfig() (*Config, error)` yang memvalidasi semua required env vars ada dan memiliki format yang benar.

---

## 6. Database Design

### 6.1 Entity Relationship Diagram

```
┌─────────────┐       ┌──────────────────┐
│    users     │       │  store_settings  │
├─────────────┤       ├──────────────────┤
│ id (PK)     │       │ id (PK)          │
│ name        │       │ name             │
│ email       │       │ address          │
│ password    │       │ phone            │
│ role        │       │ logo_url         │
│ is_active   │       │ receipt_footer   │
│ created_at  │       │ updated_at       │
│ updated_at  │       └──────────────────┘
│ deleted_at  │
└─────────────┘
                                              ┌────────────────┐
┌─────────────┐                               │  package_items │
│  products   │                               ├────────────────┤
├─────────────┤    ┌──────────────────┐       │ id (PK)        │
│ id (PK)     │◄───│ product_packages │       │ package_id(FK) │──┐
│ name        │    ├──────────────────┤   ┌──►│ product_id(FK) │  │
│ category    │    │ id (PK)          │◄──┘   │ quantity       │  │
│ hpp         │    │ name             │       └────────────────┘  │
│ sell_price  │    │ total_hpp        │                           │
│ is_active   │    │ sell_price       │◄──────────────────────────┘
│ created_at  │    │ is_active        │
│ updated_at  │    │ created_at       │
│ deleted_at  │    │ updated_at       │
└─────────────┘    │ deleted_at       │
                   └──────────────────┘

┌──────────────┐
│  customers   │
├──────────────┤
│ id (PK)      │       ┌──────────────────────────────────────────┐
│ name         │       │               orders                     │
│ phone        │       ├──────────────────────────────────────────┤
│ address      │◄──────│ id (PK)                                  │
│ created_at   │       │ invoice_no        (UNIQUE)               │
│ updated_at   │       │ invoice_token     (UNIQUE, UUID v4)      │
│ deleted_at   │       │ customer_id       (FK, NULLABLE)         │
└──────────────┘       │ user_id           (FK, kasir)            │
                       │ order_type        (DIRECT_SALE | PRE_ORDER)
                       │ status            (DRAFT|DP_PAID|PAID|...)│
                       │ pickup_date       (NULLABLE, hanya PO)   │
                       │ notes             (TEXT, NULLABLE)        │
                       │ subtotal          (DECIMAL)               │
                       │ discount_amount   (DECIMAL)               │
                       │ total_amount      (DECIMAL)               │
                       │ total_hpp         (DECIMAL)               │
                       │ total_paid        (DECIMAL)               │
                       │ payment_method    (CASH|TRANSFER|QRIS)    │
                       │ created_at                                │
                       │ updated_at                                │
                       │ deleted_at                                │
                       └───────┬─────────────┬────────────────────┘
                               │             │
                ┌──────────────▼──┐   ┌──────▼──────────┐
                │   order_items   │   │ order_payments  │
                ├─────────────────┤   ├─────────────────┤
                │ id (PK)         │   │ id (PK)         │
                │ order_id (FK)   │   │ order_id (FK)   │
                │ item_type       │   │ amount          │
                │ item_id         │   │ payment_method  │
                │ item_name       │   │ notes           │
                │ quantity        │   │ paid_at         │
                │ unit_price      │   │ created_at      │
                │ unit_hpp        │   └─────────────────┘
                │ subtotal        │
                │ created_at      │
                └─────────────────┘

┌──────────────────┐
│   audit_logs     │
├──────────────────┤
│ id (PK)          │
│ user_id (FK)     │
│ action           │   (CREATE, UPDATE, DELETE, LOGIN, STATUS_CHANGE)
│ entity_type      │   (product, order, user, dll)
│ entity_id        │
│ old_value        │   (JSON, NULLABLE)
│ new_value        │   (JSON, NULLABLE)
│ ip_address       │
│ created_at       │
└──────────────────┘
```

### 6.2 Tabel Spesifikasi Detail

#### `users`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | |
| name | VARCHAR(100) | NOT NULL | |
| email | VARCHAR(150) | NOT NULL, UNIQUE | Digunakan untuk login |
| password_hash | VARCHAR(255) | NOT NULL | Bcrypt hash |
| role | ENUM('superadmin','owner','admin') | NOT NULL | |
| is_active | TINYINT(1) | NOT NULL, DEFAULT 1 | Untuk disable akun tanpa hapus |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE | |
| deleted_at | TIMESTAMP | NULLABLE, INDEX | Soft delete |

#### `store_settings`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | Hanya 1 row |
| name | VARCHAR(100) | NOT NULL | Nama toko |
| address | TEXT | NULLABLE | Alamat toko |
| phone | VARCHAR(20) | NULLABLE | Nomor telepon toko |
| logo_url | VARCHAR(500) | NULLABLE | URL logo toko |
| receipt_footer | TEXT | NULLABLE | Teks footer di struk |
| updated_at | TIMESTAMP | NOT NULL | |

#### `customers`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | |
| name | VARCHAR(100) | NOT NULL | |
| phone | VARCHAR(20) | NOT NULL, UNIQUE | Nomor WA, identifier utama |
| address | TEXT | NULLABLE | |
| created_at | TIMESTAMP | NOT NULL | |
| updated_at | TIMESTAMP | NOT NULL | |
| deleted_at | TIMESTAMP | NULLABLE, INDEX | Soft delete |

#### `products`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | |
| name | VARCHAR(150) | NOT NULL | |
| category | VARCHAR(50) | NOT NULL | Contoh: "Roti Manis", "Kue Kering" |
| hpp | DECIMAL(12,2) | NOT NULL | Harga Pokok Penjualan |
| sell_price | DECIMAL(12,2) | NOT NULL | Harga jual ke customer |
| is_active | TINYINT(1) | NOT NULL, DEFAULT 1 | Produk bisa di-hide tanpa hapus |
| created_at | TIMESTAMP | NOT NULL | |
| updated_at | TIMESTAMP | NOT NULL | |
| deleted_at | TIMESTAMP | NULLABLE, INDEX | Soft delete |

**Validasi bisnis:** `sell_price` harus >= `hpp`. Ini divalidasi di **Service layer**, bukan di database.

#### `product_packages`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | |
| name | VARCHAR(150) | NOT NULL | |
| total_hpp | DECIMAL(12,2) | NOT NULL | Dihitung otomatis dari komponen |
| sell_price | DECIMAL(12,2) | NOT NULL | Harga paket |
| is_active | TINYINT(1) | NOT NULL, DEFAULT 1 | |
| created_at | TIMESTAMP | NOT NULL | |
| updated_at | TIMESTAMP | NOT NULL | |
| deleted_at | TIMESTAMP | NULLABLE, INDEX | Soft delete |

**Aturan:** `total_hpp` = SUM(product.hpp × package_item.quantity) dari semua komponen. Dihitung ulang setiap kali package atau komponen di-update.

#### `package_items`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | |
| package_id | BIGINT UNSIGNED | FK → product_packages.id, NOT NULL | |
| product_id | BIGINT UNSIGNED | FK → products.id, NOT NULL | |
| quantity | INT UNSIGNED | NOT NULL, MIN 1 | Jumlah produk dalam paket |

**Constraint:** UNIQUE(package_id, product_id) — satu produk tidak boleh duplikat di paket yang sama.

#### `orders`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | |
| invoice_no | VARCHAR(30) | NOT NULL, UNIQUE | Format: INV/YYYYMMDD/XXXXX |
| invoice_token | CHAR(36) | NOT NULL, UNIQUE | UUID v4, untuk public link |
| customer_id | BIGINT UNSIGNED | FK → customers.id, NULLABLE | Walk-in customer boleh NULL |
| user_id | BIGINT UNSIGNED | FK → users.id, NOT NULL | Kasir yang input |
| order_type | ENUM('DIRECT_SALE','PRE_ORDER') | NOT NULL | |
| status | ENUM('DRAFT','DP_PAID','PAID','READY','COMPLETED','CANCELLED') | NOT NULL | |
| pickup_date | DATE | NULLABLE | Wajib diisi jika order_type = PRE_ORDER |
| notes | TEXT | NULLABLE | Catatan order |
| subtotal | DECIMAL(14,2) | NOT NULL | Sum of order_items.subtotal |
| discount_amount | DECIMAL(14,2) | NOT NULL, DEFAULT 0 | Diskon manual (jika ada) |
| total_amount | DECIMAL(14,2) | NOT NULL | subtotal - discount_amount |
| total_hpp | DECIMAL(14,2) | NOT NULL | Sum of (unit_hpp × quantity) |
| total_paid | DECIMAL(14,2) | NOT NULL, DEFAULT 0 | Sum of order_payments.amount |
| payment_method | ENUM('CASH','TRANSFER','QRIS') | NOT NULL | Metode utama |
| created_at | TIMESTAMP | NOT NULL | |
| updated_at | TIMESTAMP | NOT NULL | |
| deleted_at | TIMESTAMP | NULLABLE, INDEX | Soft delete |

**Index tambahan:** INDEX pada `(status, order_type)`, `(created_at)`, `(pickup_date)`.

#### Status Flow

```
DIRECT_SALE:  DRAFT → PAID → COMPLETED
PRE_ORDER:    DRAFT → DP_PAID → PAID → READY → COMPLETED
ANY:          * → CANCELLED (kecuali COMPLETED)
```

| Status | Deskripsi |
|---|---|
| DRAFT | Order baru dibuat, belum ada pembayaran |
| DP_PAID | Down Payment sudah diterima (hanya PO) |
| PAID | Lunas penuh |
| READY | Pesanan siap diambil (hanya PO) |
| COMPLETED | Selesai / sudah diambil pelanggan |
| CANCELLED | Dibatalkan |

**Aturan transisi status (divalidasi di Service layer):**

```
Allowed transitions:
  DRAFT      → DP_PAID  (jika PO dan DP > 0)
  DRAFT      → PAID     (jika Direct Sale dan lunas, atau PO lunas langsung)
  DP_PAID    → PAID     (pelunasan PO)
  PAID       → READY    (hanya PO)
  READY      → COMPLETED
  PAID       → COMPLETED (hanya Direct Sale)
  DRAFT      → CANCELLED
  DP_PAID    → CANCELLED
  PAID       → CANCELLED (dengan catatan: refund manual)
```

#### `order_items`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | |
| order_id | BIGINT UNSIGNED | FK → orders.id, NOT NULL | |
| item_type | ENUM('PRODUCT','PACKAGE') | NOT NULL | Tipe item |
| item_id | BIGINT UNSIGNED | NOT NULL | ID product atau package |
| item_name | VARCHAR(150) | NOT NULL | Snapshot nama saat transaksi |
| quantity | INT UNSIGNED | NOT NULL, MIN 1 | |
| unit_price | DECIMAL(12,2) | NOT NULL | Snapshot harga jual |
| unit_hpp | DECIMAL(12,2) | NOT NULL | Snapshot HPP |
| subtotal | DECIMAL(14,2) | NOT NULL | quantity × unit_price |
| created_at | TIMESTAMP | NOT NULL | |

> **PENTING:** `item_name`, `unit_price`, dan `unit_hpp` adalah **snapshot** pada saat transaksi. Jika harga produk berubah kemudian, data transaksi historis tetap akurat.

#### `order_payments`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | |
| order_id | BIGINT UNSIGNED | FK → orders.id, NOT NULL | |
| amount | DECIMAL(14,2) | NOT NULL | Nominal pembayaran |
| payment_method | ENUM('CASH','TRANSFER','QRIS') | NOT NULL | |
| notes | VARCHAR(255) | NULLABLE | Keterangan (misal: "DP 50%") |
| paid_at | TIMESTAMP | NOT NULL | Waktu pembayaran |
| created_at | TIMESTAMP | NOT NULL | |

#### `audit_logs`

| Kolom | Tipe | Constraint | Keterangan |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK, AUTO_INCREMENT | |
| user_id | BIGINT UNSIGNED | FK → users.id, NULLABLE | NULL jika system action |
| action | VARCHAR(30) | NOT NULL | CREATE, UPDATE, DELETE, LOGIN, STATUS_CHANGE |
| entity_type | VARCHAR(50) | NOT NULL | "product", "order", "user", dll |
| entity_id | BIGINT UNSIGNED | NULLABLE | ID dari entity yang di-modify |
| old_value | JSON | NULLABLE | State sebelum perubahan |
| new_value | JSON | NULLABLE | State setelah perubahan |
| ip_address | VARCHAR(45) | NULLABLE | IP address pemanggil |
| created_at | TIMESTAMP | NOT NULL | |

> **Catatan:** Tabel audit_logs TIDAK menggunakan soft delete. Data audit tidak boleh dihapus.
> **Catatan:** Tabel audit_logs TIDAK memiliki foreign key constraint ke user_id agar tidak menghambat performa dan agar log tetap tersimpan meskipun user dihapus.

### 6.3 Migration File Conventions

Menggunakan `golang-migrate`. File ditempatkan di `migrations/`.

**Format nama file:**
```
{sequence}_{description}.up.sql    # Apply migration
{sequence}_{description}.down.sql  # Rollback migration
```

**Contoh:**
```
000001_create_users_table.up.sql
000001_create_users_table.down.sql
000002_create_store_settings_table.up.sql
000002_create_store_settings_table.down.sql
000003_create_customers_table.up.sql
000003_create_customers_table.down.sql
000004_create_products_table.up.sql
000004_create_products_table.down.sql
000005_create_product_packages_table.up.sql
000005_create_product_packages_table.down.sql
000006_create_package_items_table.up.sql
000006_create_package_items_table.down.sql
000007_create_orders_table.up.sql
000007_create_orders_table.down.sql
000008_create_order_items_table.up.sql
000008_create_order_items_table.down.sql
000009_create_order_payments_table.up.sql
000009_create_order_payments_table.down.sql
000010_create_audit_logs_table.up.sql
000010_create_audit_logs_table.down.sql
000011_seed_superadmin.up.sql
000011_seed_superadmin.down.sql
```

**Aturan migration:**
- Setiap `.up.sql` HARUS punya `.down.sql` yang bisa rollback sempurna.
- `.down.sql` harus `DROP TABLE IF EXISTS` atau `ALTER TABLE DROP COLUMN` yang sesuai.
- Seed data (superadmin default) masuk sebagai migration terakhir.
- Jangan pernah mengedit migration file yang sudah di-apply di production. Buat migration baru.

### 6.4 Default Seed Data

Migration `000011_seed_superadmin.up.sql` harus membuat:

```sql
INSERT INTO users (name, email, password_hash, role, is_active, created_at, updated_at)
VALUES ('Super Admin', 'superadmin@tokomakanan.com', '<bcrypt_hash_of_default_password>', 'superadmin', 1, NOW(), NOW());
```

> **Password default:** `SuperAdmin123!` (harus diganti saat pertama kali login di production).
> **Generate hash:** Gunakan Go script atau online bcrypt generator dengan cost 12.

---

## 7. User Roles & RBAC

### 7.1 Role Definitions

| Role | Level | Deskripsi |
|---|---|---|
| `superadmin` | 1 (highest) | Full system access. Mengelola semua akun, audit log, store settings. |
| `owner` | 2 | Analitik & manajemen master data. Akses dashboard, laporan, produk, paket, pelanggan. |
| `admin` | 3 (lowest) | Operasional kasir. POS, PO, pembayaran, cetak struk. **Tidak bisa** akses laporan laba rugi. |

### 7.2 Permission Matrix

| Endpoint / Fitur | Superadmin | Owner | Admin |
|---|:---:|:---:|:---:|
| Login / Me | ✅ | ✅ | ✅ |
| **User Management** | | | |
| List Users | ✅ | ❌ | ❌ |
| Create User (owner/admin) | ✅ | ❌ | ❌ |
| Update User | ✅ | ❌ | ❌ |
| Delete User | ✅ | ❌ | ❌ |
| **Store Settings** | | | |
| View Store Settings | ✅ | ✅ | ✅ |
| Update Store Settings | ✅ | ✅ | ❌ |
| **Products** | | | |
| List Products | ✅ | ✅ | ✅ |
| Create Product | ✅ | ✅ | ❌ |
| Update Product | ✅ | ✅ | ❌ |
| Delete Product | ✅ | ✅ | ❌ |
| **Packages** | | | |
| List Packages | ✅ | ✅ | ✅ |
| Create Package | ✅ | ✅ | ❌ |
| Update Package | ✅ | ✅ | ❌ |
| Delete Package | ✅ | ✅ | ❌ |
| **Customers** | | | |
| List Customers | ✅ | ✅ | ✅ |
| Create Customer | ✅ | ✅ | ✅ |
| Update Customer | ✅ | ✅ | ✅ |
| **Orders** | | | |
| Create Order | ✅ | ✅ | ✅ |
| View Order | ✅ | ✅ | ✅ |
| List Orders | ✅ | ✅ | ✅ |
| Update Order Status | ✅ | ✅ | ✅ |
| Cancel Order | ✅ | ✅ | ✅ |
| **Payments** | | | |
| Add Payment | ✅ | ✅ | ✅ |
| **Dashboard** | | | |
| View Dashboard | ✅ | ✅ | ✅ (limited*) |
| **Reports** | | | |
| Sales Report | ✅ | ✅ | ❌ |
| Profit Report (HPP/Laba) | ✅ | ✅ | ❌ |
| Export CSV | ✅ | ✅ | ❌ |
| **Audit Logs** | | | |
| View Audit Logs | ✅ | ❌ | ❌ |

> *Admin hanya melihat dashboard ringkasan (total penjualan hari ini, PO aktif). Tidak melihat data HPP dan Laba Kotor.

### 7.3 RBAC Implementation

Middleware `rbac.go` menerima daftar role yang diizinkan:

```go
// Contoh penggunaan di router:
adminGroup.PUT("/store-settings", middleware.RequireRoles("superadmin", "owner"), storeSettingHandler.Update)
```

**Logic:**
1. Middleware `auth.go` sudah melakukan JWT validation dan menyimpan user data di `gin.Context`.
2. Middleware `rbac.go` membaca role dari context, lalu cek apakah termasuk dalam daftar yang diizinkan.
3. Jika tidak diizinkan → return `403 Forbidden`.

---

## 8. API Specification

### 8.1 Base URL

```
/api/v1
```

### 8.2 Complete Endpoint List

#### Authentication (Public)

| Method | Path | Handler | Deskripsi |
|---|---|---|---|
| POST | `/api/v1/auth/login` | `AuthHandler.Login` | Login, return JWT token |
| GET | `/api/v1/auth/me` | `AuthHandler.Me` | Get current user profile (requires auth) |
| PUT | `/api/v1/auth/change-password` | `AuthHandler.ChangePassword` | Ganti password sendiri (requires auth) |

#### User Management (Superadmin only)

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| GET | `/api/v1/users` | `UserHandler.List` | superadmin | List semua user (paginated) |
| POST | `/api/v1/users` | `UserHandler.Create` | superadmin | Create user baru |
| GET | `/api/v1/users/:id` | `UserHandler.GetByID` | superadmin | Detail user |
| PUT | `/api/v1/users/:id` | `UserHandler.Update` | superadmin | Update user |
| DELETE | `/api/v1/users/:id` | `UserHandler.Delete` | superadmin | Soft delete user |

#### Store Settings

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| GET | `/api/v1/store-settings` | `StoreSettingHandler.Get` | all | Get store settings |
| PUT | `/api/v1/store-settings` | `StoreSettingHandler.Update` | superadmin, owner | Update store settings |

#### Products

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| GET | `/api/v1/products` | `ProductHandler.List` | all | List produk (paginated, filterable) |
| POST | `/api/v1/products` | `ProductHandler.Create` | superadmin, owner | Create produk baru |
| GET | `/api/v1/products/:id` | `ProductHandler.GetByID` | all | Detail produk |
| PUT | `/api/v1/products/:id` | `ProductHandler.Update` | superadmin, owner | Update produk |
| DELETE | `/api/v1/products/:id` | `ProductHandler.Delete` | superadmin, owner | Soft delete produk |

**Query params untuk List:**

| Param | Tipe | Default | Keterangan |
|---|---|---|---|
| page | int | 1 | Halaman |
| limit | int | 20 | Jumlah per halaman (max 100) |
| search | string | "" | Search by name (LIKE) |
| category | string | "" | Filter by category (exact) |
| is_active | bool | - | Filter by active status |
| sort_by | string | "created_at" | Kolom sort (name, sell_price, created_at) |
| sort_order | string | "desc" | asc / desc |

#### Packages (Bundling)

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| GET | `/api/v1/packages` | `PackageHandler.List` | all | List paket (paginated) |
| POST | `/api/v1/packages` | `PackageHandler.Create` | superadmin, owner | Create paket baru beserta items |
| GET | `/api/v1/packages/:id` | `PackageHandler.GetByID` | all | Detail paket + items |
| PUT | `/api/v1/packages/:id` | `PackageHandler.Update` | superadmin, owner | Update paket + items |
| DELETE | `/api/v1/packages/:id` | `PackageHandler.Delete` | superadmin, owner | Soft delete paket |

#### Customers

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| GET | `/api/v1/customers` | `CustomerHandler.List` | all | List customer (paginated, searchable) |
| POST | `/api/v1/customers` | `CustomerHandler.Create` | all | Create customer baru |
| GET | `/api/v1/customers/:id` | `CustomerHandler.GetByID` | all | Detail customer |
| PUT | `/api/v1/customers/:id` | `CustomerHandler.Update` | all | Update customer |

**Query params untuk List:**

| Param | Tipe | Default | Keterangan |
|---|---|---|---|
| page | int | 1 | |
| limit | int | 20 | Max 100 |
| search | string | "" | Search by name OR phone |

#### Orders

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| POST | `/api/v1/orders` | `OrderHandler.Create` | all | Create order (Direct Sale / PO) |
| GET | `/api/v1/orders` | `OrderHandler.List` | all | List orders (paginated, filterable) |
| GET | `/api/v1/orders/:id` | `OrderHandler.GetByID` | all | Detail order + items + payments |
| PUT | `/api/v1/orders/:id/status` | `OrderHandler.UpdateStatus` | all | Update status order |

**Query params untuk List:**

| Param | Tipe | Default | Keterangan |
|---|---|---|---|
| page | int | 1 | |
| limit | int | 20 | Max 100 |
| order_type | string | "" | DIRECT_SALE / PRE_ORDER |
| status | string | "" | Filter by status |
| start_date | string | "" | Format: YYYY-MM-DD |
| end_date | string | "" | Format: YYYY-MM-DD |
| search | string | "" | Search by invoice_no or customer name |

#### Payments

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| POST | `/api/v1/orders/:id/payments` | `PaymentHandler.Create` | all | Tambah pembayaran (DP/Pelunasan) |
| GET | `/api/v1/orders/:id/payments` | `PaymentHandler.ListByOrder` | all | Histori pembayaran order |

#### Dashboard

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| GET | `/api/v1/dashboard/summary` | `DashboardHandler.Summary` | all | Ringkasan hari ini |
| GET | `/api/v1/dashboard/chart` | `DashboardHandler.Chart` | superadmin, owner | Data chart penjualan |
| GET | `/api/v1/dashboard/po-reminders` | `DashboardHandler.POReminders` | all | PO yang mendekati pickup_date |

**Dashboard Summary Response:**

```json
{
  "today_sales_count": 15,
  "today_sales_amount": 2500000,
  "today_gross_profit": 750000,
  "active_po_count": 5,
  "recent_transactions": [...]
}
```

> Admin TIDAK mendapat field `today_gross_profit`.

#### Reports (Superadmin & Owner only)

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| GET | `/api/v1/reports/sales` | `ReportHandler.Sales` | superadmin, owner | Laporan penjualan periodik |
| GET | `/api/v1/reports/profit` | `ReportHandler.Profit` | superadmin, owner | Laporan laba kotor |
| GET | `/api/v1/reports/export` | `ReportHandler.ExportCSV` | superadmin, owner | Export laporan ke CSV |

**Query params untuk Reports:**

| Param | Tipe | Required | Keterangan |
|---|---|---|---|
| start_date | string | ✅ | Format: YYYY-MM-DD |
| end_date | string | ✅ | Format: YYYY-MM-DD |
| group_by | string | ❌ | "day" / "week" / "month" (default: "day") |

#### Audit Logs (Superadmin only)

| Method | Path | Handler | Roles | Deskripsi |
|---|---|---|---|---|
| GET | `/api/v1/audit-logs` | `AuditHandler.List` | superadmin | List audit logs (paginated) |

#### Public (No Auth)

| Method | Path | Handler | Deskripsi |
|---|---|---|---|
| GET | `/api/v1/public/invoice/:token` | `PublicHandler.GetInvoice` | View invoice publik (read-only) |

#### Health Check (No Auth)

| Method | Path | Handler | Deskripsi |
|---|---|---|---|
| GET | `/health` | inline | Health check endpoint |

---

## 9. Module Specifications

### 9.1 Auth Module

#### Login Flow

```
1. Client POST /api/v1/auth/login dengan {email, password}
2. Handler bind & validate request
3. Service:
   a. Cari user by email
   b. Jika tidak ditemukan → return error "Invalid credentials"
   c. Jika user.is_active == false → return error "Account is deactivated"
   d. Compare password dengan bcrypt
   e. Jika tidak cocok → return error "Invalid credentials"
   f. Generate JWT token dengan claims: {user_id, role, exp}
   g. Catat audit log (action: LOGIN)
   h. Return token + user profile
4. Handler return response
```

**Login Request DTO:**

```go
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}
```

**Login Response DTO:**

```go
type LoginResponse struct {
    Token string   `json:"token"`
    User  UserInfo `json:"user"`
}

type UserInfo struct {
    ID    uint64 `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Role  string `json:"role"`
}
```

#### JWT Claims

```go
type JWTClaims struct {
    UserID uint64 `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}
```

- **Signing method:** HS256
- **Expiry:** Dikonfigurasi via `JWT_EXPIRY_HOURS` (default 24 jam)
- **Token dikirim via:** `Authorization: Bearer <token>` header

#### Change Password Flow

```
1. User terautentikasi POST /api/v1/auth/change-password
2. Request: {old_password, new_password, confirm_password}
3. Validate: new_password == confirm_password
4. Validate: old_password benar (bcrypt compare)
5. Hash new_password, update di database
6. Catat audit log
```

### 9.2 Product Module

#### Create Product

**Request:**
```go
type CreateProductRequest struct {
    Name      string  `json:"name" binding:"required,min=2,max=150"`
    Category  string  `json:"category" binding:"required,min=2,max=50"`
    HPP       float64 `json:"hpp" binding:"required,gt=0"`
    SellPrice float64 `json:"sell_price" binding:"required,gt=0"`
}
```

**Business Rules:**
- `sell_price` harus >= `hpp`. Jika tidak → return 422 Unprocessable Entity.
- Nama produk harus unik dalam kategori yang sama (case-insensitive).
- Audit log dicatat saat create.

#### Update Product

- Hanya field yang dikirim yang di-update (partial update).
- Jika `hpp` berubah, paket yang mengandung produk ini harus di-recalculate `total_hpp` nya.
- Audit log dicatat dengan `old_value` dan `new_value`.

#### Delete Product (Soft)

- Cek apakah produk sedang digunakan di paket aktif → jika ya, tolak penghapusan atau beri warning.
- Soft delete: set `deleted_at`.

### 9.3 Package Module

#### Create Package

**Request:**
```go
type CreatePackageRequest struct {
    Name      string             `json:"name" binding:"required,min=2,max=150"`
    SellPrice float64            `json:"sell_price" binding:"required,gt=0"`
    Items     []PackageItemInput `json:"items" binding:"required,min=1,dive"`
}

type PackageItemInput struct {
    ProductID uint64 `json:"product_id" binding:"required"`
    Quantity  int    `json:"quantity" binding:"required,min=1"`
}
```

**Business Rules:**
- Setiap `product_id` harus valid dan aktif.
- Tidak boleh ada `product_id` duplikat dalam satu paket.
- `total_hpp` dihitung otomatis: `SUM(product.hpp × quantity)`.
- `sell_price` harus >= `total_hpp`.
- Semua operasi (insert package + insert items) dalam satu database transaction.

#### Update Package

- Strategi update items: **delete-and-recreate** — hapus semua `package_items` lama, insert yang baru.
- Recalculate `total_hpp`.
- Dalam satu database transaction.

### 9.4 Order Module

#### Create Order (Direct Sale)

**Request:**
```go
type CreateOrderRequest struct {
    OrderType      string           `json:"order_type" binding:"required,oneof=DIRECT_SALE PRE_ORDER"`
    CustomerID     *uint64          `json:"customer_id"`
    PickupDate     *string          `json:"pickup_date"`
    Notes          *string          `json:"notes"`
    PaymentMethod  string           `json:"payment_method" binding:"required,oneof=CASH TRANSFER QRIS"`
    DiscountAmount float64          `json:"discount_amount" binding:"gte=0"`
    Items          []OrderItemInput `json:"items" binding:"required,min=1,dive"`
    Payment        *PaymentInput    `json:"payment"`
}

type OrderItemInput struct {
    ItemType string `json:"item_type" binding:"required,oneof=PRODUCT PACKAGE"`
    ItemID   uint64 `json:"item_id" binding:"required"`
    Quantity int    `json:"quantity" binding:"required,min=1"`
}

type PaymentInput struct {
    Amount        float64 `json:"amount" binding:"required,gt=0"`
    PaymentMethod string  `json:"payment_method" binding:"required,oneof=CASH TRANSFER QRIS"`
    Notes         string  `json:"notes"`
}
```

**Create Order Flow (Service layer):**

```
1. Validate semua items exist dan aktif
2. Untuk setiap item:
   a. Jika PRODUCT → ambil sell_price dan hpp dari products table
   b. Jika PACKAGE → ambil sell_price dan total_hpp dari product_packages table
   c. Buat snapshot: item_name, unit_price, unit_hpp
   d. Hitung subtotal = quantity × unit_price
3. Hitung:
   a. subtotal = SUM(semua order_items.subtotal)
   b. total_amount = subtotal - discount_amount
   c. total_hpp = SUM(unit_hpp × quantity)
4. Generate invoice_no: INV/YYYYMMDD/XXXXX (auto-increment per hari)
5. Generate invoice_token: UUID v4
6. Tentukan initial status:
   a. Jika ada payment dan amount >= total_amount → status = PAID
   b. Jika ada payment dan order_type = PRE_ORDER → status = DP_PAID
   c. Else → status = DRAFT
7. Simpan semua dalam 1 DB transaction:
   a. INSERT order
   b. INSERT order_items (bulk)
   c. INSERT order_payment (jika ada)
   d. UPDATE order.total_paid
8. Jika customer_id diberikan dan customer belum ada → Tidak. Customer harus sudah terdaftar.
9. Catat audit log
```

**Validasi khusus PRE_ORDER:**
- `pickup_date` wajib diisi.
- `pickup_date` harus >= hari ini.

#### Invoice Number Generator

Format: `INV/YYYYMMDD/XXXXX`

```
INV/20260917/00001
INV/20260917/00002
...
```

**Implementasi:** Query last invoice_no untuk hari tersebut, increment counter. Gunakan database lock atau transaction untuk menghindari race condition.

### 9.5 Payment Module

#### Add Payment

**Request:**
```go
type CreatePaymentRequest struct {
    Amount        float64 `json:"amount" binding:"required,gt=0"`
    PaymentMethod string  `json:"payment_method" binding:"required,oneof=CASH TRANSFER QRIS"`
    Notes         string  `json:"notes" binding:"max=255"`
}
```

**Flow:**

```
1. Ambil order by ID
2. Validasi:
   a. Order tidak dalam status COMPLETED atau CANCELLED
   b. total_paid + amount tidak melebihi total_amount (overpayment prevention)
3. Dalam DB transaction:
   a. INSERT order_payment
   b. UPDATE order.total_paid += amount
   c. Jika total_paid >= total_amount:
      - Jika status DRAFT atau DP_PAID → ubah ke PAID
   d. Jika total_paid > 0 dan < total_amount dan order_type = PRE_ORDER:
      - Jika status DRAFT → ubah ke DP_PAID
4. Catat audit log
```

### 9.6 Report Module

#### Sales Report

**Response:**
```go
type SalesReportResponse struct {
    Period       string  `json:"period"`
    TotalOrders  int     `json:"total_orders"`
    TotalAmount  float64 `json:"total_amount"`
    TotalHPP     float64 `json:"total_hpp"`
    GrossProfit  float64 `json:"gross_profit"`
    ProfitMargin float64 `json:"profit_margin"`
}
```

**Query logic:**
- Hanya hitung order dengan status `PAID`, `READY`, atau `COMPLETED`.
- Group by period (day/week/month) sesuai parameter `group_by`.

#### Export CSV

- Header CSV: Date, Invoice No, Customer, Order Type, Items Count, Subtotal, Discount, Total, HPP, Profit, Payment Method, Status
- Encoding: UTF-8 with BOM (agar Excel baca dengan benar)
- Response header: `Content-Type: text/csv` dan `Content-Disposition: attachment; filename="sales_report_YYYY-MM-DD.csv"`

### 9.7 Public Invoice Module

#### GET /api/v1/public/invoice/:token

- Cari order by `invoice_token`.
- Jika tidak ditemukan → 404.
- Return data invoice lengkap: store info, customer info, items, payments, total.
- **TIDAK** termasuk HPP (ini data internal).
- Endpoint ini **TIDAK** memerlukan authentication.

**Response:**
```go
type PublicInvoiceResponse struct {
    Store      StoreInfo     `json:"store"`
    InvoiceNo  string        `json:"invoice_no"`
    Date       string        `json:"date"`
    Customer   *CustomerInfo `json:"customer"`
    OrderType  string        `json:"order_type"`
    Status     string        `json:"status"`
    PickupDate *string       `json:"pickup_date"`
    Items      []InvoiceItem `json:"items"`
    Subtotal   float64       `json:"subtotal"`
    Discount   float64       `json:"discount"`
    Total      float64       `json:"total"`
    TotalPaid  float64       `json:"total_paid"`
    Remaining  float64       `json:"remaining"`
    Payments   []PaymentInfo `json:"payments"`
}
```

### 9.8 Dashboard Module

#### Summary Endpoint

**Logic:**
```
1. Total penjualan hari ini: COUNT & SUM orders WHERE created_at = today AND status IN (PAID, READY, COMPLETED)
2. Laba Kotor hari ini: SUM(total_amount - total_hpp) WHERE same filter (HANYA untuk superadmin/owner)
3. PO aktif: COUNT orders WHERE order_type = PRE_ORDER AND status IN (DRAFT, DP_PAID, PAID, READY)
4. 5 transaksi terakhir: ORDER BY created_at DESC LIMIT 5
```

#### PO Reminders Endpoint

**Logic:**
- Ambil semua PRE_ORDER dengan `pickup_date` dalam 3 hari ke depan DAN status belum COMPLETED/CANCELLED.
- Sort by `pickup_date` ASC.

### 9.9 Audit Log Module

#### Cara Menulis Audit Log

Audit log ditulis dari **Service layer** menggunakan `AuditService`:

```go
type AuditService interface {
    Log(ctx context.Context, entry AuditEntry) error
}

type AuditEntry struct {
    UserID     *uint64
    Action     string
    EntityType string
    EntityID   *uint64
    OldValue   any
    NewValue   any
    IPAddress  string
}
```

**Kapan menulis audit log:**
- User login (berhasil)
- CRUD produk, paket, customer
- Create order, update status order
- Add payment
- User management (create, update, delete user)
- Store settings update

**Audit log TIDAK boleh menggagalkan operasi utama.** Jika insert audit log gagal, log error ke zerolog tapi jangan return error ke client. Gunakan goroutine atau channel jika perlu.

---

## 10. Middleware Stack

Middleware dieksekusi dalam urutan berikut (top to bottom):

```
1. Recovery        → Catch panic, return 500
2. Logger          → Log setiap request/response
3. CORS            → Handle preflight & allowed origins
4. RateLimiter     → Throttle requests
5. Auth (JWT)      → Validate token (skip untuk public routes)
6. RBAC            → Check role permission (per-route)
```

### 10.1 Recovery Middleware

- Catch semua panic.
- Log stack trace ke zerolog.
- Return standardized 500 response.
- **JANGAN** expose stack trace ke client di production.

### 10.2 Logger Middleware

Log setiap request dengan field:
```json
{
  "method": "POST",
  "path": "/api/v1/orders",
  "status": 201,
  "latency_ms": 45,
  "client_ip": "192.168.1.100",
  "user_id": 3,
  "request_id": "uuid-v4"
}
```

- Gunakan `X-Request-ID` header. Jika tidak ada, generate UUID v4 dan set di response header.
- **JANGAN** log request body (bisa mengandung password).
- Log response body HANYA jika status >= 400 (untuk debugging errors).

### 10.3 CORS Middleware

```go
cors.Config{
    AllowOrigins:     config.CORSAllowedOrigins,
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
    ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}
```

### 10.4 Rate Limiter Middleware

- Menggunakan `golang.org/x/time/rate` (token bucket algorithm).
- Config: `RATE_LIMIT_RPS` requests per second, `RATE_LIMIT_BURST` burst capacity.
- Rate limit **per IP address**.
- Jika limit exceeded → return `429 Too Many Requests`.
- Implementasi: `map[string]*rate.Limiter` dengan `sync.Mutex`. Cleanup stale entries secara periodik.

### 10.5 Auth Middleware

```
1. Baca header Authorization
2. Jika kosong → 401 Unauthorized
3. Parse "Bearer <token>"
4. Validate JWT (signature, expiry)
5. Jika invalid → 401 Unauthorized
6. Extract claims (user_id, role)
7. Set ke gin.Context:
   c.Set("user_id", claims.UserID)
   c.Set("user_role", claims.Role)
8. Next()
```

### 10.6 RBAC Middleware

```go
func RequireRoles(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        for _, role := range roles {
            if userRole == role {
                c.Next()
                return
            }
        }
        // Return 403 Forbidden
    }
}
```

---

## 11. Error Handling & Response Contract

### 11.1 Standard Success Response

```json
{
  "success": true,
  "message": "Product created successfully",
  "data": { ... }
}
```

**Untuk list/paginated:**
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_rows": 150,
    "total_pages": 8
  }
}
```

### 11.2 Standard Error Response

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "message": "email is required"
    },
    {
      "field": "password",
      "message": "password must be at least 8 characters"
    }
  ]
}
```

**Untuk error tanpa field-level detail:**
```json
{
  "success": false,
  "message": "Product not found"
}
```

### 11.3 HTTP Status Code Convention

| Code | Kapan Digunakan |
|---|---|
| 200 | GET success, UPDATE success |
| 201 | CREATE success |
| 204 | DELETE success (no content) |
| 400 | Bad Request — malformed JSON, invalid parameter |
| 401 | Unauthorized — token missing / invalid / expired |
| 403 | Forbidden — role tidak diizinkan |
| 404 | Not Found — resource tidak ditemukan |
| 409 | Conflict — duplikat data (email, phone, nama produk) |
| 422 | Unprocessable Entity — business rule violation (sell_price < hpp) |
| 429 | Too Many Requests — rate limited |
| 500 | Internal Server Error — unexpected error |

### 11.4 Error Handling Pattern

**Handler layer:**
```go
func (h *ProductHandler) Create(c *gin.Context) {
    var req dto.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, err)
        return
    }

    product, err := h.service.Create(c.Request.Context(), req)
    if err != nil {
        response.HandleServiceError(c, err)
        return
    }

    response.Created(c, "Product created successfully", product)
}
```

**Service layer — Custom error types:**
```go
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
```

**Response helper maps error type ke HTTP status:**
```go
func HandleServiceError(c *gin.Context, err error) {
    var serviceErr *ServiceError
    if errors.As(err, &serviceErr) {
        switch {
        case errors.Is(serviceErr.Type, ErrNotFound):
            Error(c, http.StatusNotFound, serviceErr.Message)
        case errors.Is(serviceErr.Type, ErrDuplicate):
            Error(c, http.StatusConflict, serviceErr.Message)
        case errors.Is(serviceErr.Type, ErrBusinessRule):
            Error(c, http.StatusUnprocessableEntity, serviceErr.Message)
        }
        return
    }
    Error(c, http.StatusInternalServerError, "Internal server error")
}
```

---

## 12. Security Requirements

### 12.1 Authentication & Authorization

| Requirement | Detail |
|---|---|
| Password hashing | Bcrypt dengan cost 12 (configurable via `BCRYPT_COST`) |
| JWT signing | HS256 dengan secret minimal 32 karakter |
| Token transport | `Authorization: Bearer <token>` header only |
| Token storage (client) | HttpOnly cookie atau secure local storage (frontend concern) |
| Password policy | Minimal 8 karakter (validasi di DTO) |

### 12.2 Input Security

| Requirement | Detail |
|---|---|
| Request validation | Semua input di-validate via struct tags sebelum diproses |
| SQL injection | GORM parameterized queries (tidak boleh raw string concatenation) |
| XSS | Response hanya JSON, Content-Type: application/json |
| CORS | Whitelist origins via config |
| Rate limiting | Per-IP token bucket |

### 12.3 Data Security

| Requirement | Detail |
|---|---|
| Sensitive data in logs | JANGAN log password, JWT token, atau data sensitif lainnya |
| Env vars | `.env` file HARUS masuk `.gitignore` |
| Error messages | Production TIDAK boleh expose stack trace atau internal error detail |
| HPP visibility | HPP & Laba HANYA visible untuk superadmin dan owner. Admin dan public TIDAK boleh melihat |

### 12.4 Security Headers

Set di middleware atau global:

```go
c.Header("X-Content-Type-Options", "nosniff")
c.Header("X-Frame-Options", "DENY")
c.Header("X-XSS-Protection", "1; mode=block")
c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
```

---

## 13. Observability & Logging

### 13.1 Logging Library: zerolog

**Setup:**
```go
// Production: JSON output
zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
log := zerolog.New(os.Stdout).With().Timestamp().Str("service", "tokomakanan").Logger()

// Development: Console output (pretty print)
log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
```

**Log levels:**
- `Debug` — detail untuk development (SQL queries di dev mode)
- `Info` — normal operations (request served, order created)
- `Warn` — potential issue (rate limit hit, deprecated endpoint called)
- `Error` — something failed (DB connection error, external service down)
- `Fatal` — unrecoverable (app cannot start, DB migration failed)

### 13.2 Apa yang Di-log

| Event | Level | Fields |
|---|---|---|
| Server start | Info | port, env, version |
| Request served | Info | method, path, status, latency, ip, user_id, request_id |
| Order created | Info | order_id, invoice_no, user_id, total |
| Payment added | Info | order_id, amount, method |
| Auth failed | Warn | email (masked), ip, reason |
| Validation error | Warn | path, fields, ip |
| Rate limit hit | Warn | ip, path |
| DB error | Error | query context (no raw SQL), error message |
| Panic recovery | Error | stack trace, path, ip |

### 13.3 Health Check Endpoint

```
GET /health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2026-09-17T10:30:00Z",
  "database": "connected",
  "version": "1.0.0"
}
```

**Logic:** Ping database. Jika gagal → return `503 Service Unavailable` dengan `"database": "disconnected"`.

---

## 14. Testing Strategy

### 14.1 Test Structure

```
tokomakanan/
├── internal/
│   ├── service/
│   │   ├── product_service.go
│   │   └── product_service_test.go
│   ├── handler/
│   │   ├── product_handler.go
│   │   └── product_handler_test.go
│   └── repository/
│       ├── product_repository.go
│       └── product_repository_test.go
```

### 14.2 Unit Tests (Service Layer)

- Mock repository interface menggunakan `testify/mock`.
- Test semua business rules:
  - sell_price < hpp → error
  - Status transition invalid → error
  - Overpayment → error
  - dsb.
- Naming convention: `Test<ServiceName>_<MethodName>_<Scenario>`

**Contoh:**
```go
func TestProductService_Create_ShouldReturnError_WhenSellPriceLessThanHPP(t *testing.T) { ... }
func TestOrderService_Create_ShouldSetStatusPaid_WhenFullPaymentProvided(t *testing.T) { ... }
func TestPaymentService_Create_ShouldPreventOverpayment(t *testing.T) { ... }
```

### 14.3 Integration Tests (Handler Layer)

- Gunakan `httptest` + `gin.CreateTestContext`.
- Test HTTP status codes, response format, header.
- Mock service layer.

### 14.4 Test Commands

```bash
# Run semua tests
go test ./...

# Run dengan coverage
go test ./... -cover

# Run tests untuk module tertentu
go test ./internal/service/...

# Run dengan verbose
go test -v ./internal/service/...
```

### 14.5 Minimum Coverage Target

| Layer | Target |
|---|---|
| Service | ≥ 80% |
| Handler | ≥ 70% |
| Repository | ≥ 60% (integration test) |

---

## 15. Docker & Development Setup

### 15.1 docker-compose.yml

```yaml
version: "3.8"

services:
  db:
    image: mysql:8.0
    container_name: tokomakanan-db
    restart: unless-stopped
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: tokomakanan
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: tokomakanan-api
    restart: unless-stopped
    ports:
      - "8080:8080"
    depends_on:
      db:
        condition: service_healthy
    env_file:
      - .env
    environment:
      DB_HOST: db

volumes:
  mysql_data:
```

### 15.2 Dockerfile (Multi-stage)

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /tokomakanan ./cmd/api

# Run stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /tokomakanan .
COPY migrations ./migrations
EXPOSE 8080
CMD ["./tokomakanan"]
```

### 15.3 Makefile

```makefile
.PHONY: run build test migrate-up migrate-down docker-up docker-down

run:
	go run ./cmd/api

build:
	go build -o bin/tokomakanan ./cmd/api

test:
	go test ./... -cover

migrate-up:
	migrate -path migrations -database "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)" up

migrate-down:
	migrate -path migrations -database "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)" down 1

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
```

---

## 16. Coding Conventions

### 16.1 Naming

| Element | Convention | Contoh |
|---|---|---|
| File | snake_case | `product_handler.go` |
| Package | lowercase, single word | `handler`, `service`, `repository` |
| Struct | PascalCase | `ProductService`, `CreateOrderRequest` |
| Interface | PascalCase, suffix deskriptif | `ProductRepository`, `AuditService` |
| Function | PascalCase (exported), camelCase (unexported) | `Create`, `validateStatusTransition` |
| Variable | camelCase | `totalAmount`, `orderItems` |
| Constant | PascalCase atau ALL_CAPS | `StatusPaid`, `MaxPageLimit` |
| JSON field | snake_case | `"sell_price"`, `"order_type"` |

### 16.2 Interface Placement

Interface didefinisikan **di package yang menggunakannya** (consumer), bukan di package yang mengimplementasikannya:

```
service/product_service.go  → defines ProductRepository interface
handler/product_handler.go  → defines ProductService interface
```

Ini mengikuti Go convention: "Accept interfaces, return structs."

### 16.3 Error Wrapping

Gunakan `fmt.Errorf` dengan `%w` untuk wrapping:

```go
if err != nil {
    return fmt.Errorf("failed to create product: %w", err)
}
```

### 16.4 Context Propagation

Selalu propagate `context.Context` dari handler ke service ke repository:

```go
// Handler
product, err := h.service.Create(c.Request.Context(), req)

// Service
func (s *productService) Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
    return s.repo.Create(ctx, product)
}

// Repository
func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
    return r.db.WithContext(ctx).Create(product).Error
}
```

### 16.5 DB Transaction Pattern

```go
func (s *orderService) Create(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
    var result *dto.OrderResponse

    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        order := &model.Order{...}
        if err := tx.Create(order).Error; err != nil {
            return err
        }

        for _, item := range items {
            if err := tx.Create(&item).Error; err != nil {
                return err
            }
        }

        result = mapToResponse(order)
        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }

    return result, nil
}
```

### 16.6 Project Init Commands

```bash
# Initialize Go module
go mod init github.com/<username>/tokomakanan

# Install dependencies
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
go get -u github.com/golang-jwt/jwt/v5
go get -u golang.org/x/crypto/bcrypt
go get -u github.com/rs/zerolog
go get -u github.com/joho/godotenv
go get -u github.com/google/uuid
go get -u golang.org/x/time/rate
go get -u github.com/gin-contrib/cors
go get -u github.com/stretchr/testify
go get -u github.com/golang-migrate/migrate/v4
```

---

## 17. Out-of-Scope (Phase 2)

Berikut fitur yang **TIDAK** diimplementasikan di Phase 1 dan TIDAK boleh dikerjakan:

- ❌ Manajemen stok bahan baku (Inventory/Warehouse Management)
- ❌ Recipe Bill of Materials (BOM)
- ❌ Modul akuntansi lengkap (Neraca, Arus Kas, Jurnal Umum)
- ❌ Integrasi Payment Gateway otomatis (QRIS Dynamic / VA callback)
- ❌ Notifikasi push / email
- ❌ Multi-tenant (multi-toko)
- ❌ File upload (logo toko cukup URL)
- ❌ Swagger / OpenAPI auto-generate
- ❌ WebSocket real-time update
- ❌ Caching layer (Redis)
- ❌ Background job / queue

---

> **Catatan untuk implementor:** Dokumen ini adalah sumber kebenaran (source of truth) untuk backend. Jika ada ambiguitas, tanyakan ke pemilik PRD sebelum membuat asumsi sendiri.
