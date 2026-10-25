# 02: Pluggable AI Settings, OpenRouter Model Catalog & Quick Capture Ingestion

**What to build:** Enable the candidate to configure AI providers—with first-class **OpenRouter** model discovery, bookmarked favorite models, and separate **Fast Extraction** vs. **Reasoning / Tailoring** model slots—and use a global **Quick Capture** modal to ingest a **Job Posting** from a URL, raw pasted text, or manual entry into a structured **Opportunity**.

**Blocked by:** 01: Pipeline Spine — Companies, Opportunities, Orthogonal Stage/Outcome & Stage Transitions

**Status:** ready-for-agent

- [ ] In **Settings**, the candidate can configure **OpenRouter** (as well as direct OpenAI, Anthropic, Gemini, or an OpenAI-compatible local endpoint such as Ollama), store API credentials/settings locally, and still use the entire application manually when no API key is configured.
- [ ] When using **OpenRouter**, the candidate can search the live OpenRouter model catalog, bookmark favorite models, type any custom model slug, and assign default models for **Fast Extraction** and **Reasoning / Tailoring**.
- [ ] From anywhere in the cockpit, the candidate can open the **Quick Capture** modal and submit a **Job Posting** URL, raw pasted text, or both.
- [ ] When a URL is provided, the backend fetches the page and extracts structured metadata via JSON-LD `JobPosting` schema, known ATS layouts (Greenhouse, Lever, Ashby), or readable page text; if fetching is blocked by the target site, the flow seamlessly falls back to extracting from the user's pasted text.
- [ ] The extraction flow populates structured draft fields (**Company** name, role title, seniority, location/remote policy, stated compensation range, tech stack tags, and key requirements) using the configured **Fast Extraction** model (or an on-the-fly model selected from bookmarked models), allowing the candidate to review/edit before saving the **Opportunity**.
- [ ] Backend HTTP handler tests (with injected `PageFetcher` and `LLMProvider` adapters) and frontend view tests verify OpenRouter model catalog search/bookmarking, URL + pasted text extraction, per-request model override, and manual fallback when no LLM is configured.
