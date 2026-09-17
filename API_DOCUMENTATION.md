# Dokumentasi API — Bakery POS & Management System

Dokumentasi ini disusun sebagai acuan teknis integrasi bagi **Front-End Developer** dalam mengonsumsi Backend API Bakery POS & Management System (`tokomakanan`).

> 💡 **Postman Collection**: File koleksi Postman yang siap langsung di-import tersedia di file: [**`tokomakanan.postman_collection.json`**](tokomakanan.postman_collection.json). Ketika menjalankan request `Login`, token JWT akan otomatis tersimpan di variable `{{token}}`.

---

## 1. Konvensi Umum

### 1.1 Base URL
- **Local Development**: `http://localhost:8080`
- **Prefix API**: `/api/v1`

### 1.2 Headers
Untuk seluruh request berformat JSON:
```http
Content-Type: application/json
Accept: application/json
```

Untuk seluruh endpoint yang terproteksi autentikasi:
```http
Authorization: Bearer <token_jwt>
```

---

## 2. Format Response Standar

Semua endpoint menghasilkan struktur response konsisten menggunakan pembungkus standar.

### 2.1 Single Resource / Action Response (Success)
```json
{
  "success": true,
  "message": "Pesan deskriptif keberhasilan",
  "data": { ... }
}
```

### 2.2 Paginated List Response (Success)
```json
{
  "success": true,
  "message": "Pesan deskriptif keberhasilan",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_rows": 45,
    "total_pages": 3
  }
}
```

### 2.3 Error Response
```json
{
  "success": false,
  "message": "Pesan ringkas error",
  "errors": [
    {
      "field": "email",
      "message": "required validation failed"
    }
  ]
}
```

### 2.4 Daftar HTTP Status Code
| Kode | Makna | Helper Backend | Keterangan |
|---|---|---|---|
| `200` | OK | `response.Success` / `response.Paginated` | Permintaan berhasil dibaca/diperbarui. |
| `201` | Created | `response.Created` | Data baru berhasil dibuat. |
| `400` | Bad Request | `response.ValidationError` / `response.Error` | Format payload atau parameter query tidak valid. |
| `401` | Unauthorized | `response.HandleServiceError` | Token JWT hilang, tidak valid, atau kedaluwarsa. |
| `403` | Forbidden | `response.HandleServiceError` | Role user tidak memiliki izin mengakses endpoint. |
| `404` | Not Found | `response.HandleServiceError` | Resource tidak ditemukan. |
| `409` | Conflict | `response.HandleServiceError` | Duplikasi data unik (email, no telepon, nama produk dsb). |
| `422` | Unprocessable Entity | `response.HandleServiceError` | Pelanggaran aturan bisnis (cth: `sell_price < hpp`, overpayment). |
| `500` | Internal Server Error | `response.Error` | Terjadi kesalahan sistem atau database internal. |

---

## 3. Matriks Role & Hak Akses (RBAC)

Sistem membedakan tiga role user:
1. **`superadmin`**: Hak akses tertinggi (User Management, Audit Logs, Settings, Master Data, Transaksi, Laporan Keuangan).
2. **`owner`**: Hak analitik & manajemen (Settings, Master Data, Transaksi, Laporan Keuangan & Laba Kotor).
3. **`admin`**: Operasional kasir (POS, PO, Transaksi, Pembayaran, Pelanggan). **DILARANG** melihat HPP (`unit_hpp`, `total_hpp`) dan laba kotor toko.

| Modul | Endpoint | Superadmin | Owner | Admin |
|---|---|:---:|:---:|:---:|
| Auth | `/api/v1/auth/login` | Public | Public | Public |
| Auth | `/api/v1/auth/me`, `/change-password` | ✅ | ✅ | ✅ |
| Users | `/api/v1/users/**` | ✅ | ❌ | ❌ |
| Store Settings | `GET /api/v1/store-settings` | ✅ | ✅ | ✅ |
| Store Settings | `PUT /api/v1/store-settings` | ✅ | ✅ | ❌ |
| Audit Logs | `GET /api/v1/audit-logs` | ✅ | ❌ | ❌ |
| Products | `GET /api/v1/products/**` | ✅ | ✅ | ✅ (HPP masked) |
| Products | `POST/PUT/DELETE /api/v1/products/**` | ✅ | ✅ | ❌ |
| Packages | `GET /api/v1/packages/**` | ✅ | ✅ | ✅ (HPP masked) |
| Packages | `POST/PUT/DELETE /api/v1/packages/**` | ✅ | ✅ | ❌ |
| Customers | `/api/v1/customers/**` | ✅ | ✅ | ✅ |
| Orders | `/api/v1/orders/**` | ✅ | ✅ | ✅ (HPP masked) |
| Payments | `/api/v1/orders/:id/payments` | ✅ | ✅ | ✅ |
| Public Invoice | `/api/v1/public/invoice/:token` | Public | Public | Public |
| Dashboard | `GET /api/v1/dashboard/summary` | ✅ | ✅ | ✅ (Gross Profit masked) |
| Dashboard | `GET /api/v1/dashboard/chart`, `/po-reminders` | ✅ | ✅ | ✅ |
| Reports | `/api/v1/reports/**` | ✅ | ✅ | ❌ |

