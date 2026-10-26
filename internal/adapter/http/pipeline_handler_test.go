package httpadapter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	httpadapter "job-search-cockpit/internal/adapter/http"
	"job-search-cockpit/internal/adapter/sqlite"
	"job-search-cockpit/internal/app"
)

type manualClock struct {
	now time.Time
}

func (c *manualClock) Now() time.Time {
	return c.now
}

func newTestServer(t *testing.T, clock *manualClock) http.Handler {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test-cockpit.db")
	store, err := sqlite.OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("failed to open and migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	return app.NewHTTPHandler(store, app.Dependencies{Clock: clock})
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCompaniesAPI(t *testing.T) {
	clock := &manualClock{now: time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)}
	srv := newTestServer(t, clock)

	createCases := []struct {
		name       string
		payload    map[string]any
		wantStatus int
		wantName   string
	}{
		{
			name:       "rejects blank company name with 400 validation error",
			payload:    map[string]any{"name": "   "},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "creates company and returns 201",
			payload: map[string]any{
				"name":          "Anthropic",
				"domain":        "anthropic.com",
				"headquarters":  "San Francisco, CA",
				"remote_policy": "Hybrid",
				"size_stage":    "Growth",
				"notes":         "AI safety and research",
			},
			wantStatus: http.StatusCreated,
			wantName:   "Anthropic",
		},
	}

	var createdID string
	for _, tc := range createCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, srv, http.MethodPost, "/api/companies", tc.payload)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}
			if tc.wantStatus == http.StatusCreated {
				var resp httpadapter.CompanyDTO
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode company response: %v", err)
				}
				if resp.Name != tc.wantName {
					t.Errorf("expected name %q, got %q", tc.wantName, resp.Name)
				}
				createdID = resp.ID
			}
		})
	}

	t.Run("updates company and retrieves updated details via GET", func(t *testing.T) {
		clock.now = clock.now.Add(2 * time.Hour)
		updateRec := doJSON(t, srv, http.MethodPut, "/api/companies/"+createdID, map[string]any{
			"name":          "Anthropic PBC",
			"domain":        "anthropic.com",
			"headquarters":  "San Francisco, CA",
			"remote_policy": "Hybrid (SF/Seattle/NY)",
			"size_stage":    "Late Stage",
			"notes":         "Updated notes",
		})
		if updateRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on company update, got %d: %s", updateRec.Code, updateRec.Body.String())
		}

		getRec := doJSON(t, srv, http.MethodGet, "/api/companies/"+createdID, nil)
		if getRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on company get, got %d", getRec.Code)
		}
		var got httpadapter.CompanyDTO
		if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if got.Name != "Anthropic PBC" || got.RemotePolicy != "Hybrid (SF/Seattle/NY)" {
			t.Errorf("unexpected updated company: %+v", got)
		}
	})
}

