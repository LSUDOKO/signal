# Signal — Neurodivergent Slack Agent
## Technical Specification & AI Agent Completion Guide
### Slack Agent Builder Challenge | Devpost | July 2026

---

## 1. Executive Summary

**Signal** is a Slack agent that makes workplace communication accessible for neurodivergent professionals (ADHD, autism, anxiety disorders). It transforms overwhelming, ambiguous, and interruptive Slack experiences into calm, structured, and comprehensible interactions.

**Hackathon Track:** Slack Agent for Good  
**Required Technologies:** Uses all three (Slack AI, MCP Server Integration, Real-Time Search API)  
**Prize Target:** First Place Slack Agent for Good ($8,000) + Best UX ($2,000) + Most Innovative ($2,000)  
**Deadline:** July 14, 2026 @ 5:30am GMT+5:30 (6 days remaining)  

---

## 2. Hackathon Alignment Matrix

| Judging Criteria | How Signal Satisfies It | Evidence/Demo Moment |
|---|---|---|
| **Technological Implementation** | Uses Slack AI (tone analysis, summarization), MCP (calendar/focus integration), RTS API (semantic catchup). Go backend with clean architecture. | Architecture diagram shows 3 techs. Code repo has `internal/ai/`, `internal/mcp/`, `internal/rts/`. |
| **Design** | Entire UX lives inside Slack (Block Kit, App Home, DMs). No external dashboard needed. Next.js landing page for OAuth. | Demo shows before/after: chaotic channel → clean decision tree. |
| **Potential Impact** | 15-20% of workforce is neurodivergent. Addresses real DEI gap. No competitor exists in Slack Marketplace. | Stats in submission text. Emotional user story in video. |
| **Quality of Idea** | Zero neurodivergent accessibility tools in Slack Marketplace. First-mover. Combines 3 APIs into coherent narrative. | Market research screenshot. Unique value proposition. |

---

## 3. User Personas & Pain Points

### Persona 1: Alex (ADHD, Software Engineer)
- **Pain:** Returns from lunch → 200 messages in #engineering. Spends 20 minutes scrolling. Misses a decision made at 12:47.
- **Signal Solution:** Focus Mode detects velocity spike → offers AI decision tree summarizing the last 30 minutes.

### Persona 2: Jordan (Autistic, Product Manager)
- **Pain:** Receives message: *"Per my last email..."* Spends 3 hours ruminating whether sender is angry. Avoids responding.
- **Signal Solution:** Social Translator DMs a plain-English breakdown of tone, intent, and required action.

### Persona 3: Taylor (ADHD, Designer)
- **Pain:** 12 @mentions during deep-work session. Each notification breaks 23-minute focus cycle. Turning off Slack means missing urgent requests.
- **Signal Solution:** Quiet Hours Digest batches non-urgent mentions into a structured 4 PM digest. MCP integration with calendar auto-blocks focus time.

### Persona 4: Morgan (Autistic, Data Analyst)
- **Pain:** Missed 3-hour discussion while in therapy appointment. Needs to know what was decided but doesn't know search keywords.
- **Signal Solution:** Catch-Up uses RTS semantic search to answer: *"What did we decide about the Q3 budget?"* — finds messages even without keyword "budget".

---

## 4. Feature Specifications (Detailed)

### 4.1 Focus Mode

**Trigger:** Channel receives ≥50 messages in 10 minutes (configurable per user).  
**Detection:** Redis counter with 10-minute TTL per channel.  
**Action:** Posts an ephemeral/app message to the channel with:
1. Warning: *"This channel is moving fast (50+ messages in 10 min)."*
2. AI-generated decision tree from last 30 messages.
3. Two buttons: **"Get Focus Summary"** | **"Mute for 30 min"**

**AI Prompt (focus_summary.tmpl):**
```
Extract a decision tree from this Slack channel history.
Only include: decisions made, action items with owners, and deadlines.
Ignore small talk, greetings, emojis, and reactions.

Format strictly as:
✅ [Decision made]
   ↳ [Action item] — Owner: @Name — Due: [Date or "None"]
   ↳ [Action item] — Owner: @Name — Due: [Date or "None"]

If no decisions were made, say: "No decisions found. Discussion was exploratory."

History:
{{.Messages}}
```

**Technical Flow:**
1. `OnMessage` event → Redis INCR `channel:velocity:{channel_id}`
2. If count == 50 → fetch last 50 messages via `conversations.history`
3. Send messages to OpenAI GPT-4o-mini
4. Format response as Block Kit message with buttons
5. Post to channel (public, so all users benefit)

