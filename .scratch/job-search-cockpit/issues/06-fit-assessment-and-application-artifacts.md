# 06: Fit Assessment & Application Artifacts (Tailoring, Print-to-PDF & PDF Upload)

**What to build:** Enable the candidate to evaluate an **Opportunity** with a structured **Fit Assessment** against their **Candidate Profile** and **Story Bank**, draft and edit tailored Markdown **Application Artifacts** (resumes, cover letters, screening Q&A) with AI assistance or manually, export a clean ATS-friendly PDF from the browser, and upload external PDF files as submitted artifacts.

**Blocked by:** 01: Pipeline Spine — Companies, Opportunities, Orthogonal Stage/Outcome & Stage Transitions, 02: Pluggable AI Settings, OpenRouter Model Catalog & Quick Capture Ingestion, 03: Candidate Profile & Story Bank Workspace

**Status:** ready-for-agent

- [ ] Inside an **Opportunity** workspace, the candidate can generate (via the `AssistantGateway` port using the configured **Reasoning / Tailoring** model or a selected bookmarked model) or manually edit a structured **Fit Assessment** that compares the **Job Posting** against the **Candidate Profile** and **Story Bank**, surfacing an overall fit summary/score, matched strengths, requirement gaps/risks, compensation alignment against the candidate's floor, and recommended **Stories** to emphasize.
- [ ] Inside an **Opportunity** workspace, the candidate can create, edit, and mark as submitted (`IsSubmitted`) role-specific **Application Artifacts** (`resume`, `cover_letter`, `screening_qa`, `other`) using an in-app Markdown editor, with optional LLM-assisted drafting/tailoring grounded in the **Candidate Profile**, **Story Bank**, and **Job Posting**.
- [ ] Any Markdown **Application Artifact** can be opened in a clean, print-styled ATS-friendly preview mode ready for browser print-to-PDF export.
- [ ] The candidate can upload an external PDF file (such as a resume compiled externally) to an **Application Artifact** on an **Opportunity** stored via the `ArtifactBlobStore` port (`internal/adapter/fs`), view/download it from the **Opportunity** workspace, and mark which artifact was submitted.
- [ ] Backend HTTP handler tests (including multipart PDF upload/download via `ArtifactBlobStore` and injected `AssistantGateway`) and frontend workspace tests verify **Fit Assessment** generation/editing, Markdown **Application Artifact** drafting, print preview mode, and external PDF attachment.
