# 05: Interview Rounds, Debriefs & Loop Story De-duplication

**What to build:** Allow the candidate to model any company's interview sequence as ordered **Interview Rounds** inside an **Opportunity**, link interviewer **Contacts** and planned/used **Stories** from the **Story Bank** with automatic loop-repeat warnings if a **Story** was already used in an earlier round of the same **Opportunity**, and capture structured post-interview **Debriefs** whose recorded questions populate a browsable question bank.

**Blocked by:** 01: Pipeline Spine — Companies, Opportunities, Orthogonal Stage/Outcome & Stage Transitions, 03: Candidate Profile & Story Bank Workspace, 04: Contacts CRM, Opportunity Roles, Interactions & Action Items

**Status:** ready-for-agent

- [ ] Inside an **Opportunity** workspace, the candidate can add, reorder, edit, and update the status of **Interview Rounds** of any type (`Recruiter Screen`, `Founder Chat`, `Hiring Manager`, `Take-Home`, `Technical`, `System Design`, `Behavioral`, `Team Match`, `Executive`, or custom title) with scheduled date/time, duration, and preparation notes.
- [ ] The candidate can attach participating interviewer **Contacts** and planned/used **Stories** from the **Story Bank** to each **Interview Round**.
- [ ] When viewing or selecting **Stories** for an **Interview Round**, the pure domain evaluator (`domain.DetectRepeatedStories`) automatically detects and returns a non-blocking warning badge if that **Story** is already linked to an earlier non-cancelled **Interview Round** (`Scheduled` or `Completed`) within the same **Opportunity** (indicating which round used it), while not flagging usage across different **Opportunities** or cancelled rounds.
- [ ] For any **Interview Round**, the candidate can record and update a structured 1:1 **Debrief** capturing structured questions asked (`Question`, optional `Category`, optional `AnsweredWithStoryID`, `Notes`), what went well, areas to improve, company/team signals, a 1–5 confidence score, and follow-up notes.
- [ ] Structured questions recorded across **Debriefs** are aggregated and browsable alongside the **Story Bank** (displaying the **Company**, **Opportunity**, **Interview Round** type, and linked **Story**) so the candidate can review real interview questions encountered across their search.
- [ ] Pure `internal/domain` unit tests (`DetectRepeatedStories`), backend HTTP handler tests, and frontend workspace tests verify **Interview Round** ordering, **Contact** and **Story** linking, same-Opportunity **Story** repeat detection (including ignoring `Cancelled` rounds), **Debrief** persistence, and question bank aggregation.
