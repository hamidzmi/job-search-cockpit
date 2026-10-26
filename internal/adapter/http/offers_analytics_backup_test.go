package httpadapter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	httpadapter "job-search-cockpit/internal/adapter/http"
)

func TestOffersAPI(t *testing.T) {
	t0 := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	clock := &manualClock{now: t0}
	srv := newTestServer(t, clock)

	// Setup: Candidate Profile with Comp Floor
	profilePayload := map[string]any{
		"full_name":          "Jane Candidate",
		"compensation_floor": 240000,
		"currency":           "USD",
	}
	rec := doJSON(t, srv, http.MethodPut, "/api/profile", profilePayload)
	if rec.Code != http.StatusOK {
		t.Fatalf("setup profile failed: %d: %s", rec.Code, rec.Body.String())
	}

	// Setup: Company and Opportunity
	companyPayload := map[string]any{
		"name": "Acme Corp",
	}
	rec = doJSON(t, srv, http.MethodPost, "/api/companies", companyPayload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup company failed: %d: %s", rec.Code, rec.Body.String())
	}
	var comp httpadapter.CompanyDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &comp)

	oppPayload := map[string]any{
		"company_id": comp.ID,
		"role_title": "Staff Engineer",
		"stage":      "Offer",
		"outcome":    "Active",
	}
	rec = doJSON(t, srv, http.MethodPost, "/api/opportunities", oppPayload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup opportunity failed: %d: %s", rec.Code, rec.Body.String())
	}
	var opp httpadapter.OpportunityDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &opp)

	t.Run("GET /api/opportunities/{id}/offer returns 404 before offer is created", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+opp.ID+"/offer", nil)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PUT /api/opportunities/{id}/offer rejects non-positive base salary with 400", func(t *testing.T) {
		invalidPayload := map[string]any{
			"base_salary": 0,
		}
		rec := doJSON(t, srv, http.MethodPut, "/api/opportunities/"+opp.ID+"/offer", invalidPayload)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("PUT /api/opportunities/{id}/offer creates offer and calculates annualized compensation", func(t *testing.T) {
		offerPayload := map[string]any{
			"base_salary":          200000,
			"bonus_target_pct":     15,
			"equity_type":          "RSU",
			"equity_total_value":   160000,
			"equity_vesting_years": 4,
			"sign_on_bonus":        25000,
			"currency":             "USD",
			"deadline_at":          "2026-11-05T23:59:00Z",
			"benefits_notes":       "100% health coverage, 401k 4% match",
			"negotiation_notes":    "Countered with $20k additional equity",
			"status":               "Pending",
		}
		rec := doJSON(t, srv, http.MethodPut, "/api/opportunities/"+opp.ID+"/offer", offerPayload)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var view httpadapter.OfferViewDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &view); err != nil {
			t.Fatalf("failed to parse offer response: %v", err)
		}

		if view.Offer.BaseSalary != 200000 {
			t.Errorf("base_salary = %d, want 200000", view.Offer.BaseSalary)
		}
		if view.Offer.AnnualBonus != 30000 {
			t.Errorf("annual_bonus = %d, want 30000", view.Offer.AnnualBonus)
		}
		if view.Offer.AnnualEquity != 40000 {
			t.Errorf("annual_equity = %d, want 40000", view.Offer.AnnualEquity)
		}
		if view.Offer.RecurringTotalComp != 270000 {
			t.Errorf("recurring_total_comp = %d, want 270000", view.Offer.RecurringTotalComp)
		}
		if view.Offer.Year1TotalComp != 295000 {
			t.Errorf("year1_total_comp = %d, want 295000", view.Offer.Year1TotalComp)
		}
		if view.Offer.DeltaVsCompFloor != 30000 {
			t.Errorf("delta_vs_comp_floor = %d, want 30000 (270k - 240k floor)", view.Offer.DeltaVsCompFloor)
		}
		if view.CompensationFloor != 240000 {
			t.Errorf("compensation_floor = %d, want 240000", view.CompensationFloor)
		}
		if view.CompanyName != "Acme Corp" {
			t.Errorf("company_name = %q, want 'Acme Corp'", view.CompanyName)
		}
	})

	t.Run("GET /api/offers returns comparison view with all offers", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodGet, "/api/offers", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var cmp httpadapter.OfferComparisonDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &cmp); err != nil {
			t.Fatalf("failed to parse comparison response: %v", err)
		}

		if cmp.CompensationFloor != 240000 {
			t.Errorf("compensation_floor = %d, want 240000", cmp.CompensationFloor)
		}
		if len(cmp.Offers) != 1 {
			t.Fatalf("len(offers) = %d, want 1", len(cmp.Offers))
		}
		if cmp.Offers[0].Offer.RecurringTotalComp != 270000 {
			t.Errorf("recurring_total_comp = %d, want 270000", cmp.Offers[0].Offer.RecurringTotalComp)
		}
	})

	t.Run("DELETE /api/opportunities/{id}/offer removes offer", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodDelete, "/api/opportunities/"+opp.ID+"/offer", nil)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %d: %s", rec.Code, rec.Body.String())
		}

		rec = doJSON(t, srv, http.MethodGet, "/api/opportunities/"+opp.ID+"/offer", nil)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found after deletion, got %d", rec.Code)
		}
	})
}

