# Coding Standard — tokomakanan

Stack: Go, Gin, GORM, MySQL, Zerolog, JWT

---

## 1. Arsitektur Layer

```
Handler -> Service -> Repository -> Model/DB
```

**Aturan:**
- Service TIDAK BOLEH import gorm.io/gorm
- Repository: translate ORM error, jangan normalisasi input
- Handler: parse request, tulis response — tidak ada business logic

---

## 2. HTTP Response Convention

Wajib pakai helper dari `internal/pkg/response/response.go`. Jangan menulis c.JSON() langsung di handler.

| Situasi | Status | Helper |
|---|---|---|
| Read berhasil | 200 | response.Success |
| Create berhasil | 201 | response.Created |
| Delete berhasil | 204 | response.NoContent |
| List paginated | 200 | response.Paginated |
| Input tidak valid | 400 | response.ValidationError |
| Tidak auth | 401 | response.HandleServiceError |
| Forbidden | 403 | response.HandleServiceError |
| Not found | 404 | response.HandleServiceError |
| Duplicate | 409 | response.HandleServiceError |
| Business rule | 422 | response.HandleServiceError |
| Server error | 500 | response.Error |

**ATURAN KERAS:** Handler HANYA boleh memanggil `response.HandleServiceError` untuk error dari service.

---

## 3. Error Handling

### Repository — translate ORM error

```go
func (r *userRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil // nil,nil = tidak ditemukan
        }
        return nil, fmt.Errorf("user_repository.FindByID: %w", err)
    }
    return &user, nil
}
```

### Service — gunakan ServiceError

```go
if user == nil {
    return nil, response.NewServiceError(response.ErrNotFound, "User not found")
}
if existing != nil {
    return nil, response.NewServiceError(response.ErrDuplicate, "Email already registered")
}
// Wrap error infrastruktur
return nil, fmt.Errorf("userService.Create: %w", err)
```

### Error Tipe yang Tersedia

- `response.ErrNotFound`     -> 404
- `response.ErrDuplicate`    -> 409
- `response.ErrUnauthorized` -> 401
- `response.ErrForbidden`    -> 403
- `response.ErrBusinessRule` -> 422

---

## 4. Pagination

**DILARANG** menulis ulang logika pagination. Gunakan `internal/pkg/pagination`:

```go
// Di Service
meta := pagination.FormatPagination(page, limit, total)

// Di Repository
offset := pagination.GetOffset(page, limit)

// DILARANG di repo atau handler:
// if page < 1 { page = 1 }
// if limit > 100 { limit = 100 }
// offset := (page - 1) * limit
```

---

## 5. Observability — Zerolog

**Prinsip:**
- Structured logging — selalu tambahkan field kontekstual
- Jangan log password, token JWT, atau data PII

```go
// Error — selalu sertakan .Err() dan .Str("component", ...)
log.Error().
    Err(err).
    Str("component", "userService.Create").
    Msg("failed to create user")
```

**Audit Log — jangan silence total:**

```go
// SALAH:
_ = s.auditService.Log(ctx, entry)

// BENAR:
if err := s.auditService.Log(ctx, entry); err != nil {
    log.Warn().Err(err).Msg("audit log: failed to record entry")
}
```

---

## 6. Keamanan (Security)

- JWT: gunakan HS256, secret dari config, expiry dari `JWT_EXPIRY_HOURS`
- Password: bcrypt via `jwtpkg.HashPassword`, cost dari config (default 12)
- Jangan hardcode cost bcrypt inline — buat method helper

```go
// Wajib ada — jangan inline 3x
func (s *userService) bcryptCost() int {
    if s.cfg != nil && s.cfg.BcryptCost > 0 {
        return s.cfg.BcryptCost
    }
    return 12
}
```

- Input: `ShouldBindJSON` + binding tags + `response.ValidationError`

---

## 7. Helper Reusable

Jangan duplikat — taruh di file yang tepat:

```
internal/handler/helpers.go  -> getUserID(), parseIDParam(), parsePaginationParams()
internal/service/helpers.go  -> toActorPtr(id uint64) *uint64
```

---

## 8. Konstanta & Default Value

```go
const (
    defaultStoreName  = "Toko Makanan"
    defaultBcryptCost = 12
    defaultPageSize   = 20
    maxPageSize       = 100
)
```

---

## 9. Naming Convention

| Element | Contoh |
|---|---|
| File | user_handler.go, user_service.go |
| Constructor | NewUserHandler, NewUserService |
| Converter (private) | toUserResponse |
| Repo method | FindByID, FindAll, Create, Update, Delete |
| Service/Handler | GetByID, List, Create, Update, Delete |

---

## 10. Testing

Coverage target:
- Service: >= 80%
- Handler: >= 70%
- Repository: >= 60%

Gunakan table-driven test.

---

## 11. Audit Trail — Action Label Standar

| Operasi | Action |
|---|---|
| Buat data | "CREATE" |
| Edit data | "UPDATE" |
| Hapus data | "DELETE" |
| Login | "LOGIN" |
| Ganti password | "CHANGE_PASSWORD" |

---

## 12. Git Commit Convention

Format: `type(scope): description`

Types: feat, fix, refactor, test, docs, chore, perf, style

```
feat: add user pagination
fix: resolve duplicate email in update
refactor: extract pagination guard to pkg/pagination
```

---

## 13. Checklist Sebelum PR

- [ ] Tidak ada error di-ignore tanpa alasan
- [ ] Tidak ada duplikasi logika dari pkg/
- [ ] Response pakai helper dari pkg/response
- [ ] Error dari service via response.HandleServiceError
- [ ] String literal berulang sudah jadi konstanta
- [ ] Audit log untuk semua operasi mutasi
- [ ] Test coverage cukup per layer
- [ ] Tidak ada import GORM di service layer
