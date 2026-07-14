# Signal — AI Agent Collaboration Protocol
## Git Workflow, Commit Standards, PR Automation & Merge Strategy

**Version:** 1.0  
**Last Updated:** 2026-07-11  
**Author:** Signal Engineering Team  
**Purpose:** This document governs how the AI agent (you) interacts with the Git repository, creates branches, writes commits, opens PRs, and merges code. Follow these rules exactly. No exceptions.

---

## 1. Philosophy

We practice **trunk-based development** with short-lived feature branches. Every change goes through a Pull Request. No direct pushes to `main`. No force pushes. No rewriting history on shared branches.

**Why:**
- GitHub Actions CI runs on every PR (lint, test, build)
- PRs create a review trail for hackathon judges
- Clean history tells the story of the project
- Prevents "it works on my machine" disasters

---

## 2. Branch Strategy

### 2.1 Branch Types

| Branch | Purpose | Lifetime | Protection |
|---|---|---|---|
| `main` | Production-ready code. Always deployable. | Permanent | PR required, CI must pass |
| `feat/*` | New features | 1-2 days | PR + review + CI |
| `fix/*` | Bug fixes | Hours | PR + review + CI |
| `docs/*` | Documentation, README, specs | Hours | PR + CI (skip tests) |
| `chore/*` | Tooling, config, deps | Hours | PR + CI |
| `refactor/*` | Code restructuring, no behavior change | 1 day | PR + CI |
| `test/*` | Test additions/improvements | Hours | PR + CI |
| `hotfix/*` | Critical production fixes | Minutes | PR + fast-track review |

### 2.2 Naming Convention

```
<type>/<short-kebab-description>
```

**Rules:**
- All lowercase
- Kebab-case (hyphens, not underscores)
- No spaces, no special characters except hyphen
- Max 50 characters total
- Description must be a verb phrase (what the branch DOES)

**Good Examples:**
```
feat/focus-mode-velocity-detection
fix/translator-dm-routing-edge-case
docs/mcp-server-api-reference
chore/add-golangci-lint-config
refactor/extract-slack-event-router
test/catchup-semantic-search-coverage
```

**Bad Examples:**
```
feature/new-stuff          # Wrong prefix, vague description
fix-bug                    # Missing type prefix, no description
feat/FocusMode            # Uppercase not allowed
feat/this_is_a_very_long_branch_name_that_is_hard_to_read  # Too long
jordan-fix                 # Personal names not allowed
```

### 2.3 Branch Lifecycle

```
main
  │
  ├── feat/focus-mode-velocity-detection  (Day 1, 2 hours)
  │   └── PR #1 → merged → main
  │
  ├── feat/social-translator-tone-analysis  (Day 1, 3 hours)
  │   └── PR #2 → merged → main
  │
  ├── feat/rts-semantic-catchup-search  (Day 2, 4 hours)
  │   └── PR #3 → merged → main
  │
  ├── feat/mcp-server-calendar-tools  (Day 2, 3 hours)
  │   └── PR #4 → merged → main
  │
  ├── feat/nextjs-landing-oauth  (Day 3, 5 hours)
  │   └── PR #5 → merged → main
  │
  └── docs/architecture-diagram-demo-video  (Day 5, 2 hours)
      └── PR #6 → merged → main
```

**Maximum branch lifetime:** 24 hours. If a branch lives longer than a day, it is too big. Split it.

---

## 3. Commit Standards (Conventional Commits)

### 3.1 Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Rules:**
- Type and scope are lowercase
- Subject is lowercase, imperative mood ("add" not "added" or "adds")
- No period at end of subject line
- Subject max 72 characters
- Body wraps at 72 characters
- Footer references issues/PRs

### 3.2 Types

| Type | Use When | Example |
|---|---|---|
| `feat` | New feature or capability | `feat(focus): add channel velocity detector with redis counters` |
| `fix` | Bug fix | `fix(translator): handle empty mentioned user list` |
| `docs` | Documentation only | `docs(readme): add installation instructions` |
| `style` | Formatting, semicolons, etc. (no code change) | `style(go): format imports with goimports` |
| `refactor` | Code change that neither fixes nor adds feature | `refactor(slack): extract event router into separate package` |
| `perf` | Performance improvement | `perf(ai): cache gpt responses for 5 minutes` |
| `test` | Adding or correcting tests | `test(catchup): add integration test for rts search` |
| `chore` | Build process, dependencies, tooling | `chore(ci): add github actions workflow` |
| `ci` | CI/CD changes specifically | `ci(docker): add multi-stage build for api` |
| `build` | Build system changes | `build(makefile): add migrate-up target` |
| `revert` | Revert previous commit | `revert: feat(focus) add velocity detector` |