---

## 4. Rincian Endpoint API

### 4.1 Modul Authentication

#### 1. Login
- **Method / URL**: `POST /api/v1/auth/login`
- **Auth**: Public
- **Request Body**:
```json
{
  "email": "kasir1@tokomakanan.com",
  "password": "Password123!"
}
```
- **Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 2,
      "name": "Sarah Kasir",
      "email": "kasir1@tokomakanan.com",
      "role": "admin"
    }
  }
}
```

#### 2. Get Current Profile
- **Method / URL**: `GET /api/v1/auth/me`
- **Auth**: Bearer Token (Semua Role)
- **Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "User profile retrieved successfully",
  "data": {
    "id": 2,
    "name": "Sarah Kasir",
    "email": "kasir1@tokomakanan.com",
    "role": "admin"
  }
}
```

#### 3. Ganti Password Sendiri
- **Method / URL**: `PUT /api/v1/auth/change-password`
- **Auth**: Bearer Token (Semua Role)
- **Request Body**:
```json
{
  "old_password": "Password123!",
  "new_password": "NewSecretPassword456!",
  "confirm_password": "NewSecretPassword456!"
}
```
- **Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "Password changed successfully",
  "data": null
}
```

---

### 4.2 Modul User Management (Superadmin Only)

#### 1. List Users
- **Method / URL**: `GET /api/v1/users`
- **Auth**: Superadmin
- **Query Params**: `page` (default: 1), `limit` (default: 20), `role` (`superadmin|owner|admin`), `search` (nama/email)
- **Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Super Admin",
      "email": "superadmin@tokomakanan.com",
      "role": "superadmin",
      "is_active": true,
      "created_at": "2026-09-17T09:00:00Z",
      "updated_at": "2026-09-17T09:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_rows": 1,
    "total_pages": 1
  }
}
```

#### 2. Create User
- **Method / URL**: `POST /api/v1/users`
- **Auth**: Superadmin
- **Request Body**:
```json
{
  "name": "Budi Santoso",
  "email": "budi.owner@tokomakanan.com",
  "password": "InitialPassword123!",
  "role": "owner"
}
```
- **Expected Response (201 Created)**:
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": 3,
    "name": "Budi Santoso",
    "email": "budi.owner@tokomakanan.com",
    "role": "owner",
    "is_active": true,
    "created_at": "2026-09-17T10:00:00Z",
    "updated_at": "2026-09-17T10:00:00Z"
  }
}
```

#### 3. Update User
- **Method / URL**: `PUT /api/v1/users/:id`
- **Auth**: Superadmin
- **Request Body**:
```json
{
  "name": "Budi Santoso S.E.",
  "email": "budi.owner@tokomakanan.com",
  "role": "owner",
  "is_active": true,
  "password": "" 
}
```
*(Catatan: field `password` bersifat opsional saat update; jika kosong, password lama dipertahankan)*

#### 4. Delete User (Soft Delete)
- **Method / URL**: `DELETE /api/v1/users/:id`
- **Auth**: Superadmin
- **Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "User deleted successfully",
  "data": null
}
```

---

### 4.3 Modul Store Settings

