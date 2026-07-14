# Apply Slack Manifest to Fix All Issues

Your `slack-manifest.yml` is perfect and has everything needed! But if you created the app manually, your actual Slack app might not match this file.

## **Solution: Apply the Manifest**

### **Step 1: Open Your Slack App**
1. Go to: https://api.slack.com/apps
2. Click on your **Signal** app

### **Step 2: Apply the Manifest**
1. In the left sidebar (under "Settings"), click **"App Manifest"**
2. You'll see a YAML editor with your current manifest
3. **Copy the ENTIRE contents** of `/home/arpit/Desktop/hackathon_projects/signal/slack-manifest.yml`
4. **Paste it** into the editor (replacing everything)
5. Click **"Save Changes"**
6. Slack will show you a diff of what changed
7. Click **"Confirm"**

### **Step 3: Reinstall the App**
1. In the left sidebar, click **"Install App"**
2. Click **"Reinstall to Workspace"**
3. Review permissions and click **"Allow"**

### **Step 4: Restart Your API Server**

```bash
# Kill old process
ps aux | grep "go run.*cmd/api" | grep -v grep | awk '{print $2}' | xargs kill -9

# Start fresh
cd /home/arpit/Desktop/hackathon_projects/signal/api
go run ./cmd/api/main.go
```

### **Step 5: Test in Slack**

1. **Test slash command**: `/signal`
   - Expected: Help menu in DM

2. **Test DM**: Send "hi" to Signal
   - Expected: Help menu appears

3. **Test translate**: `/translate hello world`
   - Expected: "⏳ Working..." then translation in DM

---

## **Alternative: Verify Manual Configuration**

If you prefer not to use the manifest, verify these settings manually:

### **1. Slash Commands** (must have all 5):
- `/signal`
- `/translate`
- `/catchup`
- `/focus`
- `/digest`

### **2. OAuth Scopes** (Bot Token Scopes):
- `commands` ← **CRITICAL!**
- `chat:write`
- `im:write`
- `app_mentions:read`
- `message.im`
- `channels:history`
- `search:read`

### **3. Event Subscriptions** (Bot Events):
- `message.im` ← **CRITICAL for DMs!**
- `app_mention`
- `message.channels`

### **4. Interactivity & Shortcuts**:
- Interactivity: **ON**
- Socket Mode: **Enabled**

---

## **Why This Fixes Everything**

Your manifest already includes:
- ✅ All 5 slash commands
- ✅ `commands` scope (required for slash commands)
- ✅ `message.im` event (required for DMs)
- ✅ `app_mention` event (required for @Signal)
- ✅ Socket Mode enabled
- ✅ Interactivity enabled

Applying it ensures your Slack app configuration matches your code!
