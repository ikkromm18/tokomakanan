# tokomakanan-fe Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Membangun aplikasi front-end berstandar industri berbasis React 18+, TypeScript, Vite, dan Tailwind CSS bernama `tokomakanan-fe` untuk sistem Bakery POS & Management terintegrasi penuh dengan backend Go Gin Gonic (`tokomakanan`), dilengkapi Git remote `git@github.com:ikkromm18/tokomakanan-fe.git`.

**Architecture:** Mengadopsi arsitektur *Feature-Driven Modular Architecture* yang memisahkan *Core Infrastructure* (`core/`), *Shared UI Primitives* (`components/ui/`), dan *Domain Business Modules* (`features/*`). State terbagi bersih antara *Server State* (TanStack Query v5) dan *Client State* (Zustand), diamankan dengan *DOMPurify*, *RBAC Guards*, dan dilengkapi *Structured Observability* (`X-Request-ID` tracing).

**Tech Stack:** React 18+, TypeScript, Vite, Tailwind CSS, TanStack Query v5, Zustand, Axios, React Hook Form, Zod, Lucide React, DOMPurify.

## Global Constraints
- Target folder: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe`
- Backend API Source of Truth: `http://localhost:8080/api/v1` (sesuai `tokomakanan/API_DOCUMENTATION.md` dan `tokomakanan/docs/prd.md`)
- Git Remote: `git@github.com:ikkromm18/tokomakanan-fe.git`
- Node version: `v22.18.0` / npm: `10.9.3`
- TypeScript: Strict mode enabled, zero `any`.
- Standard Coding Reference: `tokomakanan-fe/CODING_STANDARD.md` (selaras dengan backend `tokomakanan/CODING_STANDARD.md`).
- RBAC: 3 Roles (`superadmin`, `owner`, `admin`). Kasir (`admin`) dilarang melihat HPP dan laba kotor.

---

## Phase 1: Inisialisasi Proyek, Git Remote, Core Infrastructure & CODING_STANDARD

### Task 1: Inisialisasi Vite React TS, Tailwind CSS, Lucide & Git Remote
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/package.json`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/vite.config.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/tsconfig.json`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/tailwind.config.js`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/postcss.config.js`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/.env.example`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/.gitignore`

- [ ] **Step 1: Scaffold Vite project dengan template React TypeScript**
  Jalankan perintah npm create vite di direktori root workspace untuk membuat `tokomakanan-fe`.
- [ ] **Step 2: Install dependencies inti**
  Install: `axios`, `@tanstack/react-query`, `zustand`, `react-router-dom`, `react-hook-form`, `@hookform/resolvers`, `zod`, `lucide-react`, `dompurify`, `@types/dompurify`, `clsx`, `tailwind-merge`.
  Install devDependencies: `tailwindcss`, `postcss`, `autoprefixer`, `@types/node`.
- [ ] **Step 3: Konfigurasi Tailwind CSS dan Path Alias `@/*`**
  Set up `vite.config.ts` dengan `resolve.alias: { '@': path.resolve(__dirname, './src') }` dan `tsconfig.json` paths.
- [ ] **Step 4: Inisialisasi Git dan set remote GitHub**
  Inisialisasi git repository di `tokomakanan-fe`, buat branch `main`, dan pasang remote `git remote add origin git@github.com:ikkromm18/tokomakanan-fe.git`.
- [ ] **Step 5: Verifikasi build Vite awal**
  Jalankan `npm run build` untuk memastikan kompilasi awal sukses. Commit initial commit.

---

### Task 2: Pembuatan `CODING_STANDARD.md` dan Core Types
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/CODING_STANDARD.md`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/core/types/api.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/core/types/common.ts`

- [ ] **Step 1: Tulis dokumen standar koding resmi `CODING_STANDARD.md`**
  Cantumkan aturan arsitektur lapis, manajemen state, standar DTO TypeScript, aturan keamanan XSS & JWT, konvensi penamaan, dan penanganan respon HTTP backend.
- [ ] **Step 2: Definisikan type TypeScript standar**
  Buat generic interfaces: `ApiResponse<T>`, `PaginatedResponse<T>`, `PaginationMeta`, `ApiErrorResponse`, `FieldError`.
- [ ] **Step 3: Commit dokumen standar dan tipe dasar**
  Commit: `docs: add frontend CODING_STANDARD and core types`.

---

