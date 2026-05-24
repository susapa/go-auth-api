# go-auth-api

## What this is
A Go REST API for JWT authentication and slip file uploads. Backend for `angular-auth-app`.
Stack: Go 1.25, Gin 1.9, PostgreSQL 15, database/sql + lib/pq (no ORM), golang-jwt/jwt v5, bcrypt, Azure Blob Storage SDK v1.7.
Deployed on **Azure Container Apps** via GitHub Actions CI/CD → Azure Container Registry (ACR).

## Architecture: handlers → services → repository
- `handlers/` = HTTP only. Bind JSON/FormData, call service/storage, write JSON. No SQL, no business logic.
- `services/` = Business logic only. No SQL, no gin.Context.
- `repository/` = SQL only. No business logic, no HTTP types.
- `models/` = Shared structs (User, SlipUpload, Claims, DTOs). No methods except simple helpers.
- `middleware/` = Gin middleware functions only.
- `config/` = Reads env vars once at startup. Nothing else ever calls os.Getenv.
- `db/` = Opens *sql.DB singleton. Migrations are plain SQL files in `db/migrations/`.
- `jobs/` = Background goroutines. Receive `context.Context` + `*sql.DB`. No HTTP types.
- `storage/` = Azure Blob Storage client. `Init()` once in main.go, `Upload()` from handlers.

## Background jobs
- `jobs.StartTokenCleanup(ctx, db, interval)` — deletes expired rows from `refresh_tokens` every 24 hours
- Pattern: `context.WithCancel` → goroutine selects on `ticker.C` and `ctx.Done()` — stops cleanly on `cancel()`
- Wire new jobs in `main.go` by passing the same `ctx` created at startup

## API routes
| Method | Path | Auth | Handler |
|--------|------|------|---------|
| POST | /auth/register | public | handlers.Register |
| POST | /auth/login | public | handlers.Login |
| POST | /auth/refresh | public | handlers.RefreshToken |
| GET | /auth/me | Bearer | handlers.Me |
| POST | /slips/upload | Bearer | handlers.UploadSlip |
| GET | /slips/report | Bearer | handlers.GetSlipReport |

## File uploads (Azure Blob Storage)
- `storage/blob.go` — init client + upload stream ไป Azure Blob Storage
- `handlers/slip.go` — รับ multipart/form-data, stream ไป Blob, เก็บ URL ใน DB
- ไฟล์ถูกเก็บที่: `https://<account>.blob.core.windows.net/slips/<filename>`
- Blob URL ถูก store ใน `slip_uploads.path` column
- Max upload size: 10 MB. Allowed types: jpg, png, gif, webp, pdf.
- ห้ามเขียน file I/O ลง local disk ในทุก handler — ใช้ `storage.Upload()` เท่านั้น

## Config & Environment
โหลด config ตาม `APP_ENV` env var:
- `APP_ENV=development` → โหลด `.env.development`
- `APP_ENV=production` → โหลด `.env.production`
- Fallback: อ่านจาก OS environment variables โดยตรง

**Required env vars:**
| Var | Description |
|-----|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | HS256 signing key (min 64 chars) |
| `AZURE_STORAGE_CONNECTION_STRING` | Azure Storage Account connection string |
| `AZURE_STORAGE_ACCOUNT_NAME` | Storage account name |

**Optional env vars:**
| Var | Default | Description |
|-----|---------|-------------|
| `PORT` | 8080 | Server port |
| `JWT_EXPIRY_HOURS` | 24 | Access token expiry |
| `REFRESH_TOKEN_EXPIRY_DAYS` | 7 | Refresh token expiry |
| `ALLOWED_ORIGINS` | http://localhost:4200 | CORS allowed origins |
| `AZURE_STORAGE_CONTAINER` | slips | Blob container name |
| `APP_ENV` | development | Environment name |

**Env files (gitignored):**
- `.env.development` — local dev
- `.env.production` — production (ไม่ commit ค่าจริง ใส่ใน Container App settings แทน)
- `.env.example` — template สำหรับ commit

## Deployment: Azure Container Apps
- **CI**: `.github/workflows/ci.yml` — `go vet` + `go build` ทุก push/PR
- **CD**: `.github/workflows/cd.yml` — build Docker image → push ACR → update Container App (trigger on push to `main`)
- **GitHub Secrets ที่ต้องมี**: `ACR_LOGIN_SERVER`, `ACR_NAME`, `AZURE_CREDENTIALS`, `CONTAINER_APP_NAME`, `RESOURCE_GROUP`
- **Dockerfile**: multi-stage build, `golang:1.25-alpine` → `alpine:3.19`, non-root user

## Error response format
All errors return: `{"error": "human readable message"}`
Status codes: 400 bad input, 401 unauthorized, 409 duplicate email, 500 internal.

## Conventions
- Passwords: always bcrypt with cost 12. Never store plaintext.
- Config: all env vars loaded in `config/config.go`. Never call `os.Getenv` elsewhere.
- Migrations: named `NNN_description.sql`, run with `bash scripts/migrate.sh`.
- No external ORM. Use `database/sql` with lib/pq driver. Write explicit SQL with `$1 $2` placeholders.
- JWT access token: signed HS256, expiry from `config.C.JWTExpiryHours`.
- Refresh token: random 32-byte hex, stored in `refresh_tokens` table.

## Adding new features
- New endpoint → เพิ่ม handler ใน `handlers/`, route ใน `main.go`
- New DB table → เพิ่ม migration ใน `db/migrations/`
- New background job → เพิ่มใน `jobs/`, wire ใน `main.go` ด้วย `ctx` เดียวกัน

## Rules
- Never write SQL in handlers or services
- Never write HTTP response logic (gin.Context) in services or repository
- Never call `os.Getenv` outside `config/config.go`
- Never write file I/O to local disk — ใช้ `storage.Upload()` เสมอ
- Never commit `.env`, `.env.development`, `.env.production` (gitignored)
- Never use gorm or sqlx. Use `database/sql` only.
- Never store passwords in plaintext or with weak hashing (MD5, SHA1).
- golangci-lint ปิดชั่วคราว (Go 1.25 ยังไม่รองรับ v1.x) — ใช้ `go vet` + `go build` แทน
