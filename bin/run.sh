#!/usr/bin/env bash
set -euo pipefail
export DATABASE_URL="postgresql://go_project_user:5crLwQD0QYVCjkppX5Dtjn2IPWvoBz5@dpg-d5svobu3jp1c738v4g40-a.oregon-postgres.render.com/go_project_db?sslmode=require"
echo "[run.sh] Starting service"

echo "[run.sh] Running DB migrations"
goose -dir ./db/migrations postgres "${DATABASE_URL}" up

echo "[run.sh] Starting Caddy"
caddy run --config /etc/caddy/Caddyfile &

echo "[run.sh] Starting Go app"
exec /app/bin/app