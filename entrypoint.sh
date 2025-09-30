#!/bin/sh
# Start Chrome headless in background
/headless-shell/headless-shell \
  --no-sandbox \
  --remote-debugging-address=0.0.0.0 \
  --remote-debugging-port=9222 &

# Start Go server (foreground)
/app/server