**Edge Cases:**
- If AI returns empty summary → show "No decisions found" message
- If channel is DM → skip (DMs don't need focus mode)
- If user has disabled Focus Mode in preferences → skip

---

### 4.2 Social Translator

**Trigger:** Message contains ambiguous/passive-aggressive phrase OR user manually DMs Signal with `/translate [message]`  
**Detection Heuristic:** Regex scan for phrases:
- `per my last`, `as I mentioned`, `just following up`, `let's take this offline`, `moving forward`, `with all due respect`, `friendly reminder`, `circle back`, `touch base`, `loop in`, `per our conversation`

**Action:** DM the user who was mentioned (or the user who requested translation) with:
1. Original message (quoted)
2. **Tone:** [neutral / frustrated / urgent / confused / supportive]
3. **Intent:** [what the sender actually wants]
4. **Action:** [what the recipient should do]
5. **Note:** [social subtext explained in plain language]

**AI Prompt (tone_analyzer.tmpl):**
```
You are a social translator for neurodivergent adults in a workplace Slack environment.
Your job is to decode ambiguous or emotionally loaded messages into plain, direct, actionable language.

Analyze this message and provide:

- Tone: The emotional tone (neutral, frustrated, urgent, confused, supportive, passive-aggressive, etc.)
- Intent: What the sender actually wants or needs
- Action: What the recipient should do in response
- Note: Any social subtext, hidden meaning, or workplace politics explained plainly and non-judgmentally

Be direct. Do not soften. Do not use corporate euphemisms. Be kind but literal.

Message: "{{.Message}}"
```

**Privacy:** Translation is ALWAYS sent via DM. Never posted publicly. Never logged permanently (transient only).

**Technical Flow:**
1. `OnMessage` event → regex scan
2. If match → extract mentioned users via `message.text` parsing
3. For each mentioned user with Signal enabled → Open API call
4. Format as Block Kit section + action block ("Got it" button for acknowledgment)
5. DM via `chat.postMessage` to user's DM channel

---

### 4.3 Catch-Up (Semantic Search)

**Trigger:** User DMs Signal: `/catchup What did I miss about {topic}?` or clicks "Catch-Up" in App Home.  
**Technology:** Slack Real-Time Search API (`search.messages`).  
**Query Builder:** Converts natural language into Slack search query:
- User asks: *"What did we decide about the Q3 budget?"*
- Built query: `from:@user OR to:@user Q3 budget decision after:2026-07-08`
- Also searches user's accessible channels (not DMs they aren't in)

**Action:**
1. Execute RTS search
2. Take top 10 results
3. Send to AI for summarization
4. Return structured digest with message links

**AI Prompt (catchup.tmpl):**
```
Summarize these Slack messages into a "What You Missed" digest.
Organize by topic. Highlight decisions, action items, and anything requiring the user's input.

For each topic:
## [Topic Name]
- **Decision:** [What was decided]
- **Context:** [2-sentence background]
- **Your Action:** [What you need to do, or "None"]
- **Link:** [Jump to message]

Messages:
{{.Messages}}
```

**Technical Flow:**
1. Parse DM command via `message.im` event
2. Extract natural language query
3. Build Slack search query with user ID + date filter (last 7 days default)
4. Call `search.messages` with `count: 20`, `sort: timestamp`, `sort_dir: desc`
5. Extract message texts + permalinks
6. Send to OpenAI
7. Post summary back to DM with clickable message links

**Edge Cases:**
- No results → "I couldn't find anything about that in your accessible channels. Try rephrasing?"
- Results in private channels user left → filter out (respect permissions)

---

### 4.4 Quiet Hours Digest

**Trigger:** User sets "Digest Mode" in App Home preferences.  
**Mechanism:** Asynq scheduled job runs every hour. Checks Redis for users with `digest_mode: true` and `digest_time` matching current hour.  
**Content:** Batches all @mentions, thread replies, and DMs received since last digest.

**Digest Format (Block Kit):**
```
📬 Your 4:00 PM Digest

🔴 Urgent (needs response today)
• @john: "Need the report by 5 PM" → [Reply]

🟡 Action Required (this week)
• @sarah: "Review mockups when you can" → [View]

🟢 FYI (no action needed)
• @team: "Lunch tomorrow at 12" → [View]

💬 Thread Replies
• #design: 3 replies to your message → [Jump]

[Open Slack] [Update Preferences]
```

**Technical Flow:**
1. Asynq cron job: `0 * * * *` (every hour)
2. Query PostgreSQL for users where `digest_enabled = true AND digest_hour = CURRENT_HOUR`
3. For each user: fetch unread mentions from Slack (or track in our DB)
4. Categorize by urgency heuristic (contains "urgent", "asap", "deadline", "today" = red)
5. Format as Block Kit
6. DM user
7. Mark as "digested" in Redis (TTL 24h)

**MCP Integration:** When digest is sent, check user's calendar via MCP. If they have a meeting in 15 minutes, prepend: *"You have a meeting in 15 minutes — these can wait."*

---

### 4.5 Deep Work Protector (MCP Integration)

**Trigger:** User toggles "Deep Work Mode" in App Home OR MCP detects calendar event titled "Focus", "Deep Work", "Coding", etc.  
**MCP Server Tools:**
1. `block_focus_time(user_id, duration_minutes)` → Creates calendar event
2. `get_user_status(user_id)` → Returns "in_meeting", "focus_time", "available"
3. `set_slack_status(user_id, status_text, emoji, expiration)` → Updates Slack status

**Action Flow:**
1. User clicks "Start Deep Work (2 hours)" in App Home
2. Signal calls MCP tool `block_focus_time(user_id, 120)`
3. MCP server (Go) connects to Google Calendar API (or mock for hackathon)
4. Signal sets Slack status: 🧘 In Deep Work — back at 3:00 PM
5. Signal pauses non-urgent digests
6. Signal auto-responds to DMs: *"I'm in deep work mode. I'll respond at 3:00 PM. Urgent? Type URGENT to bypass."*

**MCP Server Schema:**
```json
{
  "name": "block_focus_time",
  "description": "Blocks focus time on the user's calendar and sets Slack status",
  "inputSchema": {
    "type": "object",
    "properties": {
      "user_id": { "type": "string", "description": "Slack user ID" },
      "duration_minutes": { "type": "number", "description": "Focus block duration" },
      "title": { "type": "string", "default": "Deep Work" }
    },
    "required": ["user_id", "duration_minutes"]
  }
}
```

---

## 5. Technical Architecture