### Task 3: Core Security & Token Storage Abstraction
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/core/security/tokenStorage.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/core/security/sanitizer.ts`

- [ ] **Step 1: Buat `tokenStorage.ts`**
  Implementasikan fungsi get, set, dan remove token JWT dengan fallback in-memory dan isolasi namespace `tokomakanan_token`.
- [ ] **Step 2: Buat `sanitizer.ts` menggunakan DOMPurify**
  Implementasikan wrapper sanitasi string XSS untuk catatan order dan footer nota.
- [ ] **Step 3: Commit**
  Commit: `feat(core): implement secure token storage and DOMPurify sanitizer`.

---

### Task 4: Core Structured Logger & Tracing
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/core/logger/logger.ts`

- [ ] **Step 1: Implementasi structured client logger**
  Logger mendukung level `debug`, `info`, `warn`, `error` dengan output JSON terstruktur, timestamp, nama komponen, dan requestId korelasi. Masking kata sandi & token secara otomatis.
- [ ] **Step 2: Commit**
  Commit: `feat(core): implement structured client-side logger`.

---

### Task 5: Core API Client (Axios Interceptors & Error Contract)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/core/config/env.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/core/api/client.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/core/api/errorHandler.ts`

- [ ] **Step 1: Buat `env.ts`**
  Ekspor konfigurasi URL backend (`http://localhost:8080/api/v1` default).
- [ ] **Step 2: Konfigurasi Axios Client & Interceptors**
  Request interceptor menyuntikkan `Authorization: Bearer <token>` dan `X-Request-ID: uuid()`. Response interceptor meng-unwrap `{ data, meta }` dan melempar error tertata.
- [ ] **Step 3: Buat `errorHandler.ts`**
  Menerjemahkan kode status 400 (validation), 401 (auto logout), 403 (forbidden), 409 (conflict), 422 (business violation), dan 500 menjadi pesan ramah pengguna.
- [ ] **Step 4: Commit**
  Commit: `feat(core): setup axios client with correlation ID and error mapping`.

---

### Task 6: Shared UI Primitives & Design System (Tailwind)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/ui/Button.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/ui/Input.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/ui/Card.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/ui/Badge.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/ui/Modal.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/ui/Table.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/feedback/Toast.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/feedback/LoadingSpinner.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/feedback/EmptyState.tsx`

- [ ] **Step 1: Implementasikan komponen UI dasar dengan styling bakery modern**
  Button dengan variant (`primary`, `secondary`, `danger`, `outline`, `ghost`) dan support loading spinner. Input dengan status error field. Modal dengan accessibility keyboard ESC dan backdrop blur.
- [ ] **Step 2: Implementasikan Toast notification system**
  Komponen toast reaktif untuk notifikasi sukses, peringatan, dan error.
- [ ] **Step 3: Commit**
  Commit: `feat(ui): implement reusable design system primitives and feedback components`.

---

### Task 7: Routing Shell, Guards & Global Error Boundary
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/feedback/ErrorBoundary.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/guard/ProtectedRoute.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/guard/RoleGuard.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/routes/index.tsx`
- Modify: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/App.tsx`

- [ ] **Step 1: Buat `ErrorBoundary.tsx`**
  Class component penangkap crash rendering dengan tampilan fallback ramah kasir dan tombol reload.
- [ ] **Step 2: Buat `ProtectedRoute.tsx` & `RoleGuard.tsx`**
  Melindungi rute dari akses non-login dan memfilter elemen UI sesuai role (`superadmin`, `owner`, `admin`).
- [ ] **Step 3: Setup routing dasar di `src/routes/index.tsx` dan `App.tsx`**
  Pasang `QueryClientProvider`, `ToastContainer`, dan `RouterProvider`.
- [ ] **Step 4: Commit**
  Commit: `feat(routes): implement react router with ErrorBoundary and RBAC guards`.

---

## Phase 2: Auth Module, App Shell Layout, Store Settings & User Management

### Task 8: Modul Autentikasi (`features/auth`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/auth/types/auth.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/auth/api/authApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/stores/authStore.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/auth/pages/LoginPage.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/auth/components/ChangePasswordModal.tsx`

- [ ] **Step 1: Buat DTO & API calls autentikasi**
  Koneksi ke `POST /auth/login`, `GET /auth/me`, dan `PUT /auth/change-password`.
- [ ] **Step 2: Implementasi `authStore.ts` (Zustand)**
  Kelola status login, profil user, penyimpanan token via `tokenStorage`, dan fungsi logout.
- [ ] **Step 3: Desain Halaman Login (`LoginPage.tsx`)**
  Desain elegan bernuansa bakery dengan form validasi React Hook Form + Zod, handling error credentials 401.
- [ ] **Step 4: Buat modal ganti kata sandi (`ChangePasswordModal.tsx`)**
- [ ] **Step 5: Commit**
  Commit: `feat(auth): implement login, me, and change password with authStore`.

