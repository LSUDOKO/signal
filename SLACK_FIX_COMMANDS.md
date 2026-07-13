# Fix Slack Commands Not Working

## Problem
Slash commands showing "app did not respond" error because Slack isn't sending events to your server via Socket Mode.

## Solution: Enable Interactivity & Slash Commands

### Step 1: Enable Interactivity with Socket Mode

1. Go to: https://api.slack.com/apps
2. Click your Signal app
3. Go to **Interactivity & Shortcuts** (left sidebar)
4. Toggle **Interactivity** to **ON**
5. Under **Interactivity**, select **Socket Mode**
6. Click **Save Changes**

### Step 2: Add Slash Commands

Go to **Slash Commands** (left sidebar) and add these:

#### Command 1: /signal
- Command: `/signal`
- Short Description: `Open Signal help menu`
- Usage Hint: (leave empty)
- **Important:** With Socket Mode, you DON'T need a Request URL

#### Command 2: /translate
- Command: `/translate`
- Short Description: `Translate ambiguous workplace language`
- Usage Hint: `[message to translate]`

#### Command 3: /catchup
- Command: `/catchup`
- Short Description: `Get AI summary of what you missed`
- Usage Hint: `[topic or question]`

#### Command 4: /focus
- Command: `/focus`
- Short Description: `Start deep work mode`
- Usage Hint: `[duration, e.g., 2h]`

#### Command 5: /digest
- Command: `/digest`
- Short Description: `Get instant digest of mentions`
- Usage Hint: (leave empty)

### Step 3: Verify Event Subscriptions

Go to **Event Subscriptions** (left sidebar):

1. Make sure **Enable Events** is ON
2. Under **Subscribe to bot events**, verify these are listed:
   - ✅ `app_mention`
   - ✅ `message.channels`
   - ✅ `message.im`
   - ✅ `app_home_opened`

If any are missing, click **Add Bot User Event** and add them.

3. Click **Save Changes**

### Step 4: Verify OAuth Scopes

Go to **OAuth & Permissions** (left sidebar):

Under **Bot Token Scopes**, verify these scopes:
- ✅ `app_mentions:read`
- ✅ `channels:history`
- ✅ `chat:write`
- ✅ `commands` (this is KEY for slash commands!)
- ✅ `im:history`
- ✅ `im:write`
- ✅ `search:read`
- ✅ `users:read`
- ✅ `users:read.email`
- ✅ `users.profile:write`

**If `commands` scope is missing, ADD IT!**

### Step 5: Reinstall App

After adding scopes or configuring slash commands:

1. Go to **Install App** (left sidebar)
2. Click **Reinstall to Workspace**
3. Review permissions and click **Allow**

### Step 6: Restart Your API Server

```bash
# Kill the current server
ps aux | grep "go run.*cmd/api" | grep -v grep | awk '{print $2}' | xargs kill

# Start fresh
cd /home/arpit/Desktop/hackathon_projects/signal/api
go run ./cmd/api/main.go
```

### Step 7: Test Again

In Slack:
1. Type: `/signal`
   - Expected: Help menu in DM
2. Type: `/translate hello`
   - Expected: "⏳ Working..." then translation in DM
3. DM Signal with "hi"
   - Expected: Help menu appears

---

## Still Not Working?

Check if your app is using the OLD "Request URL" method instead of Socket Mode:

1. Go to **Interactivity & Shortcuts**
2. Make sure it says **"Socket Mode"** NOT a URL
3. If you see a Request URL field, CLEAR IT and select Socket Mode
4. Save changes and reinstall

---

## Verify Socket Mode Connection

In your API server logs, you should see:
```
"msg":"slack socket mode connected"
```

This confirms Socket Mode is working. If commands still don't work, the issue is:
- Missing `commands` scope
- Slash commands not added to app
- Interactivity not enabled
