#!/bin/bash

# マイグレーションスクリプト
# Usage: ./scripts/run_migration.sh [migration_file]

set -e

# 環境変数の読み込み
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME=${DB_NAME:-point_system}

MIGRATIONS_DIR="./migrations"
MIGRATION_FILE=$1

if [ -z "$MIGRATION_FILE" ]; then
    echo "Usage: $0 <migration_file>"
    echo "Example: $0 004_add_user_settings.sql"
    exit 1
fi

MIGRATION_PATH="${MIGRATIONS_DIR}/${MIGRATION_FILE}"

if [ ! -f "$MIGRATION_PATH" ]; then
    echo "Error: Migration file not found: $MIGRATION_PATH"
    exit 1
fi

echo "Running migration: $MIGRATION_FILE"
echo "Database: $DB_NAME on $DB_HOST:$DB_PORT"

PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$MIGRATION_PATH"

echo "Migration completed successfully!"
