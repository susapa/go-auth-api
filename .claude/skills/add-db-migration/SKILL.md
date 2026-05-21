---
name: add-db-migration
description: Use when the user wants to add a new database table, add a column to an existing table, add an index, or make any schema change. Triggers include "add table", "add column", "create migration", "alter table", "add index", "new table for X".
---

# add-db-migration SOP

## When to use
User wants to change the database schema. Examples:
- "Add a posts table"
- "Add email_verified column to users"
- "Add an index on refresh_tokens.user_id"
- "Create the sessions table"

## Inputs needed
1. Table name or column being added/changed
2. Column types and constraints (NOT NULL, DEFAULT, UNIQUE, FK)
3. Is this reversible? Should I also write a DOWN migration?

## Steps

1. **Check existing sequence** — Read `db/migrations/` to find the highest number (e.g. `002`). New file must be `003_description.sql`.

2. **Write the migration file** — Create `db/migrations/NNN_description.sql`:
   - Use `TEXT` not `VARCHAR(255)` unless length matters
   - Use `TIMESTAMPTZ` not `TIMESTAMP`
   - Always `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()` on new tables
   - Use `CREATE TABLE IF NOT EXISTS`
   - Indexes: `CREATE INDEX IF NOT EXISTS idx_tablename_col ON tablename(col)`
   - Foreign keys: always explicit `ON DELETE CASCADE` or `ON DELETE SET NULL`
   - Primary key: `id BIGSERIAL PRIMARY KEY`

3. **Go struct** — If new table, add a struct to `models/`. Field names in CamelCase matching column names. Use `sql.NullString` / `sql.NullTime` for nullable columns.

4. **Run the migration** — Tell user:
   ```bash
   psql $DATABASE_URL -f db/migrations/NNN_description.sql
   ```
   Or: `bash scripts/migrate.sh`

5. **Verify** — Ask user to run:
   ```bash
   psql $DATABASE_URL -c "\d table_name"
   ```

## Output checklist
- [ ] Migration file created with correct sequence number
- [ ] Correct SQL types (TEXT, TIMESTAMPTZ, BIGSERIAL)
- [ ] Go struct updated in models/ (if new table)
- [ ] Run command provided
- [ ] Verify command provided

## Rules
- Never modify an existing migration file that may have already been applied
- Migration files are append-only once committed
- Always TIMESTAMPTZ, never TIMESTAMP
- Always TEXT unless there is a specific length reason
- Never DROP columns or tables without explicit user confirmation first
- New tables must have a primary key
