# 08: Offers Comparison, Funnel Analytics, Local Backup & Sample Seeding

**What to build:** Enable the candidate to record and compare structured **Offers** against their **Candidate Profile** compensation floor, analyze funnel conversion rates and stage durations computed from **Stage Transitions**, export a full local data backup, and seed realistic sample data for demo/testing.

**Blocked by:** 01: Pipeline Spine — Companies, Opportunities, Orthogonal Stage/Outcome & Stage Transitions, 03: Candidate Profile & Story Bank Workspace

**Status:** ready-for-agent

- [ ] Inside an **Opportunity** workspace, the candidate can create and update its `0..1` structured **Offer** (base salary, target bonus % or amount, equity type, total equity grant value, vesting years defaulting to 4 when unset, sign-on bonus, currency, decision deadline, benefits notes, negotiation log, and status), with pure `internal/domain` methods calculating `AnnualBonus`, `AnnualEquity`, `RecurringAnnualTotalComp`, `Year1TotalComp`, and `DeltaVsCompFloor`.
- [ ] In the **Analytics & Offers** workspace (`/analytics`), the candidate can compare multiple **Offers** side by side across base, bonus, annualized equity, recurring and Year-1 total compensation, deadline, and delta against the **Candidate Profile** compensation floor.
- [ ] In the **Analytics & Offers** workspace, the candidate can inspect funnel analytics computed by the pure domain service (`domain.ComputeFunnelAnalytics`) over **Opportunities** and **Stage Transitions**: count of opportunities reaching each **Stage**, stage-to-stage conversion percentages, breakdown by **Outcome** at each drop-off **Stage**, and average days spent per **Stage**.
- [ ] In **Settings**, the candidate can trigger a one-click workspace backup export (structured JSON snapshot and SQLite `VACUUM INTO` snapshot via `BackupPort`) and optionally run the transactional sample-data seeder use case (constructing valid domain aggregates across **Companies**, **Opportunities** with **Stage Transitions**, **Stories**, **Contacts**, **Interactions**, **Interview Rounds** with a repeated-story loop warning and missing **Debrief**, **Action Items**, and **Offers** via `TxManager`).
- [ ] Pure `internal/domain` unit tests (`Offer` compensation math and `ComputeFunnelAnalytics`), backend HTTP handler tests, and frontend workspace tests verify annualized **Offer** compensation calculations, side-by-side comparison vs. compensation floor, funnel conversion analytics across **Stage Transitions**, backup export, and sample-data seeding.