---

### Task 9: App Shell Layout & Navigasi Responsif
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/layout/AppLayout.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/layout/Sidebar.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/components/layout/Navbar.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/stores/uiStore.ts`

- [ ] **Step 1: Buat `Sidebar.tsx`**
  Menu navigasi dinamis yang memfilter link sesuai role user (contoh: menu *Laporan Keuangan* hanya tampil untuk `superadmin` dan `owner`).
- [ ] **Step 2: Buat `Navbar.tsx`**
  Menampilkan nama toko aktif, informasi kasir/user bertugas dengan badge role berwarna khusus, tombol ganti password, dan tombol logout.
- [ ] **Step 3: Satukan di `AppLayout.tsx`**
- [ ] **Step 4: Commit**
  Commit: `feat(layout): implement responsive app layout with role-filtered sidebar`.

---

### Task 10: Modul Pengaturan Toko (`features/settings`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/settings/types/settings.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/settings/api/settingsApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/settings/pages/SettingsPage.tsx`

- [ ] **Step 1: Buat types dan API calls store settings**
  Koneksi ke `GET /store-settings` dan `PUT /store-settings`.
- [ ] **Step 2: Buat `SettingsPage.tsx`**
  Form edit nama toko, alamat, nomor telepon, logo URL, dan teks footer struk. Role `admin` hanya dapat melihat (read-only), sedangkan `superadmin` & `owner` dapat menyimpan perubahan.
- [ ] **Step 3: Commit**
  Commit: `feat(settings): implement store settings view and update page`.

---

### Task 11: Modul Manajemen Pengguna (`features/users`, Superadmin Only)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/users/types/users.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/users/api/usersApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/users/pages/UsersPage.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/users/components/UserModal.tsx`

- [ ] **Step 1: Buat types dan API calls user CRUD**
  Koneksi ke `GET /users`, `POST /users`, `PUT /users/:id`, `DELETE /users/:id`.
- [ ] **Step 2: Buat tabel user dengan pagination & filter role**
- [ ] **Step 3: Buat modal tambah & edit akun user**
  Validasi role (`owner` atau `admin`), password minimal 8 karakter.
- [ ] **Step 4: Commit**
  Commit: `feat(users): implement user management for superadmin`.

---

## Phase 3: Master Data Management (Products, Packages & Customers)

### Task 12: Modul Master Produk (`features/products`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/products/types/product.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/products/api/productApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/products/pages/ProductsPage.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/products/components/ProductModal.tsx`

- [ ] **Step 1: Buat types dan API calls produk**
  Koneksi ke `GET /products`, `POST /products`, `PUT /products/:id`, `DELETE /products/:id`.
- [ ] **Step 2: Buat `ProductsPage.tsx`**
  Tabel produk dengan pencarian nama, filter kategori, badge status aktif, dan kolom HPP yang otomatis disembunyikan jika role adalah `admin`.
- [ ] **Step 3: Buat `ProductModal.tsx`**
  Validasi aturan bisnis front-end: `sell_price >= hpp` dengan pesan error interaktif.
- [ ] **Step 4: Commit**
  Commit: `feat(products): implement product master data with RBAC HPP masking`.

---

### Task 13: Modul Master Paket Bundling (`features/packages`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/packages/types/package.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/packages/api/packageApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/packages/pages/PackagesPage.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/packages/components/PackageModal.tsx`

- [ ] **Step 1: Buat types dan API calls package**
  Koneksi ke `GET /packages`, `POST /packages`, `PUT /packages/:id`, `DELETE /packages/:id`.
- [ ] **Step 2: Buat `PackagesPage.tsx`**
  Menampilkan daftar paket beserta rincian produk penyusunnya.
- [ ] **Step 3: Buat `PackageModal.tsx`**
  Selector produk multi-baris dinamis. Menghitung estimasi `total_hpp` secara langsung di browser dan memvalidasi `sell_price >= total_hpp`.
- [ ] **Step 4: Commit**
  Commit: `feat(packages): implement bundling package management with dynamic items`.

---

### Task 14: Modul Master Pelanggan (`features/customers`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/customers/types/customer.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/customers/api/customerApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/customers/pages/CustomersPage.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/customers/components/CustomerModal.tsx`

- [ ] **Step 1: Buat types dan API calls customer**
  Koneksi ke `GET /customers`, `POST /customers`, `PUT /customers/:id`.
- [ ] **Step 2: Buat `CustomersPage.tsx`**
  Tabel pelanggan dengan pencarian cepat nama atau nomor WhatsApp.