func TestFunnelAnalyticsAPI(t *testing.T) {
	t0 := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	clock := &manualClock{now: t0}
	srv := newTestServer(t, clock)

	// Create company
	rec := doJSON(t, srv, http.MethodPost, "/api/companies", map[string]any{"name": "TechCo"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create company failed: %d: %s", rec.Code, rec.Body.String())
	}
	var comp httpadapter.CompanyDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &comp)

	// Create opportunity in Discovered
	rec = doJSON(t, srv, http.MethodPost, "/api/opportunities", map[string]any{
		"company_id": comp.ID,
		"role_title": "Backend Lead",
		"stage":      "Discovered",
		"outcome":    "Active",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create opportunity failed: %d: %s", rec.Code, rec.Body.String())
	}
	var opp httpadapter.OpportunityDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &opp)

	// Transition to Outreach
	clock.now = clock.now.Add(24 * time.Hour)
	rec = doJSON(t, srv, http.MethodPost, "/api/opportunities/"+opp.ID+"/transition", map[string]any{
		"stage":   "Outreach",
		"outcome": "Active",
		"note":    "Reached out to recruiter",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("transition failed: %d: %s", rec.Code, rec.Body.String())
	}

	// Fetch Funnel Analytics
	rec = doJSON(t, srv, http.MethodGet, "/api/analytics/funnel", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var funnel httpadapter.FunnelAnalyticsDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &funnel); err != nil {
		t.Fatalf("parse funnel response failed: %v", err)
	}

	if funnel.TotalOpportunities != 1 {
		t.Errorf("total_opportunities = %d, want 1", funnel.TotalOpportunities)
	}
	if funnel.ActiveOpportunities != 1 {
		t.Errorf("active_opportunities = %d, want 1", funnel.ActiveOpportunities)
	}
	if len(funnel.Stages) != 6 {
		t.Fatalf("len(stages) = %d, want 6", len(funnel.Stages))
	}
}

func TestBackupExportAPI(t *testing.T) {
	t0 := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	clock := &manualClock{now: t0}
	srv := newTestServer(t, clock)

	// Add a company
	_ = doJSON(t, srv, http.MethodPost, "/api/companies", map[string]any{"name": "BackupCo"})

	t.Run("GET /api/backup/export?format=json exports valid JSON snapshot", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodGet, "/api/backup/export?format=json", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
			t.Errorf("expected application/json content-type, got %s", rec.Header().Get("Content-Type"))
		}
		var parsed map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("invalid json output: %v", err)
		}
		if _, ok := parsed["companies"]; !ok {
			t.Errorf("expected companies key in backup json")
		}
	})

	t.Run("GET /api/backup/export?format=sqlite exports SQLite VACUUM snapshot", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodGet, "/api/backup/export?format=sqlite", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		bodyBytes := rec.Body.Bytes()
		if len(bodyBytes) < 16 {
			t.Fatalf("sqlite backup too small: %d bytes", len(bodyBytes))
		}
		// SQLite header starts with "SQLite format 3\000"
		if !bytes.HasPrefix(bodyBytes, []byte("SQLite format 3\x00")) {
			t.Errorf("expected SQLite format 3 magic header, got %q", string(bodyBytes[:16]))
		}
	})
}

func TestSeedSampleDataAPI(t *testing.T) {
	t0 := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	clock := &manualClock{now: t0}
	srv := newTestServer(t, clock)

	rec := doJSON(t, srv, http.MethodPost, "/api/seed", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from seed, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify companies were seeded
	rec = doJSON(t, srv, http.MethodGet, "/api/companies", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get companies failed: %d: %s", rec.Code, rec.Body.String())
	}
	var companies []httpadapter.CompanyDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &companies)
	if len(companies) == 0 {
		t.Fatalf("expected seeded companies, got 0")
	}

	// Verify opportunities were seeded
	rec = doJSON(t, srv, http.MethodGet, "/api/opportunities", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get opportunities failed: %d: %s", rec.Code, rec.Body.String())
	}
	var opps []httpadapter.OpportunityDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &opps)
	if len(opps) == 0 {
		t.Fatalf("expected seeded opportunities, got 0")
	}

	// Verify offers were seeded
	rec = doJSON(t, srv, http.MethodGet, "/api/offers", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get offers failed: %d: %s", rec.Code, rec.Body.String())
	}
	var cmp httpadapter.OfferComparisonDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &cmp)
	if len(cmp.Offers) == 0 {
		t.Fatalf("expected seeded offers in comparison, got 0")
	}
}
