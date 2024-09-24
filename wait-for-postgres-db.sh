#!/bin/bash
set -e

until pg_isready -h postgres -U ${POSTGRES_USER}; do
  echo "Waiting for PostgreSQL to be ready..."
  sleep 2
done

echo "PostgreSQL is ready!"
exec "$@"