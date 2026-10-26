# 02: Pluggable AI Settings, OpenRouter Model Catalog & Quick Capture Ingestion

**What to build:** Enable the candidate to configure AI providers—with first-class **OpenRouter** model discovery, bookmarked favorite models, and separate **Fast Extraction** vs. **Reasoning / Tailoring** model slots—and use a global **Quick Capture** modal to ingest a **Job Posting** from a URL, raw pasted text, or manual entry into a structured **Opportunity**.

**Blocked by:** 01: Pipeline Spine — Companies, Opportunities, Orthogonal Stage/Outcome & Stage Transitions

**Status:** resolved

- [x] In **Settings**, the candidate can configure **OpenRouter** (as well as direct OpenAI, Anthropic, Gemini, or an OpenAI-compatible local endpoint such as Ollama), store API credentials/settings locally, and still use the entire application manually when no API key is configured.
- [x] When using **OpenRouter**, the candidate can search the live OpenRouter model catalog, bookmark favorite models, type any custom model slug, and assign default models for **Fast Extraction** and **Reasoning / Tailoring**.
- [x] From anywhere in the cockpit, the candidate can open the **Quick Capture** modal and submit a **Job Posting** URL, raw pasted text, or both.
- [x] When a URL is provided, the `JobPageParser` port (`internal/adapter/scraper`) fetches the page and extracts deterministic metadata via JSON-LD `JobPosting` schema or known ATS layouts (Greenhouse, Lever, Ashby) plus cleaned readable text; if fetching is blocked or no LLM key is configured, the flow seamlessly falls back to deterministic fields and/or the user's pasted text without blocking capture.
- [x] The extraction flow populates structured draft fields (**Company** name, role title, seniority, location/remote policy, stated compensation range, tech stack tags, and key requirements) via the semantic `AssistantGateway` port (`internal/adapter/ai`) using the configured **Fast Extraction** model (or an on-the-fly model selected from bookmarked models), allowing the candidate to review/edit before saving the **Opportunity** atomically via `TxManager`.
- [x] Backend HTTP handler tests (with injected `JobPageParser` and `AssistantGateway` adapters) and frontend view tests verify OpenRouter model catalog search/bookmarking, URL + pasted text extraction, per-request model override, and deterministic/manual fallback when no LLM is configured.
