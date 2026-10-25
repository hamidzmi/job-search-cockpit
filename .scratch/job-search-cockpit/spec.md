# Job Search Cockpit Specification

**Status:** ready-for-agent

## Problem Statement

Managing a serious job search across dozens of target companies involves juggling disconnected spreadsheets, browser tabs, PDF resume variants, scattered interview prep docs, LinkedIn messages, and mental follow-up timers. Existing job trackers fail in several ways:

1. **Flat, brittle pipeline models**: Treating terminal outcomes (`Rejected`, `Withdrawn`, `Ghosted`) as columns at the end of a board overwrites how far an opportunity progressed, making funnel conversion analysis impossible. Conversely, baking rigid company-specific interview steps into board columns breaks when companies use different interview sequences.
2. **Application-only myopia**: Not every pursuit starts with a formal job posting—many begin with a warm intro or recruiter conversation, or involve applying to two different roles at the same company simultaneously.
3. **Disconnected tailoring & interview prep**: Candidate experience, reusable STAR stories, role-specific fit analysis, tailored resumes/cover letters, and post-interview debriefs live outside the tracker, making it easy to forget which resume was sent or accidentally repeat the same STAR story to two interviewers in the same loop.
4. **Passive storage instead of active execution**: Spreadsheets don't warn the candidate when a referral request has gone unanswered for 4 days, when an applied role has gone silent for 10 days, when an interview is 24 hours away, or when a completed interview is missing a debrief while memories are fresh.
5. **Privacy & vendor lock-in**: Compensation floors, candid interview debriefs, and contact notes are deeply personal; storing them in a multi-tenant cloud SaaS or paying monthly subscriptions adds friction without local file access.

## Solution

Build the **Job Search Cockpit**: a local-first, single-user web application (Go backend with SQLite + React frontend) that serves as an active daily command center across the candidate's entire job search lifecycle:

- **Flight Deck**: A daily command center surfacing imminent **Interview Rounds** (within 48 hours), completed **Interview Rounds** missing a **Debrief**, due **Action Items**, and **Opportunities** flagged by stage-specific staleness rules.
- **Orthogonal Pipeline Tracking**: Tracks **Companies**, **Opportunities**, and **Job Postings** across 6 company-agnostic macro **Stages** (`Discovered`, `Outreach`, `Applied`, `Screening`, `Interviewing`, `Offer`) kept orthogonal to **Outcome** (`Active`, `On Hold`, `Won`, `Declined`, `Rejected`, `Withdrawn`, `Ghosted`), automatically recording an immutable **Stage Transition** log on every state change.
- **Quick Capture + Pluggable AI (OpenRouter-First)**: Captures roles from a URL, pasted text, or manual entry, extracting structured **Job Posting** metadata via ATS/JSON-LD parsers and a pluggable LLM client (with first-class **OpenRouter** model discovery, bookmarked models, and separate fast-extraction vs. reasoning/tailoring model slots, plus full manual fallback).
- **Candidate Profile, Story Bank & Tailoring**: Maintains a canonical **Candidate Profile** and a **Story Bank** of structured STAR **Stories** to power per-Opportunity **Fit Assessments** and **Application Artifacts** (in-app Markdown editor with clean print-to-PDF and external PDF file attachment).
- **Interview Loop Intelligence**: Tracks company-specific **Interview Rounds** in any sequence under an **Opportunity**, linking interviewer **Contacts**, planned/used **Stories** (with automatic loop-repeat warnings when a **Story** was already used in an earlier round of the same **Opportunity**), and structured post-interview **Debriefs**.
- **Networking CRM & Offers**: Manages a global **Contact** directory joined to **Opportunities** by role (`Referrer`, `Recruiter`, `Hiring Manager`, `Interviewer`, `Champion`), logs timestamped **Interactions**, tracks **Action Items**, and evaluates structured **Offers** side by side.

## User Stories

