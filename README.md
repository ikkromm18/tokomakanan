# Bakery POS & Management System — Backend API

Backend API produksi untuk sistem Point of Sale (POS) dan manajemen operasional toko roti/pastry berbasis Go 1.22+, Gin Gonic, GORM, dan MySQL 8.0 dengan Clean 3-Layer Architecture.

---

> ## 📖 Panduan Integrasi Front-End & Dokumentasi API
> Untuk rincian lengkap mengenai seluruh endpoint, autentikasi, expected request body, query parameter, contoh response payload JSON, dan kode status HTTP yang dihasilkan, silakan buka:
> 
> 👉 [**API_DOCUMENTATION.md**](API_DOCUMENTATION.md)
> 
> File tersebut dirancang khusus sebagai panduan acuan bagi tim Front-End (Web/Mobile/Desktop).

---

## 1. Fitur Utama

- **Authentication & RBAC**: Autentikasi berbasis JWT (HS256) dengan 3 tingkatan peran (`superadmin`, `owner`, `admin`/kasir).
- **POS & Direct Sales**: Transaksi penjualan kasir langsung dengan cetak nota dan pembayaran lunas di tempat.
- **Pre-Order (PO) Katering**: Pemesanan untuk tanggal mendatang dengan alur Down Payment (DP), status transisi ketat, dan pelunasan bertahap.
- **Perlindungan Margin & Kerahasiaan HPP**:
  - Validasi aturan bisnis: harga jual tidak boleh di bawah HPP (`sell_price >= hpp`).
  - Penyamaran data (masking): role `admin` (kasir) tidak dapat melihat HPP produk/paket maupun laba kotor toko.
- **Nota Digital Publik (Public Invoice)**:
  - Pelanggan dapat membuka nota digital melalui tautan token acak UUIDv4 (misalnya via link WhatsApp/QR).
  - Akses terbuka tanpa login, namun **terjamin 100% bebas dari informasi HPP atau margin keuntungan toko**.
- **Audit Logging**: Pencatatan riwayat setiap aksi mutasi data (CREATE, UPDATE, DELETE, perubahan status) secara terstruktur.
- **Dashboard & Reporting**:
  - Grafik penjualan harian & reminder pesanan PO yang harus diambil hari ini.
  - Laporan penjualan dan laporan laba kotor harian.
  - Export laporan ke format CSV ber-standard **UTF-8 Byte Order Mark (BOM)** agar langsung rapi saat dibuka di Microsoft Excel.

---

## 2. Arsitektur Kode (Clean 3-Layer Architecture)

Aplikasi dibangun menggunakan pola arsitektur bersih 3 lapis untuk menjamin *separation of concerns*, kemudahan pemeliharaan, serta skalabilitas pengujian.

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Request                         │
└───────────────────────────┬─────────────────────────────┘
                            │
               ┌────────────▼────────────┐
               │    Middleware Stack     │
               │  - Panic Recovery       │
               │  - Zerolog HTTP Logger  │
               │  - CORS                 │
               │  - Rate Limiter (Token) │
               │  - JWT Authentication   │
               │  - RBAC (Role-Based)    │
               └────────────┬────────────┘
                            │
               ┌────────────▼────────────┐
               │      Handler Layer      │  (internal/handler)
               │ - Parse HTTP Request    │
               │ - DTO Validation        │
               │ - Standard JSON Helpers │
               └────────────┬────────────┘
                            │ (Go Interface)
               ┌────────────▼────────────┐
               │      Service Layer      │  (internal/service)
               │ - Core Business Logic   │
               │ - Role-Based Masking    │
               │ - Audit Trail Trigger   │
               │ - Error Classification  │
               └────────────┬────────────┘
                            │ (Go Interface)
               ┌────────────▼────────────┐
               │    Repository Layer     │  (internal/repository)
               │ - GORM Query Execution  │
               │ - Database Transaction  │
               │ - Translate ORM Errors  │
               └────────────┬────────────┘
                            │
               ┌────────────▼────────────┐
               │      MySQL Database     │
               └─────────────────────────┘
