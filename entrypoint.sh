#!/bin/sh
set -e

# Start Chrome headless in background
/headless-shell/headless-shell \
  --no-sandbox \
  --remote-debugging-address=0.0.0.0 \
  --remote-debugging-port=9223 &

# Start Go server (chạy foreground)
exec /app/server
