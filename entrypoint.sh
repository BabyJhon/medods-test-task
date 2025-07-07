#!/bin/sh
set -e

echo "Waiting PostgreSQL"
until pg_isready -h db -U postgres; do
  sleep 1
done

echo "ptrforming migrations"
goose -dir /app/migrations postgres "user=postgres password=$DB_PASSWORD host=db dbname=postgres sslmode=disable" up

echo "Service started"
exec "$@"