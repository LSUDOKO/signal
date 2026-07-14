# Slack App Configuration Checklist

## Quick Fix for @Signal Mentions Not Working

Go to: https://api.slack.com/apps → Your Signal App

---

## 1. Event Subscriptions (Required for @mentions)

Navigate to: **Event Subscriptions** → **Subscribe to bot events**

Make sure these are enabled:
- ✅ `app_mention` - When someone @mentions your bot
- ✅ `message.channels` - Public channel messages
- ✅ `message.im` - Direct messages
- ✅ `app_home_opened` - App Home tab

**If any are missing:** Add them and save changes.

---

## 2. OAuth Scopes (Required permissions)

Navigate to: **OAuth & Permissions** → **Bot Token Scopes**

Make sure these are enabled:
- ✅ `app_mentions:read` - Read @mentions
- ✅ `chat:write` - Send messages
- ✅ `im:write` - Send DMs
- ✅ `channels:history` - Read channel history
- ✅ `search:read` - Use RTS API
- ✅ `users:read` - Get user info
- ✅ `users.profile:write` - Set status (for Deep Work)

**If any are missing:** Add them and **reinstall the app** (see step 4).

---

## 3. Socket Mode (Must be ON)

Navigate to: **Socket Mode**

- ✅ **Enable Socket Mode** - Should be ON
- ✅ **App-Level Token** - Should exist (starts with `xapp-`)

**If Socket Mode is OFF:** Turn it on, generate an App-Level Token, and update your `.env`:
```bash
SLACK_APP_TOKEN=xapp-your-token-here
```

---

## 4. Reinstall App (Required after adding scopes/events)

If you just added scopes or events:
1. Go to **Install App** in the left sidebar
2. Click **Reinstall to Workspace**
3. Review permissions and click **Allow**
4. Restart your Signal API server

---

## 5. Verify Installation in Slack

1. Open your Slack workspace
2. Go to **Apps** in the left sidebar
3. Find **Signal** in your apps list
4. The app should show as "Active"

---

## 6. Test Commands

### Test @mention (in any channel):
```
@Signal help
```
**Expected:** Signal responds in the channel with a greeting.

### Test /translate command:
```
/translate per my last email
```
**Expected:** You see "⏳ Working on your request..." then get a DM with the translation.

### Test DM:
Send a direct message to Signal:
```
hi
```
**Expected:** Signal responds with the help menu.

---

## Common Issues

### Issue: "@Signal doesn't respond"
**Fix:** 
1. Make sure `app_mention` event is subscribed
2. Make sure `app_mentions:read` scope is granted
3. Reinstall the app
4. Restart your API server

### Issue: "Slash commands timeout"
**Fix:** 
1. Make sure your API server is running (`go run cmd/api/main.go`)
2. Make sure PostgreSQL and Redis are running (`docker-compose ps`)
3. Check API logs for errors

### Issue: "App did not respond"
**Fix:**
1. Check if Socket Mode is enabled
2. Verify `SLACK_APP_TOKEN` in `.env` is correct
3. Restart API server

---

## Quick Restart Commands

```bash
# Restart database services
cd /home/arpit/Desktop/hackathon_projects/signal
docker-compose restart postgres redis

# Restart API server
cd api
go run cmd/api/main.go
```