### Flight Deck & Daily Execution
1. As a candidate, I want to open the **Flight Deck** as my default home view, so that I immediately see what needs my attention today without hunting through boards.
2. As a candidate, I want any **Interview Round** scheduled within the next 48 hours pinned at the top of the **Flight Deck**, so that I can jump straight into its preparation workspace and linked **Stories**.
3. As a candidate, I want completed **Interview Rounds** that do not yet have a **Debrief** surfaced on the **Flight Deck**, so that I record questions asked and signals while my memory is fresh.
4. As a candidate, I want `Active` **Opportunities** in `Discovered` for more than 3 days with no activity flagged on the **Flight Deck**, so that I decide whether to request a referral, apply, or archive before the role ages.
5. As a candidate, I want `Active` **Opportunities** in `Outreach` with no **Interaction** for more than 4 days flagged on the **Flight Deck**, so that I follow up with my **Contact** or apply directly.
6. As a candidate, I want `Active` **Opportunities** in `Applied` with no activity for more than 10 days flagged on the **Flight Deck**, so that I send a recruiter nudge or mark the **Opportunity** as `Ghosted`.
7. As a candidate, I want `Active` **Opportunities** in `Screening` or `Interviewing` where 5 days have elapsed since the last completed **Interview Round** with no next round scheduled flagged on the **Flight Deck**, so that I follow up on next steps.
8. As a candidate, I want to create, complete, snooze, or dismiss manual and suggested **Action Items** with due dates linked to an **Opportunity** or **Contact**, so that no follow-up slips through the cracks.

### Quick Capture & Job Posting Ingestion
9. As a candidate, I want to open a **Quick Capture** modal from anywhere in the cockpit to create a new **Opportunity** from a **Job Posting** URL, raw pasted text, or manual fields, so that capturing a role takes seconds.
10. As a candidate, I want the cockpit to automatically fetch and parse structured metadata when I provide a public **Job Posting** URL (including Greenhouse, Lever, Ashby, and JSON-LD `JobPosting` pages), so that I don't have to copy-paste every field manually.
11. As a candidate, I want the **Quick Capture** flow to seamlessly fall back to extracting from raw pasted text when a job board blocks automated fetching (such as LinkedIn or Workday), so that blocked URLs never stop me from capturing a role.
12. As a candidate, I want the LLM extractor to populate the **Company** name, role title, seniority level, location/remote policy, stated compensation range, tech stack tags, and key requirements from a **Job Posting**, so that the **Opportunity** starts rich and structured.
13. As a candidate, I want to create an **Opportunity** at a **Company** even before a formal **Job Posting** exists (such as a warm intro or exploratory founder chat) and attach a **Job Posting** later if one opens, so that unposted leads are tracked alongside formal applications.
14. As a candidate, I want to pursue multiple distinct **Opportunities** at the same **Company** simultaneously or sequentially, so that applying to two teams at one organization never causes data collisions.

### Pipeline Board, Stages, Outcomes & Transitions
15. As a candidate, I want to view and move `Active` **Opportunities** across 6 company-agnostic macro **Stages** (`Discovered`, `Outreach`, `Applied`, `Screening`, `Interviewing`, `Offer`) on a Kanban board, so that I have a clean visual overview of my active funnel.
16. As a candidate, I want to update an **Opportunity**'s **Outcome** (`Active`, `On Hold`, `Won`, `Declined`, `Rejected`, `Withdrawn`, `Ghosted`) independently of its **Stage**, so that closing an **Opportunity** preserves the exact **Stage** it reached.
17. As a candidate, I want every change to an **Opportunity**'s **Stage** or **Outcome** to automatically record an immutable **Stage Transition** with the previous state, new state, timestamp, and optional context note, so that I have a complete historical timeline and accurate funnel analytics.
18. As a candidate, I want to switch between the Kanban board and a filterable table view of **Opportunities** (filterable by **Stage**, **Outcome**, priority, remote policy, **Company**, and text search), so that I can inspect closed, paused, or filtered subsets of my search.