func TestOpportunitiesAPI_LifecycleAndTransitions(t *testing.T) {
	t0 := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	clock := &manualClock{now: t0}
	srv := newTestServer(t, clock)

	compRec := doJSON(t, srv, http.MethodPost, "/api/companies", map[string]any{
		"name":          "Stripe",
		"domain":        "stripe.com",
		"remote_policy": "Hybrid",
	})
	if compRec.Code != http.StatusCreated {
		t.Fatalf("setup company failed: %d", compRec.Code)
	}
	var company httpadapter.CompanyDTO
	_ = json.Unmarshal(compRec.Body.Bytes(), &company)

	createCases := []struct {
		name           string
		payload        map[string]any
		wantStatus     int
		wantStage      string
		wantOutcome    string
		wantHasPosting bool
	}{
		{
			name: "rejects opportunity with unknown company_id with 404",
			payload: map[string]any{
				"company_id": "non-existent",
				"role_title": "Backend Engineer",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "creates unposted opportunity under Stripe defaulting to Discovered and Active",
			payload: map[string]any{
				"company_id":    company.ID,
				"role_title":    "Staff Core Payments Engineer",
				"priority":      "High",
				"remote_policy": "Remote",
				"note":          "Warm intro from staff engineer",
			},
			wantStatus:     http.StatusCreated,
			wantStage:      "Discovered",
			wantOutcome:    "Active",
			wantHasPosting: false,
		},
		{
			name: "creates second distinct posted opportunity under Stripe without collision",
			payload: map[string]any{
				"company_id":    company.ID,
				"role_title":    "Staff Ruby Infrastructure Engineer",
				"seniority":     "Staff",
				"location":      "Seattle, WA",
				"remote_policy": "Hybrid",
				"priority":      "Medium",
				"stage":         "Applied",
				"note":          "Applied directly",
				"job_posting": map[string]any{
					"source_url":      "https://stripe.com/jobs/infra-1",
					"raw_text":        "Lead developer productivity and Sorbet infrastructure.",
					"requirements":    []string{"8+ years infrastructure", "Distributed systems"},
					"tech_stack_tags": []string{"Go", "Ruby", "Sorbet"},
					"stated_min_comp": 230000,
					"stated_max_comp": 280000,
					"currency":        "USD",
				},
			},
			wantStatus:     http.StatusCreated,
			wantStage:      "Applied",
			wantOutcome:    "Active",
			wantHasPosting: true,
		},
	}

	var oppIDs []string
	for _, tc := range createCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, srv, http.MethodPost, "/api/opportunities", tc.payload)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}
			if tc.wantStatus == http.StatusCreated {
				var opp httpadapter.OpportunityDTO
				if err := json.Unmarshal(rec.Body.Bytes(), &opp); err != nil {
					t.Fatalf("decode failed: %v", err)
				}
				if opp.Stage != tc.wantStage || opp.Outcome != tc.wantOutcome {
					t.Errorf("expected stage=%s outcome=%s, got %s/%s", tc.wantStage, tc.wantOutcome, opp.Stage, opp.Outcome)
				}
				if opp.CompanyName != "Stripe" {
					t.Errorf("expected CompanyName 'Stripe', got %q", opp.CompanyName)
				}
				if opp.JobPosting.HasContent != tc.wantHasPosting {
					t.Errorf("expected HasContent=%v, got %v", tc.wantHasPosting, opp.JobPosting.HasContent)
				}
				oppIDs = append(oppIDs, opp.ID)
			}
		})
	}

	firstOppID := oppIDs[0]

	t.Run("attaches JobPosting later via PUT without advancing LastActivityAt", func(t *testing.T) {
		clock.now = t0.Add(24 * time.Hour)
		rec := doJSON(t, srv, http.MethodPut, "/api/opportunities/"+firstOppID, map[string]any{
			"company_id":    company.ID,
			"role_title":    "Staff Core Payments Engineer",
			"seniority":     "Staff",
			"location":      "San Francisco, CA",
			"remote_policy": "Remote",
			"priority":      "High",
			"notes":         "Recruiter sent official req link",
			"job_posting": map[string]any{
				"source_url":      "https://stripe.com/jobs/payments-99",
				"raw_text":        "Architect global money movement ledger.",
				"requirements":    []string{"High-throughput ledgers", "Go/Java"},
				"tech_stack_tags": []string{"Go", "PostgreSQL", "Kafka"},
				"stated_min_comp": 240000,
				"stated_max_comp": 290000,
				"currency":        "USD",
			},
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on PUT opportunity, got %d: %s", rec.Code, rec.Body.String())
		}
		var updated httpadapter.OpportunityDTO
		_ = json.Unmarshal(rec.Body.Bytes(), &updated)
		if !updated.JobPosting.HasContent {
			t.Errorf("expected JobPosting.HasContent=true after attaching posting")
		}
		if !updated.LastActivityAt.Equal(t0) {
			t.Errorf("expected LastActivityAt to remain %v on detail edit, got %v", t0, updated.LastActivityAt)
		}
		if !updated.UpdatedAt.Equal(clock.now) {
			t.Errorf("expected UpdatedAt to advance to %v, got %v", clock.now, updated.UpdatedAt)
		}
	})

	transitionCases := []struct {
		name        string
		advanceBy   time.Duration
		payload     map[string]any
		wantStatus  int
		wantStage   string
		wantOutcome string
	}{
		{
			name:      "moves stage from Discovered to Outreach",
			advanceBy: 48 * time.Hour,
			payload: map[string]any{
				"stage": "Outreach",
				"note":  "Requested referral",
			},
			wantStatus:  http.StatusOK,
			wantStage:   "Outreach",
			wantOutcome: "Active",
		},
		{
			name:      "moves stage from Outreach to Interviewing",
			advanceBy: 96 * time.Hour,
			payload: map[string]any{
				"stage": "Interviewing",
				"note":  "Passed recruiter screen, scheduled loop",
			},
			wantStatus:  http.StatusOK,
			wantStage:   "Interviewing",
			wantOutcome: "Active",
		},
		{
			name:      "updates Outcome to Rejected while preserving Interviewing Stage",
			advanceBy: 144 * time.Hour,
			payload: map[string]any{
				"outcome": "Rejected",
				"note":    "Filled internally after final round",
			},
			wantStatus:  http.StatusOK,
			wantStage:   "Interviewing",
			wantOutcome: "Rejected",
		},
		{
			name:      "rejects unchanged Stage and Outcome with 422",
			advanceBy: 150 * time.Hour,
			payload: map[string]any{
				"stage":   "Interviewing",
				"outcome": "Rejected",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range transitionCases {
		t.Run(tc.name, func(t *testing.T) {
			clock.now = t0.Add(tc.advanceBy)
			rec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+firstOppID+"/transition", tc.payload)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}
			if tc.wantStatus == http.StatusOK {
				var opp httpadapter.OpportunityDTO
				_ = json.Unmarshal(rec.Body.Bytes(), &opp)
				if opp.Stage != tc.wantStage || opp.Outcome != tc.wantOutcome {
					t.Errorf("expected stage=%s outcome=%s, got %s/%s", tc.wantStage, tc.wantOutcome, opp.Stage, opp.Outcome)
				}
				if !opp.LastActivityAt.Equal(clock.now) {
					t.Errorf("expected LastActivityAt=%v, got %v", clock.now, opp.LastActivityAt)
				}
			}
		})
	}

	t.Run("returns full chronological StageTransition history on opportunity detail endpoint", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+firstOppID, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on GET opportunity, got %d", rec.Code)
		}
		var detail httpadapter.OpportunityDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
			t.Fatalf("decode opportunity detail failed: %v", err)
		}
		history := detail.StageTransitions
		if len(history) != 4 {
			t.Fatalf("expected 4 transitions (1 initial + 3 changes), got %d: %+v", len(history), history)
		}
		if history[0].FromStage != "" || history[0].ToStage != "Discovered" || history[0].ToOutcome != "Active" {
			t.Errorf("unexpected initial transition: %+v", history[0])
		}
		if history[3].FromStage != "Interviewing" || history[3].ToStage != "Interviewing" || history[3].FromOutcome != "Active" || history[3].ToOutcome != "Rejected" {
			t.Errorf("unexpected final transition: %+v", history[3])
		}
	})
}

