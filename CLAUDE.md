# go-auth-api

## What this is
A Go REST API for JWT authentication. Backend for `angular-auth-app`.
Stack: Go 1.22, Gin 1.9, PostgreSQL 15, database/sql + lib/pq (no ORM), golang-jwt/jwt v5, bcrypt for passwords.

## Background jobs
- `jobs/` package runs goroutines that live for the server lifetime
- `jobs.StartTokenCleanup(ctx, db, interval)` — deletes expired rows from `refresh_tokens` every 24 hours
- Pattern: `context.WithCancel` → goroutine selects on `ticker.C` and `ctx.Done()` — stops cleanly on `cancel()`
- Wire new jobs in `main.go` by passing the same `ctx` created at startup
- Cleanup SQL lives in `repository/cleanup.go`

## Architecture: handlers → services → repository
- `handlers/` = HTTP only. Bind JSON, call service, write JSON. No SQL, no business logic.
- `services/` = Business logic only. No SQL, no gin.Context.
- `repository/` = SQL only. No business logic, no HTTP types.
- `models/` = Shared structs (User, Claims, request/response DTOs). No methods except simple helpers.
- `middleware/` = Gin middleware functions only.
- `config/` = Reads env vars once. Nothing else ever calls os.Getenv.
- `db/` = Opens *sql.DB singleton. Migrations are plain SQL files in db/migrations/.
- `jobs/` = Background goroutines. Receive `context.Context` + `*sql.DB`. No HTTP types.

## API routes
| Method | Path           | Auth     | Handler          |
|--------|----------------|----------|------------------|
| POST   | /auth/register | public   | handlers.Register |
| POST   | /auth/login    | public   | handlers.Login   |
| POST   | /auth/refresh  | public   | handlers.RefreshToken |
| GET    | /auth/me       | Bearer   | handlers.Me      |

## Error response format
All errors return: `{"error": "human readable message"}`
Status codes: 400 bad input, 401 unauthorized, 409 duplicate email, 500 internal.

## Conventions
- Passwords: always bcrypt with cost 12. Never store plaintext.
- Config: all env vars loaded in config/config.go. Never call os.Getenv elsewhere.
- Migrations: named NNN_description.sql, run with `bash scripts/migrate.sh` or psql directly.
- No external ORM. Use database/sql with lib/pq driver. Write explicit SQL with $1 $2 placeholders.
- JWT access token: signed HS256, expiry from config.C.JWTExpiryHours.
- Refresh token: random 32-byte hex, stored in refresh_tokens table.

## Adding new features
Use the skills in .claude/skills/ for common tasks:
- New endpoint: use `add-api-endpoint` skill
- New DB table or column: use `add-db-migration` skill

## Rules
- Never write SQL in handlers or services
- Never write HTTP response logic (gin.Context) in services or repository
- Never call os.Getenv outside config/config.go
- Never commit .env (it is gitignored). .env.example is the source of truth.
- Never use gorm or sqlx. Use database/sql only.
- Never store passwords in plaintext or with weak hashing (MD5, SHA1).