### 3.3 Scopes

| Scope | Area | Example |
|---|---|---|
| `focus` | Focus Mode feature | `feat(focus): add velocity threshold config` |
| `translator` | Social Translator feature | `fix(translator): correct tone detection regex` |
| `catchup` | Catch-Up / RTS feature | `feat(catchup): implement semantic query builder` |
| `digest` | Quiet Hours Digest feature | `feat(digest): add asynq scheduler` |
| `deepwork` | Deep Work / MCP feature | `feat(deepwork): connect mcp client to calendar` |
| `mcp` | MCP server or client | `feat(mcp): add block_focus_time tool schema` |
| `slack` | Slack SDK, events, Socket Mode | `fix(slack): reconnect on websocket error` |
| `ai` | OpenAI client, prompts | `perf(ai): reduce token usage in summary prompt` |
| `rts` | Real-Time Search API | `feat(rts): add search query builder` |
| `store` | Database, repositories | `feat(store): add user preferences migration` |
| `api` | HTTP API, handlers | `feat(api): add oauth callback handler` |
| `web` | Next.js frontend | `feat(web): add app home preferences page` |
| `docs` | Mintlify documentation | `docs(api): document rts search endpoint` |
| `ci` | GitHub Actions | `ci(test): add race detection to go tests` |
| `config` | Environment, settings | `chore(config): add redis connection string` |
| `deps` | Dependencies | `chore(deps): upgrade slack-go to v0.14` |

### 3.4 Full Commit Examples

**Simple feature commit:**
```
feat(focus): add channel velocity detector with redis counters

Implements Redis-based message counting per channel with 10-minute TTL.
When threshold (default 50 messages) is reached, triggers Focus Mode
offer with AI-generated decision tree summary.

Closes #3
```

**Bug fix with context:**
```
fix(translator): handle empty mentioned user list in tone analysis

Previously, messages with no @mentions would crash the regex parser
when trying to find target users for DM. Now gracefully skips
translation when no users are mentioned and message is not a direct
reply to a thread.

Fixes #7
```

**Documentation:**
```
docs(architecture): add system diagram and data flow explanation

Includes Excalidraw source file and exported PNG for Devpost
submission. Documents the interaction between Socket Mode, API,
Worker, MCP Server, and external services.
```

**Breaking change (rare in hackathon, but good to know):**
```
feat(api)!: change preferences endpoint from PUT to PATCH

BREAKING CHANGE: The /api/v1/users/{id}/preferences endpoint now
accepts partial updates instead of requiring full resource.
Frontend must be updated to send only changed fields.
```

### 3.5 Commit Frequency

**Rule:** Commit early, commit often. Minimum 3 commits per feature branch. Maximum 10 commits per branch (squash if more).

**Good rhythm:**
```
feat(focus): scaffold focus mode controller and redis client
feat(focus): implement velocity counter with 10-minute ttl
feat(focus): add conversations.history fetch for last 50 messages
feat(focus): integrate openai for decision tree summarization
feat(focus): add block kit message builder with action buttons
feat(focus): write integration test for velocity threshold trigger
```

**Bad rhythm (everything in one commit):**
```
feat: focus mode done
```

---

## 4. Pull Request Workflow (The ONLY Way to Merge)

### 4.1 PR Creation Rules

**Every code change MUST go through a PR.** No exceptions. Even documentation. Even one-line fixes.

**PR Title Format:**
```
<type>(<scope>): <imperative description>
```

Same as commit subject line, but can be slightly longer (max 100 chars).

**Examples:**
```
feat(focus): implement channel velocity detection and ai summarization
fix(translator): correct dm routing when no users are mentioned
feat(web): add next.js landing page with slack oauth integration
docs: add architecture diagram and demo video assets
```

### 4.2 PR Description Template

Every PR description MUST follow this template. Copy-paste and fill in.