- [ ] **Step 3: Buat `CustomerModal.tsx`**
  Form tambah/update pelanggan (Nama, Nomor Telepon/WA, Alamat).
- [ ] **Step 4: Commit**
  Commit: `feat(customers): implement customer management`.

---

## Phase 4: High-Speed POS & Kasir Engine, Thermal Printing & Orders

### Task 15: Store Keranjang Kasir Reaktif (`stores/posCartStore.ts`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/stores/posCartStore.ts`

- [ ] **Step 1: Implementasi Zustand POS Cart Store**
  Method: `addItem(item, type)`, `removeItem(id)`, `updateQuantity(id, qty)`, `setDiscount(amount)`, `setOrderType(type)`, `setCustomer(customer)`, `setPickupDate(date)`, `setNotes(notes)`, `clearCart()`.
  Computed: `subtotal`, `totalAmount`, `totalItems`.
- [ ] **Step 2: Commit**
  Commit: `feat(pos): implement reactive high-speed posCartStore`.

---

### Task 16: Layar Utama POS Kasir (`features/pos/pages/PosPage.tsx`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/pos/pages/PosPage.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/pos/components/ProductCatalog.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/pos/components/CartSidebar.tsx`

- [ ] **Step 1: Buat `ProductCatalog.tsx`**
  Tab filter kategori (Semua, Roti, Pastry, Paket Bundling), grid kartu produk dengan foto/icon, harga jual, dan aksi sekali klik untuk menambahkan ke keranjang.
- [ ] **Step 2: Buat `CartSidebar.tsx`**
  Panel kanan dengan rincian item keranjang, tombol kuantitas besar, toggle order type `DIRECT_SALE` vs `PRE_ORDER`, selector pelanggan, dan tombol "Lanjut ke Pembayaran".
- [ ] **Step 3: Gabungkan di `PosPage.tsx`**
- [ ] **Step 4: Commit**
  Commit: `feat(pos): implement touch-friendly split-screen POS layout`.

---

### Task 17: Modal Pembayaran Kasir & Transaksi POS
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/pos/components/PaymentModal.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/orders/api/ordersApi.ts`

- [ ] **Step 1: Buat `ordersApi.ts`**
  Koneksi ke `POST /orders`, `GET /orders`, `GET /orders/:id`, `PATCH /orders/:id/status`.
- [ ] **Step 2: Buat `PaymentModal.tsx`**
  Pilihan metode (`CASH`, `TRANSFER`, `QRIS`), tombol nominal cepat (*Uang Pas*, *Rp 20.000*, *Rp 50.000*, *Rp 100.000*), perhitungan kembalian instan, dan penanganan DP pada Pre-Order.
- [ ] **Step 3: Integrasikan submit order dengan API backend**
  Setelah sukses, trigger dialog konfirmasi cetak nota thermal dan reset keranjang kasir.
- [ ] **Step 4: Commit**
  Commit: `feat(pos): implement quick payment modal with change calculator`.

---

### Task 18: Engine Cetak Struk Thermal (80mm & 58mm) & Nota WhatsApp
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/pos/components/ThermalReceipt.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/hooks/useThermalPrint.ts`
- Modify: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/styles/index.css`

- [ ] **Step 1: Tambahkan stylesheet print thermal di `index.css`**
  Aturan `@media print` khusus ukuran lebar 80mm dan 58mm tanpa margin halaman dan tanpa header URL browser.
- [ ] **Step 2: Buat komponen `ThermalReceipt.tsx`**
  Format struk rapi: Header toko, no invoice, nama kasir, daftar belanja, subtotal, diskon, total bayar, kembalian, dan footer.
- [ ] **Step 3: Buat hook `useThermalPrint.ts` & generator link WhatsApp**
  Trigger print otomatis via browser dan buat tautan WA instan dengan `invoice_token`.
- [ ] **Step 4: Commit**
  Commit: `feat(receipt): implement 80mm/58mm thermal print engine and WhatsApp share`.

---

### Task 19: Modul Manajemen Pesanan (`features/orders`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/orders/types/order.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/orders/pages/OrdersListPage.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/orders/pages/OrderDetailPage.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/orders/components/OrderStatusBadge.tsx`

- [ ] **Step 1: Buat `OrdersListPage.tsx`**
  Tabel pesanan dengan filter tanggal (*date picker*), tipe order (`DIRECT_SALE`/`PRE_ORDER`), status order, dan pencarian no invoice/nama pelanggan.
- [ ] **Step 2: Buat `OrderDetailPage.tsx`**
  Menampilkan detail pesanan lengkap: data pelanggan, item transaksi, histori pembayaran, dan tombol update status pesanan (`DRAFT` ➔ `DP_PAID` ➔ `PAID` ➔ `READY` ➔ `COMPLETED`).
