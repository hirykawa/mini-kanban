#!/bin/bash
set -e

# Install dependencies if needed
cd web && npm install && cd ..

# Start frontend watch build in background
(cd web && npm run build -- --watch) &
WATCH_PID=$!

# Start Go server in dev mode with debug logging
go run ./cmd web --dev --debug

# Cleanup on exit
trap "kill $WATCH_PID 2>/dev/null" EXIT
