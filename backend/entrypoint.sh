#!/bin/sh
set -e

echo "🔄 Running database migrations..."

# マイグレーションファイルを番号順に実行
# IF NOT EXISTS や ON CONFLICT DO NOTHING で冪等性を確保しつつ、
# 既存オブジェクトのエラーは無視して続行
for f in /app/migrations/*.sql; do
  echo "  Applying: $(basename $f)"
  PGPASSWORD="${DB_PASSWORD}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -f "$f" 2>&1 | grep -v "already exists\|duplicate key\|NOTICE" || true
done

echo "✅ Migrations complete"

# サーバー起動
exec ./server
