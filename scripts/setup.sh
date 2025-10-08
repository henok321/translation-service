#!/bin/bash
set -e

echo "Install git hooks..."

pre-commit install --hook-type pre-commit --hook-type pre-push

echo "Setup database..."
docker-compose up -d

export DATABASE_URL="postgres://postgres:secret@localhost:5432/postgres?sslmode=disable"
export GOOSE_DRIVER=postgres
export GOOSE_MIGRATION_DIR="./db_migration"
export GOOSE_DBSTRING="postgres://postgres:secret@localhost:5432/postgres?sslmode=disable"

until docker exec translation-service-db-1 pg_isready -h localhost -p 5432; do
	echo "Waiting for PostgreSQL to start..."
	sleep 2
done

go tool goose validate
go tool goose up

echo "Init .env..."

echo "DATABASE_URL=$DATABASE_URL" >>.env
