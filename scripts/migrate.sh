
#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "================================"
echo " Unified Data Platform"
echo " Running database migrations"
echo "================================"

echo "[1/3] MySQL Cadastro"

docker compose exec -T mysql-cadastro \
  sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysql -u root "$MYSQL_DATABASE"' \
  < database/source-a/001_init.sql

echo "[OK] MySQL Cadastro"

echo "[2/3] MySQL Financeiro"

docker compose exec -T mysql-financeiro \
  sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysql -u root "$MYSQL_DATABASE"' \
  < database/source-b/001_init.sql

echo "[OK] MySQL Financeiro"

echo "[3/3] PostgreSQL Warehouse"

docker compose exec -T postgres-warehouse \
  sh -c 'exec psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < database/warehouse/001_init.sql

echo "[OK] PostgreSQL Warehouse"

echo "================================"
echo " All migrations completed!"
echo "================================"