### Candidate Profile, Story Bank & Fit Assessment
19. As a candidate, I want to maintain a canonical **Candidate Profile** containing my work history, quantified impact bullets, skills, target roles, location preferences, and compensation floor, so that all AI analysis and tailoring uses a single source of truth.
20. As a candidate, I want to build and search a **Story Bank** of structured STAR **Stories** (`Title`, `Hook`, `Situation`, `Task`, `Action`, `Result`, `Competency Tags`, and `Skill Tags`), so that my best career narratives are organized and ready for any interview or application.
21. As a candidate, I want to generate or edit a structured **Fit Assessment** for any **Opportunity** comparing its **Job Posting** against my **Candidate Profile** and **Story Bank**, so that I can see matched strengths, requirement gaps, compensation alignment, and recommended **Stories** to emphasize.

### Application Artifacts & Tailoring
22. As a candidate, I want to draft, edit, and store role-specific **Application Artifacts** (tailored resume, cover letter, and screening question responses) inside an **Opportunity**'s workspace, either assisted by the configured LLM or written manually, so that all materials for that pursuit live in one place.
23. As a candidate, I want to preview and export any Markdown **Application Artifact** via a clean, print-ready ATS-friendly PDF view directly from the browser, so that I can produce a polished submission without leaving the cockpit.
24. As a candidate, I want to upload and attach external PDF files (such as a resume compiled in LaTeX, Typst, or Word) to an **Opportunity** as an **Application Artifact**, so that I always have the exact file I submitted on record.

### Interview Rounds, Loop Story De-duplication & Debriefs
25. As a candidate, I want to add ordered **Interview Rounds** of any type (`Recruiter Screen`, `Founder Chat`, `Hiring Manager`, `Take-Home`, `Technical`, `System Design`, `Behavioral`, `Team Match`, `Executive`, or custom label) to an **Opportunity**, so that any company's unique interview sequence is accurately represented.
26. As a candidate, I want to link participating interviewer **Contacts** and planned/used **Stories** from my **Story Bank** to each **Interview Round**, so that I know who I am speaking with and which examples I plan to share.
27. As a candidate, I want the cockpit to display a clear warning badge when I select or view a **Story** on an **Interview Round** that was already used in an earlier **Interview Round** of the same **Opportunity**, so that I don't accidentally repeat the same story to multiple interviewers in the same loop.
28. As a candidate, I want to record a structured **Debrief** for each completed **Interview Round** (capturing exact questions asked, what went well, areas to improve, company/team signals, confidence score, and follow-up actions), so that I build institutional memory from every interview.
29. As a candidate, I want questions recorded across my **Debriefs** to be browsable alongside my **Story Bank**, so that I can prepare for recurring interview questions over time.

### Contacts CRM & Interactions
30. As a candidate, I want to maintain a global directory of **Contacts** (with optional current **Company** affiliation, title, email, LinkedIn URL, warmth level, and notes), so that relationships persist even when a person works across companies or refers me to multiple roles.
31. As a candidate, I want to link **Contacts** to an **Opportunity** with an explicit role (`Referrer`, `Recruiter`, `Hiring Manager`, `Interviewer`, `Champion`, `Other`), so that I can see every stakeholder involved in that pursuit.
32. As a candidate, I want to log timestamped **Interactions** (email, DM, referral request, coffee chat, phone call) against a **Contact** and/or **Opportunity** and optionally schedule a follow-up **Action Item** in the same step, so that my networking outreach stays organized and resets staleness timers accurately.

### Offers & Funnel Analytics
33. As a candidate, I want to record a structured **Offer** on an **Opportunity** (base salary, target bonus, equity/RSUs and vesting notes, sign-on bonus, benefits notes, decision deadline, and negotiation log), so that I can calculate annualized total compensation and track negotiation progress.
34. As a candidate, I want to compare multiple **Offers** side by side against each other and against my **Candidate Profile** compensation floor, so that I can make an informed decision.
35. As a candidate, I want to view funnel analytics computed from my **Stage Transitions** and **Outcomes** (stage-to-stage conversion rates, drop-off by stage, average days spent per stage, and active pipeline counts), so that I can diagnose bottlenecks in my job search.