#### 1. Get Store Settings
- **Method / URL**: `GET /api/v1/store-settings`
- **Auth**: Bearer Token (Semua Role)
- **Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "Store settings retrieved successfully",
  "data": {
    "id": 1,
    "name": "Bakery Delights",
    "address": "Jl. Merdeka No. 18, Bandung",
    "phone": "081234567890",
    "logo_url": "https://storage.googleapis.com/tokomakanan/logo.png",
    "receipt_footer": "Terima kasih telah berbelanja di Bakery Delights!",
    "updated_at": "2026-09-17T09:00:00Z"
  }
}
```

#### 2. Update Store Settings
- **Method / URL**: `PUT /api/v1/store-settings`
- **Auth**: Superadmin, Owner
- **Request Body**:
```json
{
  "name": "Bakery Delights Bandung",
  "address": "Jl. Merdeka No. 18, Bandung",
  "phone": "081234567890",
  "logo_url": "https://storage.googleapis.com/tokomakanan/logo.png",
  "receipt_footer": "Follow Instagram kami @bakerydelights.bdg"
}
```

---

### 4.4 Modul Products (Master Data Produk)

#### 1. List Products
- **Method / URL**: `GET /api/v1/products`
- **Auth**: Bearer Token (Semua Role)
- **Query Params**: `page`, `limit`, `category`, `is_active` (`true|false`), `search`
- **Expected Response (200 OK - Superadmin/Owner)**:
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Roti Tawar Gandum",
      "category": "Bakery",
      "hpp": 12000,
      "sell_price": 20000,
      "is_active": true,
      "created_at": "2026-09-17T09:00:00Z",
      "updated_at": "2026-09-17T09:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_rows": 1,
    "total_pages": 1
  }
}
```
*(Catatan Keamanan: Jika diakses role `admin`, field `hpp` tidak dikembalikan / `null`)*

#### 2. Create Product
- **Method / URL**: `POST /api/v1/products`
- **Auth**: Superadmin, Owner
- **Aturan Bisnis**: `sell_price >= hpp`
- **Request Body**:
```json
{
  "name": "Croissant Coklat",
  "category": "Pastry",
  "hpp": 8000,
  "sell_price": 16000
}
```

#### 3. Update Product
- **Method / URL**: `PUT /api/v1/products/:id`
- **Auth**: Superadmin, Owner
- **Request Body**:
```json
{
  "name": "Croissant Coklat Almond",
  "category": "Pastry",
  "hpp": 9000,
  "sell_price": 18000,
  "is_active": true
}
```

#### 4. Delete Product
- **Method / URL**: `DELETE /api/v1/products/:id`
- **Auth**: Superadmin, Owner
- **Aturan Bisnis**: Gagal (422) jika produk masih terdaftar dalam paket aktif.

---

### 4.5 Modul Packages (Bundling)

#### 1. List Packages
- **Method / URL**: `GET /api/v1/packages`
- **Auth**: Bearer Token (Semua Role)
- **Query Params**: `page`, `limit`, `is_active`, `search`
- **Expected Response (200 OK - Superadmin/Owner)**:
```json
{
  "success": true,
  "message": "Packages retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Paket Sarapan Pagi",
      "total_hpp": 20000,
      "sell_price": 28000,
      "is_active": true,
      "items": [
        {
          "id": 10,
          "product_id": 1,
          "product_name": "Roti Tawar Gandum",
          "quantity": 1,
          "unit_hpp": 12000,
          "unit_sell": 20000
        },
        {
          "id": 11,
          "product_id": 2,
          "product_name": "Croissant Coklat",
          "quantity": 1,
          "unit_hpp": 8000,
          "unit_sell": 16000
        }
      ],
      "created_at": "2026-09-17T09:00:00Z",
      "updated_at": "2026-09-17T09:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total_rows": 1, "total_pages": 1 }
}
```
*(Catatan: Untuk role `admin`, `total_hpp` dan `unit_hpp` di-mask / `null`)*

#### 2. Create Package
- **Method / URL**: `POST /api/v1/packages`
- **Auth**: Superadmin, Owner
- **Aturan Bisnis**: 
  - `total_hpp` dihitung otomatis dari database: `SUM(product.hpp * quantity)`.
  - `sell_price` harus `>= total_hpp`.
- **Request Body**:
```json
{
  "name": "Paket Sarapan Pagi",
  "sell_price": 28000,
  "items": [
    { "product_id": 1, "quantity": 1 },
    { "product_id": 2, "quantity": 1 }
  ]
}
```

---

### 4.6 Modul Customers

#### 1. List Customers
- **Method / URL**: `GET /api/v1/customers`
- **Auth**: Bearer Token (Semua Role)
- **Query Params**: `page`, `limit`, `search` (mencari nama atau no telepon)

#### 2. Create Customer
- **Method / URL**: `POST /api/v1/customers`
- **Auth**: Bearer Token (Semua Role)
- **Request Body**:
```json
{
  "name": "Ibu Kartika",
  "phone": "081987654321",
  "address": "Komplek Dago Asri No. 12"
}
```

