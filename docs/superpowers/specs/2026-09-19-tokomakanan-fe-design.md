# Design Spec: Front-End Bakery POS & Management System (`tokomakanan-fe`)

> **Tanggal:** 19 September 2026  
> **Status:** Approved by User  
> **Backend Source of Truth:** `tokomakanan/docs/prd.md`, `tokomakanan/API_DOCUMENTATION.md`, `tokomakanan/CODING_STANDARD.md`  
> **Stack:** React 18+, TypeScript, Vite, Tailwind CSS, TanStack Query v5, Zustand, Axios, React Hook Form, Zod, Lucide React  

---

## 1. Executive Summary & Goals

### 1.1 Objective
Membangun aplikasi web front-end produksi berstandar industri dengan nama **`tokomakanan-fe`** yang beroperasi secara mulus dengan Backend API Go Gin Gonic (`tokomakanan`). Front-end ini melayani tiga kelompok pengguna utama:
1. **Kasir (`admin`)**: Pengalaman POS layar sentuh/desktop berkecepatan tinggi untuk *Direct Sale* dan *Pre-Order (PO)*, kalkulasi kembalian otomatis, cetak struk thermal (80mm & 58mm), dan pembuatan tautan nota WhatsApp.
2. **Manajemen (`owner` & `superadmin`)**: Analitik dashboard, manajemen master data (produk, paket bundling, pelanggan, akun user), pengaturan toko, serta pelaporan keuangan (HPP, laba kotor, margin, dan ekspor CSV).
3. **Pelanggan Umum**: Halaman nota digital publik (`/invoice/:token`) tanpa autentikasi yang aman dan responsif di smartphone.

### 1.2 Non-Functional Requirements
- **Scalability**: Arsitektur modular *Feature-Driven* memisahkan domain secara tegas. Penambahan modul baru tidak mengganggu modul yang sudah ada.
- **Maintainability**: Pengetikan ketat TypeScript (zero `any`), skema validasi Zod kembar dengan struct validation backend, pemisahan jelas antara Server State (TanStack Query) dan Client State (Zustand).
- **Observability**: Client-side structured logger (level-based), propagasi header korelasi `X-Request-ID` ke backend Zerolog, dan React Error Boundary untuk mencegah crash layar putih (*white screen of death*).
- **Security**: Sanitasi input bebas menggunakan *DOMPurify*, penyimpanan token aman dengan *auto-purge* saat status 401, *Role-Based Access Control* (RBAC) pada router dan komponen, serta masking data finansial rahasia (HPP dan laba kotor) dari role kasir.

---

## 2. Directory Structure & Architecture

Proyek ditempatkan di root workspace sejajar dengan backend: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe`.

```
tokomakanan-fe/
├── public/
├── src/
│   ├── assets/                 # SVGs, logo, icons
│   ├── components/             # Reusable UI Shared across features
│   │   ├── ui/                 # Headless Design system primitives (Button, Input, Modal, Badge, Card, Table, Select)
│   │   ├── feedback/           # LoadingSpinner, ErrorState, EmptyState, SkeletonLoader
│   │   ├── layout/             # AppLayout, Sidebar, Navbar, PageHeader
│   │   └── guard/              # ProtectedRoute, RoleGuard
│   │
│   ├── core/                   # Pondasi infrastruktur aplikasi
│   │   ├── api/                # Axios instance, request/response interceptors, error mapper
│   │   ├── config/             # Environment variables & constants
│   │   ├── logger/             # Structured logger with correlation ID
│   │   ├── security/           # Token storage abstraction, DOMPurify sanitizer
│   │   └── types/              # Base ApiResponse<T>, PaginatedResponse<T>, ApiError
│   │
│   ├── features/               # Domain-driven feature modules
│   │   ├── auth/               # Login, Me, Change Password
│   │   ├── dashboard/          # Summary metrics, sales chart, PO reminders
│   │   ├── pos/                # POS Kasir, reactive cart, payment modal, thermal receipt
│   │   ├── orders/             # Order list, filters, detail, status update
│   │   ├── payments/           # Payment history & add payment modal
│   │   ├── products/           # Product CRUD & category filters
│   │   ├── packages/           # Package bundling CRUD & HPP calculation
│   │   ├── customers/          # Customer CRUD & quick selector
│   │   ├── reports/            # Sales & Profit reports, CSV export
│   │   ├── users/              # User management (Superadmin only)
│   │   ├── settings/           # Store settings & receipt footer
│   │   └── public-invoice/     # Public customer digital receipt
│   │
│   ├── hooks/                  # Global hooks (useThermalPrint, useDebounce, useCurrency)
│   ├── routes/                 # React Router configuration
│   ├── stores/                 # Zustand stores (authStore, posCartStore, uiStore)
│   ├── styles/                 # Tailwind CSS & print media queries
│   ├── utils/                  # Formatters (formatRupiah, formatDate)
│   ├── App.tsx                 # Root providers
│   └── main.tsx                # Entry point
│
├── CODING_STANDARD.md          # Frontend Coding Standards
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