```markdown
## What
<!-- 1-2 sentences describing the change -->

## Why
<!-- Link to issue or explain the problem this solves -->

## How to Test
<!-- Step-by-step for reviewers to validate -->
1. 
2. 
3. 

## Screenshots / Demo
<!-- If UI change, include screenshot or Loom link -->

## Checklist
- [ ] `make lint` passes
- [ ] `make test` passes (or tests added for new code)
- [ ] Feature tested in Slack sandbox (if applicable)
- [ ] Commit messages follow Conventional Commits
- [ ] No direct pushes to main (this is a PR)
- [ ] Branch is up to date with main before merge
```

### 4.3 PR Size Rules

| Metric | Maximum | Why |
|---|---|---|
| Files changed | 15 | Large PRs are hard to review |
| Lines added | 500 | Beyond this, split into stacked PRs |
| Lines deleted | 300 | Refactors can be bigger, but document why |
| Commits | 10 | Squash if more |
| Review time | 24 hours | Merge within a day of opening |

**If a PR is too big, split it:**
```
# Instead of one giant PR "feat: all features done"
# Do:
PR 1: feat(focus): scaffold controller and redis client
PR 2: feat(focus): implement velocity detection logic
PR 3: feat(focus): integrate ai summarization
PR 4: feat(focus): add block kit ui and action handlers
```

### 4.4 PR Labels (GitHub)

Always add at least one label:

| Label | Use When |
|---|---|
| `enhancement` | New feature |
| `bug` | Bug fix |
| `documentation` | Docs only |
| `chore` | Tooling/config |
| `breaking-change` | Changes API/behavior |
| `needs-review` | Ready for human review |
| `wip` | Work in progress, do not merge |
| `hackathon` | Related to hackathon submission |

### 4.5 PR Automation Commands (GitHub CLI)

```bash
# Create PR from current branch (after pushing)
gh pr create   --title "feat(focus): add channel velocity detector"   --body-file .github/pull_request_template.md   --base main   --label "enhancement,hackathon"

# View PR status
gh pr view
gh pr checks

# Add reviewer
gh pr edit --add-reviewer username

# Mark ready for review (if draft)
gh pr ready

# Convert to draft (if needs more work)
gh pr ready --undo
```

---

## 5. Merge Strategy

### 5.1 Merge Methods

| Method | When to Use | Command |
|---|---|---|
| **Squash and Merge** | Default for all PRs | `gh pr merge --squash --delete-branch` |
| **Rebase and Merge** | Long-lived feature branch with clean commits | `gh pr merge --rebase --delete-branch` |
| **Create Merge Commit** | Never (pollutes history) | Do not use |

**Squash and Merge is the default.** It combines all branch commits into one clean commit on main with the PR title as the commit message.

### 5.2 Merge Requirements

Before merging ANY PR, the following MUST be true:

```
□ CI passes (GitHub Actions green checkmarks)
□ make lint passes locally
□ make test passes locally
□ Branch is up to date with main (rebase if behind)
□ PR has at least 1 approval (if team member available)
□ PR title follows Conventional Commits format
□ PR description is complete (What, Why, How to Test)
□ No WIP label
□ No merge conflicts
```

### 5.3 Merge Commands

```bash
# Standard merge (squash + delete branch)
gh pr merge --squash --delete-branch

# If branch is behind main, rebase first
git fetch origin
git rebase origin/main
# Resolve conflicts if any
git push --force-with-lease  # Safe force push on feature branch
gh pr merge --squash --delete-branch

# After merge, verify main is healthy
git checkout main
git pull origin main
make lint
make test
```

### 5.4 Post-Merge Verification

After EVERY merge to main:

```bash
# 1. Pull latest main
git checkout main
git pull origin main

# 2. Verify CI is green on main
gh run list --branch main --limit 5

# 3. Run full test suite locally
make test

# 4. If anything fails, revert immediately
git revert HEAD --no-edit
git push origin main
# Then fix in a new branch and open new PR
```

---

## 6. Daily Workflow (Step-by-Step)

### 6.1 Morning Start (Every Day)

```bash
# 1. Switch to main and pull latest
git checkout main
git pull origin main

# 2. Verify everything is healthy
make lint
make test

# 3. Start infrastructure
make dev

# 4. Check for any open PRs that need attention
gh pr list --author "@me"
```

### 6.2 Starting a New Feature