```

### Prinsip-Prinsip Desain Kode:
1. **Komunikasi Berbasis Interface**: Handler berinteraksi dengan Service melalui interface; Service berinteraksi dengan Repository melalui interface. Hal ini memungkinkan setiap layer diuji dengan unit mock independen.
2. **Isolasi Service Layer**: Service layer **tidak boleh** mengimpor `gorm.io/gorm` atau package HTTP seperti `net/http` maupun `github.com/gin-gonic/gin`.
3. **Penerjemahan Error Konsisten**:
   - Repository menerjemahkan `gorm.ErrRecordNotFound` menjadi `nil, nil` agar tidak membocorkan detail ORM.
   - Service menggunakan `response.NewServiceError` untuk error berjenis domain (NotFound: 404, Duplicate: 409, Unauthorized: 401, Forbidden: 403, BusinessRule: 422).
   - Handler memanggil helper tunggal `response.HandleServiceError(c, err)`.
4. **Data Hashing & Security**: Password di-hash menggunakan `bcrypt` dengan configurable work factor (default: 12) via helper terisolasi.

---

## 3. Struktur Direktori Proyek

```text
tokomakanan/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point aplikasi & dependency injection
├── internal/
│   ├── config/
│   │   └── config.go               # Loader environment variable (.env)
│   ├── database/
│   │   ├── mysql.go                # Inisialisasi koneksi GORM & pool connection
│   │   └── migrator.go             # Runner migrasi skema database
│   ├── dto/                        # Data Transfer Object (Request & Response structs)
│   │   ├── auth_dto.go
│   │   ├── user_dto.go
│   │   ├── product_dto.go
│   │   ├── package_dto.go
│   │   ├── customer_dto.go
│   │   ├── order_dto.go
│   │   ├── payment_dto.go
│   │   ├── dashboard_dto.go
│   │   ├── report_dto.go
│   │   ├── public_invoice_dto.go
│   │   └── common_dto.go           # Standard Response & Pagination DTO
│   ├── handler/                    # Controller HTTP Gin
│   │   ├── helpers.go              # Helper parse ID, user_id, pagination
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── product_handler.go
│   │   ├── package_handler.go
│   │   ├── customer_handler.go
│   │   ├── order_handler.go
│   │   ├── payment_handler.go
│   │   ├── dashboard_handler.go
│   │   ├── report_handler.go
│   │   ├── store_setting_handler.go
│   │   ├── public_handler.go
│   │   └── audit_handler.go
│   ├── middleware/                 # Middleware HTTP
│   │   ├── auth.go                 # JWT bearer token verification
│   │   ├── rbac.go                 # Role-based access control guard
│   │   ├── cors.go                 # Cross-Origin Resource Sharing
│   │   ├── rate_limiter.go         # Token bucket IP rate limiter
│   │   ├── logger.go               # Zerolog structured access logger
│   │   └── recovery.go             # Safe panic recovery
│   ├── model/                      # Definisi GORM entities / schema model
│   │   ├── user.go
│   │   ├── customer.go
│   │   ├── product.go
│   │   ├── product_package.go
│   │   ├── package_item.go
│   │   ├── order.go
│   │   ├── order_item.go
│   │   ├── order_payment.go
│   │   ├── store_setting.go
│   │   └── audit_log.go
│   ├── pkg/                        # Utility mandiri dan reusable
│   │   ├── jwt/                    # Helper token JWT & hash bcrypt
│   │   ├── pagination/             # Format pagination & kalkulasi offset
│   │   ├── response/               # Standard JSON response helper & error mapping
│   │   └── invoice/                # Generator format nomor faktur INV/YYYYMMDD/XXXXX
│   ├── repository/                 # Data access layer (GORM)
│   │   ├── user_repository.go
│   │   ├── product_repository.go
│   │   ├── package_repository.go
│   │   ├── customer_repository.go
│   │   ├── order_repository.go
│   │   ├── payment_repository.go
│   │   ├── dashboard_repository.go
│   │   ├── report_repository.go
│   │   ├── store_setting_repository.go
│   │   └── audit_repository.go
│   ├── router/
│   │   └── router.go               # Registrasi seluruh rute & middleware grouping
│   └── service/                    # Business logic layer
│       ├── auth_service.go
│       ├── user_service.go
│       ├── product_service.go
│       ├── package_service.go
│       ├── customer_service.go
│       ├── order_service.go
│       ├── payment_service.go
│       ├── dashboard_service.go
│       ├── report_service.go
│       ├── store_setting_service.go
│       ├── public_service.go
│       └── audit_service.go
├── migrations/                     # SQL migration up & down berurutan
├── test/
│   └── e2e_test.go                 # End-to-End integration test
├── Dockerfile                      # Multi-stage Docker build
├── docker-compose.yml              # Service API & MySQL 8.0
├── Makefile                        # CLI shortcut task runner
├── CODING_STANDARD.md              # Standar baku penulisan kode
├── API_DOCUMENTATION.md            # Dokumentasi acuan Front-End
├── go.mod
└── go.sum
```

---

## 4. Tech Stack & Library

| Kategori | Komponen / Pustaka | Keterangan |
|---|---|---|
| **Bahasa** | Go 1.22+ | Backend language |
| **HTTP Framework** | `github.com/gin-gonic/gin` v1.9+ | Router & engine HTTP |
| **ORM** | `gorm.io/gorm` v1.25+ | Database ORM |
| **DB Driver** | `gorm.io/driver/mysql` v1.5+ | Driver MySQL 8.0 |
| **Database Migration** | `github.com/golang-migrate/migrate/v4` | Migrasi skema SQL terversi |
| **Autentikasi** | `github.com/golang-jwt/jwt/v5` | Token JWT HS256 |
| **Password Hashing** | `golang.org/x/crypto/bcrypt` | Bcrypt hashing |
| **Logging** | `github.com/rs/zerolog` | Structured JSON log |
| **Rate Limiter** | `golang.org/x/time/rate` | Token-bucket per IP |
| **Testing** | `github.com/stretchr/testify` | Assertions & test mocks |

---

## 5. Panduan Instalasi & Menjalankan Aplikasi

### 5.1 Prasyarat
- Go versi 1.22 atau lebih baru
- MySQL Server versi 8.0
- Docker & Docker Compose *(opsional jika ingin menjalankan via container)*

### 5.2 Menjalankan Secara Lokal (Local Machine)

1. **Clone repository dan masuk ke direktori**:
   ```bash
   git clone git@github.com:ikkromm18/tokomakanan.git
   cd tokomakanan
   ```

2. **Siapkan konfigurasi `.env`**:
   Salin dari template `.env.example`:
   ```bash
   cp .env.example .env
   ```
   Pastikan kredensial database di `.env` sudah sesuai dengan MySQL lokal Anda:
   ```env
   DB_HOST=localhost
   DB_PORT=3306
   DB_USER=root
   DB_PASSWORD=your_password
   DB_NAME=tokomakanan
   JWT_SECRET=super-secret-jwt-key-with-minimum-32-chars
   ```

3. **Buat Database MySQL**:
   ```sql
   CREATE DATABASE IF NOT EXISTS tokomakanan;
   ```

4. **Jalankan Database Migration**:
   Eksekusi migrasi skema tabel beserta seed akun superadmin:
   ```bash
   make migrate
   # atau langsung via go:
   go run cmd/api/main.go --migrate
   ```

5. **Jalankan Server API**:
   ```bash
   make run
   # atau:
   go run cmd/api/main.go
   ```
   Server akan berjalan di `http://localhost:8080`.

