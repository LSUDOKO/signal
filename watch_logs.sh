#!/bin/bash
echo "Watching Signal logs in real-time..."
echo "Test your Slack app NOW and watch for events!"
echo "Press Ctrl+C to stop"
echo "=========================================="
tail -f /proc/$(pgrep -f "go run ./cmd/api/main.go")/fd/1 2>/dev/null || echo "Server not running!"