```bash
# 1. Ensure you're on main and up to date
git checkout main
git pull origin main

# 2. Create branch using the naming convention
# Format: type/short-kebab-description
git checkout -b feat/focus-mode-velocity-detector

# 3. Work on code...
# ... edit files ...

# 4. Stage changes
git add .

# 5. Commit with Conventional Commits format
git commit -m "feat(focus): add redis velocity counter with 10-minute ttl

Implements atomic INCR with TTL pipeline to track messages per
channel. Threshold configurable per user preference."

# 6. Continue working, commit often...
# ... more edits ...
git add .
git commit -m "feat(focus): integrate conversations.history fetch"

git add .
git commit -m "feat(focus): add openai decision tree summarization"

git add .
git commit -m "feat(focus): implement block kit message builder"

# 7. Push branch to remote
git push -u origin feat/focus-mode-velocity-detector

# 8. Create PR
gh pr create   --title "feat(focus): implement channel velocity detection and ai summarization"   --body "## What
Implements Focus Mode feature. When a channel exceeds 50 messages in 10 minutes, Signal offers an AI-generated decision tree summary.

## Why
Closes #3. Addresses information overload for ADHD users.

## How to Test
1. Install app to test workspace
2. Send 50 messages in #test-channel within 10 minutes
3. Verify Focus Mode message appears with summary and buttons

## Checklist
- [x] make lint passes
- [x] make test passes
- [x] Tested in Slack sandbox
- [x] Commit messages follow Conventional Commits"   --base main   --label "enhancement,hackathon"

# 9. Wait for CI to pass
gh pr checks --watch

# 10. Merge when ready
gh pr merge --squash --delete-branch

# 11. Clean up local branch
git branch -d feat/focus-mode-velocity-detector

# 12. Return to main
git checkout main
git pull origin main
```

### 6.3 Fixing a Bug (Hotfix Flow)

```bash
# 1. Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b fix/translator-dm-crash-empty-mentions

# 2. Fix the bug, write test for it
git add .
git commit -m "fix(translator): handle empty mentioned user list

Messages with no @mentions previously caused nil pointer dereference
when building DM recipient list. Now returns early with debug log."

# 3. Push and create PR immediately
git push -u origin fix/translator-dm-crash-empty-mentions
gh pr create --title "fix(translator): handle empty mentioned user list" --body "Fixes #7. Critical bug causing panic on messages without mentions." --base main --label "bug,hackathon"

# 4. Fast-track merge (if CI passes)
gh pr checks --watch
gh pr merge --squash --delete-branch
```

### 6.4 End of Day (Every Day)

```bash
# 1. Check status
git status

# 2. If uncommitted work exists, decide:
#    - Complete enough to commit? → commit and push
#    - Not complete? → stash or leave in branch (never commit broken code to main)

# 3. Ensure all branches are pushed
git push --all origin

# 4. Check open PRs
gh pr list --author "@me"

# 5. Update issue board if using GitHub Projects
# (Optional: gh project item-edit ...)

# 6. Write a brief standup note in a file or Slack
# "Today: completed Focus Mode velocity detection + AI summarization.
#  Tomorrow: start Social Translator regex + tone analysis."
```

---

## 7. Emergency Procedures

### 7.1 CI Fails on Main After Merge

```bash
# 1. Identify the bad commit
git log --oneline -5

# 2. Revert immediately (do not try to fix on main)
git revert HEAD --no-edit
# Or revert specific commit:
git revert abc1234 --no-edit

# 3. Push the revert
git push origin main

# 4. Create new branch to fix properly
git checkout -b fix/revert-velocity-detector-bug

# 5. Fix the issue, test thoroughly
# 6. Open new PR, merge when CI passes
```

### 7.2 Merge Conflict During Rebase

```bash
# While rebasing feature branch onto main:
git rebase origin/main
# CONFLICT in internal/features/focusmode.go

# 1. Open conflicted file, resolve markers
# 2. Stage resolved file
git add internal/features/focusmode.go

# 3. Continue rebase
git rebase --continue

# 4. If rebase is too messy, abort and merge instead
git rebase --abort
git merge origin/main  # Creates merge commit (acceptable on feature branch)
```

### 7.3 Accidental Direct Push to Main

