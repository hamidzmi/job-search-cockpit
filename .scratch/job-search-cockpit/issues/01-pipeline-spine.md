# 01: Pipeline Spine — Companies, Opportunities, Orthogonal Stage/Outcome & Stage Transitions

**What to build:** Stand up the local-first Go + SQLite backend and React + TypeScript frontend application shell with persistent sidebar navigation, enabling the candidate to create and manage **Companies** and **Opportunities** (with or without a formal **Job Posting**), track active pursuits across the 6 company-agnostic macro **Stages** (`Discovered`, `Outreach`, `Applied`, `Screening`, `Interviewing`, `Offer`) on a Kanban board, filter all **Opportunities** across every **Outcome** (`Active`, `On Hold`, `Won`, `Declined`, `Rejected`, `Withdrawn`, `Ghosted`) in a table view, and inspect the immutable **Stage Transition** timeline inside an **Opportunity** workspace.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] Starting the application runs automatic SQLite schema migrations and exposes the HTTP API seam and React workspace navigation (`Flight Deck`, `Pipeline`, `Network & Companies`, `Candidate Profile & Story Bank`, `Analytics & Offers`, `Settings`).
- [ ] A candidate can create and edit a **Company** (name, website/domain, headquarters, remote policy, size/stage, notes) and create multiple distinct **Opportunities** under the same **Company** without data collisions.
- [ ] A candidate can create an **Opportunity** either with structured **Job Posting** details (role title, seniority, location/remote policy, stated compensation range, tech stack tags, requirements, source URL, raw text) or as an unposted pursuit where a **Job Posting** can be added later.
- [ ] The **Pipeline** Kanban board displays `Active` **Opportunities** grouped into the 6 macro **Stages** (`Discovered`, `Outreach`, `Applied`, `Screening`, `Interviewing`, `Offer`) and lets the candidate move an **Opportunity** between **Stages**.
- [ ] The **Pipeline** table view allows filtering **Opportunities** by **Stage**, **Outcome**, priority, remote policy, **Company**, and text search.
- [ ] Updating an **Opportunity**'s **Outcome** (`Active`, `On Hold`, `Won`, `Declined`, `Rejected`, `Withdrawn`, `Ghosted`) preserves its current **Stage** rather than overwriting it.
- [ ] Every **Opportunity** creation and every change to **Stage** or **Outcome** atomically appends an immutable **Stage Transition** record (`from_stage`, `to_stage`, `from_outcome`, `to_outcome`, `transitioned_at`, optional `note`) that is displayed chronologically in the **Opportunity** detail workspace.
- [ ] Backend HTTP handler tests (against real SQLite) and frontend view tests verify **Company** and **Opportunity** creation, orthogonal **Stage**/**Outcome** updates, board/table filtering, and **Stage Transition** history.
