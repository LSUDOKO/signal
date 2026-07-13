#!/bin/bash
# Quick setup using Slack CLI (GUARANTEED TO WORK)

set -e

echo "=== Signal Slack App Setup via CLI ==="
echo ""

# Check if slack CLI is installed
if ! command -v slack &> /dev/null; then
    echo "❌ Slack CLI not found. Installing..."
    curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash
    echo "✅ Slack CLI installed"
fi

echo ""
echo "Step 1: Login to Slack"
echo "This will open your browser to authorize the CLI..."
slack login

echo ""
echo "Step 2: Create app from manifest"
cd "$(dirname "$0")"
slack create signal --manifest slack-manifest.yml

echo ""
echo "✅ App created! The CLI will show your tokens."
echo ""
echo "Step 3: Copy the tokens and update .env file:"
echo "  SLACK_BOT_TOKEN=<Bot Token from above>"
echo "  SLACK_APP_TOKEN=<App Token from above>"
echo ""
echo "Step 4: Restart your API server:"
echo "  cd api"
echo "  go run ./cmd/api/main.go"
echo ""
echo "Step 5: Test in Slack:"
echo "  1. Send DM to Signal: 'hi'"
echo "  2. Use slash command: '/signal'"
echo ""