### 5.1 Monorepo Structure
```
signal/
├── api/                    ← Go backend (all Slack logic)
│   ├── cmd/api/            ← HTTP server + Socket Mode handler
│   ├── cmd/worker/         ← Asynq background processor
│   ├── cmd/mcp-server/     ← MCP server (calendar/tools)
│   ├── internal/
│   │   ├── domain/         ← Business entities & interfaces (no deps)
│   │   ├── config/         ← Viper/cleanenv config loader
│   │   ├── slack/          ← Slack SDK wrapper, Socket Mode, event routing
│   │   ├── ai/             ← OpenAI client + prompt templates
│   │   ├── rts/            ← Real-Time Search API client
│   │   ├── mcp/            ← MCP host client (connects to MCP servers)
│   │   ├── features/       ← focusmode.go, translator.go, catchup.go, digest.go, deepwork.go
│   │   ├── store/          ← Postgres (sqlc) + Redis repositories
│   │   ├── httpapi/        ← OAuth handlers, health checks
│   │   └── observability/  ← slog, Prometheus metrics
│   ├── db/migrations/      ← golang-migrate SQL files
│   ├── prompts/            ← *.tmpl files for AI prompts
│   └── go.mod
├── frontend/               ← Next.js 14 App Router (landing + OAuth + App Home)
│   ├── src/app/            ← Routes: /, /oauth/callback, /app-home
│   ├── src/components/     ← shadcn/ui components
│   └── package.json
├── docs/                   ← Mintlify documentation
│   ├── mint.json
│   └── *.mdx
├── docker-compose.yml
├── Makefile
└── .github/workflows/      ← CI/CD
```

### 5.2 Data Flow Diagram
```
Slack Events (Socket Mode WebSocket)
    ↓
[api/cmd/api] → Route to Feature Controller
    ↓
[internal/features/] → Business Logic
    ↓
[internal/ai/] → OpenAI GPT-4o-mini (tone/summary/catchup)
    ↓
[internal/rts/] → Slack Search API (semantic catchup)
    ↓
[internal/mcp/] → MCP Client → [cmd/mcp-server] → Calendar API
    ↓
[internal/store/] → PostgreSQL (users, prefs, channels) + Redis (cache, velocity, sessions)
    ↓
Slack Web API (post messages, DMs, status updates)
```

### 5.3 Technology Stack

| Layer | Technology | Purpose |
|---|---|---|
| Language | Go 1.23 | Backend, MCP server, worker |
| HTTP Router | chi/v5 | REST API, OAuth callbacks, health checks |
| Slack SDK | slack-go/slack | Web API + Socket Mode |
| Database | PostgreSQL 16 | Users, channels, preferences, message metadata |
| DB Access | sqlc + pgx | Type-safe SQL code generation |
| Migrations | golang-migrate | Versioned schema changes |
| Cache/Queue | Redis 7 | Velocity counters, sessions, Asynq job queue |
| Background Jobs | asynq | Digest delivery, focus-mode batching |
| AI | OpenAI GPT-4o-mini | Tone analysis, summarization, semantic search results |
| MCP | mcp-go | Host + server implementation |
| Config | cleanenv | Typed environment variable loading |
| Logging | log/slog (stdlib) | Structured JSON logging |
| Metrics | Prometheus | `/metrics` endpoint for Grafana |
| Frontend | Next.js 14 + TypeScript | Landing page, OAuth, App Home preferences |
| UI | shadcn/ui + Tailwind | Accessible, dark-mode ready components |
| State | TanStack Query + Zustand | Server + client state |
| Docs | Mintlify | Hosted documentation with search |
| CI/CD | GitHub Actions | Lint, test, build on every PR |
| Deploy | Docker + Fly.io/Railway | Containerized deployment |

---

## 6. Database Schema

### 6.1 PostgreSQL (via sqlc)

```sql
-- db/migrations/0001_init.sql

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slack_user_id TEXT NOT NULL UNIQUE,
    slack_team_id TEXT NOT NULL,
    email TEXT,
    display_name TEXT,
    neurotype TEXT CHECK (neurotype IN ('adhd', 'autism', 'anxiety', 'unspecified', 'ally')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    focus_mode_enabled BOOLEAN DEFAULT true,
    focus_threshold INTEGER DEFAULT 50, -- messages per 10 min
    translator_enabled BOOLEAN DEFAULT true,
    digest_enabled BOOLEAN DEFAULT false,
    digest_hour INTEGER CHECK (digest_hour BETWEEN 0 AND 23) DEFAULT 16,
    deep_work_auto_detect BOOLEAN DEFAULT false,
    quiet_hours_start TIME DEFAULT '22:00',
    quiet_hours_end TIME DEFAULT '08:00',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slack_channel_id TEXT NOT NULL UNIQUE,
    slack_team_id TEXT NOT NULL,
    name TEXT,
    is_dm BOOLEAN DEFAULT false,
    focus_mode_enabled BOOLEAN DEFAULT true, -- can be disabled per channel
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE channel_subscriptions (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    muted BOOLEAN DEFAULT false,
    PRIMARY KEY (user_id, channel_id)
);

CREATE TABLE digests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    mention_count INTEGER DEFAULT 0,
    thread_reply_count INTEGER DEFAULT 0,
    content JSONB, -- structured digest content
    status TEXT CHECK (status IN ('pending', 'sent', 'read')) DEFAULT 'pending'
);

CREATE TABLE focus_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    triggered_at TIMESTAMPTZ DEFAULT NOW(),
    message_count INTEGER,
    summary_text TEXT,
    ai_model TEXT DEFAULT 'gpt-4o-mini',
    raw_messages JSONB -- store for debugging/improvement
);

CREATE TABLE translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    original_message_ts TEXT, -- Slack message timestamp
    original_channel_id TEXT,
    original_text TEXT,
    translation_text TEXT,
    tone TEXT,
    intent TEXT,
    action TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_slack_id ON users(slack_user_id);
CREATE INDEX idx_channels_slack_id ON channels(slack_channel_id);
CREATE INDEX idx_digests_user_sent ON digests(user_id, sent_at);
CREATE INDEX idx_translations_user ON translations(user_id, created_at);
```

