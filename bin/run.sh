#!/usr/bin/env bash
set -euo pipefail
export DATABASE_URL="postgresql://go_project_user:Fj2SbLdlar3a4l48bXHObp5r6ewZEzpO@dpg-d5u8jbh4tr6s739dbca0-a/go_project_db_h0do"
echo "[run.sh] Starting service"

echo "[run.sh] Running DB migrations"
goose -dir ./db/migrations postgres "${DATABASE_URL}" up

echo "[run.sh] Starting Caddy"
caddy run --config /etc/caddy/Caddyfile &

echo "[run.sh] Starting Go app"
exec /app/bin/app