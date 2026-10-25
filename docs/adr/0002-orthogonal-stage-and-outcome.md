# Orthogonal Stage and Outcome with Immutable Stage Transitions

Each Opportunity tracks funnel progress (`Stage`) separately from resolution state (`Outcome`), and every change to either field appends an immutable `Stage Transition` record.

## Context & Trade-offs

In a single-column status model, moving an Opportunity to `Rejected`, `Withdrawn`, or `Ghosted` overwrites the funnel stage it reached (`Applied` vs. `Screening` vs. `Interviewing`), destroying conversion-rate and drop-off analytics. Keeping company-agnostic macro `Stage` (`Discovered`, `Outreach`, `Applied`, `Screening`, `Interviewing`, `Offer`) orthogonal to `Outcome` (`Active`, `On Hold`, `Won`, `Declined`, `Rejected`, `Withdrawn`, `Ghosted`)—while tracking company-specific evaluation steps as ordered `Interview Round` records—keeps the active pipeline board clean and comparable across companies while preserving full stage-duration and funnel metrics via `Stage Transition` records.