---

### 4.7 Modul Orders & Kasir (Direct Sale & Pre-Order)

#### 1. Create Order
- **Method / URL**: `POST /api/v1/orders`
- **Auth**: Bearer Token (Semua Role)
- **Penting**:
  - `order_type`: `DIRECT_SALE` atau `PRE_ORDER`
  - Jika `DIRECT_SALE`: `initial_paid` harus lunas penuh (`initial_paid == total_amount`).
  - Jika `PRE_ORDER`: `pickup_date` wajib diisi (`YYYY-MM-DD` dan tidak boleh tanggal lampau). `initial_paid` dapat berupa 0 (status `DRAFT`), sebagian DP (status `DP_PAID`), atau lunas penuh (status `PAID`).
- **Request Body (Direct Sale)**:
```json
{
  "customer_id": null,
  "order_type": "DIRECT_SALE",
  "pickup_date": null,
  "notes": "Pelanggan minta bon dipisah",
  "discount_amount": 0,
  "payment_method": "CASH",
  "initial_paid": 36000,
  "items": [
    { "item_type": "PRODUCT", "item_id": 1, "quantity": 1 },
    { "item_type": "PRODUCT", "item_id": 2, "quantity": 1 }
  ]
}
```
- **Expected Response (201 Created)**:
```json
{
  "success": true,
  "message": "Order created successfully",
  "data": {
    "id": 15,
    "invoice_no": "INV/20260917/00001",
    "invoice_token": "a87998b6-90fe-42e5-a337-14fc75ba4a41",
    "customer_id": null,
    "user_id": 2,
    "user_name": "Sarah Kasir",
    "order_type": "DIRECT_SALE",
    "status": "PAID",
    "subtotal": 36000,
    "discount_amount": 0,
    "total_amount": 36000,
    "total_paid": 36000,
    "remaining_paid": 0,
    "payment_method": "CASH",
    "items": [
      {
        "id": 28,
        "item_type": "PRODUCT",
        "item_id": 1,
        "item_name": "Roti Tawar Gandum",
        "quantity": 1,
        "unit_price": 20000,
        "subtotal": 20000
      }
    ],
    "payments": [
      {
        "id": 12,
        "amount": 36000,
        "payment_method": "CASH",
        "paid_at": "2026-09-17T10:15:00Z"
      }
    ],
    "created_at": "2026-09-17T10:15:00Z"
  }
}
```

#### 2. Update Order Status
- **Method / URL**: `PATCH /api/v1/orders/:id/status`
- **Auth**: Bearer Token (Semua Role)
- **State Machine Transisi**:
  - `DRAFT` ➔ `DP_PAID`, `PAID`, `CANCELLED`
  - `DP_PAID` ➔ `PAID`, `CANCELLED`
  - `PAID` ➔ `READY`, `COMPLETED`, `CANCELLED`
  - `READY` ➔ `COMPLETED`, `CANCELLED`
  - `COMPLETED` & `CANCELLED` = Terminal (tidak dapat diubah lagi)
- **Request Body**:
```json
{
  "status": "READY"
}
```

---

### 4.8 Modul Payments (Pelunasan & Cicilan PO)

#### 1. Tambah Pembayaran Baru
- **Method / URL**: `POST /api/v1/orders/:id/payments`
- **Auth**: Bearer Token (Semua Role)
- **Aturan Bisnis**:
  - Overpayment dicegah: `amount <= remaining_bill`.
  - Otomatis mengubah status order ke `PAID` jika sisa tagihan telah lunas.
- **Request Body**:
```json
{
  "amount": 250000,
  "payment_method": "TRANSFER",
  "notes": "Pelunasan PO via transfer BCA"
}
```

#### 2. Riwayat Pembayaran Order
- **Method / URL**: `GET /api/v1/orders/:id/payments`
- **Auth**: Bearer Token (Semua Role)

---

### 4.9 Modul Public Digital Invoice (Tanpa Autentikasi)

Endpoint ini digunakan untuk membuka nota digital bagi pelanggan (misalnya melalui link tautan yang dikirim via WhatsApp atau QR Code).

