#!/bin/sh
set -eu

if [ "${SKIP_MIGRATIONS:-false}" = "true" ]; then
  echo "Skipping migrations because SKIP_MIGRATIONS=true"
  exec "$@"
fi

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is not set; cannot run migrations" >&2
  exit 1
fi

MIGRATION_RETRIES="${MIGRATION_RETRIES:-30}"
MIGRATION_RETRY_DELAY="${MIGRATION_RETRY_DELAY:-2}"

attempt=1
while [ "$attempt" -le "$MIGRATION_RETRIES" ]; do
  if goose -dir /migrations postgres "$DATABASE_URL" up; then
    echo "Migrations applied successfully"
    exec "$@"
  fi

  if [ "$attempt" -eq "$MIGRATION_RETRIES" ]; then
    echo "Failed to apply migrations after ${MIGRATION_RETRIES} attempts" >&2
    exit 1
  fi

  echo "Migration attempt ${attempt}/${MIGRATION_RETRIES} failed; retrying in ${MIGRATION_RETRY_DELAY}s..."
  sleep "$MIGRATION_RETRY_DELAY"
  attempt=$((attempt + 1))
done

exit 1