### 6.2 Redis Key Schema

```
channel:velocity:{channel_id}           → String (counter, TTL 10 min)
channel:velocity:{channel_id}:offered  → String (flag, TTL 30 min) -- prevent spam
user:session:{slack_user_id}             → Hash (access_token, team_id, TTL 24h)
user:digest:last:{user_id}               → String (timestamp)
user:deepwork:{user_id}                  → Hash (start_time, duration, status)
user:status:{user_id}                    → String (online, away, deep_work)
rate_limit:ai:{user_id}                  → String (counter, TTL 1 min) -- 20 requests/min
```

---

## 7. API Specifications

### 7.1 Internal REST API (Go + Chi)

```yaml
openapi: 3.0.0
info:
  title: Signal API
  version: 1.0.0
paths:
  /health:
    get:
      summary: Health check
      responses:
        200:
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  status: { type: string, example: "ok" }
                  version: { type: string }

  /oauth/slack:
    get:
      summary: Slack OAuth callback
      parameters:
        - name: code
          in: query
          required: true
          schema: { type: string }
      responses:
        302:
          description: Redirect to frontend with token

  /api/v1/users/{user_id}/preferences:
    get:
      summary: Get user preferences
      responses:
        200:
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserPreferences'
    put:
      summary: Update preferences
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UserPreferences'
      responses:
        200:
          description: Updated

  /api/v1/catchup:
    post:
      summary: Semantic catchup query
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                user_id: { type: string }
                query: { type: string }
                days_back: { type: integer, default: 7 }
      responses:
        200:
          content:
            application/json:
              schema:
                type: object
                properties:
                  summary: { type: string }
                  sources: { type: array, items: { type: string } }
                  message_count: { type: integer }

  /metrics:
    get:
      summary: Prometheus metrics
      responses:
        200:
          content:
            text/plain: {}

components:
  schemas:
    UserPreferences:
      type: object
      properties:
        focus_mode_enabled: { type: boolean }
        focus_threshold: { type: integer }
        translator_enabled: { type: boolean }
        digest_enabled: { type: boolean }
        digest_hour: { type: integer }
        deep_work_auto_detect: { type: boolean }
        quiet_hours_start: { type: string }
        quiet_hours_end: { type: string }
```

### 7.2 Slack Socket Mode Events Handled

| Event Type | Feature | Handler |
|---|---|---|
| `message` (channel) | Focus Mode | `features/focusmode.go` |
| `message` (channel) | Translator | `features/translator.go` |
| `message` (im) | Catch-Up | `features/catchup.go` |
| `app_home_opened` | App Home preferences | `slack/events.go` |
| `block_actions` | Button clicks (summary, mute, etc.) | `slack/events.go` |
| `member_joined_channel` | Auto-enable for new channels | `slack/events.go` |
| `reaction_added` | (Future: sentiment tracking) | — |

### 7.3 Slack Commands

| Command | Feature | Description |
|---|---|---|
| `/signal` | General | Opens App Home preferences |
| `/translate [message]` | Translator | Manual translation request |
| `/catchup [query]` | Catch-Up | Semantic search query |
| `/focus [duration]` | Deep Work | Start focus mode (e.g., `/focus 2h`) |
| `/digest` | Digest | Force-send digest now |

---

## 8. Slack Block Kit UI Specifications

### 8.1 Focus Mode Message (Posted to Channel)

```json
{
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "🧘 *This channel is moving fast*
50+ messages in the last 10 minutes."
      }
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "✅ *Decisions & Actions*
• Moved Q3 budget to cloud servers
  ↳ Cost breakdown due Friday — Owner: @sarah
• Approved new design system
  ↳ No action needed"
      }
    },
    {
      "type": "actions",
      "elements": [
        {
          "type": "button",
          "text": { "type": "plain_text", "text": "Get Full Summary", "emoji": true },
          "style": "primary",
          "value": "get_summary",
          "action_id": "focus_get_summary"
        },
        {
          "type": "button",
          "text": { "type": "plain_text", "text": "Mute 30 min", "emoji": true },
          "value": "mute_30",
          "action_id": "focus_mute_30"
        },
        {
          "type": "button",
          "text": { "type": "plain_text", "text": "Open in Thread", "emoji": true },
          "value": "open_thread",
          "action_id": "focus_open_thread"
        }
      ]
    }
  ]
}
```

### 8.2 Social Translator DM

```json
{
  "blocks": [
    {
      "type": "header",
      "text": { "type": "plain_text", "text": "🔍 Signal Translation", "emoji": true }
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*Original message in <#CHANNEL_ID>:*
> Per my last email, we need this by EOD."
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "section",
      "fields": [
        { "type": "mrkdwn", "text": "*Tone:*
Frustrated / Urgent" },
        { "type": "mrkdwn", "text": "*Intent:*
They want you to complete a task you were already told about" }
      ]
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*Action:* Reply confirming completion time or ask for clarification if you don't know which task.

*Note:* "Per my last email" often means the sender feels ignored. A quick acknowledgment defuses tension."
      }
    },
    {
      "type": "actions",
      "elements": [
        {
          "type": "button",
          "text": { "type": "plain_text", "text": "Got it ✓", "emoji": true },
          "style": "primary",
          "value": "ack",
          "action_id": "translator_ack"
        }
      ]
    }
  ]
}
```

### 8.3 App Home (User Preferences)