6. **Verifikasi Healthcheck**:
   ```bash
   curl http://localhost:8080/health
   ```
   Output:
   ```json
   {
     "success": true,
     "message": "Service is healthy",
     "data": {
       "app": "tokomakanan",
       "env": "development"
     }
   }
   ```

### 5.3 Menjalankan Menggunakan Docker Compose

Jika menggunakan Docker, seluruh dependency (MySQL 8.0 & Go API) akan berjalan otomatis:
```bash
make docker-up
# atau:
docker compose up -d
```
Untuk menghentikan:
```bash
make docker-down
```

---

## 6. Akun Default Bawaan (Seed Data)

Setelah migrasi awal dijalankan, sistem secara otomatis menyediakan akun awal:
- **Email**: `superadmin@tokomakanan.com`
- **Password**: `SuperAdmin123!`
- **Role**: `superadmin`

> ⚠️ **Penting**: Harap segera perbarui password akun superadmin ini pada lingkungan production melalui endpoint `PUT /api/v1/auth/change-password`!

---

## 7. Testing & Quality Assurance

Sistem ini dilengkapi cakupan unit test dan integration test yang ketat pada setiap layer:

```bash
# Menjalankan seluruh test suite
make test

# Menjalankan test dengan report coverage per file
make test-coverage
```

### Rekap Cakupan Pengujian (Test Coverage):
- **Service Layer**: **`80.7%`** (Memenuhi target standar `>= 80%`)
- **Handler Layer**: **`82.2%`** (Memenuhi target standar `>= 70%`)
- **Status Build**: **100% PASS** tanpa kegagalan.

---

## 8. Lisensi & Kontributor

- **Project**: Bakery POS & Management System (`tokomakanan`)
- **Developer**: Muhammad Ikrom (`ikkromm18`)
- **Repository**: [github.com/ikkromm18/tokomakanan](https://github.com/ikkromm18/tokomakanan)