### Pluggable AI (OpenRouter-First), Local Data & Settings
36. As a candidate, I want to configure **OpenRouter** as a first-class AI provider (alongside direct OpenAI, Anthropic, Gemini, and local OpenAI-compatible endpoints like Ollama) via environment variables or the in-app **Settings** view, so that I can use any model through one unified setup.
37. As a candidate, I want to search the live **OpenRouter** model catalog from the cockpit, bookmark favorite models, configure separate default models for **Fast Extraction** vs. **Reasoning / Tailoring**, and switch models on the fly when triggering an AI action, so that I can balance speed, cost, and writing quality.
38. As a candidate, I want every feature of the cockpit to work 100% manually when no LLM API key is configured, so that I am never blocked by API availability or offline use.
39. As a candidate, I want all SQLite data and uploaded PDF **Application Artifacts** stored locally in a gitignored data directory, with one-click JSON/SQLite backup export and an optional sample-data seed action, so that my private career data stays under my control.

## Implementation Decisions

### Architectural Shape & Packaging
- **Single-user local-first runtime (ADR-0001)**: The application consists of a Go HTTP backend (`net/http`) and a React + TypeScript single-page frontend (Vite, Tailwind CSS, TanStack Query, React Router). During development, Vite proxies `/api/*` requests to the Go backend. For production/local binary usage, the Go server can serve the compiled frontend static assets via `embed.FS`.
- **Local storage layout**: All mutable state lives under a configurable local data directory (defaulting to `./data/`, gitignored): a single SQLite database file (`cockpit.db`) using the pure-Go `modernc.org/sqlite` driver (with WAL mode and foreign keys enabled) and an `artifacts/` directory for uploaded binary files (such as external PDF resumes).
- **Automatic schema migrations**: Embedded SQL migrations run automatically on Go server startup to keep the SQLite schema up to date.

### Core Domain Modules & Schema Invariants
- **Companies & Opportunities**:
  - `companies` stores name, domain/website, headquarters/remote policy, size/stage, notes, and timestamps.
  - `opportunities` references `company_id`, holds role title, seniority, location/remote policy, priority, `stage` (`Discovered`, `Outreach`, `Applied`, `Screening`, `Interviewing`, `Offer`), `outcome` (`Active`, `On Hold`, `Won`, `Declined`, `Rejected`, `Withdrawn`, `Ghosted`), last activity timestamp, and embedded/linked **Job Posting** fields (source URL, raw text, parsed requirements, tech stack tags, stated min/max compensation, currency).
- **Orthogonal Stage & Outcome + Stage Transitions (ADR-0002)**:
  - Updating an **Opportunity**'s `stage` or `outcome` happens inside a single database transaction that updates the `opportunities` row, bumps `last_activity_at` (when appropriate), and inserts an immutable row into `stage_transitions` (`opportunity_id`, `from_stage`, `to_stage`, `from_outcome`, `to_outcome`, `transitioned_at`, `note`).
  - Creating a new **Opportunity** records its initial `stage_transitions` entry (`from_stage = NULL`, `to_stage = initial_stage`, `to_outcome = 'Active'`).
- **Flight Deck Staleness & Triage Engine**:
  - The **Flight Deck** module computes actionable items deterministically using an injected clock and configurable staleness thresholds (defaults: `Discovered` > 3 days, `Outreach` > 4 days since last activity/interaction, `Applied` > 10 days, `Screening` / `Interviewing` > 5 days after the latest completed **Interview Round** with no upcoming round scheduled).
  - It also returns: (a) **Interview Rounds** scheduled within the next 48 hours, (b) past **Interview Rounds** lacking a **Debrief**, and (c) open **Action Items** ordered by due date.