- [ ] **Step 3: Commit**
  Commit: `feat(orders): implement orders list and order detail with status transitions`.

---

### Task 20: Modul Pembayaran Tambahan / Pelunasan PO (`features/payments`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/payments/types/payment.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/payments/api/paymentsApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/payments/components/AddPaymentModal.tsx`

- [ ] **Step 1: Buat API calls payments**
  Koneksi ke `POST /orders/:id/payments` dan `GET /orders/:id/payments`.
- [ ] **Step 2: Buat `AddPaymentModal.tsx`**
  Modal pelunasan cicilan atau sisa tagihan PO. Validasi ketat mencegah overpayment (`amount <= remaining_bill`).
- [ ] **Step 3: Commit**
  Commit: `feat(payments): implement add payment modal with overpayment guard`.

---

## Phase 5: Dashboard Analytics, Laporan Keuangan & Nota Digital Publik

### Task 21: Dashboard Analitik (`features/dashboard`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/dashboard/types/dashboard.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/dashboard/api/dashboardApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/dashboard/pages/DashboardPage.tsx`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/dashboard/components/POReminderWidget.tsx`

- [ ] **Step 1: Buat types dan API calls dashboard**
  Koneksi ke `GET /dashboard/summary`, `GET /dashboard/chart`, `GET /dashboard/po-reminders`.
- [ ] **Step 2: Buat `DashboardPage.tsx`**
  Kartu ringkasan harian (total penjualan, jumlah transaksi, PO aktif, dan laba kotor hari ini yang di-guard dari role admin). Grafik tren penjualan sederhana berbasis SVG/Canvas.
- [ ] **Step 3: Buat widget pengingat PO hari ini (`POReminderWidget.tsx`)**
- [ ] **Step 4: Commit**
  Commit: `feat(dashboard): implement dashboard analytics and PO reminder widget`.

---

### Task 22: Laporan Keuangan & Ekspor CSV (`features/reports`, Superadmin & Owner Only)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/reports/types/reports.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/reports/api/reportsApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/reports/pages/ReportsPage.tsx`

- [ ] **Step 1: Buat types dan API calls reports**
  Koneksi ke `GET /reports/sales`, `GET /reports/profit`, dan `GET /reports/export`.
- [ ] **Step 2: Buat `ReportsPage.tsx`**
  Filter rentang tanggal (*start_date*, *end_date*), tabel breakdown harian pendapatan vs HPP vs Laba Kotor beserta persentase margin keuntungan.
- [ ] **Step 3: Tombol Ekspor CSV dengan BOM**
  Fungsi download file CSV dari endpoint `/reports/export` dengan penanganan header otomatis.
- [ ] **Step 4: Commit**
  Commit: `feat(reports): implement sales and profit reports with CSV export`.

---

### Task 23: Nota Digital Publik Pelanggan (`features/public-invoice`)
**Files:**
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/public-invoice/types/publicInvoice.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/public-invoice/api/publicInvoiceApi.ts`
- Create: `/home/muhammad-ikrom/Toko Makanan/tokomakanan-fe/src/features/public-invoice/pages/PublicInvoicePage.tsx`

- [ ] **Step 1: Buat API call public invoice**
  Koneksi ke `GET /public/invoice/:token` (tanpa token JWT).
- [ ] **Step 2: Desain `PublicInvoicePage.tsx` (Mobile-First)**
  Tampilan nota digital bersih untuk pelanggan: status pesanan, rincian barang, tanggal pengambilan PO, riwayat bayar, tombol cetak nota PDF & share WhatsApp. Zero kebocoran data HPP.
- [ ] **Step 3: Commit**
  Commit: `feat(public-invoice): implement public digital invoice page for customers`.

---

### Task 24: Verifikasi Akhir, Linting, Build Check & Git Remote Push
**Files:**
- All created files

- [ ] **Step 1: Jalankan TypeScript typecheck**
  Perintah: `npm run build` (menjalankan `tsc -b && vite build`). Pastikan 0 type error.
- [ ] **Step 2: Verifikasi routing dan akses RBAC**
  Pastikan seluruh 3 role (`superadmin`, `owner`, `admin`) dan public invoice route terverifikasi dengan benar.
- [ ] **Step 3: Setup Git remote dan push branch main**
  Verifikasi remote `git@github.com:ikkromm18/tokomakanan-fe.git`.
- [ ] **Step 4: Buat Walkthrough Dokumen**
  Dokumentasikan seluruh komponen yang telah berhasil dibangun.