```bash
# If you accidentally pushed to main directly:

# 1. Do NOT force push to "fix" it
# 2. Revert the bad commit on main
git revert HEAD --no-edit
git push origin main

# 3. Create proper feature branch with the work
git checkout -b feat/recovered-work HEAD~1
git cherry-pick <commit-hash>

# 4. Open PR and merge properly
```

### 7.4 Lost Work (Uncommitted Changes)

```bash
# If you accidentally checked out another branch without committing:

# 1. Check git stash list
git stash list

# 2. If stashed, pop it
git stash pop

# 3. If not stashed, check reflog
git reflog
# Find the commit hash before checkout, then:
git checkout <hash>
git checkout -b recovery-branch

# 4. Commit the recovered work properly
```

---

## 8. GitHub Actions CI Integration

### 8.1 CI Runs on Every PR

The `.github/workflows/ci.yml` runs:
1. Go lint (`golangci-lint`)
2. Go test (`go test -race -cover`)
3. Go build (api, worker, mcp-server)
4. Next.js lint (`npm run lint`)
5. Next.js type check (`tsc --noEmit`)
6. Next.js build (`npm run build`)

### 8.2 CI Must Pass Before Merge

**Never merge a PR with failing CI.** Even if "it's just a test." Even if "it works locally." Fix the CI first.

Common CI failures and fixes:

| Failure | Likely Cause | Fix |
|---|---|---|
| `golangci-lint` fails | Unused imports, missing error checks | Run `make lint` locally, fix issues, push |
| `go test -race` fails | Data race in goroutines | Add mutex locks, use channels properly |
| `go build` fails | Syntax error, missing dependency | Run `go mod tidy`, fix imports |
| `npm run lint` fails | ESLint rule violation | Run `cd frontend && npm run lint -- --fix` |
| `tsc --noEmit` fails | Type error in TypeScript | Fix type annotations, check interfaces |
| `npm run build` fails | Next.js build error | Check `next.config.ts`, fix imports |

### 8.3 Skipping CI (Emergency Only)

```bash
# If you MUST push without CI (e.g., updating README during demo):
git commit -m "docs: update demo video link [skip ci]"
# CI will not run. Use sparingly.
```

---

## 9. Release & Tagging (Post-Hackathon)

### 9.1 Hackathon Submission Tag

After final submission to Devpost, tag the commit:

```bash
# 1. Ensure main is clean and all PRs merged
git checkout main
git pull origin main
make lint
make test

# 2. Create annotated tag
git tag -a v1.0.0-hackathon -m "Slack Agent Builder Challenge submission

Features: Focus Mode, Social Translator, Catch-Up, Quiet Hours Digest, Deep Work Protector
Technologies: Slack AI, MCP, Real-Time Search API
Track: Slack Agent for Good
Date: 2026-07-14"

# 3. Push tag
git push origin v1.0.0-hackathon

# 4. Create GitHub release
gh release create v1.0.0-hackathon   --title "Signal v1.0.0 — Hackathon Submission"   --notes-file RELEASE_NOTES.md   --generate-notes
```

### 9.2 Version Format

During hackathon: `v0.x.x` (pre-release)  
After submission: `v1.0.0-hackathon`  
Future iterations: `v1.1.0`, `v2.0.0` following semver

---

## 10. Collaboration with Human Teammates

### 10.1 If Working with Others

| Scenario | Action |
|---|---|
| Teammate opens PR | Review within 4 hours. Approve if CI passes and code is reasonable. |
| Teammate requests changes on your PR | Address within 2 hours. Do not argue about style (follow conventions). |
| Merge conflict with teammate's branch | Communicate in Slack. Rebase onto main, resolve together. |
| Teammate pushes to main directly | Revert their commit, message them politely, share this doc. |
| Disagreement on architecture | Open GitHub Discussion or issue. Do not block PRs over style. |

### 10.2 Code Review Checklist (When Reviewing)

```markdown
## Review Checklist

### Correctness
- [ ] Feature works as described in PR
- [ ] Edge cases handled (empty input, nil pointers, errors)
- [ ] No obvious bugs or security issues

### Testing
- [ ] Tests added for new code
- [ ] Tests pass locally (`make test`)
- [ ] CI passes on PR

### Style
- [ ] Conventional Commits format followed
- [ ] Code follows project conventions (naming, structure)
- [ ] No commented-out code or debug prints
- [ ] Error messages are helpful

### Performance
- [ ] No N+1 queries (check sqlc generated code)
- [ ] Redis operations use pipelines where appropriate
- [ ] No unnecessary API calls (cache where possible)

### Documentation
- [ ] Complex logic has comments explaining WHY
- [ ] Public functions have doc comments
- [ ] PR description is complete
```

