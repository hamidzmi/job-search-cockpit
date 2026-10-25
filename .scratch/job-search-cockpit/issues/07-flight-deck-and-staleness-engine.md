# 07: Flight Deck Daily Command Center & Staleness Engine

**What to build:** Deliver the **Flight Deck** home view (`/`) that acts as the candidate's daily command center, automatically surfacing imminent **Interview Rounds**, completed **Interview Rounds** missing a **Debrief**, due **Action Items**, and `Active` **Opportunities** that have exceeded stage-specific staleness thresholds, with inline actions to resolve each item.

**Blocked by:** 04: Contacts CRM, Opportunity Roles, Interactions & Action Items, 05: Interview Rounds, Debriefs & Loop Story De-duplication

**Status:** ready-for-agent

- [ ] Opening the **Flight Deck** (`/`) displays all **Interview Rounds** scheduled within the next 48 hours pinned at the top, showing the **Company**, **Opportunity**, round type/time, linked interviewer **Contacts**, linked **Stories**, and a one-click link into the round's preparation view.
- [ ] The **Flight Deck** surfaces any past/completed **Interview Round** that does not yet have a recorded **Debrief**, prompting the candidate to log their reflection while fresh.
- [ ] The **Flight Deck** lists open **Action Items** ordered by due date (highlighting overdue and due-today items) and allows completing or snoozing them inline.
- [ ] The **Flight Deck** uses the pure domain evaluator (`domain.EvaluateFlightDeck`) with an injected `Clock` to deterministically flag `Active` **Opportunities** that exceed stage-based staleness thresholds:
  - `Discovered` > 3 days measured from `max(LatestStageTransitionAt, LastActivityAt)`
  - `Outreach` > 4 days measured from `max(LatestStageTransitionAt, LatestInteractionAt)`
  - `Applied` > 10 days measured from `max(LatestStageTransitionAt, LastActivityAt)`
  - `Screening` or `Interviewing` > 5 days after the latest completed **Interview Round** (`ScheduledAt`) with no upcoming **Interview Round** scheduled
- [ ] From any staleness nudge on the **Flight Deck**, the candidate can immediately log an **Interaction** (which advances `LastActivityAt` and resets the staleness timer), advance the **Stage**, or update the **Outcome** (e.g., mark `Ghosted` or `Withdrawn`).
- [ ] Pure `internal/domain` unit tests (`EvaluateFlightDeck`), backend HTTP handler tests (using an injected deterministic `Clock`), and frontend **Flight Deck** tests verify all four staleness rules, 48-hour upcoming interview pinning, missing **Debrief** alerts, and inline resolution actions.