- **Method / URL**: `GET /api/v1/public/invoice/:token`
- **Auth**: **PUBLIC (Tidak butuh token JWT)**
- **Jaminan Keamanan**: Field HPP dan Laba Kotor toko **ditiadakan secara permanen** dari response ini.
- **Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "Invoice retrieved successfully",
  "data": {
    "store": {
      "name": "Bakery Delights",
      "address": "Jl. Merdeka No. 18, Bandung",
      "phone": "081234567890",
      "logo_url": "https://storage.googleapis.com/tokomakanan/logo.png",
      "receipt_footer": "Terima kasih atas kunjungan Anda!"
    },
    "invoice_no": "INV/20260917/00001",
    "customer_name": "Ibu Kartika",
    "cashier_name": "Sarah Kasir",
    "order_type": "PRE_ORDER",
    "status": "PAID",
    "pickup_date": "2026-09-20",
    "notes": "Kotak kue warna gold",
    "subtotal": 500000,
    "discount_amount": 20000,
    "total_amount": 480000,
    "total_paid": 480000,
    "remaining_paid": 0,
    "items": [
      {
        "item_name": "Paket Sarapan Pagi",
        "quantity": 10,
        "unit_price": 48000,
        "subtotal": 480000
      }
    ],
    "payments": [
      {
        "amount": 200000,
        "payment_method": "TRANSFER",
        "paid_at": "2026-09-17T09:00:00Z"
      },
      {
        "amount": 280000,
        "payment_method": "TRANSFER",
        "paid_at": "2026-09-17T11:00:00Z"
      }
    ],
    "created_at": "2026-09-17T09:00:00Z"
  }
}
```

---

### 4.10 Modul Dashboard Analitik

#### 1. Ringkasan Harian (Summary)
- **Method / URL**: `GET /api/v1/dashboard/summary`
- **Auth**: Bearer Token (Semua Role)
- **Expected Response (200 OK - Superadmin/Owner)**:
```json
{
  "success": true,
  "message": "Dashboard summary retrieved successfully",
  "data": {
    "today_revenue": 3450000,
    "today_transactions": 28,
    "active_pre_orders": 7,
    "today_gross_profit": 1250000
  }
}
```
*(Catatan: Untuk role `admin`, field `today_gross_profit` bernilai `null`)*

#### 2. Grafik Penjualan (Chart)
- **Method / URL**: `GET /api/v1/dashboard/chart`
- **Auth**: Bearer Token (Semua Role)
- **Query Params**: `start_date`, `end_date` (Format: `YYYY-MM-DD`)

#### 3. Reminder Pre-Order Hari Ini
- **Method / URL**: `GET /api/v1/dashboard/po-reminders`
- **Auth**: Bearer Token (Semua Role)
- Mengembalikan daftar pesanan Pre-Order yang memiliki `pickup_date` jatuh pada hari ini yang belum `COMPLETED` atau `CANCELLED`.

---

### 4.11 Modul Laporan & Export CSV (Superadmin & Owner)

#### 1. Laporan Penjualan (Sales Report)
- **Method / URL**: `GET /api/v1/reports/sales`
- **Auth**: Superadmin, Owner
- **Query Params**: `start_date` (wajib), `end_date` (wajib), `order_type` (`DIRECT_SALE|PRE_ORDER`)

#### 2. Laporan Laba Kotor (Profit Report)
- **Method / URL**: `GET /api/v1/reports/profit`
- **Auth**: Superadmin, Owner
- **Query Params**: `start_date` (wajib), `end_date` (wajib)
- **Expected Response (200 OK)**:
```json
{
  "success": true,
  "message": "Profit report retrieved successfully",
  "data": {
    "total_revenue": 15000000,
    "total_hpp": 9000000,
    "total_gross_profit": 6000000,
    "average_margin_pct": 40.0,
    "daily_breakdown": [
      {
        "date": "2026-09-17",
        "revenue": 3450000,
        "total_hpp": 2070000,
        "gross_profit": 1380000,
        "margin_pct": 40.0
      }
    ]
  }
}
```

#### 3. Export CSV Laporan
- **Method / URL**: `GET /api/v1/reports/export`
- **Auth**: Superadmin, Owner
- **Query Params**:
  - `start_date` (wajib, YYYY-MM-DD)
  - `end_date` (wajib, YYYY-MM-DD)
  - `type`: `sales` (default) atau `profit`
- **Response**: File streaming `text/csv; charset=utf-8` dengan header `Content-Disposition: attachment; filename=...`. File dilengkapi **UTF-8 Byte Order Mark (BOM)** sehingga langsung rapi saat dibuka di Microsoft Excel.
