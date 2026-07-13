# FINAL FIX: Slack Not Sending Events

## Problem
Your server is running perfectly, Socket Mode is connected, but NO events are coming through (no DMs, no slash commands, nothing).

## Root Cause
One of these issues:
1. Testing in wrong Slack workspace
2. Old/invalid tokens
3. App not properly installed in the workspace

## Solution (5 minutes)

### Step 1: Verify You're in the Right Workspace

Your tokens are for workspace: **`slack_hackathon_test`**

Make sure you're testing in a workspace called "slack_hackathon_test" (or similar).

### Step 2: Reinstall App & Get Fresh Tokens

1. Go to: https://api.slack.com/apps
2. Click your **Signal** app
3. Click **"Install App"** in left sidebar
4. Click **"Reinstall to Workspace"**
5. Choose your workspace and click **"Allow"**
6. **COPY THE NEW TOKENS:**
   - **Bot User OAuth Token** (starts with `xoxb-`)
   - Go to **"Basic Information"** → **"App-Level Tokens"**
   - **Copy the App Token** (starts with `xapp-`)

### Step 3: Update .env File

```bash
cd /home/arpit/Desktop/hackathon_projects/signal
nano .env
```

Update these lines with your NEW tokens:
```
SLACK_BOT_TOKEN=xoxb-YOUR-NEW-BOT-TOKEN
SLACK_APP_TOKEN=xapp-YOUR-NEW-APP-TOKEN
```

Save and exit (Ctrl+X, Y, Enter)

### Step 4: Restart Server

```bash
# Kill old server
ps aux | grep "go run.*cmd/api" | grep -v grep | awk '{print $2}' | xargs kill -9

# Start fresh
cd api
go run ./cmd/api/main.go
```

### Step 5: Test Immediately

In Slack (make sure you're in the RIGHT workspace):

1. **Send a DM to Signal:** Type `hi`
2. **Use a slash command:** Type `/signal`

You should see logs appear IMMEDIATELY in your server showing:
```
"msg":"handling message event"
OR
"msg":"handling slash command"
```

---

## Alternative: Create a NEW Test App

If reinstalling doesn't work, create a completely fresh app:

### Method 1: Using Your Manifest

1. Go to: https://api.slack.com/apps
2. Click **"Create New App"**
3. Click **"From an app manifest"**
4. Select your workspace
5. Paste the contents of `slack-manifest.yml`
6. Click **"Create"**
7. Go to **"Install App"** → **"Install to Workspace"**
8. Copy the new tokens and update `.env`

### Method 2: Using Slack CLI (Fastest!)

```bash
# Install Slack CLI if you don't have it
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash

# Login
slack login

# Create app from manifest
cd /home/arpit/Desktop/hackathon_projects/signal
slack create signal --manifest slack-manifest.yml

# The CLI will give you the tokens automatically!
```

---

## How to Verify It's Working

After reinstalling/updating tokens, when you send a DM or slash command, you should see in the logs:

```
{"msg":"handling message event","channel":"D...","user":"U..."}
```

NOT just:
```
{"msg":"slack socket mode connected"}
```

The "connected" messages are just keep-alive pings. Actual events have different log messages.

---

## Still Not Working?

Check these:

1. **Is Signal app visible in your Slack workspace?**
   - Go to Apps → Should see "Signal" listed

2. **Can you open Signal's App Home?**
   - Click on Signal in Apps list
   - Should see the home tab

3. **Are you using a free Slack workspace?**
   - Socket Mode works on ALL Slack plans (free, pro, enterprise)

4. **Check App Tokens page:**
   - Go to https://api.slack.com/apps → Your App → Basic Information
   - Under "App-Level Tokens", make sure token has:
     - ✅ `connections:write` scope
     - ✅ Token is not expired/revoked

If none of this works, share:
- Your workspace name
- Screenshot of "Install App" page showing installation status
- Server logs from the moment you send a command
