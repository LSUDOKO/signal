# Signal — Demo Workspace Setup Guide
## Complete Step-by-Step: Real World Demo for Hackathon Judges

---

## Part 1: Feature Completion Status

| Feature | Status | Required Setup |
|---|---|---|
| Tone Translation (`/translate`) | ✅ LIVE | None |
| ADHD Mode (`/mode adhd`) | ✅ LIVE | None |
| Autism Mode (`/mode autism`) | ✅ LIVE | None |
| Focus Guard (`/focus`) | ✅ LIVE | None |
| RTS Memory (`/catchup`) | ⚠️ Needs scope | Add `search:read` user scope (Step 2 below) |
| Decision Search (`/decisions`) | ✅ LIVE | None |
| Thread Summaries (📋 reaction) | ✅ LIVE | None |
| Smart Notifications | ✅ LIVE | Auto-active during deep work |
| GitHub MCP (`/github`) | ✅ LIVE | Add `GITHUB_TOKEN` + `GITHUB_ORG` to `.env` |
| Calendar MCP (`/focus`) | ✅ LIVE | Add `MCP_CALENDAR_CREDENTIALS_PATH` to `.env` |
| Docs MCP (`/docs`) | ✅ LIVE | Add `NOTION_TOKEN` to `.env` |
| Action Planner (`/plan`) | ✅ LIVE | None |
| Daily Digest (`/digest`) | ✅ LIVE | None |
| What Did I Miss? (`/catchup`) | ✅ LIVE | None |
| AI Workspace Memory | ✅ LIVE | None (auto, Redis-backed) |

---

## Part 2: One-Time Setup Steps

### Step 1: Add New Slack Commands to Your App

You added code for `/mode`, `/decisions`, `/plan`, `/github`, `/docs`.
Now register them in Slack:

1. Go to https://api.slack.com/apps → Signal app
2. Click **"Slash Commands"** → **"Create New Command"** for each:

   | Command | Description | Usage Hint |
   |---|---|---|
   | `/mode` | Set neurotype mode | `adhd \| autism \| anxiety \| ally` |
   | `/decisions` | Find all decisions in a channel | `[#channel] [7\|14\|30]` |
   | `/plan` | AI action planner | `describe your goal` |
   | `/github` | Search GitHub PRs and issues | `open PRs \| issues \| repo name` |
   | `/docs` | Search Notion workspace | `search query` |

3. For **Request URL** on each: put any valid URL (e.g. `https://slack.com`) — Signal uses Socket Mode so this URL is never called.

4. Click **"Save"** after each command.

---

### Step 2: Fix RTS Search (for /catchup to use real Slack data)

1. Go to https://api.slack.com/apps → Signal → **"OAuth & Permissions"**
2. Under **"User Token Scopes"**, click **"Add an OAuth Scope"**
3. Add: `search:read`
4. Also add: `users.profile:write` (for status updates)
5. Scroll down → Click **"Reinstall to Workspace"**
6. After reinstalling, copy the new **User OAuth Token** (`xoxp-...`)
7. Update `.env`:
   ```
   SLACK_USER_TOKEN=xoxp-NEW-TOKEN-HERE
   ```
8. Restart the server

---

### Step 3: GitHub Token (for /github)

1. Go to https://github.com/settings/tokens
2. Click **"Generate new token (classic)"**
3. Select scopes: `repo` (read-only is fine)
4. Copy the token
5. Update `.env`:
   ```
   GITHUB_TOKEN=ghp_your_token_here
   GITHUB_ORG=your-github-username-or-org
   ```
   For your personal repos: `GITHUB_ORG=LSUDOKO`
6. Restart server

---

### Step 4: Notion Token (for /docs)

1. Go to https://www.notion.so/my-integrations
2. Click **"New integration"**
3. Name it **"Signal"**, select your workspace
4. Under **"Capabilities"**: enable **"Read content"**
5. Copy the **"Internal Integration Token"** (starts with `secret_`)
6. Update `.env`:
   ```
   NOTION_TOKEN=secret_your_token_here
   ```
7. **IMPORTANT**: For each Notion page Signal should search:
   - Open the page in Notion
   - Click **"..."** (top right) → **"Add connections"** → select **"Signal"**
8. Restart server

---

### Step 5: Google Calendar (for /focus to block calendar time)

1. Go to https://console.cloud.google.com
2. Create a project (or use existing)
3. Enable **"Google Calendar API"**
4. Go to **"Credentials"** → **"Create Credentials"** → **"Service Account"**
5. Name it "signal-calendar", create
6. Click on the service account → **"Keys"** → **"Add Key"** → **"JSON"**
7. Download the JSON file, save it (e.g. `~/signal-calendar-key.json`)
8. In your Google Calendar: Settings → **"Share with specific people"** → add the service account email
9. Update `.env`:
   ```
   MCP_CALENDAR_CREDENTIALS_PATH=/home/you/signal-calendar-key.json
   ```
10. Restart the MCP server: `go run ./cmd/mcp-server/main.go`

---

## Part 3: Demo Workspace Channels Setup