- **Candidate Profile, Story Bank & Loop Story De-duplication**:
  - `candidate_profile` stores the singleton record of experience entries, skills, target titles, location preferences, compensation floor/target, and summary.
  - `stories` stores reusable STAR narratives (`title`, `hook`, `situation`, `task`, `action`, `result`, JSON arrays for `competency_tags` and `skill_tags`).
  - `interview_round_stories` links `interview_round_id` and `story_id`. When querying or attaching **Stories** for an **Opportunity**'s **Interview Rounds**, the backend computes a `used_in_earlier_rounds` list (referencing the round title/sequence within the same `opportunity_id`) so the UI immediately flags duplicate story usage across the same company's interview loop.
- **Fit Assessments & Application Artifacts**:
  - `fit_assessments` stores structured match evaluations per **Opportunity** (overall score/verdict, matched strengths, gaps/risks, compensation alignment note, recommended `story_ids`, and model used).
  - `application_artifacts` stores role-specific materials per **Opportunity** (`kind`: `resume`, `cover_letter`, `screening_qa`, `other`; `title`; `content_markdown`; optional `file_path` and `mime_type` for uploaded PDFs; `is_submitted` flag).
- **Interview Rounds & Debriefs**:
  - `interview_rounds` belongs to an **Opportunity** (`round_order`, `round_type`, `title`, `scheduled_at`, `duration_minutes`, `status`, `prep_notes`).
  - `debriefs` has a 1:1 relationship with `interview_rounds` (`questions_asked` JSON list, `went_well`, `to_improve`, `signals`, `confidence_score` 1–5, `follow_up_notes`).
- **Contacts, Opportunity Roles, Interactions & Action Items**:
  - `contacts` stores global network entries (`name`, optional `company_id`, `title`, `email`, `linkedin_url`, `warmth`, `notes`).
  - `opportunity_contacts` joins `opportunity_id` and `contact_id` with a `role` (`Referrer`, `Recruiter`, `Hiring Manager`, `Interviewer`, `Champion`, `Other`).
  - `interview_round_contacts` joins `interview_round_id` and `contact_id`.
  - `interactions` logs timestamped touchpoints (`contact_id`, optional `opportunity_id`, `kind`, `direction`, `occurred_at`, `summary`), automatically updating `last_activity_at` on the linked **Opportunity**.
  - `action_items` stores tasks (`title`, optional `opportunity_id`, optional `contact_id`, `due_at`, `status`: `open` | `completed` | `dismissed`).
- **Offers**:
  - `offers` belongs to an **Opportunity** (`base_salary`, `bonus_target_pct`, `bonus_amount`, `equity_type`, `equity_total_value`, `equity_vesting_years`, `sign_on_bonus`, `currency`, `deadline_at`, `benefits_notes`, `negotiation_notes`, `status`), exposing computed annualized Year-1 and recurring total compensation.
- **Pluggable AI & OpenRouter Integration Module**:
  - A single deep `LLMClient` interface in Go supports `OpenRouter` (`https://openrouter.ai/api/v1`), `OpenAI`, `Anthropic`, `Gemini`, and custom OpenAI-compatible base URLs (e.g., local Ollama).
  - Settings are persisted in a `settings` table (with environment variable fallbacks such as `OPENROUTER_API_KEY`), including `provider`, `fast_model`, `reasoning_model`, and `bookmarked_models`.
  - A dedicated endpoint proxies/caches `GET https://openrouter.ai/api/v1/models` so the frontend can search and bookmark OpenRouter models, and every AI endpoint (`extract-job-posting`, `assess-fit`, `draft-artifact`) accepts an optional `model` override parameter.
