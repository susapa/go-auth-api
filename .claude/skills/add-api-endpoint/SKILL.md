---
name: add-api-endpoint
description: Use when the user wants to add a new API endpoint or route to the Go backend. Triggers include "add endpoint", "create route", "add a POST /foo", "add handler for X", "new API for Y", "add a GET /something".
---

# add-api-endpoint SOP

## When to use
User wants a new HTTP endpoint. Examples:
- "Add a GET /users/:id endpoint"
- "Create a route to update the user's profile"
- "Add endpoint for changing password"

## Inputs needed
Confirm before writing code:
1. HTTP method and path (e.g. `PUT /users/:id`)
2. Protected (requires JWT) or public?
3. Request body shape (if POST/PUT/PATCH)
4. Success response shape and status code
5. Failure cases (not found, unauthorized, validation error)

## Steps

1. **Model** — If new request/response structs are needed, add to `models/user.go` (or a new models file for the domain). Keep structs small. Do not reuse structs across layers if shapes differ.

2. **Repository** — Only if DB access is needed. Add a function to the relevant file in `repository/`. Use explicit SQL with `$1 $2` placeholders. Return `(result, error)`. Example:
   ```go
   func UpdateUserName(db *sql.DB, id int64, name string) error {
       _, err := db.Exec(`UPDATE users SET name=$1, updated_at=NOW() WHERE id=$2`, name, id)
       return err
   }
   ```

3. **Service** — Add a function to `services/`. Call the repository. Apply business logic (validation, authorization, transformation). Return `(result, error)` with descriptive sentinel errors.

4. **Handler** — Add a function to `handlers/`. Use this exact shape:
   ```go
   func HandlerName(c *gin.Context) {
       // 1. Extract claims if protected: claims, _ := c.Get("claims")
       // 2. Bind input: c.ShouldBindJSON / c.Param / c.Query
       // 3. Call service
       // 4. If error: c.JSON(statusCode, gin.H{"error": msg}); return
       // 5. c.JSON(statusCode, response)
   }
   ```

5. **Register route** — In `main.go`, add to the correct group:
   - Public: inside `auth := r.Group("/auth")`
   - Protected: inside `protected := r.Group("/auth"); protected.Use(middleware.AuthRequired())`
   - New resource: create a new group with `r.Group("/resource-name")`

6. **Test** — Provide a curl command for the user to verify.

## Output checklist
- [ ] Model struct added (if needed)
- [ ] Repository function added (if DB access needed)
- [ ] Service function added
- [ ] Handler function added
- [ ] Route registered in main.go
- [ ] curl command provided

## Rules
- Never put SQL in a handler or service
- Never put gin.Context in a service or repository
- Error responses always: `gin.H{"error": "message"}` with a specific HTTP status code
- Handler always returns after writing a response (never fall through to success after error)
- Protected endpoints must be under the group that uses `middleware.AuthRequired()`
