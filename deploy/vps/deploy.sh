#!/usr/bin/env bash
# Build the API for linux and install it as its own systemd service on the meme-bot VPS.
# Touches only /opt/danh-birthday, /var/lib/danh-birthday and the new unit file;
# never /opt/project-test or the bot's .env / database. The HTTPS route is added by enable-caddy-route.sh.
set -euo pipefail

HOST=root@103.121.90.43
SSH=(ssh -i ~/.ssh/memebot_vps -p 26266 -o BatchMode=yes "$HOST")
SCP=(scp -i ~/.ssh/memebot_vps -P 26266)
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

echo "== build"
(cd "$ROOT/api" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /tmp/danh-birthday-api .)

echo "== upload"
"${SSH[@]}" 'mkdir -p /opt/danh-birthday'
"${SCP[@]}" /tmp/danh-birthday-api "$HOST:/opt/danh-birthday/api.new"
"${SCP[@]}" "$ROOT/deploy/vps/danh-birthday-api.service" "$HOST:/etc/systemd/system/danh-birthday-api.service"

echo "== install + restart"
"${SSH[@]}" 'set -e
  chmod 755 /opt/danh-birthday/api.new
  mv /opt/danh-birthday/api.new /opt/danh-birthday/api
  systemctl daemon-reload
  systemctl enable --now danh-birthday-api
  systemctl restart danh-birthday-api
  sleep 1
  systemctl is-active danh-birthday-api
  curl -fsS http://127.0.0.1:8091/healthz && echo'