Create these channels in your Slack workspace for the demo:

### Channels to Create
```
#engineering     - Main engineering discussions
#design          - Design team channel
#general         - Company-wide channel (already exists)
#demo            - Your demo channel (already exists)
#deep-work-test  - For testing focus mode
```

### Invite Signal to Each Channel
In each channel, type:
```
/invite @Signal
```

---

## Part 4: Demo Scenario — "A Day at Work with Signal"

This is the complete demo script for judges. Follow in order.

---

### Scene 1: ADHD Mode Setup (30 seconds)
**Show:** Personalization

**In any channel, type:**
```
/mode adhd
```

**Expected result:**
- ⚡ ADHD Mode Activated card
- Lists what changed (shorter summaries, 30 msg threshold, urgent-only digest)

**What to say:**
> "Alex is an engineer with ADHD. The first thing he does is set his neurotype. Signal immediately adjusts every feature to his cognitive needs."

---

### Scene 2: Social Translator — Passive Aggressive Message (45 seconds)
**Show:** Tone Translation + Autism Mode

**In #engineering, send this message:**
```
Per my last email, we need the deployment done by EOD today. Going forward, let's make sure we're aligned on deadlines.
```

**Expected result:**
- Signal auto-detects the ambiguous phrase
- Sends Signal a DM with tone analysis: Frustrated/Urgent, action: reply with ETA

**Then run:**
```
/translate Per my last email, we need this done today.
```

**What to say:**
> "Jordan is autistic and received this passive-aggressive message. 'Per my last email' signals frustration — but Jordan couldn't tell. Signal decoded it instantly: the sender is frustrated, they want confirmation. Jordan now knows exactly what to do."

---

### Scene 3: Focus Guard + Smart Notifications (1 minute)
**Show:** Deep Work + AI urgency detection

**Type:**
```
/focus 1h
```

**Expected result:**
- 🧘 Deep Work Mode Activated
- Status updated (if Calendar configured: event blocked)
- "Extend 1h" and "End Early" buttons

**Then send a DM TO Signal:**
```
hey can you review this when you get a chance?
```

**Expected result:**
- Auto-reply: "I'm in Deep Work mode. I'll respond at [time]."

**Then send another DM:**
```
URGENT the production server is down
```

**Expected result:**
- Signal detects URGENT keyword
- Sends: "🚨 Urgent message detected" notification

**What to say:**
> "Taylor starts a deep work session. Non-urgent DMs get an auto-reply. But when someone types URGENT, Signal's AI classifier bypasses the block immediately. No missed incidents, no broken focus cycles."

---

### Scene 4: Channel Velocity / Focus Mode (45 seconds)
**Show:** Focus Mode auto-detection

**In #engineering, rapidly send 15+ messages (copy/paste these):**
```
hey team
standup in 5 min
we need to decide on the API changes
should we use REST or GraphQL?
REST is simpler to implement
but GraphQL is more flexible
@sarah what do you think?
I vote REST
agreed, REST it is
also the deploy is broken
jakub is fixing it
ETA 30 minutes
ok thanks
will the Q3 deadline still be met?
yes, we're on track
```

**Expected result:**
- Signal posts a Focus Mode summary card
- ✅ Decision: Use REST API
- Action items listed
- "Get Full Summary" and "Mute 30 min" buttons

**What to say:**
> "While Taylor is in deep work, the engineering channel is exploding. Signal automatically detected the velocity spike and generated a decision tree. When Taylor comes back, they don't need to scroll 50 messages."

---

### Scene 5: What Did I Miss? / Catchup (30 seconds)
**Show:** RTS Memory + AI Summary

**Type:**
```
/catchup what happened in engineering today?
```

**Expected result:**
- "What You Missed" card with AI-organized summary
- Topics, decisions, action items

**What to say:**
> "Morgan was in therapy and missed the discussion. Instead of scrolling for an hour, she types one command. Signal searched across all accessible channels and summarized everything."

---

### Scene 6: Thread Summary Reaction (30 seconds)
**Show:** Thread Summaries

**In #engineering, find a thread with multiple replies.
React to the parent message with the 📋 (memo) emoji.**

**Expected result:**
- Ephemeral message appears ONLY for you:
  - "📋 Thread Summary"
  - What happened: 1 sentence
  - Key points: bullet list
  - Your action: specific next step

**What to say:**
> "React with 📋 on any thread to get an instant summary. Only you see it. Takes 2 seconds instead of 5 minutes of reading."

---

### Scene 7: Decision Search (30 seconds)
**Show:** Decision Search

**Type:**
```
/decisions #engineering 7
```

**Expected result:**
- ✅ Decision Log card
- Lists formal decisions from the last 7 days
- Each decision has context + action items

**What to say:**
> "Every decision Signal has ever seen in this channel, surfaced in seconds. No more 'wait, what did we decide about the API?' conversations."

---

### Scene 8: Action Planner (30 seconds)
**Show:** Action Planner + Memory

**Type:**
```
/plan I need to finish the Q3 report, review Alex's PR, and prepare for tomorrow's design review
```