- **Job Posting URL Fetcher & Parser Module**:
  - Given a URL, the backend fetches the page with a timeout and attempts structured extraction in priority order: (1) JSON-LD `<script type="application/ld+json">` matching `@type: JobPosting`, (2) known ATS structures (Greenhouse, Lever, Ashby), and (3) readable text extraction from the HTML body, before passing the extracted text (plus any user-pasted text) to the configured `fast_model` for structured JSON extraction.

### Frontend Workspace Structure
- Six top-level workspaces sharing a persistent sidebar and global **Quick Capture** (`Cmd+K` / button) modal:
  1. **Flight Deck (`/`)**
  2. **Pipeline (`/opportunities`)** — Kanban board by **Stage** + filterable table view
  3. **Opportunity Workspace (`/opportunities/:id`)** — Tabbed/sectioned deep view for **Job Posting**, **Fit Assessment**, **Application Artifacts** (including browser print-to-PDF view), **Interview Rounds** & **Debriefs**, **Contacts** & **Interactions**, **Offer**, and **Stage Transitions**
  4. **Network & Companies (`/network`)** — **Companies** and **Contacts** CRM with **Interactions**
  5. **Candidate Profile & Story Bank (`/profile`)** — **Candidate Profile**, **Story Bank**, and interview questions aggregated from **Debriefs**
  6. **Analytics, Offers & Settings (`/analytics`, `/settings`)** — Funnel conversion metrics, side-by-side **Offer** comparison, AI/OpenRouter configuration, backup export, and sample data seeding

## Testing Decisions

### Philosophy
Tests verify observable behavior through public interfaces at pre-agreed seams—never internal implementation details, private helpers, or tautological assertions.

### Pre-Agreed Seams
1. **Primary Backend Seam — Go HTTP Handler (`http.Handler` over `/api/*` with real SQLite)**:
   - Backend integration tests instantiate the full server `http.Handler` backed by a fresh real SQLite database (with all migrations applied) and inject deterministic test adapters only at true external boundaries:
     - `Clock` (for deterministic **Flight Deck** staleness thresholds and 48-hour **Interview Round** windows)
     - `LLMProvider` (for deterministic **Job Posting** extraction, **Fit Assessment**, and **Application Artifact** drafting tests, plus verifying per-request `model` overrides)
     - `PageFetcher` (for testing JSON-LD and ATS HTML parsing without live network calls)
   - All domain behaviors (**Stage** vs. **Outcome** orthogonality, automatic **Stage Transition** creation, **Story** repeat-in-loop warnings, **Flight Deck** staleness computation, **Interaction** activity-timestamp updates, **Offer** annualized total-comp calculation, and **Analytics** conversion rates) are verified through HTTP requests and responses at this seam.
2. **Frontend Seam — Top-Level Workspace Views (`@testing-library/react` + Vitest over `/api/*`)**:
   - Frontend tests render the workspace views with a real router and query client against stubbed `/api/*` HTTP handlers, asserting user-visible text, roles, form submissions, filtering, and warning badges.

## Out of Scope

- Multi-user authentication, user accounts, billing, or multi-tenant cloud deployment (per ADR-0001).
- Automated background scraping/crawling of job boards without user-initiated **Quick Capture**.
- Two-way OAuth syncing with Gmail or Google Calendar.
- Native mobile (iOS/Android) applications.
- Custom per-company Kanban board columns (per ADR-0002, macro **Stages** are fixed and company-agnostic while company-specific steps live as **Interview Rounds**).

## Further Notes

- All domain terminology in code, API payloads, UI labels, and tests must follow `GLOSSARY.md` (`Company`, `Opportunity`, `Job Posting`, `Stage`, `Outcome`, `Stage Transition`, `Offer`, `Candidate Profile`, `Story`, `Story Bank`, `Fit Assessment`, `Application Artifact`, `Interview Round`, `Debrief`, `Contact`, `Interaction`, `Action Item`, `Flight Deck`).
- Architectural constraints in `docs/adr/0001-local-first-single-user-architecture.md` and `docs/adr/0002-orthogonal-stage-and-outcome.md` govern the implementation.
