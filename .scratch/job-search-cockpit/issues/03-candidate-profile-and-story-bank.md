# 03: Candidate Profile & Story Bank Workspace

**What to build:** Provide a dedicated **Candidate Profile & Story Bank** workspace where the candidate maintains their canonical background (work history, quantified achievements, skills, target preferences, and compensation floor) alongside a searchable library of structured STAR **Stories** tagged by competency and skill.

**Blocked by:** 01: Pipeline Spine — Companies, Opportunities, Orthogonal Stage/Outcome & Stage Transitions

**Status:** ready-for-agent

- [ ] In the **Candidate Profile & Story Bank** workspace, the candidate can view and save their canonical **Candidate Profile**, including summary, work history entries with impact bullets, skills, target role titles, location/remote preferences, and minimum/target compensation floor.
- [ ] The candidate can create, edit, and delete structured **Stories** in the **Story Bank**, capturing `Title`, `Hook`, `Situation`, `Task`, `Action`, `Result` (quantified impact), `Competency Tags` (e.g., Leadership, System Design, Conflict, Ownership), and `Skill Tags`.
- [ ] The candidate can search and filter the **Story Bank** by keyword, `Competency Tags`, and `Skill Tags` to quickly locate relevant narratives during application prep or live interviews.
- [ ] Backend HTTP handler tests and frontend workspace tests verify **Candidate Profile** persistence and **Story Bank** CRUD, tagging, and search filtering.
