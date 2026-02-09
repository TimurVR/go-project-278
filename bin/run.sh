#!/usr/bin/env bash
set -euo pipefail
export DATABASE_URL
echo "[run.sh] Starting service"

echo "[run.sh] Running DB migrations"
goose -dir ./db/migrations postgres "${DATABASE_URL}" up

echo "[run.sh] Starting Caddy"
caddy start --config /etc/caddy/Caddyfile

echo "[run.sh] Starting Go app"
exec /app/bin/app