---

## 11. AI Agent Specific Instructions

### 11.1 When Generating Code

1. **Always create a branch first.** Never commit directly to main.
2. **Write commit messages as you go.** Do not batch 20 files into one commit.
3. **Run `make lint` before every commit.** Fix issues immediately.
4. **Write tests for new features.** Even minimal tests are better than none.
5. **Update this doc if conventions change.** Add a `docs` PR.

### 11.2 When Reviewing Your Own Work

Before opening a PR, ask yourself:
- [ ] Would a human teammate understand this PR in 2 minutes?
- [ ] Are the commit messages descriptive enough to reconstruct the work later?
- [ ] Is the PR small enough to review in 10 minutes?
- [ ] Did I test the feature in the Slack sandbox?
- [ ] Are there any secrets or API keys in the code? (Use `.env`)

### 11.3 When Stuck

If you encounter a problem:
1. **Search existing issues/PRs** for similar problems
2. **Check Section 18 (Common Pitfalls) in SIGNAL_PROJECT_SPEC.md**
3. **Ask for help** by opening a GitHub issue with:
   - What you tried
   - Error message
   - Expected vs actual behavior
   - Minimal reproduction steps

### 11.4 When Completing a Feature

After implementing a feature:
1. Run full test suite: `make test`
2. Run linter: `make lint`
3. Test in Slack sandbox
4. Write or update documentation in `docs/`
5. Open PR with complete description
6. Merge only after CI passes
7. Move to next feature in roadmap

---

## 12. Quick Reference Card

### Commands You Will Use 100 Times

```bash
# Daily start
git checkout main && git pull origin main && make dev

# New feature
git checkout -b feat/short-description
# ... work ...
git add . && git commit -m "feat(scope): description"
git push -u origin feat/short-description
gh pr create --title "feat(scope): description" --body "..." --base main
gh pr checks --watch
gh pr merge --squash --delete-branch

# Bug fix
git checkout -b fix/short-description
# ... fix ...
git add . && git commit -m "fix(scope): description"
git push -u origin fix/short-description
gh pr create --title "fix(scope): description" --body "Fixes #N" --base main
gh pr merge --squash --delete-branch

# End of day
git status && git push --all origin && gh pr list --author "@me"

# Emergency revert on main
git checkout main && git pull origin main
git revert HEAD --no-edit && git push origin main
```

### Branch Prefix Cheat Sheet

| Type | Prefix | Example |
|---|---|---|
| Feature | `feat/` | `feat/focus-mode-detection` |
| Bug fix | `fix/` | `fix/translator-dm-crash` |
| Documentation | `docs/` | `docs/architecture-diagram` |
| Chore | `chore/` | `chore/add-ci-workflow` |
| Refactor | `refactor/` | `refactor/extract-slack-client` |
| Test | `test/` | `test/catchup-integration` |
| Hotfix | `hotfix/` | `hotfix/api-panic` |

### Commit Type Cheat Sheet

| Type | When | Example |
|---|---|---|
| `feat` | New capability | `feat(focus): add velocity counter` |
| `fix` | Bug fix | `fix(translator): handle nil pointer` |
| `docs` | Documentation | `docs(readme): add install steps` |
| `style` | Formatting | `style(go): run goimports` |
| `refactor` | Restructure | `refactor(store): extract interface` |
| `perf` | Optimize | `perf(ai): cache responses` |
| `test` | Tests | `test(focus): add threshold test` |
| `chore` | Tooling | `chore(deps): upgrade pgx` |
| `ci` | CI/CD | `ci(docker): add build stage` |
| `build` | Build system | `build(makefile): add test target` |
| `revert` | Undo | `revert: feat(focus) velocity` |

---

## 13. Example: Complete Day 2 Workflow

**Context:** Day 2 of the hackathon. Yesterday merged Focus Mode. Today building Social Translator.