```json
{
  "type": "home",
  "blocks": [
    {
      "type": "header",
      "text": { "type": "plain_text", "text": "⚙️ Signal Preferences", "emoji": true }
    },
    {
      "type": "section",
      "text": { "type": "mrkdwn", "text": "*Neurotype:* ADHD
*Focus Mode:* ✅ On (threshold: 50 msg/10min)
*Translator:* ✅ On
*Digest:* ⏰ Daily at 4:00 PM
*Deep Work:* Auto-detect from calendar" }
    },
    {
      "type": "actions",
      "elements": [
        { "type": "button", "text": { "type": "plain_text", "text": "Edit Preferences" }, "value": "edit_prefs", "action_id": "open_prefs_modal" }
      ]
    }
  ]
}
```

---

## 9. MCP Server Specification

### 9.1 Server Details
- **Name:** `signal-calendar`
- **Version:** `1.0.0`
- **Transport:** SSE (Server-Sent Events) on port 3001
- **Binary:** `cmd/mcp-server/main.go`

### 9.2 Tools Exposed

#### Tool 1: `block_focus_time`
```json
{
  "name": "block_focus_time",
  "description": "Blocks focus time on the user's calendar and updates Slack status",
  "inputSchema": {
    "type": "object",
    "properties": {
      "user_id": { "type": "string", "description": "Slack user ID" },
      "duration_minutes": { "type": "number", "description": "Duration in minutes" },
      "title": { "type": "string", "default": "Deep Work" },
      "calendar_id": { "type": "string", "default": "primary" }
    },
    "required": ["user_id", "duration_minutes"]
  }
}
```
**Returns:** `{"blocked": true, "event_id": "abc123", "end_time": "2026-07-11T18:00:00Z"}`

#### Tool 2: `get_user_status`
```json
{
  "name": "get_user_status",
  "description": "Checks if user is in a meeting, focus time, or available",
  "inputSchema": {
    "type": "object",
    "properties": {
      "user_id": { "type": "string" },
      "check_next_minutes": { "type": "number", "default": 60 }
    },
    "required": ["user_id"]
  }
}
```
**Returns:** `{"status": "in_meeting", "event_title": "Sprint Planning", "ends_at": "2026-07-11T17:00:00Z"}`

#### Tool 3: `set_slack_status`
```json
{
  "name": "set_slack_status",
  "description": "Sets the user's Slack status text and emoji",
  "inputSchema": {
    "type": "object",
    "properties": {
      "user_id": { "type": "string" },
      "status_text": { "type": "string", "maxLength": 100 },
      "status_emoji": { "type": "string", "example": ":brain:" },
      "expiration_minutes": { "type": "number", "default": 120 }
    },
    "required": ["user_id", "status_text", "status_emoji"]
  }
}
```

### 9.3 MCP Host Client (in API)
The API service (`internal/mcp/client.go`) connects to the MCP server via SSE and calls these tools when:
- User starts deep work mode
- Digest is about to send but user is in a meeting
- Auto-detect calendar events for focus mode

---

## 10. RTS (Real-Time Search) Integration

### 10.1 API Endpoint
```go
func (c *RTSSearcher) SemanticCatchup(ctx context.Context, userID, query string, daysBack int) (*SearchResult, error) {
    // Build Slack search query
    dateFilter := time.Now().AddDate(0, 0, -daysBack).Format("2006-01-02")
    searchQuery := fmt.Sprintf(
        "from:@%s OR to:@%s %s after:%s",
        userID, userID, query, dateFilter,
    )

    params := slack.NewSearchMessagesParameters()
    params.Count = 20
    params.Sort = "timestamp"
    params.SortDir = "desc"
    params.Highlight = true

    results, err := c.slackClient.SearchMessagesContext(ctx, searchQuery, params)
    // ... process results
}
```

### 10.2 Query Examples
| User Question | Generated Slack Query |
|---|---|
| "What did we decide about the budget?" | `from:@U123 OR to:@U123 budget decision after:2026-07-04` |
| "Did anyone mention my design?" | `from:@U123 OR to:@U123 design mockup after:2026-07-04` |
| "What did I miss in engineering?" | `from:@U123 OR to:@U123 in:#engineering after:2026-07-04` |

---

## 11. Frontend Specification (Next.js)

### 11.1 Routes
| Route | Purpose | Auth |
|---|---|---|
| `/` | Landing page (value prop, features, "Add to Slack" button) | Public |
| `/oauth/callback` | Slack OAuth redirect handler | Public (temp code) |
| `/app-home` | User preferences dashboard (neurotype, toggles, digest time) | Slack OAuth |
| `/privacy` | Privacy policy (required by Slack Marketplace) | Public |
| `/support` | Contact/support info | Public |

### 11.2 Landing Page Sections
1. **Hero:** "Slack is overwhelming. Signal makes it calm." + "Add to Slack" button
2. **Features Grid:** 4 cards (Focus Mode, Translator, Catch-Up, Deep Work)
3. **User Story:** Quote from "Alex" (ADHD engineer) about information overload
4. **How It Works:** 3-step diagram (Install → Configure → Relax)
5. **Security:** "Your data stays in Slack. Translations are never logged."
6. **Footer:** Links to Docs, GitHub, Privacy, Contact

### 11.3 App Home Dashboard (`/app-home`)
- **Neurotype Selector:** Radio cards (ADHD, Autism, Anxiety, Unspecified, Ally)
- **Toggle Switches:** Focus Mode, Translator, Digest, Deep Work Auto-Detect
- **Sliders:** Focus threshold (10-100 messages), Digest hour (0-23)
- **Time Pickers:** Quiet hours start/end
- **Save Button:** PUT to `/api/v1/users/{id}/preferences`
- **Toast:** "Preferences saved" via Sonner

