package httpadapter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	aiadapter "job-search-cockpit/internal/adapter/ai"
	httpadapter "job-search-cockpit/internal/adapter/http"
	"job-search-cockpit/internal/adapter/scraper"
	"job-search-cockpit/internal/adapter/sqlite"
	"job-search-cockpit/internal/app"
	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type fakeAssistantGateway struct {
	catalog       []domain.OpenRouterModel
	extractFn     func(ctx context.Context, settings domain.AISettings, in domain.AssistantExtractInput) (domain.ExtractedJobPostingDraft, error)
	extractCalls  []domain.AssistantExtractInput
	settingsCalls []domain.AISettings
}

func (f *fakeAssistantGateway) ExtractJobPosting(ctx context.Context, settings domain.AISettings, in domain.AssistantExtractInput) (domain.ExtractedJobPostingDraft, error) {
	f.extractCalls = append(f.extractCalls, in)
	f.settingsCalls = append(f.settingsCalls, settings)
	if f.extractFn != nil {
		return f.extractFn(ctx, settings, in)
	}
	return domain.ExtractedJobPostingDraft{}, domain.ErrAssistantUnavailable
}

func (f *fakeAssistantGateway) ListOpenRouterModels(_ context.Context, _ domain.AISettings) ([]domain.OpenRouterModel, error) {
	return f.catalog, nil
}

type fakeJobPageParser struct {
	parseFn func(ctx context.Context, rawURL string) (domain.ParsedJobPage, error)
	calls   []string
}

func (f *fakeJobPageParser) ParseJobPage(ctx context.Context, rawURL string) (domain.ParsedJobPage, error) {
	f.calls = append(f.calls, rawURL)
	if f.parseFn != nil {
		return f.parseFn(ctx, rawURL)
	}
	return domain.ParsedJobPage{}, nil
}

func newTestServerWithDeps(t *testing.T, clock *manualClock, assistant outbound.AssistantGateway, parser outbound.JobPageParser) http.Handler {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test-cockpit-ai.db")
	store, err := sqlite.OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("failed to open and migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	return app.NewHTTPHandler(store, app.Dependencies{
		Clock:            clock,
		AssistantGateway: assistant,
		JobPageParser:    parser,
	})
}

