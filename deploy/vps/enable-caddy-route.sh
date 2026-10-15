#!/usr/bin/env bash
# Add the /danh-bd-api/* route to the meme-bot VPS Caddy, without breaking the bot's media serving.
# Safe order: probe a real media file -> backup -> validate new config -> reload only if valid
# -> re-probe the same file. Rollback: cp /etc/caddy/Caddyfile.bak_261015 /etc/caddy/Caddyfile && systemctl reload caddy
set -euo pipefail

ssh -i ~/.ssh/memebot_vps -p 26266 root@103.121.90.43 'set -e
F=$(ls -t /opt/project-test/assets/temp_videos | head -1)
echo "probe file: $F"
echo "media before: $(curl -s -o /dev/null -w "%{http_code} %{content_type}" "https://test.sunmedia.net.vn/$F")"
cp -a /etc/caddy/Caddyfile /etc/caddy/Caddyfile.bak_261015
cat > /etc/caddy/Caddyfile.new <<EOF
test.sunmedia.net.vn {
    # Danh birthday API (guestbook). handle_path strips the prefix before proxying.
    handle_path /danh-bd-api/* {
        reverse_proxy 127.0.0.1:8091
    }
    root * /opt/project-test/assets/temp_videos
    file_server
}
EOF
if caddy validate --adapter caddyfile --config /etc/caddy/Caddyfile.new >/tmp/caddy-validate.log 2>&1; then
  mv /etc/caddy/Caddyfile.new /etc/caddy/Caddyfile
  systemctl reload caddy
  sleep 2
  echo "caddy: $(systemctl is-active caddy)"
else
  echo "VALIDATE FAILED - Caddyfile untouched"; tail -5 /tmp/caddy-validate.log; rm -f /etc/caddy/Caddyfile.new; exit 1
fi
echo "media after:  $(curl -s -o /dev/null -w "%{http_code} %{content_type}" "https://test.sunmedia.net.vn/$F")"
echo "api via https: $(curl -s https://test.sunmedia.net.vn/danh-bd-api/healthz)"'