### 11.4 Component Requirements
- All forms use `react-hook-form` + `zod` validation
- All API calls use `TanStack Query` with loading states
- Dark mode support via `next-themes`
- Accessible: ARIA labels, keyboard navigation, focus indicators
- Mobile-responsive (Tailwind breakpoints)

---

## 12. Testing Strategy

### 12.1 Go Tests
```bash
# Unit tests (table-driven)
go test ./internal/features/... -v

# Integration tests (with testcontainers)
go test ./test/integration/... -v

# Race detection
go test -race ./...

# Coverage target: >70%
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Cases:**
- Focus Mode: Channel gets 49 messages → no trigger. 50th message → trigger.
- Translator: Message with "per my last" → DM sent. Message without → no DM.
- Catch-Up: Query "budget" → search results contain "Q3 spending" (semantic match).
- MCP: `block_focus_time` returns valid event ID.

### 12.2 Frontend Tests
```bash
npm run test        # Vitest unit tests
npm run test:e2e    # Playwright E2E (optional, skip if time-constrained)
```

### 12.3 Manual Testing Checklist
- [ ] Install app to Slack sandbox
- [ ] Send 50 messages in test channel → Focus Mode triggers
- [ ] Post "Per my last email" → Translator DM received
- [ ] DM `/catchup What about the website?` → Relevant results returned
- [ ] Set digest time → Digest arrives at correct hour
- [ ] Click "Start Deep Work" → Slack status updates
- [ ] App Home renders preferences correctly
- [ ] OAuth flow completes without error

---

## 13. 6-Day Implementation Roadmap

### Day 1 — Foundation (July 9)
**Goal:** Infrastructure running. One Slack command responds.
**Tasks:**
1. Run `make install` and `make dev` (Postgres + Redis up)
2. Create Slack app at https://api.slack.com/apps
3. Enable Socket Mode, get App-Level Token (`xapp-`)
4. Add Bot Token Scopes: `chat:write`, `im:write`, `users:read`, `channels:history`, `groups:history`, `search:read`, `users:read.email`, `reactions:write`
5. Write `cmd/api/main.go` with basic Socket Mode event loop
6. Implement `/signal` slash command → responds "Hello from Signal"
7. Commit: `feat: slack socket mode connection and basic command`

**Deliverable:** `git log` shows 3+ commits. `docker-compose ps` shows all services healthy.

---

### Day 2 — Core Features (July 10)
**Goal:** Focus Mode + Translator working in sandbox.
**Tasks:**
1. Implement Redis velocity counter in `internal/features/focusmode.go`
2. Implement `conversations.history` fetch + AI summary
3. Implement regex-based trigger + DM sender in `internal/features/translator.go`
4. Write prompt templates (`prompts/focus_summary.tmpl`, `prompts/tone_analyzer.tmpl`)
5. Test in Slack sandbox: spam 50 messages, verify Focus Mode triggers. Post ambiguous message, verify DM.
6. Commit: `feat: focus mode velocity detection and AI summarization`
7. Commit: `feat: social translator with tone analysis`

**Deliverable:** Demo-able in Slack. Screenshot of Focus Mode message and Translator DM.

---

### Day 3 — Search + MCP (July 11)
**Goal:** Catch-Up semantic search + MCP server skeleton.
**Tasks:**
1. Implement `internal/rts/search.go` with `search.messages` API
2. Implement `/catchup` command handler
3. Write `cmd/mcp-server/main.go` with 3 tools (block_focus_time, get_user_status, set_slack_status)
4. Implement `internal/mcp/client.go` host connection
5. Implement Deep Work button in App Home → calls MCP → sets Slack status
6. Test: `/catchup budget` returns relevant messages. Deep Work button updates status.
7. Commit: `feat: RTS semantic search for catchup feature`
8. Commit: `feat: MCP server with calendar and status tools`

**Deliverable:** All 3 required technologies (AI, RTS, MCP) are functional.

---

### Day 4 — Frontend + Polish (July 12)
**Goal:** Landing page + App Home preferences + OAuth.
**Tasks:**
1. Build Next.js landing page (`/`) with "Add to Slack" OAuth button
2. Implement `/oauth/callback` handler (exchanges code for token, stores user)
3. Build `/app-home` preferences dashboard with shadcn/ui
4. Connect frontend to Go API (`/api/v1/users/{id}/preferences`)
5. Add dark mode, mobile responsiveness
6. Write `docs/introduction.mdx` and `docs/quickstart.mdx`
7. Commit: `feat: next.js landing page and oauth flow`
8. Commit: `feat: app home preferences dashboard`

**Deliverable:** Vercel preview URL works. OAuth completes. Preferences save.

---

### Day 5 — Video + Docs + Submission Prep (July 13)
**Goal:** 3-minute demo video + complete submission text.
**Tasks:**
1. Record 3-minute demo video (screen capture + voiceover):
   - 0:00-0:30: Problem statement (Alex with ADHD, 200 messages)
   - 0:30-1:00: Focus Mode demo (spam channel → AI summary)
   - 1:00-1:30: Translator demo (ambiguous message → DM translation)
   - 1:30-2:00: Catch-Up demo ("/catchup budget" → semantic results)
   - 2:00-2:30: Deep Work demo (button click → calendar block + status update)
   - 2:30-3:00: Impact statement + GitHub link
2. Create architecture diagram (draw.io or Excalidraw → PNG)
3. Write Devpost submission text (features, tech stack, impact)
4. Ensure `slackhack@salesforce.com` and `testing@devpost.com` are invited to sandbox
5. Final lint + test: `make lint && make test`
6. Commit: `docs: architecture diagram and demo video assets`

**Deliverable:** Video file ready. Architecture diagram PNG. Devpost draft complete.

---

### Day 6 — Submit (July 14)
**Goal:** Submit before 5:30am GMT+5:30.
**Tasks:**
1. Final end-to-end test in clean Slack workspace
2. Check all submission requirements:
   - [ ] Project Track selected: "Slack Agent for Good"
   - [ ] Text description complete (explain social impact!)
   - [ ] 3-minute demo video uploaded
   - [ ] Architecture diagram attached
   - [ ] Slack developer sandbox URL provided
   - [ ] `slackhack@salesforce.com` added to sandbox
   - [ ] `testing@devpost.com` added to sandbox
   - [ ] GitHub repo link included
   - [ ] Mintlify docs link included
3. Submit on Devpost
4. Tweet/LinkedIn post about submission (optional, for visibility)
5. Commit: `chore: final submission preparation`

**Deliverable:** Devpost submission page confirmed. Confirmation email received.

---

## 14. Submission Checklist (Devpost)

### Required Fields
- [ ] **Project Track:** Slack Agent for Good
- [ ] **Text Description:**
  - Summarize features (Focus Mode, Translator, Catch-Up, Quiet Hours, Deep Work)
  - Explain social impact: "15-20% of workforce is neurodivergent. Slack's real-time design creates daily anxiety for ADHD and autistic professionals. Signal is the first Slack agent designed for cognitive accessibility, transforming overwhelming channels into structured summaries, ambiguous messages into plain translations, and interruptions into batched digests."
  - Mention all 3 technologies: Slack AI, MCP, RTS API
- [ ] **Demo Video:** ~3 minutes, shows working project, includes all 5 features
- [ ] **Architecture Diagram:** PNG showing API, Worker, MCP Server, Postgres, Redis, OpenAI, Slack APIs
- [ ] **Slack Developer Sandbox URL:** `https://your-team.slack.com/`
- [ ] **Access Granted:** `slackhack@salesforce.com` and `testing@devpost.com` are workspace members