func TestOpportunitiesAPI_Filtering(t *testing.T) {
	clock := &manualClock{now: time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)}
	srv := newTestServer(t, clock)

	createCompany := func(name string) string {
		rec := doJSON(t, srv, http.MethodPost, "/api/companies", map[string]any{"name": name})
		var c httpadapter.CompanyDTO
		_ = json.Unmarshal(rec.Body.Bytes(), &c)
		return c.ID
	}

	compDatadog := createCompany("Datadog")
	compCloudflare := createCompany("Cloudflare")

	seedOpps := []map[string]any{
		{
			"company_id":    compDatadog,
			"role_title":    "Staff Go Ingestion Engineer",
			"priority":      "High",
			"remote_policy": "Remote",
			"stage":         "Interviewing",
			"outcome":       "Active",
			"job_posting": map[string]any{
				"tech_stack_tags": []string{"Go", "Kafka", "Rust"},
			},
		},
		{
			"company_id":    compDatadog,
			"role_title":    "Senior Storage Engineer",
			"priority":      "Low",
			"remote_policy": "Hybrid",
			"stage":         "Screening",
		},
		{
			"company_id":    compCloudflare,
			"role_title":    "Principal Edge Runtime Engineer",
			"priority":      "High",
			"remote_policy": "Remote",
			"stage":         "Offer",
			"notes":         "Workers runtime V8 isolates",
		},
	}

	var seededIDs []string
	for _, payload := range seedOpps {
		clock.now = clock.now.Add(time.Minute)
		rec := doJSON(t, srv, http.MethodPost, "/api/opportunities", payload)
		if rec.Code != http.StatusCreated {
			t.Fatalf("seed opportunity failed: %d %s", rec.Code, rec.Body.String())
		}
		var created httpadapter.OpportunityDTO
		_ = json.Unmarshal(rec.Body.Bytes(), &created)
		seededIDs = append(seededIDs, created.ID)
	}

	// Transition Senior Storage Engineer to Ghosted
	trRec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+seededIDs[1]+"/transition", map[string]any{
		"outcome": "Ghosted",
		"note":    "No response after screening",
	})
	if trRec.Code != http.StatusOK {
		t.Fatalf("seed transition failed: %d %s", trRec.Code, trRec.Body.String())
	}

	filterCases := []struct {
		name       string
		query      string
		wantStatus int
		wantTitles []string
	}{
		{
			name:       "filters Active opportunities for Kanban board",
			query:      "?outcome=Active",
			wantStatus: http.StatusOK,
			wantTitles: []string{"Principal Edge Runtime Engineer", "Staff Go Ingestion Engineer"},
		},
		{
			name:       "filters by Stage=Screening and Outcome=Ghosted",
			query:      "?stage=Screening&outcome=Ghosted",
			wantStatus: http.StatusOK,
			wantTitles: []string{"Senior Storage Engineer"},
		},
		{
			name:       "filters by priority and remote_policy",
			query:      "?priority=High&remote_policy=Remote",
			wantStatus: http.StatusOK,
			wantTitles: []string{"Principal Edge Runtime Engineer", "Staff Go Ingestion Engineer"},
		},
		{
			name:       "filters by company_id",
			query:      "?company_id=" + compCloudflare,
			wantStatus: http.StatusOK,
			wantTitles: []string{"Principal Edge Runtime Engineer"},
		},
		{
			name:       "filters by text query matching tech stack or notes",
			query:      "?q=kafka",
			wantStatus: http.StatusOK,
			wantTitles: []string{"Staff Go Ingestion Engineer"},
		},
		{
			name:       "rejects invalid stage filter with 400",
			query:      "?stage=NotAStage",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range filterCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, srv, http.MethodGet, "/api/opportunities"+tc.query, nil)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}
			if tc.wantStatus == http.StatusOK {
				var got []httpadapter.OpportunityDTO
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("decode failed: %v", err)
				}
				if len(got) != len(tc.wantTitles) {
					t.Fatalf("expected %d opportunities, got %d: %+v", len(tc.wantTitles), len(got), got)
				}
				for i, wantTitle := range tc.wantTitles {
					if got[i].RoleTitle != wantTitle {
						t.Errorf("result[%d]: expected title %q, got %q", i, wantTitle, got[i].RoleTitle)
					}
				}
			}
		})
	}
}


