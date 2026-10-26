package httpadapter_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	httpadapter "job-search-cockpit/internal/adapter/http"
)

func TestFlightDeckAPI(t *testing.T) {
	t0 := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	clock := &manualClock{now: t0}
	srv := newTestServer(t, clock)

	// 1. Seed Company, Contact, Story
	coRec := doJSON(t, srv, http.MethodPost, "/api/companies", map[string]any{"name": "Anthropic"})
	var company httpadapter.CompanyDTO
	_ = json.Unmarshal(coRec.Body.Bytes(), &company)

	ctRec := doJSON(t, srv, http.MethodPost, "/api/contacts", map[string]any{
		"company_id": company.ID,
		"name":       "Dario Amodei",
		"title":      "CEO",
	})
	var contact httpadapter.ContactDTO
	_ = json.Unmarshal(ctRec.Body.Bytes(), &contact)

	stRec := doJSON(t, srv, http.MethodPost, "/api/stories", map[string]any{
		"title":           "Scaling Alignment Research",
		"hook":            "Led 20-person interpretability team.",
		"competency_tags": []string{"Leadership"},
		"skill_tags":      []string{"AI", "Python"},
	})
	var story httpadapter.StoryDTO
	_ = json.Unmarshal(stRec.Body.Bytes(), &story)

	// 2. Create Opportunity in Discovered
	oppRec := doJSON(t, srv, http.MethodPost, "/api/opportunities", map[string]any{
		"company_id": company.ID,
		"role_title": "Member of Technical Staff",
		"stage":      "Discovered",
		"priority":   "High",
	})
	var opp httpadapter.OpportunityDTO
	_ = json.Unmarshal(oppRec.Body.Bytes(), &opp)

	// 3. Create Action Items: one overdue, one due today, one future
	doJSON(t, srv, http.MethodPost, "/api/action-items", map[string]any{
		"title":          "Overdue action",
		"opportunity_id": opp.ID,
		"due_at":         t0.Add(-5 * time.Hour).Format(time.RFC3339),
	})
	doJSON(t, srv, http.MethodPost, "/api/action-items", map[string]any{
		"title":          "Due today action",
		"opportunity_id": opp.ID,
		"due_at":         t0.Add(4 * time.Hour).Format(time.RFC3339),
	})
	futActionRec := doJSON(t, srv, http.MethodPost, "/api/action-items", map[string]any{
		"title":          "Future action to complete",
		"opportunity_id": opp.ID,
		"due_at":         t0.Add(40 * time.Hour).Format(time.RFC3339),
	})
	var futAction httpadapter.ActionItemDTO
	_ = json.Unmarshal(futActionRec.Body.Bytes(), &futAction)

	// 4. Create Interview Rounds:
	// Round A: in 20 hours (within 48h) -> upcoming
	doJSON(t, srv, http.MethodPost, "/api/opportunities/"+opp.ID+"/rounds", map[string]any{
		"round_order":      1,
		"round_type":       "Technical",
		"title":            "Interpretability Coding",
		"scheduled_at":     t0.Add(20 * time.Hour).Format(time.RFC3339),
		"duration_minutes": 60,
		"contact_ids":      []string{contact.ID},
		"story_ids":        []string{story.ID},
	})
	// Round B: in 72 hours (> 48h) -> NOT in upcoming
	doJSON(t, srv, http.MethodPost, "/api/opportunities/"+opp.ID+"/rounds", map[string]any{
		"round_order":      2,
		"round_type":       "System Design",
		"title":            "Model Architecture",
		"scheduled_at":     t0.Add(72 * time.Hour).Format(time.RFC3339),
		"duration_minutes": 60,
	})
	// Round C: 3 hours ago (past, completed) missing debrief
	missingRoundRec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+opp.ID+"/rounds", map[string]any{
		"round_order":      3,
		"round_type":       "Founder Chat",
		"title":            "Chat with Dario",
		"scheduled_at":     t0.Add(-3 * time.Hour).Format(time.RFC3339),
		"status":           "Completed",
		"duration_minutes": 45,
	})
	var missingRound httpadapter.InterviewRoundDTO
	_ = json.Unmarshal(missingRoundRec.Body.Bytes(), &missingRound)

	// 5. Test initial Flight Deck response
	fdRec := doJSON(t, srv, http.MethodGet, "/api/flight-deck", nil)
	if fdRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from GET /api/flight-deck, got %d: %s", fdRec.Code, fdRec.Body.String())
	}
	var fd httpadapter.FlightDeckDTO
	if err := json.Unmarshal(fdRec.Body.Bytes(), &fd); err != nil {
		t.Fatalf("failed to decode flight deck response: %v", err)
	}

	// Verify Upcoming Rounds
	if len(fd.UpcomingRounds) != 1 {
		t.Fatalf("expected 1 upcoming round, got %d", len(fd.UpcomingRounds))
	}
	if fd.UpcomingRounds[0].Title != "Interpretability Coding" {
		t.Errorf("expected title 'Interpretability Coding', got %q", fd.UpcomingRounds[0].Title)
	}
	if fd.UpcomingRounds[0].CompanyName != "Anthropic" {
		t.Errorf("expected company 'Anthropic', got %q", fd.UpcomingRounds[0].CompanyName)
	}
	if len(fd.UpcomingRounds[0].InterviewerNames) != 1 || fd.UpcomingRounds[0].InterviewerNames[0] != "Dario Amodei" {
		t.Errorf("expected interviewer Dario Amodei, got %v", fd.UpcomingRounds[0].InterviewerNames)
	}
	if len(fd.UpcomingRounds[0].StoryTitles) != 1 || fd.UpcomingRounds[0].StoryTitles[0] != "Scaling Alignment Research" {
		t.Errorf("expected story title, got %v", fd.UpcomingRounds[0].StoryTitles)
	}

	// Verify Missing Debriefs
	if len(fd.MissingDebriefs) != 1 {
		t.Fatalf("expected 1 missing debrief, got %d", len(fd.MissingDebriefs))
	}
	if fd.MissingDebriefs[0].ID != missingRound.ID {
		t.Errorf("expected missing debrief round ID %s, got %s", missingRound.ID, fd.MissingDebriefs[0].ID)
	}

	// Verify Action Items
	if len(fd.ActionItems) != 3 {
		t.Fatalf("expected 3 open action items, got %d", len(fd.ActionItems))
	}
	if !fd.ActionItems[0].IsOverdue {
		t.Errorf("expected first action item to be overdue")
	}
	if !fd.ActionItems[1].IsDueToday {
		t.Errorf("expected second action item to be due today")
	}

	// Complete the third action item inline and verify it's removed
	doJSON(t, srv, http.MethodPut, "/api/action-items/"+futAction.ID, map[string]any{
		"title":          futAction.Title,
		"opportunity_id": opp.ID,
		"due_at":         futAction.DueAt,
		"status":         "completed",
	})
	fdRec2 := doJSON(t, srv, http.MethodGet, "/api/flight-deck", nil)
	var fd2 httpadapter.FlightDeckDTO
	_ = json.Unmarshal(fdRec2.Body.Bytes(), &fd2)
	if len(fd2.ActionItems) != 2 {
		t.Errorf("expected 2 open action items after completing one, got %d", len(fd2.ActionItems))
	}

	// Record Debrief on missingRound and verify it disappears from missing_debriefs
	doJSON(t, srv, http.MethodPut, "/api/rounds/"+missingRound.ID+"/debrief", map[string]any{
		"went_well":        "Vision alignment",
		"signals":          "Strong culture fit",
		"confidence_score": 5,
	})
	fdRec3 := doJSON(t, srv, http.MethodGet, "/api/flight-deck", nil)
	var fd3 httpadapter.FlightDeckDTO
	_ = json.Unmarshal(fdRec3.Body.Bytes(), &fd3)
	if len(fd3.MissingDebriefs) != 0 {
		t.Errorf("expected 0 missing debriefs after recording debrief, got %d", len(fd3.MissingDebriefs))
	}

	// 6. Test Stage Staleness & Inline Resolution Actions
	// Advance clock by 4 days -> Discovered > 3 days -> Stale!
	clock.now = t0.Add(4 * 24 * time.Hour)
	fdRec4 := doJSON(t, srv, http.MethodGet, "/api/flight-deck", nil)
	var fd4 httpadapter.FlightDeckDTO
	_ = json.Unmarshal(fdRec4.Body.Bytes(), &fd4)
	if len(fd4.StaleOpportunities) != 1 || fd4.StaleOpportunities[0].OpportunityID != opp.ID {
		t.Fatalf("expected opp %s to be stale in Discovered, got %v", opp.ID, fd4.StaleOpportunities)
	}

	// Test Inline Action 1: Log Interaction resets staleness timer for Discovered
	doJSON(t, srv, http.MethodPost, "/api/interactions", map[string]any{
		"contact_id":     contact.ID,
		"opportunity_id": opp.ID,
		"kind":           "Email",
		"direction":      "Outbound",
		"occurred_at":    clock.now.Format(time.RFC3339),
		"summary":        "Reached out regarding alignment team",
	})
	fdRec5 := doJSON(t, srv, http.MethodGet, "/api/flight-deck", nil)
	var fd5 httpadapter.FlightDeckDTO
	_ = json.Unmarshal(fdRec5.Body.Bytes(), &fd5)
	if len(fd5.StaleOpportunities) != 0 {
		t.Errorf("expected opportunity to no longer be stale after logging interaction, got %d", len(fd5.StaleOpportunities))
	}

	// Test Inline Action 2: Advance stage to Outreach
	doJSON(t, srv, http.MethodPost, "/api/opportunities/"+opp.ID+"/transition", map[string]any{
		"stage": "Outreach",
		"note":  "Recruiter contacted",
	})
	// Advance clock by 5 days -> Outreach > 4 days -> Stale!
	clock.now = clock.now.Add(5 * 24 * time.Hour)
	fdRec6 := doJSON(t, srv, http.MethodGet, "/api/flight-deck", nil)
	var fd6 httpadapter.FlightDeckDTO
	_ = json.Unmarshal(fdRec6.Body.Bytes(), &fd6)
	if len(fd6.StaleOpportunities) != 1 || fd6.StaleOpportunities[0].Stage != "Outreach" {
		t.Fatalf("expected opp to be stale in Outreach, got %v", fd6.StaleOpportunities)
	}

	// Test Inline Action 3: Update Outcome to Ghosted
	doJSON(t, srv, http.MethodPost, "/api/opportunities/"+opp.ID+"/transition", map[string]any{
		"outcome": "Ghosted",
		"note":    "No response after 5 days",
	})
	fdRec7 := doJSON(t, srv, http.MethodGet, "/api/flight-deck", nil)
	var fd7 httpadapter.FlightDeckDTO
	_ = json.Unmarshal(fdRec7.Body.Bytes(), &fd7)
	if len(fd7.StaleOpportunities) != 0 {
		t.Errorf("expected inactive (Ghosted) opportunity to not be flagged as stale, got %d", len(fd7.StaleOpportunities))
	}
}
