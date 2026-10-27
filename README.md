# 🛫 Job Search Cockpit

> **A local-first, private career command center for managing applications, networking, interview loops, and offers with zero vendor lock-in.**

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react)](https://react.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org)
[![SQLite](https://img.shields.io/badge/SQLite-Pure--Go%20(Zero--CGO)-003B57?style=flat&logo=sqlite)](https://modernc.org/sqlite)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Managing a high-caliber job search across dozens of companies involves juggling disconnected spreadsheets, browser tabs, PDF resume variants, scattered prep documents, and mental follow-up timers. 

**Job Search Cockpit** replaces fragmented tools with an active, daily command center that runs entirely on your local machine. Your candid interview debriefs, compensation numbers, private contact notes, and career stories remain strictly under your control.

---

## ✨ Key Features

### 🧭 1. Flight Deck (Daily Command Center)
- **Heads-Up Display (HUD)**: Instant snapshot of due tasks, imminent interview rounds, debriefs needed, and stale pipeline alerts.
- **Two-Column Command Grid**: Prioritizes today's agenda (interviews within 48h, due follow-up tasks) alongside a triage desk for stalled applications.
- **Pure Staleness Engine**: Alerts you when an outreach has gone unanswered for 4 days, when an application has stalled for 10 days, or when an interview completed 5 days ago without next steps.
- **One-Click Triage**: Inline resolution chips to snooze reminders, log interactions, or update stages in seconds.

### 📊 2. Orthogonal Pipeline & Drag-and-Drop Kanban
- **Orthogonal Stages vs. Outcomes**: Separates progress milestones (**Discovered**, **Outreach**, **Applied**, **Screening**, **Interviewing**, **Offer**) from terminal outcomes (**Active**, **On Hold**, **Won**, **Declined**, **Rejected**, **Withdrawn**, **Ghosted**). Closing an opportunity never destroys where it reached in the funnel.
- **Immutable Stage Transitions**: Automatically records historical audit logs with timestamps and context notes for precise funnel conversion metrics.
- **Modern Kanban Board**: High-signal cards, semantic color accents, drop-target indicators, and native HTML5 drag-and-drop with optimistic updates.
- **Filterable Table View**: Switch between Kanban columns and a structured table to inspect historical or filtered pipeline subsets.

### ⚡ 3. Quick Capture & Intelligent Ingestion
- **Global Shortcut (`⌘K` / `Ctrl+K`)**: Rapidly capture open roles from anywhere in the app.
- **Automated ATS Extraction**: Automatically parses job postings from public URLs (Greenhouse, Lever, Ashby, and JSON-LD schema) into structured fields.
- **Pluggable AI (OpenRouter-First)**: Extract structured requirements, compensation ranges, tech stack tags, and seniority using OpenRouter, Anthropic, OpenAI, Gemini, or local Ollama endpoints.
- **100% Offline / Manual Fallback**: Fully operational without an API key or internet connection.

### 🧩 4. Interview Loop Intelligence & De-duplication
- **Chronological Loop Tracker**: Order company-specific interview rounds (Recruiter Screen, Technical, System Design, Behavioral, Executive).
- **Story Bank & Loop Repeat Warnings**: Links STAR stories (Situation, Task, Action, Result) from your canonical Story Bank to each round, warning you if you accidentally repeat a story to two interviewers in the same loop.
- **Structured Debriefs**: Record candid post-interview reflections (what went well, areas to improve, company signals, confidence score, and exact questions asked) while your memory is fresh.
- **Aggregated Question Bank**: Browses all historical interview questions grouped alongside your Story Bank for targeted interview preparation.

### 📝 5. Candidate Profile, Fit Assessment & Artifacts
- **Canonical Candidate Profile**: Source of truth for your work history, impact bullets, technical skills, target preferences, and compensation floor.
- **Fit Assessment**: Evaluates job descriptions against your profile to surface matched strengths, requirement gaps, and recommended STAR stories.
- **Application Artifacts**: Write role-tailored resumes and cover letters in Markdown with ATS-friendly print-to-PDF export, or attach external compiled PDF documents (LaTeX, Typst, Word).

### 🤝 6. Networking CRM & Outreach
- **Global Contact Directory**: Manage recruiters, referrers, hiring managers, interviewers, and champions across organizations.
- **Touchpoint Logging**: Log timestamped interactions (coffee chats, referral requests, emails, LinkedIn messages) and schedule automatic follow-up action items.

### 💰 7. Structured Offers & Funnel Analytics
- **Side-by-Side Offer Comparison**: Break down base salary, target bonus, equity/RSUs, vesting schedules, and sign-on bonuses against your compensation floor.
- **Funnel Conversion Analytics**: Inspect stage-to-stage conversion rates, average days per stage, and drop-off bottlenecks computed from immutable transition logs.
- **Local Backup & Seeding**: One-click JSON export, SQLite `VACUUM INTO` snapshots, and optional sample data seeding.

---

## 🏛️ Architecture & Clean Domain Boundaries

Job Search Cockpit is built using **Concentric Clean Architecture** on the backend and modular, decoupled state management on the frontend:

```
Job-Search-Cockpit/
├── cmd/cockpit/               # Application entry point (main.go)
├── internal/
│   ├── domain/               # Pure Go domain entities, value objects, domain errors, and evaluators
│   ├── usecase/              # Application services & ports (inbound & outbound interfaces)
│   ├── adapter/
│   │   ├── http/             # REST controllers, DTOs, and error mapping
│   │   ├── sqlite/           # Pure-Go SQLite migrations & repositories (zero-CGO)
│   │   ├── ai/               # OpenRouter / LLM prompt construction & extraction
│   │   ├── scraper/          # JSON-LD & ATS job posting extraction
│   │   └── fs/               # Local binary artifact blob storage (PDFs)
│   └── app/                  # Composition root & HTTP server wiring
└── web/                      # React 19 + TypeScript SPA
    ├── src/
    │   ├── api/              # Strongly typed API client & context
    │   ├── components/       # Shared UI primitives & AppShell
    │   ├── features/         # Feature modules (flightdeck, pipeline, profile, network, analytics)
    │   └── lib/              # Formatting & timezone utilities (UTC-to-local)
```

- **Zero CGO**: Uses [`modernc.org/sqlite`](https://modernc.org/sqlite), enabling effortless cross-compilation across macOS, Linux, and Windows without requiring a C compiler.
- **Pure Domain Evaluators**: Business rules (staleness evaluation, story repeat detection, funnel analytics) are pure Go functions tested with zero I/O or mocking.
- **Decoupled Frontend**: React components interact through a centralized, typed API gateway backed by TanStack Query for caching and optimistic updates.

---

## 🚀 Getting Started

### Prerequisites
- **Go**: `1.24` or later ([Download Go](https://golang.org/dl/))
- **Node.js**: `20.x` or later ([Download Node.js](https://nodejs.org/))

### 1. Clone & Install Dependencies
```bash
git clone https://github.com/your-username/Job-Search-Cockpit.git
cd Job-Search-Cockpit

# Install frontend dependencies
cd web && npm install && cd ..
```

### 2. Run in Development Mode
You can run the Go backend and Vite development server in parallel:

**Terminal 1 (Backend):**
```bash
go run cmd/cockpit/main.go
# Server listens on :8080 with SQLite database initialized in ./data/cockpit.db
```

**Terminal 2 (Frontend):**
```bash
cd web
npm run dev
# Frontend runs at http://localhost:5173 (proxies /api to http://localhost:8080)
```

Open [http://localhost:5173](http://localhost:5173) in your browser.

---

## ⚙️ Configuration & Environment Variables

Job Search Cockpit is configured via environment variables or the in-app **Settings** page:

| Variable | Default | Description |
| :--- | :--- | :--- |
| `COCKPIT_ADDR` | `:8080` | HTTP listen address for the Go backend |
| `COCKPIT_DATA_DIR` | `./data` | Local directory for `cockpit.db` and uploaded artifacts |
| `OPENROUTER_API_KEY`| *empty* | Optional OpenRouter API key for LLM-assisted features |

*(Note: AI features are optional. All extraction, tailoring, and debriefing features support 100% manual entry without an API key).*

---

## 🧪 Testing & Verification

Both the backend and frontend maintain comprehensive test suites:

```bash
# Run all Go backend unit & integration tests
go test -v ./...

# Run all React workspace Vitest test suites
npm --prefix web test

# Run frontend production build & type check
npm --prefix web run build
```

---

## 📖 Ubiquitous Language

The system strictly adheres to the domain terms defined in [`GLOSSARY.md`](GLOSSARY.md):

- **Opportunity**: A distinct pursuit of a role at a Company, whether or not a formal Job Posting exists yet.
- **Stage**: Company-agnostic macro funnel milestones (`Discovered`, `Outreach`, `Applied`, `Screening`, `Interviewing`, `Offer`).
- **Outcome**: The resolution state (`Active`, `On Hold`, `Won`, `Declined`, `Rejected`, `Withdrawn`, `Ghosted`), orthogonal to Stage.
- **Flight Deck**: The daily triage dashboard highlighting immediate tasks, interview prep, and stale pursuits.
- **Debrief**: Structured post-interview reflection capturing questions, signals, and follow-ups.
- **Story Bank**: The repository of reusable STAR narratives linked across interview rounds.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
