<div align="center">

# 🧘 Signal
### Calm Slack for Neurodivergent Professionals

**The first Slack agent designed for cognitive accessibility**

Built for the [Slack Agent Builder Challenge](https://slack-agent-builder-challenge.devpost.com/) · July 2026

[![Slack Agent for Good](https://img.shields.io/badge/Track-Slack%20Agent%20for%20Good-4f46e5?style=flat-square)](https://slack.com)
[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Groq AI](https://img.shields.io/badge/AI-Groq%20LLaMA%203.3-f97316?style=flat-square)](https://groq.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

</div>

---

## The Problem

**15–20% of the global workforce is neurodivergent** — people with ADHD, autism, anxiety disorders, and other cognitive differences. Slack is the default communication tool at most companies. It was not designed for them.

Here's what their day looks like:

> **Alex (ADHD, Software Engineer):** Returns from a 45-minute meeting. 127 unread messages. Which ones matter? He spends 40 minutes scrolling. Misses the deployment decision made at 2:13 PM. His manager asks why he wasn't prepared. He has no answer.

> **Jordan (Autistic, Product Manager):** Receives this message: *"Per my last email, we really need to align on this before EOD."* She spends three hours wondering if she's about to be fired. Was that passive-aggressive? Did she do something wrong? She's afraid to respond.

> **Taylor (ADHD, Designer):** Gets 14 @mentions during a 2-hour deep work session. Each notification breaks her focus. She tries turning off Slack. She misses a production incident.

**There is not a single tool in the Slack Marketplace that addresses these problems for neurodivergent professionals.** Until now.

---

## The Solution

Signal is a Slack agent that transforms overwhelming, ambiguous, and interruptive Slack experiences into calm, structured, and comprehensible interactions — with zero extra apps, zero new tools, and zero learning curve.

Everything happens inside Slack. One `/signal` command to see everything it can do.

---
<img width="1536" height="1024" alt="ChatGPT Image Jul 14, 2026, 02_17_05 AM" src="https://github.com/user-attachments/assets/3ddd7711-f83e-4723-a49c-d17972ea002e" />

## Features

### 🔍 Social Translator
`/translate [message]`

Decodes passive-aggressive, ambiguous, or emotionally loaded workplace messages into plain language. Tells you the **tone**, **what the sender actually wants**, **what you should do**, and the **hidden social subtext** — explained directly, without euphemism.

```
/translate Per my last email, the deployment needs to be done by EOD.
```
→ **Tone:** Frustrated / Urgent  
→ **Intent:** They want confirmation the task is being done today  
→ **Action:** Reply with your ETA or ask for clarification  
→ **Note:** "Per my last email" signals the sender feels ignored. A quick reply defuses the tension.

Auto-detects passive-aggressive phrases in channels and sends translations privately — no one else sees it.

---

### 🎛️ Neurotype Mode
`/mode [adhd|autism|anxiety|ally]`

Personalizes every Signal feature to your cognitive profile. One command, and the entire system adapts.

| Mode | What Changes |
|------|-------------|
| **ADHD** | Focus triggers at 30 messages (not 50), summaries are bullet-only, digest shows urgent items only, deep work auto-detects |
| **Autism** | Translator always on, AI responses are literal and explicit, passive-aggressive language is named directly |
| **Anxiety** | Calm reassuring tone in all AI responses, translator confirms neutral messages as non-threatening, Focus Guard always active |
| **Ally** | Optimized for teammates supporting neurodivergent colleagues — includes "why this matters" context |

---

### 📋 What Did I Miss? (Catch-Up)
`/catchup [question]`

Semantic search across your accessible Slack channels. Answer any question about what happened while you were away — without keyword matching, without scrolling.

```
/catchup what did we decide about the Q3 budget?
/catchup any updates on the API migration?
/catchup what happened in engineering today?
```

Uses Slack's **Real-Time Search API** + Groq AI summarization to return structured "What You Missed" digests organized by topic, decision, and required action.

---

### ⚡ Focus Guard
`/focus [duration]`

Starts a deep work session. Signal:
- Sets your Slack status to 🧘 *In Deep Work — back at [time]*
- Blocks the time on your **Google Calendar** via MCP integration
- Activates smart DM auto-responder with AI urgency classification
- Pauses non-urgent digest delivery

**Smart URGENT bypass:** Incoming DMs during deep work are classified by AI as URGENT / NORMAL / LOW. If someone types "URGENT" or their message is classified as a production incident, the deep work block is immediately bypassed and you're notified.

```
/focus 2h      → 2-hour deep work session
/focus 90      → 90-minute session
/focus          → default 2 hours
```

---

### ✅ Decision Search
`/decisions [#channel] [days]`

Scans a channel's recent history and extracts every formal decision, who made it, and what action items followed. Stops the "wait, what did we decide about X?" conversation.

```
/decisions #engineering 7    → last 7 days of decisions
/decisions #design 14        → last 14 days
```

AI extracts decisions, context, and action items — formatted as a clean log.

---

### 📋 Thread Summaries
React with **📋** on any Slack message

Instantly summarizes the full thread. Only you see the summary (ephemeral). Neurotype-aware framing — ADHD mode gives bullet-only summaries, autism mode is explicit and literal.

No commands. No typing. Just react.

---

### 📋 AI Action Planner
`/plan [your goal]`

Turns any natural language goal into a structured, neurotype-aware task list.

```
/plan I need to finish the Q3 report, review Alex's PR, and prepare for demo
```

- **ADHD mode:** Energy labels (🔋 Low / ⚡ Medium / 🚀 High), short 2-minute tasks
- **Autism mode:** Literal step-by-step instructions, exact completion criteria
- **Anxiety mode:** Starts with easiest task, includes ✅ checkpoints

Uses **AI Workspace Memory** — remembers your last 10 interactions to personalize plans.

---

### 📬 Daily Digest
`/digest`

Batches all your @mentions, thread replies, and DMs into a structured digest organized by urgency:

- 🔴 **Urgent** — needs response today
- 🟡 **Action Required** — this week
- 🟢 **FYI** — no action needed

Configurable delivery time. If you have a meeting in 15 minutes (via Calendar MCP), digest delivery is intelligently postponed.

---

### 🐙 GitHub MCP
`/github [query]`

Search your GitHub organization directly from Slack.

```
/github open PRs           → your open pull requests
/github issues             → open issues in the org
/github repo signal        → repository info and stats
/github bugs assigned to me → issues labeled 'bug' assigned to you
```

Powered by real GitHub Search API — not a mock.

---

### 📚 Docs MCP (Notion)
`/docs [search query]`

Search your Notion workspace from Slack.

```
/docs onboarding guide
/docs how to deploy to production
/docs Q3 roadmap
```

Returns matching pages with titles, types, last-edited dates, and direct links.

---

### 🧠 AI Workspace Memory
Automatic — no command needed

Signal remembers your last 10 interactions (7-day rolling window, stored in Redis). This context is injected into `/plan`, action recommendations, and AI responses — making every feature more personalized over time.

---

## Architecture
<img width="2752" height="1536" alt="Signal_AI_Slack_Architecture_Dia…_202607140222" src="https://github.com/user-attachments/assets/99282deb-8cd3-459d-b965-55fff951c8a4" />

```
┌─────────────────────────────────────────────────────────────┐
│                        Slack Workspace                       │
│  User types /translate, /focus, /plan, reacts with 📋, etc. │
└──────────────────────┬──────────────────────────────────────┘
                       │ Socket Mode WebSocket
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                     Signal API (Go)                          │
│                                                              │
│  EventHandler → FeatureController → Feature Services        │
│                                                              │
│  ├── /translate  → AI Tone Analysis (Groq LLaMA 3.3)       │
│  ├── /catchup    → RTS API + AI Summary                     │
│  ├── /focus      → Redis State + MCP Calendar               │
│  ├── /plan       → AI Planner + Memory (Redis)              │
│  ├── /decisions  → Channel History + AI Extraction          │
│  ├── /mode       → PostgreSQL Preferences Update            │
│  ├── /github     → GitHub Search API                        │
│  ├── /docs       → Notion Search API                        │
│  ├── /digest     → RTS + AI Categorization                  │
│  └── 📋 react    → Thread Fetch + AI Summary               │
└──────┬──────────────────────┬──────────────────────────────┘
       │                      │
       ▼                      ▼
┌──────────────┐    ┌─────────────────────┐
│  PostgreSQL  │    │  MCP Server (Go)     │
│              │    │                      │
│  Users       │    │  block_focus_time    │
│  Preferences │    │  get_user_status     │
│  Digests     │    │  set_slack_status    │
│  Summaries   │    │       │              │
│  Translations│    │       ▼              │
└──────────────┘    │  Google Calendar API │
                    └─────────────────────┘
       │
       ▼
┌──────────────┐
│    Redis     │
│              │
│  AI Memory   │
│  Deep Work   │
│  Velocity    │
│  Rate Limits │
└──────────────┘
```

**Technologies used:**

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22 |
| Slack Integration | Socket Mode (slack-go/slack) |
| AI | Groq API (LLaMA 3.3 70B) via OpenAI-compatible SDK |
| Real-Time Search | Slack RTS API (`search.messages`) |
| MCP | Custom Go MCP server (HTTP/SSE) |
| Calendar | Google Calendar API (Service Account) |
| Database | PostgreSQL 16 (pgx driver) |
| Cache / Memory | Redis 7 |
| Background Jobs | Asynq |
| HTTP Router | chi/v5 |
| Config | cleanenv |
| Observability | Prometheus + structured slog |
| Deployment | Render (API + Worker + Postgres + Redis) |

---

## Required Technologies (Hackathon)

Signal uses all three required technologies:

### 1. Slack AI Capabilities
- Tone analysis via Groq LLaMA 3.3 70B (OpenAI-compatible)
- Focus Mode decision tree summaries
- Catch-Up semantic digests
- Thread summarization
- Action planning with neurotype-aware framing
- Urgency classification for smart notifications

### 2. MCP Server Integration
- Custom Go MCP server exposing `block_focus_time`, `get_user_status`, `set_slack_status`
- Google Calendar API integration for real focus time blocking
- GitHub API integration via `/github` command
- Notion API integration via `/docs` command

### 3. Real-Time Search API
- `/catchup` uses `search.messages` to find relevant messages semantically
- `/digest` uses RTS to surface recent mentions
- Results are passed to AI for structured summarization

---

## Getting Started

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- A Slack workspace (free works)
- Groq API key (free tier available at [console.groq.com](https://console.groq.com))

### 1. Clone and configure

```bash
git clone https://github.com/LSUDOKO/signal.git
cd signal
cp .env.example .env
```

Edit `.env` and fill in your credentials:

```bash
# Required
SLACK_BOT_TOKEN=xoxb-...
SLACK_APP_TOKEN=xapp-...
SLACK_SIGNING_SECRET=...
OPENAI_API_KEY=gsk_...         # Your Groq API key
OPENAI_BASE_URL=https://api.groq.com/openai/v1
OPENAI_MODEL=llama-3.3-70b-versatile

# Optional (enables GitHub + Notion + Calendar MCP)
GITHUB_TOKEN=ghp_...
GITHUB_ORG=your-org
NOTION_TOKEN=secret_...
MCP_CALENDAR_CREDENTIALS_PATH=/path/to/google-credentials.json
```

### 2. Start infrastructure

```bash
make dev
# Starts PostgreSQL and Redis via Docker Compose
```

### 3. Create your Slack app

1. Go to [api.slack.com/apps](https://api.slack.com/apps) → "Create New App" → "From manifest"
2. Paste the contents of `slack-manifest.yml`
3. Install to your workspace
4. Copy the Bot Token, App Token, and Signing Secret to `.env`

### 4. Run Signal

```bash
# Terminal 1: API server + Slack bot
cd api && go run ./cmd/api/main.go

# Terminal 2: MCP server (Calendar integration)
cd api && go run ./cmd/mcp-server/main.go
```

### 5. Test it

In your Slack workspace:
```
/signal          → see all commands
/mode adhd       → activate ADHD mode
/translate per my last email, we need this done today
/plan prepare for the demo tomorrow
```

---

## Slash Commands Reference

| Command | Description |
|---------|-------------|
| `/signal` | Help menu — lists all commands |
| `/mode [adhd\|autism\|anxiety\|ally]` | Set neurotype mode |
| `/translate [message]` | Decode ambiguous workplace language |
| `/catchup [question]` | AI summary of what you missed |
| `/focus [duration]` | Start deep work session |
| `/decisions [#channel] [days]` | Find all decisions in a channel |
| `/plan [goal]` | AI action planner |
| `/digest` | Instant digest of recent mentions |
| `/github [query]` | Search GitHub PRs, issues, repos |
| `/docs [query]` | Search Notion workspace |
| React 📋 | Summarize any thread (ephemeral) |

---

## Project Structure

```
signal/
├── api/                          Go backend
│   ├── cmd/
│   │   ├── api/main.go           Main API server + Slack Socket Mode
│   │   ├── worker/main.go        Asynq background job worker
│   │   └── mcp-server/main.go   MCP server (Calendar, tools)
│   ├── internal/
│   │   ├── ai/                   Groq/OpenAI client + all prompts
│   │   ├── config/               Environment config (cleanenv)
│   │   ├── domain/               Business entities and types
│   │   ├── features/             All feature implementations
│   │   │   ├── controller.go     Event router + command dispatcher
│   │   │   ├── translator.go     Social Translator
│   │   │   ├── focusmode.go      Channel velocity + AI summaries
│   │   │   ├── catchup.go        RTS semantic search
│   │   │   ├── deepwork.go       Deep Work + MCP calendar
│   │   │   ├── digest.go         Quiet Hours Digest
│   │   │   ├── mode.go           Neurotype mode personalization
│   │   │   ├── decisions.go      Decision extraction
│   │   │   ├── planner.go        AI Action Planner
│   │   │   ├── threadsummary.go  Thread summary on reaction
│   │   │   ├── memory.go         AI Workspace Memory (Redis)
│   │   │   ├── github.go         GitHub MCP
│   │   │   └── docs.go           Docs MCP (Notion)
│   │   ├── httpapi/              REST API (OAuth, preferences)
│   │   ├── mcp/                  MCP server + Google Calendar client
│   │   ├── rts/                  Real-Time Search client
│   │   ├── slack/                Socket Mode event handler
│   │   ├── store/                PostgreSQL + Redis repositories
│   │   └── observability/        Prometheus metrics + structured logging
│   └── db/migrations/            SQL schema migrations
├── slack-manifest.yml            Slack app configuration
├── render.yaml                   Render deployment config
├── docker-compose.yml            Local development infrastructure
└── Makefile                      Development commands
```

---

## Impact

**Who this helps:**

- **ADHD professionals** who lose hours to information overload every day
- **Autistic professionals** who spend enormous energy decoding social subtext
- **Anxiety-affected workers** who ruminate over ambiguous messages
- **All neurodivergent people** who feel Slack was built for someone else's brain

**The numbers:**

- 1 in 7 people worldwide are neurodivergent (approximately 1 billion people)
- 15–20% of the global workforce
- Neurodivergent employees report Slack as one of their top 3 sources of workplace anxiety
- Zero competitors exist in the Slack Marketplace for this use case

**Signal is not an accommodation tool. It's a productivity tool that happens to be built around how neurodivergent brains actually work.** Everyone benefits from clearer communication. Signal just makes it explicit.

---

## Hackathon Submission

**Track:** Slack Agent for Good  
**Technologies:** Slack AI · MCP Server Integration · Real-Time Search API  
**Sandbox:** [slackhackathontest.slack.com](https://slackhackathontest.slack.com)

**Judge access:**
- `slackhack@salesforce.com` — invited to workspace
- `testing@devpost.com` — invited to workspace

---

## License

MIT — see [LICENSE](LICENSE)

---

<div align="center">

Built with 🧘 for the people Slack forgot to design for.

**[Demo Video](#) · [Devpost Submission](#) · [Try Signal](#)**

</div>