func TestAISettingsAndOpenRouterModelsAPI(t *testing.T) {
	clock := &manualClock{now: time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)}
	fakeAI := &fakeAssistantGateway{
		catalog: []domain.OpenRouterModel{
			{
				ID:                "anthropic/claude-3.5-haiku",
				Name:              "Anthropic: Claude 3.5 Haiku",
				Description:       "Fast structured extraction model",
				ContextLength:     200000,
				PromptPricing:     "$0.80/M",
				CompletionPricing: "$4.00/M",
			},
			{
				ID:                "anthropic/claude-3.7-sonnet",
				Name:              "Anthropic: Claude 3.7 Sonnet",
				Description:       "Deep reasoning and resume tailoring model",
				ContextLength:     200000,
				PromptPricing:     "$3.00/M",
				CompletionPricing: "$15.00/M",
			},
			{
				ID:                "openai/gpt-4o-mini",
				Name:              "OpenAI: GPT-4o Mini",
				Description:       "Lightweight fast extraction",
				ContextLength:     128000,
				PromptPricing:     "$0.15/M",
				CompletionPricing: "$0.60/M",
			},
		},
	}
	srv := newTestServerWithDeps(t, clock, fakeAI, &fakeJobPageParser{})

	t.Run("GET /api/settings/ai returns default unconfigured OpenRouter settings initially", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodGet, "/api/settings/ai", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		var got httpadapter.AISettingsDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if got.Provider != "openrouter" || got.IsConfigured || got.HasAPIKey {
			t.Errorf("expected default unconfigured openrouter settings, got %+v", got)
		}
		if len(got.BookmarkedModels) != 0 {
			t.Errorf("expected empty bookmarks, got %v", got.BookmarkedModels)
		}
	})

	t.Run("PUT /api/settings/ai persists provider, credentials, slots, and deduplicated bookmarked models", func(t *testing.T) {
		clock.now = clock.now.Add(10 * time.Minute)
		rec := doJSON(t, srv, http.MethodPut, "/api/settings/ai", map[string]any{
			"provider":                  "openrouter",
			"base_url":                  "https://openrouter.ai/api/v1",
			"api_key":                   "sk-or-v1-test-key",
			"fast_extraction_model":     "anthropic/claude-3.5-haiku",
			"reasoning_tailoring_model": "anthropic/claude-3.7-sonnet",
			"bookmarked_models": []string{
				"anthropic/claude-3.5-haiku",
				"anthropic/claude-3.7-sonnet",
				"openai/gpt-4o-mini",
				"anthropic/claude-3.5-haiku",
			},
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on PUT /api/settings/ai, got %d: %s", rec.Code, rec.Body.String())
		}
		var updated httpadapter.AISettingsDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if !updated.IsConfigured || !updated.HasAPIKey {
			t.Errorf("expected IsConfigured=true and HasAPIKey=true, got %+v", updated)
		}
		if updated.FastExtractionModel != "anthropic/claude-3.5-haiku" || updated.ReasoningTailoringModel != "anthropic/claude-3.7-sonnet" {
			t.Errorf("unexpected model slots: %+v", updated)
		}
		if len(updated.BookmarkedModels) != 3 {
			t.Errorf("expected 3 deduplicated bookmarks, got %v", updated.BookmarkedModels)
		}

		getRec := doJSON(t, srv, http.MethodGet, "/api/settings/ai", nil)
		if getRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on GET, got %d", getRec.Code)
		}
		var reloaded httpadapter.AISettingsDTO
		_ = json.Unmarshal(getRec.Body.Bytes(), &reloaded)
		if reloaded.APIKey != "sk-or-v1-test-key" || len(reloaded.BookmarkedModels) != 3 {
			t.Errorf("expected persisted settings on reload, got %+v", reloaded)
		}
	})

	t.Run("GET /api/settings/ai/openrouter-models searches live OpenRouter catalog", func(t *testing.T) {
		allRec := doJSON(t, srv, http.MethodGet, "/api/settings/ai/openrouter-models", nil)
		if allRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", allRec.Code, allRec.Body.String())
		}
		var allModels []httpadapter.OpenRouterModelDTO
		if err := json.Unmarshal(allRec.Body.Bytes(), &allModels); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(allModels) != 3 {
			t.Fatalf("expected 3 catalog models, got %d", len(allModels))
		}

		searchRec := doJSON(t, srv, http.MethodGet, "/api/settings/ai/openrouter-models?q=sonnet", nil)
		if searchRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on filtered catalog, got %d", searchRec.Code)
		}
		var filtered []httpadapter.OpenRouterModelDTO
		_ = json.Unmarshal(searchRec.Body.Bytes(), &filtered)
		if len(filtered) != 1 || filtered[0].ID != "anthropic/claude-3.7-sonnet" {
			t.Errorf("expected sonnet match, got %+v", filtered)
		}
	})
}

