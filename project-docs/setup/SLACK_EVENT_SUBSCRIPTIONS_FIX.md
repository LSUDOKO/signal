# Slack Event Subscriptions Configuration

## Problem
Socket Mode is connected but NO events are being received (no messages, mentions, slash commands).

## Root Cause
Event Subscriptions might not be enabled or configured correctly in the Slack app settings.

## Solution: Enable Event Subscriptions

### Step 1: Go to Slack App Settings
1. Open: https://api.slack.com/apps
2. Select your "Signal" app
3. Click **"Event Subscriptions"** in the left sidebar

### Step 2: Enable Events
1. Toggle **"Enable Events"** to **ON**
2. You should see: "✓ Events are enabled for this app"

### Step 3: Subscribe to Bot Events
Under **"Subscribe to bot events"** section, add these events:

**REQUIRED Events:**
- `app_mention` — When someone @mentions your bot
- `message.channels` — Messages in public channels
- `message.im` — Direct messages to your bot
- `app_home_opened` — When user opens App Home tab

**Click "Add Bot User Event"** and search for each one.

### Step 4: Subscribe to Events on Behalf of Users (if needed)
Under **"Subscribe to events on behalf of users"**, you might also need:
- `message` — All messages (optional, might need user token)

### Step 5: Interactivity & Shortcuts
1. Click **"Interactivity & Shortcuts"** in left sidebar
2. Toggle **"Interactivity"** to **ON**
3. For **Request URL**, you can use a placeholder: `https://your-app.com/slack/events`
   - Since you're using Socket Mode, this URL won't be called
   - But Slack still requires it to be set

### Step 6: Slash Commands
1. Click **"Slash Commands"** in left sidebar
2. Verify these commands exist:
   - `/signal`
   - `/translate`
   - `/catchup`
   - `/focus`
   - `/digest`

3. For each command, click "Edit" and verify:
   - **Request URL**: Can be a placeholder like `https://your-app.com/slack/commands`
   - **Short Description**: Must be filled
   - **Usage Hint**: Optional but recommended

### Step 7: Socket Mode
1. Click **"Socket Mode"** in left sidebar
2. Verify: **"Enable Socket Mode"** is **ON**
3. You should see your App-Level Token listed

### Step 8: OAuth & Permissions
1. Click **"OAuth & Permissions"** in left sidebar
2. Verify **Bot Token Scopes** include:
   - `app_mentions:read`
   - `channels:history`
   - `channels:read`
   - `chat:write`
   - `commands`
   - `im:history`
   - `im:read`
   - `im:write`
   - `search:read`
   - `users:read`
   - `users:read.email`

### Step 9: Reinstall App
**CRITICAL: After making ANY changes, you MUST reinstall the app!**

1. Click **"Install App"** in left sidebar
2. Click **"Reinstall to Workspace"**
3. Review permissions
4. Click **"Allow"**
5. Wait for "✓ App installed successfully"

### Step 10: Invite Bot to Channels
1. In Slack workspace, go to a public channel (e.g., `#new-channel`)
2. Type: `/invite @Signal`
3. Press Enter
4. You should see: "Signal was added to #new-channel"

### Step 11: Test Again
1. In the channel, type: `@Signal hi`
2. Try: `/translate test message`
3. Check your server logs for:
   ```
   📥 INCOMING SLACK EVENT | type=events_api
   🔔 EventsAPI event received
   ```

## Verification Checklist

- [ ] Event Subscriptions enabled
- [ ] Bot events subscribed (app_mention, message.channels, message.im, app_home_opened)
- [ ] Interactivity enabled
- [ ] Slash commands created
- [ ] Socket Mode enabled
- [ ] App reinstalled to workspace
- [ ] Bot invited to at least one channel
- [ ] Tested @mention in channel
- [ ] Tested slash command
- [ ] Server logs show incoming events

## If Still Not Working

### Check App Installation
1. In Slack workspace, click workspace name → **Settings & administration** → **Manage apps**
2. Search for "Signal"
3. If not found, the app is NOT installed
4. Go back to https://api.slack.com/apps → Your App → **Install App** → **Install to Workspace**

### Check Channel Membership
1. In channel, type: `/who`
2. Verify Signal bot is listed as a member
3. If not, type: `/invite @Signal`

### Check Socket Mode Connection
Your server logs should show:
```
connected to slack | bot_user=U0BHJQD4G2U | team=slack_hackathon_test
✅ slack socket mode connected
```

If you see connection errors, check:
- `SLACK_APP_TOKEN` starts with `xapp-`
- `SLACK_BOT_TOKEN` starts with `xoxb-`
- Both tokens are from the SAME app in https://api.slack.com/apps

## Common Mistakes

1. **Forgot to reinstall** after changing event subscriptions
2. **Bot not invited** to the channel you're testing in
3. **Interactivity disabled** (required for buttons and modals)
4. **Request URLs empty** (even though Socket Mode doesn't use them, Slack requires them)
5. **Testing in DMs** before enabling `message.im` event subscription

## Expected Log Output (Success)

When you `@Signal hi` in a channel, you should see:

```json
📥 INCOMING SLACK EVENT | type=events_api | has_request=true
🔔 EventsAPI event received
📨 handling app_mention event | user=U0BHJNYV3J4 | channel=C0BGM4GD4GM
```

When you `/translate test`, you should see:

```json
📥 INCOMING SLACK EVENT | type=slash_command | has_request=true
⚡ Slash command received
📝 handling slash command | command=/translate | user=U0BHJNYV3J4
```

## Next Steps

After completing all steps above:
1. Restart your Signal server
2. Test @mention in channel
3. Test slash command
4. Share new server logs

If events are STILL not arriving:
- Try creating a **brand new Slack app** from scratch
- Use `slack create signal --manifest slack-manifest.yml` (Slack CLI)
- This ensures all settings are correct from the start