**Expected result:**
- 📋 Action Plan with 3-4 tasks
- ADHD-aware: energy labels (⚡🔋🚀)
- Time estimates per task
- "Done when" criteria

**What to say:**
> "Alex has ADHD and 3 things due tomorrow. One command turns it into a structured plan with energy labels and completion criteria. No executive function required."

---

### Scene 9: GitHub Integration (45 seconds)
**Show:** GitHub MCP

*(Requires GITHUB_TOKEN set)*

**Type:**
```
/github open PRs
```

**Expected result:**
- 🔀 Open Pull Requests list
- PR titles with authors and labels
- Clickable links

**Then type:**
```
/github issues assigned to me
```

**What to say:**
> "Signal connects to GitHub. Engineers can see their open PRs and assigned issues without leaving Slack."

---

### Scene 10: Docs Search (30 seconds)
**Show:** Docs MCP

*(Requires NOTION_TOKEN set)*

**Type:**
```
/docs onboarding guide
```

**Expected result:**
- 📚 Notion search results
- Page titles with links
- Last edited dates

**What to say:**
> "The team's documentation is in Notion. Signal searches it directly from Slack — no tab switching."

---

### Scene 11: Daily Digest (30 seconds)
**Show:** Daily Digest

**Type:**
```
/digest
```

**Expected result:**
- 📬 On-Demand Digest
- Mentions from today organized by channel
- Footer with preferences tip

**What to say:**
> "At 4 PM, instead of checking every channel, Signal sends a structured digest. Urgent items at the top. FYI at the bottom. Everything batched, nothing missed."

---

## Part 5: Submission Checklist

Before submitting to Devpost:

### Required
- [ ] All slash commands working in sandbox
- [ ] `/translate` shows real tone analysis
- [ ] `/catchup` returns AI summary
- [ ] `/focus` sets Slack status
- [ ] `/mode adhd` changes behavior
- [ ] Thread summary (📋) works
- [ ] Focus Mode auto-triggers on fast channels
- [ ] `slackhack@salesforce.com` invited to workspace
- [ ] `testing@devpost.com` invited to workspace
- [ ] 3-minute demo video recorded (follow Scene 1-11 above)

### Optional but high-impact
- [ ] GitHub token set → `/github` works
- [ ] Notion token set → `/docs` works
- [ ] Calendar credentials set → `/focus` blocks calendar

### Video Script (3 minutes)
```
0:00-0:15  Problem: "Slack overwhelms 15-20% of the workforce — the neurodivergent"
0:15-0:45  Scene 2: Social Translator (passive aggressive message decoded)
0:45-1:15  Scene 3: Focus Guard + Smart Notifications (URGENT bypass)
1:15-1:45  Scene 4: Channel Velocity / Focus Mode auto-trigger
1:45-2:00  Scene 6: Thread Summary reaction
2:00-2:20  Scene 8: Action Planner (ADHD-aware task breakdown)
2:20-2:45  Scenes 9+10: GitHub + Docs MCP (workplace integrations)
2:45-3:00  "Signal is the first neurodivergent accessibility tool in the Slack Marketplace"
```

---

## Part 6: Judging Criteria Mapping

| Criterion | Evidence to Show |
|---|---|
| **Tech Implementation** | Socket Mode, RTS API, MCP integrations, Go+sqlc+asynq architecture |
| **Design** | Block Kit cards, neurotype-adaptive UI, no external dashboard |
| **Potential Impact** | 15-20% workforce, zero Slack Marketplace competitors |
| **Quality of Idea** | Social Translator (novel), neurotype modes (novel), AI memory (novel) |

---

## Part 7: Architecture Diagram Description

For your Devpost submission diagram, show:

```
User (Slack)
    ↓ Socket Mode WebSocket
[Signal API — Go]
    ├── /translate → Groq AI (tone analysis)
    ├── /catchup → RTS API + Groq AI
    ├── /focus → Redis + MCP Server → Google Calendar
    ├── /github → GitHub REST API
    ├── /docs → Notion API
    ├── /plan → Groq AI + Redis (memory)
    └── 📋 reaction → Groq AI (thread summary)
    
[Data Layer]
    ├── PostgreSQL — users, preferences, digests, summaries
    └── Redis — velocity counters, deep work state, AI memory

[MCP Server — Go]
    └── Google Calendar API
```

---

## Commands Quick Reference

| Command | Example | What it does |
|---|---|---|
| `/signal` | `/signal` | Help menu |
| `/mode` | `/mode adhd` | Set neurotype |
| `/translate` | `/translate per my last email...` | Decode ambiguous message |
| `/catchup` | `/catchup what happened today` | AI summary of missed messages |
| `/focus` | `/focus 2h` | Start deep work session |
| `/digest` | `/digest` | Instant mentions digest |
| `/decisions` | `/decisions #engineering 7` | Find decisions in channel |
| `/plan` | `/plan finish report by Friday` | AI action plan |
| `/github` | `/github open PRs` | List GitHub PRs/issues |
| `/docs` | `/docs onboarding guide` | Search Notion |
| React `📋` | On any thread | Instant thread summary |