### Optional but Recommended
- [ ] GitHub repo link (public)
- [ ] Mintlify docs link
- [ ] Live landing page URL (Vercel)
- [ ] Screenshots of each feature
- [ ] Team member names and roles

---

## 15. Judging Criteria — How to Maximize Score

### 15.1 Technological Implementation (25%)
**What judges look for:** Quality code, proper use of required techs, architecture decisions.
**How to win:**
- Clean architecture (`internal/domain` with interfaces, `internal/store` with implementations)
- All 3 required techs are deeply integrated, not bolted on
- `sqlc` for type-safe DB access (shows engineering maturity)
- `asynq` for background jobs (shows you understand production concerns)
- MCP server is a separate binary with proper tool schemas
- Include `docs/architecture.md` explaining why you chose Go + sqlc + asynq

### 15.2 Design (25%)
**What judges look for:** UX polish, frontend/backend balance, visual appeal.
**How to win:**
- Block Kit messages look professional (colors, emojis, clear hierarchy)
- App Home is intuitive (not a wall of text)
- Next.js landing page is modern, dark-mode, mobile-responsive
- Demo video has clear before/after contrast
- Include accessibility notes (ARIA labels, keyboard nav in frontend)

### 15.3 Potential Impact (25%)
**What judges look for:** How big is the problem? How many people benefit?
**How to win:**
- Open with a statistic: "15-20% of the workforce is neurodivergent"
- Quote a real pain point: "I spend 2 hours every morning catching up on Slack"
- Explain why Slack specifically: "It's the default workplace OS, but built for neurotypical processing speed"
- Mention no existing Slack Marketplace solution for this
- Include a "future roadmap" section: voice captioning, screen reader support, team analytics

### 15.4 Quality of Idea (25%)
**What judges look for:** Creativity, uniqueness, improvement over existing solutions.
**How to win:**
- Emphasize "first neurodivergent Slack agent in the Marketplace"
- The combination of 5 features into one coherent narrative (not 5 separate bots)
- The "Social Translator" is genuinely novel — no tool exists that translates workplace subtext for autistic adults
- The MCP integration isn't just "connect to a tool" — it's "protect cognitive resources by connecting calendar + Slack status"
- Show you did research: screenshot Slack Marketplace showing only generic bots, no accessibility tools

---

## 16. Prompt Engineering Reference

### 16.1 Focus Summary Prompt
```
You are a workplace communication summarizer for neurodivergent professionals.
Your job is to extract ONLY decisions and action items from a fast-moving Slack channel.

Rules:
1. Ignore greetings, small talk, emojis, reactions, and off-topic jokes.
2. If a decision was made, state it clearly with a ✅.
3. For each decision, list action items with owners and deadlines.
4. If no decisions were made, say: "No formal decisions found. Discussion was exploratory."
5. Use plain language. No corporate buzzwords.
6. Format as a decision tree, not paragraphs.

Input (last {{.MessageCount}} messages):
{{.Messages}}

Output format:
✅ [Decision]
   ↳ [Action] — Owner: @Name — Due: [Date or "None"]
```

### 16.2 Tone Analysis Prompt
```
You are a direct, kind translator of workplace subtext for autistic and ADHD adults.
You decode ambiguous messages into literal meaning without judgment.

Analyze: "{{.Message}}"

Respond in this exact format:
- Tone: [single word or short phrase]
- Intent: [1 sentence, what they want]
- Action: [1 sentence, what you should do]
- Note: [1-2 sentences explaining any hidden social context]

Rules:
- Never say "they might be" — be confident but not rude.
- Never use "just" or "simply" — these are patronizing.
- If the message is genuinely neutral, say so clearly.
- If the message is passive-aggressive, name it directly but kindly.
```