```bash
# === 09:00 — Morning Start ===
git checkout main
git pull origin main
make dev
# Verify: all services running

# === 09:15 — Create Branch ===
git checkout -b feat/social-translator-tone-analysis

# === 09:30 — Scaffold ===
# Create internal/features/translator.go
# Create prompts/tone_analyzer.tmpl
# Write tests

git add internal/features/translator.go prompts/tone_analyzer.tmpl
git commit -m "feat(translator): scaffold tone analysis controller and prompt template

Adds regex-based trigger detection for ambiguous workplace phrases.
Defines OpenAI prompt template for tone/intent/action extraction."

# === 11:00 — Implement Core Logic ===
# Write OnMessage handler, DM builder, regex patterns

git add internal/features/translator.go
git commit -m "feat(translator): implement message scanning and dm routing

Scans incoming messages for passive-aggressive/ambiguous phrases.
Routes translation DMs to mentioned users or requester."

# === 13:00 — Integrate AI ===
# Connect to OpenAI client, format response, build Block Kit message

git add internal/features/translator.go internal/ai/client.go
git commit -m "feat(translator): integrate openai for tone analysis

Sends matched messages to GPT-4o-mini with tone_analyzer prompt.
Formats response into structured Block Kit DM with tone, intent,
action, and note fields."

# === 15:00 — Test in Sandbox ===
# Install app, post "Per my last email...", verify DM received
# Fix edge case: message with no mentions

git add internal/features/translator.go
git commit -m "fix(translator): handle messages with no user mentions

Previously crashed when trying to find DM recipients for messages
without @mentions. Now skips translation for non-mention messages
unless explicitly requested via /translate command."

# === 16:00 — Write Tests ===
# Add unit tests for regex matching, DM formatting, edge cases

git add internal/features/translator_test.go
git commit -m "test(translator): add unit tests for tone detection and dm formatting

Tests regex matching for 10 ambiguous phrases. Tests DM Block Kit
builder with various tone outputs. Tests edge case: empty mentions."

# === 17:00 — Lint and Verify ===
make lint
make test
# All green

# === 17:15 — Push and Create PR ===
git push -u origin feat/social-translator-tone-analysis

gh pr create   --title "feat(translator): implement social tone analysis and dm translation"   --body "## What
Implements Social Translator feature. Detects ambiguous/passive-aggressive workplace messages and sends plain-English translations via DM to neurodivergent users.

## Why
Closes #4. Addresses social anxiety for autistic professionals who struggle with workplace subtext.

## How to Test
1. Install app to sandbox
2. Post "Per my last email, we need this by EOD" in #general
3. Verify Signal DMs you with tone analysis (frustrated/urgent)
4. Test /translate command with custom message

## Screenshots
[Attach screenshot of DM]

## Checklist
- [x] make lint passes
- [x] make test passes
- [x] Tested in Slack sandbox
- [x] Commit messages follow Conventional Commits
- [x] No direct pushes to main"   --base main   --label "enhancement,hackathon"

# === 17:20 — Wait for CI ===
gh pr checks --watch
# All checks pass

# === 17:30 — Merge ===
gh pr merge --squash --delete-branch

# === 17:35 — Verify Main ===
git checkout main
git pull origin main
make test
# Green

# === 17:40 — Plan Tomorrow ===
# Write standup note: "Tomorrow: Catch-Up RTS semantic search + MCP server skeleton"

# === 18:00 — End of Day ===
git status
# Clean
gh pr list --author "@me"
# No open PRs
# Done.
```

---

## 14. Rules Summary (The Non-Negotiables)

1. **NO direct pushes to `main`.** Ever. Every change goes through a PR.
2. **NO force pushes to `main`.** Ever. History is sacred.
3. **ALL commits use Conventional Commits format.** Every single one.
4. **ALL PRs have descriptive titles and complete descriptions.** No empty PRs.
5. **CI must pass before merge.** No exceptions, no "I'll fix it later."
6. **Branches live max 24 hours.** Merge or abandon. No long-lived branches.
7. **Squash and merge is default.** One clean commit per feature on main.
8. **Delete branches after merge.** No orphan branches.
9. **Run `make lint` before every commit.** Fix issues immediately.
10. **Test in Slack sandbox before opening PR.** Code that hasn't been clicked is broken.

---

**End of Protocol**

*This document is law. Violations require immediate correction and a `docs` PR explaining why the rule was broken and how to prevent recurrence.*
