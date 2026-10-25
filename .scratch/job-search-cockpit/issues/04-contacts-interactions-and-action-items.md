# 04: Contacts CRM, Opportunity Roles, Interactions & Action Items

**What to build:** Enable the candidate to manage a global directory of **Contacts**, link **Contacts** to specific **Opportunities** with explicit stakeholder roles (`Referrer`, `Recruiter`, `Hiring Manager`, `Interviewer`, `Champion`, `Other`), log timestamped **Interactions** that keep opportunity activity fresh, and track due-dated **Action Items**.

**Blocked by:** 01: Pipeline Spine — Companies, Opportunities, Orthogonal Stage/Outcome & Stage Transitions

**Status:** ready-for-agent

- [ ] In the **Network & Companies** workspace, the candidate can create, edit, search, and filter global **Contacts** (name, optional current **Company**, title, email, LinkedIn URL, warmth level, and notes) and inspect all **Opportunities** and **Interactions** associated with each **Contact**.
- [ ] Inside an **Opportunity** workspace (and from the **Contact** view), the candidate can link a **Contact** to an **Opportunity** with an explicit role (`Referrer`, `Recruiter`, `Hiring Manager`, `Interviewer`, `Champion`, `Other`), supporting scenarios where one **Contact** is linked to multiple **Opportunities** or multiple **Contacts** play different roles on one **Opportunity**.
- [ ] The candidate can log timestamped **Interactions** (email, DM, referral request, coffee chat, phone call, inbound/outbound direction, summary) linked to a **Contact** and optionally an **Opportunity**, which atomically (via `TxManager`) advances the linked **Opportunity**'s `LastActivityAt` timestamp and optionally schedules a follow-up **Action Item**.
- [ ] The candidate can create, edit, complete, and dismiss **Action Items** with due dates (optionally linked to an **Opportunity** and/or **Contact**), including a quick option to schedule a follow-up **Action Item** right when logging an **Interaction**.
- [ ] Pure `internal/domain` unit tests, backend HTTP handler tests, and frontend workspace tests verify global **Contact** management, multi-Opportunity role linking, atomic **Interaction** logging with **Opportunity** `LastActivityAt` updates, and **Action Item** lifecycle.