---

## 3. Data Flow & Integration Contracts

### 3.1 HTTP Client & Interceptors
- **Base URL**: `http://localhost:8080/api/v1` (configurable via `VITE_API_BASE_URL`).
- **Outgoing Headers**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <token>` (otomatis disuntik jika user login)
  - `X-Request-ID: <uuid-v4>` (otomatis digenerate per request untuk *distributed tracing*)

### 3.2 Error Mapping Table
| HTTP Code | Backend Trigger | Front-End Handling |
|---|---|---|
| `200 / 201` | Sukses | Unpack payload `data` dan return ke caller |
| `400` | Format JSON salah / validasi gagal | Parse `errors: [{ field, message }]` dan attach ke form fields |
| `401` | Token expired / invalid | Clear session, notifikasi toast, redirect ke `/login` |
| `403` | Role tidak diizinkan | Toast "Akses ditolak untuk role Anda", redirect ke `/pos` |
| `404` | Data tidak ditemukan | Tampilkan `EmptyState` / pesan not found |
| `409` | Duplikasi data unik | Tampilkan pesan duplikasi (email/no HP/nama produk) |
| `422` | Business rule violation | Alert pesan bisnis (misal `sell_price < hpp` atau overpayment) |
| `500` | Internal server error | Pesan error umum + tampilkan `requestId` untuk pelaporan |

---

## 4. Security & Observability

### 4.1 Keamanan (Security)
1. **XSS Mitigation**:
   - Teks catatan transaksi dan footer struk disaring dengan `DOMPurify.sanitize()` sebelum dirender.
2. **Token Lifecycle**:
   - `tokenStorage` mengisolasi token JWT di browser storage dengan namespace aplikasi.
   - Tidak pernah mencatat token atau password ke console log.
3. **Data Masking (HPP & Laba)**:
   - Role `admin` tidak dapat mengakses kolom HPP dan Laba Kotor. Komponen UI dibungkus `<RoleGuard allowedRoles={['superadmin', 'owner']}>`.

### 4.2 Observabilitas (Observability)
1. **Structured Logging**:
   - Format: `{ timestamp, level, component, requestId, message, details }`.
   - Hanya error dan warning yang aktif di mode production.
2. **Crash Resilience (Error Boundary)**:
   - Global Error Boundary membungkus rute utama dengan tombol pemulihan agar kasir tidak terhenti saat terjadi error runtime JavaScript tak terduga.

---

## 5. UI/UX & Fitur Kasir (POS)

### 5.1 Alur Kasir (POS)
1. **Katalog & Pencarian**:
   - Grid produk dengan thumbnail/icon kategori.
   - Search bar auto-focus untuk pencarian nama produk atau scan barcode.
2. **Keranjang Kasir Reaktif (`usePosCartStore`)**:
   - Tambah/kurang item instan tanpa lag.
   - Switch tipe order: `DIRECT_SALE` vs `PRE_ORDER`.
   - Pada `PRE_ORDER`, field pelanggan dan `pickup_date` (dengan validasi min hari ini) wajib diisi.
3. **Modal Pembayaran**:
   - Tombol nominal cepat (*Uang Pas*, *Rp 20.000*, *Rp 50.000*, *Rp 100.000*).
   - Input nominal bayar dengan kalkulasi kembalian otomatis.
   - Validasi overpayment pencegahan kesalahan input kasir.
4. **Struk Thermal & Nota WhatsApp**:
   - Cetak struk otomatis dengan layout `@media print` untuk lebar 80mm dan 58mm.
   - Tombol satu klik untuk generate tautan WhatsApp berisi link nota publik: `https://.../invoice/:token`.

