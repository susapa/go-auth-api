#!/bin/bash
# Run all migrations in order against DATABASE_URL from .env
set -e

if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

if [ -z "$DATABASE_URL" ]; then
  echo "ERROR: DATABASE_URL is not set"
  exit 1
fi

echo "Running migrations against: $DATABASE_URL"

for f in db/migrations/*.sql; do
  echo "  applying $f..."
  psql "$DATABASE_URL" -f "$f"
done

echo "All migrations applied."
