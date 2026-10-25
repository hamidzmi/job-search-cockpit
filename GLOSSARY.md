# Job Search Cockpit

A personal workspace for managing a candidate's job search across opportunity tracking, application tailoring, networking outreach, and interview preparation.

## Language

### Pipeline & Core Entities

**Company**:
An organization that employs or may employ the candidate.
_Avoid_: Employer, account, organization, client

**Opportunity**:
A distinct pursuit of a role at a Company, tracked through pipeline stages from initial lead or interest to a final outcome, whether or not a formal Job Posting exists yet.
_Avoid_: Application, lead, deal, card, position

**Job Posting**:
A published description of an open role from a Company or job board, including its stated requirements, compensation range, and source URL.
_Avoid_: Job description, JD, listing, req, requisition

**Stage**:
The furthest macro funnel milestone an Opportunity has reached (`Discovered`, `Outreach`, `Applied`, `Screening`, `Interviewing`, `Offer`), kept company-agnostic while company-specific interview steps are tracked as Interview Rounds.
_Avoid_: Status, column, phase, state

**Outcome**:
The resolution state of an Opportunity (`Active`, `On Hold`, `Won`, `Declined`, `Rejected`, `Withdrawn`, `Ghosted`), kept separate from Stage so funnel progress is preserved when an Opportunity closes.
_Avoid_: Stage, result, disposition

**Stage Transition**:
An immutable historical record created whenever an Opportunity's Stage or Outcome changes, capturing the prior state, new state, timestamp, and context note for timeline and funnel conversion analysis.
_Avoid_: Audit log, status change, history event

**Offer**:
A formal compensation package proposed by a Company for an Opportunity, breaking down base salary, bonus, equity, sign-on bonus, decision deadline, and negotiation history.
_Avoid_: Deal, package, contract, proposal

### Candidate & Tailoring

**Candidate Profile**:
The canonical record of the candidate's work history, achievements, skills, compensation floor, and role preferences used as the source of truth for fit analysis and tailoring.
_Avoid_: Master resume, bio, user profile

**Story**:
A single structured career narrative (capturing Situation, Task, Action, Result, and competency tags) stored in the Story Bank and linked to the Interview Rounds where it is used.
_Avoid_: Example, anecdote, snippet, case study

**Story Bank**:
The collection of reusable Stories maintained by the candidate to prepare for behavioral and technical interviews and prevent accidental story repetition across rounds of the same Opportunity.
_Avoid_: Experience bank, answer templates, snippets

**Fit Assessment**:
A structured evaluation comparing a Job Posting against the Candidate Profile and Story Bank, surfacing matched strengths, skill or requirement gaps, compensation alignment, and recommended Stories to emphasize.
_Avoid_: Match score, resume check, compatibility report

**Application Artifact**:
A role-specific document or set of responses prepared for an Opportunity, such as a tailored resume, cover letter, or screening question answers.
_Avoid_: Attachment, submission, asset

### Interviewing

**Interview Round**:
A scheduled or completed evaluation session within an Opportunity (e.g., Recruiter Screen, Technical, System Design, Behavioral), linked to participating Contacts, planned Stories, and preparation materials.
_Avoid_: Interview, meeting, event, stage

**Debrief**:
A structured post-interview reflection recorded for an Interview Round, capturing questions asked, signals observed, self-assessment, and follow-up actions.
_Avoid_: Post-mortem, review, feedback, recap

### Networking & Execution

**Contact**:
A person associated with a Company or Opportunity, such as a recruiter, hiring manager, interviewer, or referral connection.
_Avoid_: Lead, connection, person, stakeholder

**Interaction**:
A logged communication or touchpoint with a Contact or regarding an Opportunity, such as an email, DM, referral request, or coffee chat.
_Avoid_: Activity, message, event, touch

**Action Item**:
A concrete task with a due date (manual or suggested by staleness/stage rules) surfaced on the Flight Deck to keep an Opportunity or Contact relationship moving forward.
_Avoid_: Todo, reminder, ticket, notification

**Flight Deck**:
The daily command-center view that surfaces imminent Interview Rounds, missing Debriefs, due Action Items, pending referral follow-ups, and stale Opportunities requiring attention.
_Avoid_: Dashboard, home page, inbox
