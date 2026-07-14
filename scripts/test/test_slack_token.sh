#!/bin/bash

# Test Slack Bot Token and diagnose connection issues

BOT_TOKEN="xoxb-11568451159650-11596829152096-oauAP68wF42cU7P08v4XcVl0"
APP_TOKEN="xapp-1-A0BGU5ZBG20-11553124269399-bbe8ada1ab224b258792e6817a853cd8d3d0db8d3acfbba257a1984994edea12"

echo "=========================================="
echo "Signal Slack Token Diagnostics"
echo "=========================================="
echo ""

echo "1. Testing Bot Token with auth.test..."
echo "----------------------------------------"
curl -s -X POST https://slack.com/api/auth.test \
  -H "Authorization: Bearer $BOT_TOKEN" \
  -H "Content-Type: application/json" | jq '.'
echo ""

echo "2. Testing Bot Info..."
echo "----------------------------------------"
curl -s -X POST https://slack.com/api/auth.teams.list \
  -H "Authorization: Bearer $BOT_TOKEN" \
  -H "Content-Type: application/json" | jq '.'
echo ""

echo "3. Testing Bot Scopes..."
echo "----------------------------------------"
curl -s -X POST https://slack.com/api/apps.permissions.info \
  -H "Authorization: Bearer $BOT_TOKEN" \
  -H "Content-Type: application/json" | jq '.'
echo ""

echo "4. Testing if bot can read conversations..."
echo "----------------------------------------"
curl -s -X POST https://slack.com/api/conversations.list \
  -H "Authorization: Bearer $BOT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "limit=5" | jq '.'
echo ""

echo "5. Testing bot user info..."
echo "----------------------------------------"
curl -s -X POST https://slack.com/api/users.info \
  -H "Authorization: Bearer $BOT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "user=U0BHJQD4G2U" | jq '.'
echo ""

echo "=========================================="
echo "NEXT STEPS:"
echo "=========================================="
echo "If auth.test shows 'ok: false' with 'not_authed':"
echo "  → Bot token is invalid or workspace mismatch"
echo "  → Go to: https://api.slack.com/apps"
echo "  → Select your app"
echo "  → Go to 'OAuth & Permissions'"
echo "  → Copy the FULL Bot User OAuth Token (should be 50+ chars)"
echo ""
echo "If auth.test shows 'ok: true':"
echo "  → Token is valid!"
echo "  → Check that Socket Mode is enabled in app settings"
echo "  → Check that Event Subscriptions has the right events"
echo "  → Reinstall the app to workspace"
echo ""
