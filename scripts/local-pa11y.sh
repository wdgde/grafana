#!/usr/bin/env bash

# set -x

. scripts/grafana-server/variables

# Cleanup function to kill the backgrounded server
cleanup() {
  if [ -f "$PIDFILE" ]; then
    echo "Killing grafana-server..."
    kill $(cat "$PIDFILE") 2>/dev/null || true
    rm -f "$PIDFILE"
  fi
}

# Set trap to cleanup on script exit
trap cleanup EXIT

rm -rf ./screenshots/*.png

# Check if server is already running
if curl -s -f http://${HOST:-$DEFAULT_HOST}:${PORT:-$DEFAULT_PORT}/api/health > /dev/null 2>&1; then
  echo "Server is already running"
else
  ./scripts/grafana-server/start-server > scripts/grafana-server/server.log &
  ./scripts/grafana-server/wait-for-grafana
fi

echo ""

yarn run pa11y-ci --config ./.pa11yci-pr.conf.js

# yarn run pa11y --config pa11y.json http://localhost:3001/login

# cat pa11y-ci-results.json



echo "Total: $(jq '.total' pa11y-ci-results.json)"
echo "Passes: $(jq '.passes' pa11y-ci-results.json)"
echo "Errors: $(jq '.errors' pa11y-ci-results.json)"

