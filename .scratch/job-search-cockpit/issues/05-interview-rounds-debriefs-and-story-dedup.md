# 05: Interview Rounds, Debriefs & Loop Story De-duplication

**What to build:** Allow the candidate to model any company's interview sequence as ordered **Interview Rounds** inside an **Opportunity**, link interviewer **Contacts** and planned/used **Stories** from the **Story Bank** with automatic loop-repeat warnings if a **Story** was already used in an earlier round of the same **Opportunity**, and capture structured post-interview **Debriefs** whose recorded questions populate a browsable question bank.

**Blocked by:** 01: Pipeline Spine — Companies, Opportunities, Orthogonal Stage/Outcome & Stage Transitions, 03: Candidate Profile & Story Bank Workspace, 04: Contacts CRM, Opportunity Roles, Interactions & Action Items

**Status:** ready-for-agent

- [ ] Inside an **Opportunity** workspace, the candidate can add, reorder, edit, and update the status of **Interview Rounds** of any type (`Recruiter Screen`, `Founder Chat`, `Hiring Manager`, `Take-Home`, `Technical`, `System Design`, `Behavioral`, `Team Match`, `Executive`, or custom title) with scheduled date/time, duration, and preparation notes.
- [ ] The candidate can attach participating interviewer **Contacts** and planned/used **Stories** from the **Story Bank** to each **Interview Round**.
- [ ] When viewing or selecting **Stories** for an **Interview Round**, the cockpit automatically detects and displays a warning badge if that **Story** is already linked to an earlier **Interview Round** within the same **Opportunity** (indicating which round used it), while not flagging usage across different **Opportunities**.
- [ ] For any **Interview Round**, the candidate can record and update a structured **Debrief** capturing exact questions asked, what went well, areas to improve, company/team signals, a 1–5 confidence score, and follow-up notes.
- [ ] Questions recorded across **Debriefs** are aggregated and browsable alongside the **Story Bank** so the candidate can review real interview questions encountered across their search.
- [ ] Backend HTTP handler tests and frontend workspace tests verify **Interview Round** ordering, **Contact** and **Story** linking, same-Opportunity **Story** repeat detection, **Debrief** persistence, and question bank aggregation.