---

## 6. Phased Implementation Plan

Pengerjaan proyek akan dibagi menjadi **5 Phase Terstruktur**:

### Phase 1: Project Initialization, Foundation, Core Infrastructure & CODING_STANDARD
- Inisialisasi proyek Vite + React TS (`tokomakanan-fe`).
- Konfigurasi Tailwind CSS, Lucide Icons, TypeScript paths (`@/*`).
- Implementasi `core/api` (Axios client, `X-Request-ID`, error parser, standard response types).
- Implementasi `core/logger` & `core/security` (`tokenStorage`, `DOMPurify`).
- Implementasi shared UI primitives (`Button`, `Input`, `Card`, `Badge`, `Modal/Dialog`, `Table`, `Select`, `Toast`).
- Implementasi React Router, `ProtectedRoute`, `RoleGuard`, dan `ErrorBoundary`.
- Penyusunan dokumen `tokomakanan-fe/CODING_STANDARD.md`.

### Phase 2: Auth Module, App Shell Layout, Store Settings & User Management
- Modul Auth: Login, Profil User (`/auth/me`), Ganti Password, Logout, auto-clear session saat 401.
- Layout Aplikasi: Sidebar navigasi responsif, Header dengan profil user & badge role, breadcrumb.
- Modul Pengaturan Toko (`store-settings`): Tampilan dan edit nama toko, alamat, kontak, dan footer nota.
- Modul Manajemen User (`users`, Superadmin only): List user paginated, form tambah/edit akun (`owner`/`admin`), toggle status aktif, soft delete.

### Phase 3: Master Data Management (Products, Packages & Customers)
- Modul Produk (`products`): List dengan pagination, pencarian nama & filter kategori, form modal create & edit (validasi `sell_price >= hpp`), soft delete. Guard HPP untuk role kasir.
- Modul Paket Bundling (`packages`): List paket, modal create/edit paket dengan pemilihan produk multi-item dinamis dan kalkulasi otomatis `total_hpp`.
- Modul Pelanggan (`customers`): List pelanggan dengan pencarian nama/telepon, form tambah dan update data pelanggan.

### Phase 4: High-Speed POS Kasir, Alur Transaksi, Struk Thermal & Manajemen Pesanan
- Layar POS Kasir: Split-screen katalog produk/paket + Keranjang Belanja reaktif (`usePosCartStore`).
- Modal Pembayaran POS: Dukungan Direct Sale & Pre-Order, tombol nominal cepat, kalkulasi kembalian real-time.
- Thermal Print Engine: Komponen struk belanja siap cetak ukuran 80mm & 58mm.
- Modul Pesanan (`orders`): Daftar riwayat pesanan (filter tanggal, tipe pesanan, status), halaman detail pesanan, update status pesanan (`DRAFT` ➔ `DP_PAID` ➔ `PAID` ➔ `READY` ➔ `COMPLETED`).
- Modul Pembayaran Tambahan (`payments`): Modal cicilan/pelunasan PO dengan proteksi overpayment.

### Phase 5: Dashboard Analitik, Laporan Keuangan & Nota Digital Publik
- Dashboard: Widget ringkasan harian (omzet, transaksi, PO aktif, laba kotor), grafik tren penjualan, daftar pengingat PO hari ini.
- Laporan Keuangan (`reports`, Superadmin/Owner): Laporan penjualan periodik, laporan laba kotor & margin, ekspor file CSV ber-BOM untuk Excel.
- Halaman Nota Publik (`/invoice/:token`): Tampilan nota digital bersih tanpa autentikasi, status pesanan real-time, tombol cetak PDF & share WhatsApp.
- Verifikasi komprehensif, build check TypeScript (`tsc`), dan pengujian integrasi end-to-end.

---

## 7. Approval
- User approval diberikan pada 19 September 2026.
- Transisi ke tahap penulisan rencana implementasi detail (*Implementation Plan*).