func TestQuickCaptureAPI_ExtractionFallbacksAndAtomicSave(t *testing.T) {
	clock := &manualClock{now: time.Date(2026, 10, 26, 11, 0, 0, 0, time.UTC)}

	// Local test server hosting JSON-LD, Greenhouse, Lever, Ashby, and blocked 403 job pages
	atsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/jsonld-role":
			w.Header().Set("Content-Type", "text/html")
			_, _ = fmt.Fprint(w, `<!doctype html>
<html>
<head>
<title>Staff Distributed Systems Engineer - Temporal</title>
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "JobPosting",
  "title": "Staff Distributed Systems Engineer",
  "hiringOrganization": {
    "@type": "Organization",
    "name": "Temporal Technologies",
    "sameAs": "https://temporal.io"
  },
  "jobLocation": {
    "@type": "Place",
    "address": {
      "addressLocality": "Seattle",
      "addressRegion": "WA"
    }
  },
  "jobLocationType": "TELECOMMUTE",
  "baseSalary": {
    "@type": "MonetaryAmount",
    "currency": "USD",
    "value": {
      "@type": "QuantitativeValue",
      "minValue": 230000,
      "maxValue": 275000
    }
  },
  "description": "<p>Build durable execution workflows in Go and Cassandra.</p><ul><li>8+ years distributed systems</li><li>Consensus and workflow engines</li></ul>"
}
</script>
</head>
<body><h1>Staff Distributed Systems Engineer</h1></body>
</html>`)
		case "/greenhouse-role":
			w.Header().Set("Content-Type", "text/html")
			_, _ = fmt.Fprint(w, `<!doctype html>
<html>
<head><title>Job Application for Principal Storage Engineer at PlanetScale</title></head>
<body>
  <h1 class="app-title">Principal Storage Engineer</h1>
  <span class="company-name">at PlanetScale</span>
  <div class="location">San Francisco, CA (Hybrid)</div>
  <div id="content">
    <p>Compensation: $240,000 - $290,000 USD</p>
    <ul>
      <li>Deep expertise in Go, MySQL, and Vitess</li>
      <li>Experience operating sharded databases at scale</li>
    </ul>
  </div>
</body>
</html>`)
		case "/lever-role":
			w.Header().Set("Content-Type", "text/html")
			_, _ = fmt.Fprint(w, `<!doctype html>
<html>
<head><title>Tailscale - Senior Kernel Networking Engineer</title></head>
<body>
  <div class="posting-headline">
    <h2>Senior Kernel Networking Engineer</h2>
    <div class="posting-categories">
      <div class="sort-by-time posting-category location">Remote (North America)</div>
    </div>
  </div>
  <div class="section-wrapper">
    <p>Salary range: $200,000 - $245,000 USD</p>
    <ul>
      <li>Experience with Go, Rust, and WireGuard</li>
    </ul>
  </div>
</body>
</html>`)
		case "/ashby-role":
			w.Header().Set("Content-Type", "text/html")
			_, _ = fmt.Fprint(w, `<!doctype html>
<html>
<head><title>Staff AI Platform Engineer @ Ramp</title></head>
<body>
  <h1>Staff AI Platform Engineer</h1>
  <p>Location: New York, NY (Hybrid)</p>
  <p>Compensation: $235,000 - $285,000 USD</p>
  <ul>
    <li>Build LLM routing and evaluation infrastructure in Python, Go, and PostgreSQL</li>
  </ul>
</body>
</html>`)
		case "/blocked-role":
			w.WriteHeader(http.StatusForbidden)
			_, _ = fmt.Fprint(w, "403 Forbidden - Anti-bot challenge")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer atsServer.Close()

	fakeAI := &fakeAssistantGateway{
		extractFn: func(_ context.Context, _ domain.AISettings, in domain.AssistantExtractInput) (domain.ExtractedJobPostingDraft, error) {
			return domain.ExtractedJobPostingDraft{
				CompanyName:      firstNonEmptyStr(in.DeterministicHint.CompanyName, "Vercel"),
				CompanyDomain:    firstNonEmptyStr(in.DeterministicHint.CompanyDomain, "vercel.com"),
				RoleTitle:        firstNonEmptyStr(in.DeterministicHint.RoleTitle, "Staff Edge Network Engineer"),
				Seniority:        "Staff",
				Location:         firstNonEmptyStr(in.DeterministicHint.Location, "Remote"),
				RemotePolicy:     domain.RemotePolicyRemote,
				StatedMinComp:    225000,
				StatedMaxComp:    265000,
				Currency:         "USD",
				TechStackTags:    []string{"Go", "Rust", "Anycast"},
				Requirements:     []string{"7+ years networking and edge proxies", "High-throughput Rust/Go services"},
				ExtractionSource: "llm",
				ModelUsed:        in.Model,
			}, nil
		},
	}

	realParser := scraper.NewJobPageParser(atsServer.Client())
	srv := newTestServerWithDeps(t, clock, fakeAI, realParser)

	t.Run("extracts deterministic JSON-LD and Greenhouse metadata when no LLM key is configured", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodPost, "/api/quick-capture/extract", map[string]any{
			"source_url": atsServer.URL + "/jsonld-role",
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		var draft httpadapter.ExtractedJobPostingDraftDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &draft); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if draft.CompanyName != "Temporal Technologies" || draft.RoleTitle != "Staff Distributed Systems Engineer" {
			t.Errorf("unexpected JSON-LD company/role: %+v", draft)
		}
		if draft.Seniority != "Staff" || draft.RemotePolicy != "Remote" {
			t.Errorf("expected Seniority=Staff and RemotePolicy=Remote, got %s / %s", draft.Seniority, draft.RemotePolicy)
		}
		if draft.StatedMinComp != 230000 || draft.StatedMaxComp != 275000 || draft.Currency != "USD" {
			t.Errorf("unexpected salary: %d-%d %s", draft.StatedMinComp, draft.StatedMaxComp, draft.Currency)
		}
		if draft.ExtractionSource != "deterministic" {
			t.Errorf("expected extraction_source='deterministic', got %q", draft.ExtractionSource)
		}
		if len(fakeAI.extractCalls) != 0 {
			t.Errorf("expected 0 LLM calls when unconfigured, got %d", len(fakeAI.extractCalls))
		}

		ghRec := doJSON(t, srv, http.MethodPost, "/api/quick-capture/extract", map[string]any{
			"source_url": atsServer.URL + "/greenhouse-role",
		})
		if ghRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on greenhouse role, got %d: %s", ghRec.Code, ghRec.Body.String())
		}
		var ghDraft httpadapter.ExtractedJobPostingDraftDTO
		_ = json.Unmarshal(ghRec.Body.Bytes(), &ghDraft)
		if ghDraft.CompanyName != "PlanetScale" || ghDraft.RoleTitle != "Principal Storage Engineer" {
			t.Errorf("unexpected Greenhouse company/role: %+v", ghDraft)
		}
		if ghDraft.StatedMinComp != 240000 || ghDraft.StatedMaxComp != 290000 {
			t.Errorf("expected Greenhouse comp 240000-290000, got %d-%d", ghDraft.StatedMinComp, ghDraft.StatedMaxComp)
		}

		leverRec := doJSON(t, srv, http.MethodPost, "/api/quick-capture/extract", map[string]any{
			"source_url": atsServer.URL + "/lever-role",
		})
		if leverRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on lever role, got %d: %s", leverRec.Code, leverRec.Body.String())
		}
		var leverDraft httpadapter.ExtractedJobPostingDraftDTO
		_ = json.Unmarshal(leverRec.Body.Bytes(), &leverDraft)
		if leverDraft.CompanyName != "Tailscale" || leverDraft.RoleTitle != "Senior Kernel Networking Engineer" {
			t.Errorf("unexpected Lever company/role: %+v", leverDraft)
		}

		ashbyRec := doJSON(t, srv, http.MethodPost, "/api/quick-capture/extract", map[string]any{
			"source_url": atsServer.URL + "/ashby-role",
		})
		if ashbyRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on ashby role, got %d: %s", ashbyRec.Code, ashbyRec.Body.String())
		}
		var ashbyDraft httpadapter.ExtractedJobPostingDraftDTO
		_ = json.Unmarshal(ashbyRec.Body.Bytes(), &ashbyDraft)
		if ashbyDraft.CompanyName != "Ramp" || ashbyDraft.RoleTitle != "Staff AI Platform Engineer" {
			t.Errorf("unexpected Ashby company/role: %+v", ashbyDraft)
		}
	})

	t.Run("falls back seamlessly to pasted raw_text when URL fetching is blocked (403)", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodPost, "/api/quick-capture/extract", map[string]any{
			"source_url": atsServer.URL + "/blocked-role",
			"raw_text":   "Company: Figma\nRole: Staff Infrastructure Engineer\nLocation: San Francisco, CA (Hybrid)\nCompensation: $230,000 - $280,000\n- 8+ years Go and TypeScript\n- Distributed storage systems",
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on blocked URL fallback, got %d: %s", rec.Code, rec.Body.String())
		}
		var draft httpadapter.ExtractedJobPostingDraftDTO
		_ = json.Unmarshal(rec.Body.Bytes(), &draft)
		if draft.FetchWarning == "" {
			t.Errorf("expected non-empty fetch_warning when URL returned 403")
		}
		if draft.CompanyName != "Figma" || draft.RoleTitle != "Staff Infrastructure Engineer" {
			t.Errorf("expected pasted text fallback fields for Figma, got %+v", draft)
		}
		if draft.StatedMinComp != 230000 || draft.StatedMaxComp != 280000 {
			t.Errorf("expected fallback comp 230000-280000, got %d-%d", draft.StatedMinComp, draft.StatedMaxComp)
		}
	})

	t.Run("uses configured Fast Extraction model by default and respects per-request model_override", func(t *testing.T) {
		// Configure OpenRouter API key and default models
		cfgRec := doJSON(t, srv, http.MethodPut, "/api/settings/ai", map[string]any{
			"provider":                  "openrouter",
			"api_key":                   "sk-or-v1-live",
			"fast_extraction_model":     "anthropic/claude-3.5-haiku",
			"reasoning_tailoring_model": "anthropic/claude-3.7-sonnet",
			"bookmarked_models": []string{
				"anthropic/claude-3.5-haiku",
				"openai/gpt-4o-mini",
			},
		})
		if cfgRec.Code != http.StatusOK {
			t.Fatalf("failed to configure AI settings: %d", cfgRec.Code)
		}

		// 1. Default fast extraction model
		rec1 := doJSON(t, srv, http.MethodPost, "/api/quick-capture/extract", map[string]any{
			"raw_text": "Vercel is hiring a Staff Edge Network Engineer ($225k-$265k).",
		})
		if rec1.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec1.Code, rec1.Body.String())
		}
		var draft1 httpadapter.ExtractedJobPostingDraftDTO
		_ = json.Unmarshal(rec1.Body.Bytes(), &draft1)
		if draft1.ExtractionSource != "llm" || draft1.ModelUsed != "anthropic/claude-3.5-haiku" {
			t.Errorf("expected llm extraction with default model anthropic/claude-3.5-haiku, got source=%q model=%q", draft1.ExtractionSource, draft1.ModelUsed)
		}

		// 2. Per-request model override from bookmarked models
		rec2 := doJSON(t, srv, http.MethodPost, "/api/quick-capture/extract", map[string]any{
			"source_url":     atsServer.URL + "/jsonld-role",
			"model_override": "openai/gpt-4o-mini",
		})
		if rec2.Code != http.StatusOK {
			t.Fatalf("expected 200 OK with model override, got %d: %s", rec2.Code, rec2.Body.String())
		}
		var draft2 httpadapter.ExtractedJobPostingDraftDTO
		_ = json.Unmarshal(rec2.Body.Bytes(), &draft2)
		if draft2.ModelUsed != "openai/gpt-4o-mini" {
			t.Errorf("expected model_used='openai/gpt-4o-mini', got %q", draft2.ModelUsed)
		}
		if draft2.CompanyName != "Temporal Technologies" {
			t.Errorf("expected deterministic hint passed to LLM for Temporal Technologies, got %q", draft2.CompanyName)
		}
	})

	t.Run("atomically creates Company, Opportunity, and initial StageTransition via POST /api/quick-capture", func(t *testing.T) {
		saveRec := doJSON(t, srv, http.MethodPost, "/api/quick-capture", map[string]any{
			"company_name":    "Vercel",
			"company_domain":  "vercel.com",
			"role_title":      "Staff Edge Network Engineer",
			"seniority":       "Staff",
			"location":        "Remote",
			"remote_policy":   "Remote",
			"priority":        "High",
			"stage":           "Discovered",
			"transition_note": "Captured via Quick Capture modal",
			"notes":           "Edge proxy team",
			"job_posting": map[string]any{
				"source_url":      "https://vercel.com/careers/edge-eng",
				"raw_text":        "Build global Anycast edge network in Rust and Go.",
				"requirements":    []string{"7+ years networking and edge proxies"},
				"tech_stack_tags": []string{"Go", "Rust", "Anycast"},
				"stated_min_comp": 225000,
				"stated_max_comp": 265000,
				"currency":        "USD",
			},
		})
		if saveRec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created on POST /api/quick-capture, got %d: %s", saveRec.Code, saveRec.Body.String())
		}
		var created httpadapter.OpportunityDTO
		if err := json.Unmarshal(saveRec.Body.Bytes(), &created); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if created.CompanyName != "Vercel" || created.RoleTitle != "Staff Edge Network Engineer" {
			t.Errorf("unexpected created opportunity: %+v", created)
		}
		if !created.JobPosting.HasContent {
			t.Errorf("expected JobPosting.HasContent=true")
		}

		// Verify StageTransition was recorded
		detailRec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+created.ID, nil)
		var detail httpadapter.OpportunityDTO
		_ = json.Unmarshal(detailRec.Body.Bytes(), &detail)
		if len(detail.StageTransitions) != 1 || detail.StageTransitions[0].ToStage != "Discovered" {
			t.Errorf("expected 1 initial StageTransition to Discovered, got %+v", detail.StageTransitions)
		}

		// Saving a second Opportunity for "Vercel" reuses the existing Company without duplicating it
		saveRec2 := doJSON(t, srv, http.MethodPost, "/api/quick-capture", map[string]any{
			"company_name":  "vercel",
			"role_title":    "Principal Build Systems Engineer",
			"remote_policy": "Remote",
			"priority":      "Medium",
		})
		if saveRec2.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created on second quick capture, got %d: %s", saveRec2.Code, saveRec2.Body.String())
		}
		var created2 httpadapter.OpportunityDTO
		_ = json.Unmarshal(saveRec2.Body.Bytes(), &created2)
		if created2.CompanyID != created.CompanyID {
			t.Errorf("expected same CompanyID %q for Vercel, got %q", created.CompanyID, created2.CompanyID)
		}
	})

	t.Run("exercises real HTTPAssistantGateway against stubbed OpenRouter server", func(t *testing.T) {
		openRouterSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/models":
				_, _ = fmt.Fprint(w, `{"data":[{"id":"anthropic/claude-3.5-haiku","name":"Claude 3.5 Haiku","description":"Fast extractor","context_length":200000,"pricing":{"prompt":"$0.80/M","completion":"$4.00/M"}}]}`)
			case "/chat/completions":
				_, _ = fmt.Fprint(w, `{"model":"anthropic/claude-3.5-haiku","choices":[{"message":{"content":"{\"company_name\":\"Fly.io\",\"company_domain\":\"fly.io\",\"role_title\":\"Staff Kernel Engineer\",\"seniority\":\"Staff\",\"location\":\"Remote\",\"remote_policy\":\"Remote\",\"stated_min_comp\":210000,\"stated_max_comp\":250000,\"currency\":\"USD\",\"tech_stack_tags\":[\"Rust\",\"Go\",\"KVM\"],\"requirements\":[\"Linux kernel and Firecracker experience\"]}"}}]}`)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer openRouterSrv.Close()

		realAI := aiadapter.NewAssistantGateway(openRouterSrv.Client())
		realSrv := newTestServerWithDeps(t, clock, realAI, realParser)

		doJSON(t, realSrv, http.MethodPut, "/api/settings/ai", map[string]any{
			"provider":              "openrouter",
			"base_url":              openRouterSrv.URL,
			"api_key":               "sk-or-test",
			"fast_extraction_model": "anthropic/claude-3.5-haiku",
		})

		modelsRec := doJSON(t, realSrv, http.MethodGet, "/api/settings/ai/openrouter-models", nil)
		if modelsRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK from real HTTPAssistantGateway models, got %d", modelsRec.Code)
		}

		extRec := doJSON(t, realSrv, http.MethodPost, "/api/quick-capture/extract", map[string]any{
			"raw_text": "Fly.io is hiring a Staff Kernel Engineer working on Firecracker microVMs.",
		})
		if extRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK from real HTTPAssistantGateway extract, got %d: %s", extRec.Code, extRec.Body.String())
		}
		var extDraft httpadapter.ExtractedJobPostingDraftDTO
		_ = json.Unmarshal(extRec.Body.Bytes(), &extDraft)
		if extDraft.CompanyName != "Fly.io" || extDraft.RoleTitle != "Staff Kernel Engineer" {
			t.Errorf("unexpected draft from real HTTPAssistantGateway: %+v", extDraft)
		}
	})
}

func firstNonEmptyStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