### 16.3 Catch-Up Prompt
```
You are a "What You Missed" assistant for a neurodivergent professional returning to Slack after time away.

Summarize these messages into topics. For each topic:
1. State the topic name.
2. State any decision made (or "No decision").
3. State any action required by the user (or "None").
4. Provide a 2-sentence context summary.

Messages:
{{.Messages}}

Format:
## [Topic]
- Decision: [X]
- Your Action: [Y]
- Context: [Z]
```

---

## 17. Environment Variables (.env.example)

```bash
# Slack
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_APP_TOKEN=xapp-your-app-level-token
SLACK_SIGNING_SECRET=your-signing-secret
SLACK_CLIENT_ID=your-client-id
SLACK_CLIENT_SECRET=your-client-secret

# OpenAI
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4o-mini

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=signal
DB_PASSWORD=signal
DB_NAME=signal
DB_SSLMODE=disable

# Redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=

# MCP
MCP_SERVER_URL=http://mcp-server:3001/sse
MCP_SERVER_TIMEOUT=30s

# App
APP_ENV=development
APP_PORT=8080
LOG_LEVEL=info

# Frontend (Next.js)
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Signal
```

---

## 18. Common Pitfalls & How to Avoid Them

| Pitfall | Why It Happens | Solution |
|---|---|---|
| **Socket Mode disconnects** | Network blip or wrong token | Implement exponential backoff reconnect in `internal/slack/socket.go` |
| **AI rate limits** | Too many messages trigger OpenAI 429 | Implement Redis rate limiter (`rate_limit:ai:{user_id}`). Cache summaries for 5 minutes. |
| **Missing scopes** | Slack API returns `missing_scope` | Double-check Bot Token Scopes in Slack app settings. Need `search:read` for RTS. |
| **Permission denied on search** | RTS searches channels user left | Filter results by checking `channel.is_member` before including in summary. |
| **MCP server won't start** | Port 3001 conflict or SSE CORS | Use `0.0.0.0:3001` binding. Add CORS headers if testing from local API. |
| **Frontend OAuth fails** | Redirect URI mismatch | Ensure `http://localhost:3000/oauth/callback` is in Slack app OAuth settings. |
| **Docker build fails** | `go.mod` not copied before `go mod download` | In Dockerfile, copy `go.mod go.sum` BEFORE copying source code. |
| **Digest sends at wrong time** | Timezone mismatch | Store `digest_hour` in UTC. Convert to user's local time before sending. |
| **Focus Mode spam** | Counter resets, re-triggers immediately | Use `channel:velocity:{id}:offered` flag with 30-min TTL to prevent re-offer. |

---

## 19. Git Commit Log (Target)

Your final `git log --oneline` should look like this (chronological):
```
chore: final submission preparation
feat: quiet hours digest with asynq scheduler
feat: app home preferences dashboard
feat: next.js landing page and oauth flow
feat: MCP server with calendar and status tools
feat: RTS semantic search for catchup feature
feat: deep work mode with MCP calendar integration
feat: social translator with tone analysis
feat: focus mode velocity detection and AI summarization
feat: slack socket mode connection and basic command
chore: scaffold docker, ci/cd, makefile, and project tooling
chore: initial repository structure
```

This commit history tells a story: infrastructure → Slack connection → core features → integrations → frontend → polish → submission.

---

## 20. Quick Reference Commands

```bash
# Start everything
make dev

# Create a new feature branch and PR
make pr

# Run all tests
make test

# Lint everything
make lint

# Build all Docker images
make build-all

# Database migrations
make migrate-up
make migrate-down
make migrate-create

# Generate sqlc code
make sqlc

# Start docs locally
make docs

# View Git history
git log --oneline --graph --all

# Check test coverage
cd api && go test -cover ./...
```

---

## 21. Final Notes for AI Agent

1. **Priority Order:** Focus Mode → Translator → Catch-Up → MCP/Deep Work → Digest → Frontend. If time runs out, drop Digest and Frontend polish. The core 3 features + MCP are what judges score.

2. **Demo Video is 40% of your score.** A polished 3-minute video beats a perfect codebase. Allocate Day 5 entirely to video production.

3. **Slack sandbox must be real.** Create a free Slack workspace, invite test accounts, and actually test every feature. Judges will click around.

4. **Impact statement wins hackathons.** In every sentence of your submission, ask: "How does this help a real person?" Not "What tech did I use?" but "What pain did I remove?"

5. **Be ready to explain MCP.** Judges may ask why you used MCP instead of direct API calls. Answer: "MCP standardizes tool interfaces, making Signal extensible to any calendar or task system without code changes."

6. **Go + sqlc + asynq is a flex.** These are trendy, professional choices in 2026. Mention them in your submission text. It signals senior engineering judgment.

7. **If stuck on a feature, mock it.** A mocked MCP server that returns realistic data is better than a broken real integration. Judges care about the UX flow, not whether your Google Calendar API key works.

8. **The "Social Translator" is your secret weapon.** It is emotionally compelling, technically simple (regex + OpenAI), and genuinely unique. Feature it prominently in your demo video.

9. **Block Kit is your frontend.** Don't build a complex React dashboard for the Slack experience. The App Home tab + Block Kit messages ARE the product surface. The Next.js site is just for OAuth and marketing.

10. **Submit early, not at the deadline.** Devpost servers can lag. Aim to submit by July 13, 11:59 PM GMT+5:30. Use the final hours for polish, not panic.

---

**End of Specification**

*This document is the single source of truth for the Signal project. Any deviation from these specs should be documented in a GitHub issue or PR description before implementation.*
