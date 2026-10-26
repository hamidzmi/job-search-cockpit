package httpadapter_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	httpadapter "job-search-cockpit/internal/adapter/http"
)

func TestInterviewRoundsAndDebriefsAPI(t *testing.T) {
	t0 := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	clock := &manualClock{now: t0}
	srv := newTestServer(t, clock)

	// Seed two Companies and two Opportunities
	stripeRec := doJSON(t, srv, http.MethodPost, "/api/companies", map[string]any{"name": "Stripe"})
	var stripe httpadapter.CompanyDTO
	_ = json.Unmarshal(stripeRec.Body.Bytes(), &stripe)

	datadogRec := doJSON(t, srv, http.MethodPost, "/api/companies", map[string]any{"name": "Datadog"})
	var datadog httpadapter.CompanyDTO
	_ = json.Unmarshal(datadogRec.Body.Bytes(), &datadog)

	oppStripeRec := doJSON(t, srv, http.MethodPost, "/api/opportunities", map[string]any{
		"company_id": stripe.ID,
		"role_title": "Staff Core Payments Engineer",
		"stage":      "Interviewing",
	})
	var oppStripe httpadapter.OpportunityDTO
	_ = json.Unmarshal(oppStripeRec.Body.Bytes(), &oppStripe)

	oppDatadogRec := doJSON(t, srv, http.MethodPost, "/api/opportunities", map[string]any{
		"company_id": datadog.ID,
		"role_title": "Principal Ingestion Engineer",
		"stage":      "Interviewing",
	})
	var oppDatadog httpadapter.OpportunityDTO
	_ = json.Unmarshal(oppDatadogRec.Body.Bytes(), &oppDatadog)

	// Seed two Contacts and two Stories
	aliceRec := doJSON(t, srv, http.MethodPost, "/api/contacts", map[string]any{
		"company_id": stripe.ID,
		"name":       "Alice Chen",
		"title":      "Staff Payments Engineer",
		"warmth":     "Advocate",
	})
	var alice httpadapter.ContactDTO
	_ = json.Unmarshal(aliceRec.Body.Bytes(), &alice)

	bobRec := doJSON(t, srv, http.MethodPost, "/api/contacts", map[string]any{
		"company_id": stripe.ID,
		"name":       "Bob Martinez",
		"title":      "Engineering Director",
		"warmth":     "Warm",
	})
	var bob httpadapter.ContactDTO
	_ = json.Unmarshal(bobRec.Body.Bytes(), &bob)

	storyLedgerRec := doJSON(t, srv, http.MethodPost, "/api/stories", map[string]any{
		"title":           "Zero-Downtime Ledger Sharding",
		"hook":            "Sharded $40B/yr ledger across 16 clusters with zero downtime.",
		"competency_tags": []string{"System Design", "Leadership"},
		"skill_tags":      []string{"Go", "PostgreSQL"},
	})
	var storyLedger httpadapter.StoryDTO
	_ = json.Unmarshal(storyLedgerRec.Body.Bytes(), &storyLedger)

	storyConflictRec := doJSON(t, srv, http.MethodPost, "/api/stories", map[string]any{
		"title":           "Unblocking Cross-Team Schema Deadlock",
		"hook":            "Aligned 3 platform teams on unified event schema.",
		"competency_tags": []string{"Conflict", "Ownership"},
		"skill_tags":      []string{"Kafka", "Rust"},
	})
	var storyConflict httpadapter.StoryDTO
	_ = json.Unmarshal(storyConflictRec.Body.Bytes(), &storyConflict)

	var round1ID, round2ID, round3ID string

	t.Run("creates ordered InterviewRounds, links Contacts and Stories, advances Opportunity.LastActivityAt, and detects repeated Stories across non-cancelled rounds", func(t *testing.T) {
		// Round 1: Recruiter Screen (Cancelled) using storyLedger -> should NOT trigger repeat warnings
		clock.now = t0.Add(24 * time.Hour)
		r1Rec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+oppStripe.ID+"/rounds", map[string]any{
			"round_order":      1,
			"round_type":       "Recruiter Screen",
			"title":            "Initial Recruiter Screen",
			"scheduled_at":     t0.Add(26 * time.Hour).Format(time.RFC3339),
			"duration_minutes": 30,
			"status":           "Cancelled",
			"prep_notes":       "Rescheduled directly to HM",
			"contact_ids":      []string{},
			"story_ids":        []string{storyLedger.ID},
		})
		if r1Rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 creating round 1, got %d: %s", r1Rec.Code, r1Rec.Body.String())
		}
		var r1 httpadapter.InterviewRoundDTO
		_ = json.Unmarshal(r1Rec.Body.Bytes(), &r1)
		round1ID = r1.ID

		// Round 2: Hiring Manager (Completed) using storyLedger -> first non-cancelled use of storyLedger
		clock.now = t0.Add(48 * time.Hour)
		r2ScheduledAt := t0.Add(47 * time.Hour)
		r2Rec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+oppStripe.ID+"/rounds", map[string]any{
			"round_order":      2,
			"round_type":       "Hiring Manager",
			"title":            "HM Deep Dive",
			"scheduled_at":     r2ScheduledAt.Format(time.RFC3339),
			"duration_minutes": 45,
			"status":           "Completed",
			"prep_notes":       "Discuss ledger reliability",
			"contact_ids":      []string{alice.ID},
			"story_ids":        []string{storyLedger.ID},
		})
		if r2Rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 creating round 2, got %d: %s", r2Rec.Code, r2Rec.Body.String())
		}
		var r2 httpadapter.InterviewRoundDTO
		_ = json.Unmarshal(r2Rec.Body.Bytes(), &r2)
		round2ID = r2.ID
		if len(r2.Contacts) != 1 || r2.Contacts[0].Name != "Alice Chen" {
			t.Errorf("expected Alice linked to round 2, got %+v", r2.Contacts)
		}
		if len(r2.Stories) != 1 || len(r2.Stories[0].UsedInEarlierRounds) != 0 {
			t.Errorf("expected storyLedger in round 2 to have 0 prior warnings (round 1 was Cancelled), got %+v", r2.Stories)
		}

		// Verify Opportunity.LastActivityAt advanced when Round 2 was created/completed
		oppAfterR2Rec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+oppStripe.ID, nil)
		var oppAfterR2 httpadapter.OpportunityDTO
		_ = json.Unmarshal(oppAfterR2Rec.Body.Bytes(), &oppAfterR2)
		if !oppAfterR2.LastActivityAt.Equal(clock.now) {
			t.Errorf("expected Opportunity.LastActivityAt to advance to %v, got %v", clock.now, oppAfterR2.LastActivityAt)
		}

		// Round 3: System Design (Scheduled) using storyLedger and storyConflict -> storyLedger must have repeat warning referencing Round 2
		clock.now = t0.Add(72 * time.Hour)
		r3Rec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+oppStripe.ID+"/rounds", map[string]any{
			"round_order":      3,
			"round_type":       "System Design",
			"title":            "Distributed Ledger Architecture",
			"scheduled_at":     t0.Add(96 * time.Hour).Format(time.RFC3339),
			"duration_minutes": 60,
			"status":           "Scheduled",
			"prep_notes":       "Whiteboard multi-region settlement",
			"contact_ids":      []string{alice.ID, bob.ID},
			"story_ids":        []string{storyLedger.ID, storyConflict.ID},
		})
		if r3Rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 creating round 3, got %d: %s", r3Rec.Code, r3Rec.Body.String())
		}
		var r3 httpadapter.InterviewRoundDTO
		_ = json.Unmarshal(r3Rec.Body.Bytes(), &r3)
		round3ID = r3.ID
		if len(r3.Stories) != 2 {
			t.Fatalf("expected 2 stories on round 3, got %+v", r3.Stories)
		}
		if len(r3.Stories[0].UsedInEarlierRounds) != 1 || r3.Stories[0].UsedInEarlierRounds[0].RoundID != round2ID || r3.Stories[0].UsedInEarlierRounds[0].RoundOrder != 2 {
			t.Errorf("expected storyLedger on round 3 to warn about prior use in round 2, got %+v", r3.Stories[0].UsedInEarlierRounds)
		}
		if len(r3.Stories[1].UsedInEarlierRounds) != 0 {
			t.Errorf("expected storyConflict on round 3 to have no prior warnings, got %+v", r3.Stories[1].UsedInEarlierRounds)
		}

		// Create Round 1 on Datadog Opportunity using storyLedger -> should NOT warn across different Opportunities
		ddRoundRec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+oppDatadog.ID+"/rounds", map[string]any{
			"round_order":      1,
			"round_type":       "Technical",
			"title":            "Datadog Systems Round",
			"scheduled_at":     t0.Add(100 * time.Hour).Format(time.RFC3339),
			"duration_minutes": 60,
			"status":           "Scheduled",
			"story_ids":        []string{storyLedger.ID},
		})
		if ddRoundRec.Code != http.StatusCreated {
			t.Fatalf("expected 201 creating Datadog round, got %d: %s", ddRoundRec.Code, ddRoundRec.Body.String())
		}
		var ddRound httpadapter.InterviewRoundDTO
		_ = json.Unmarshal(ddRoundRec.Body.Bytes(), &ddRound)
		if len(ddRound.Stories) != 1 || len(ddRound.Stories[0].UsedInEarlierRounds) != 0 {
			t.Errorf("expected Datadog round to have 0 warnings for storyLedger used in Stripe loop, got %+v", ddRound.Stories)
		}
	})

	t.Run("reorders and updates InterviewRound via PUT and re-evaluates repeat warnings", func(t *testing.T) {
		// Swap Round 2 (order -> 3) and Round 3 (order -> 2)
		clock.now = clock.now.Add(time.Hour)
		updR2 := doJSON(t, srv, http.MethodPut, "/api/rounds/"+round2ID, map[string]any{
			"round_order":      3,
			"round_type":       "Hiring Manager",
			"title":            "HM Deep Dive",
			"scheduled_at":     t0.Add(120 * time.Hour).Format(time.RFC3339),
			"duration_minutes": 45,
			"status":           "Completed",
			"prep_notes":       "Discuss ledger reliability",
			"contact_ids":      []string{alice.ID},
			"story_ids":        []string{storyLedger.ID},
		})
		if updR2.Code != http.StatusOK {
			t.Fatalf("expected 200 updating round 2 order, got %d: %s", updR2.Code, updR2.Body.String())
		}

		updR3 := doJSON(t, srv, http.MethodPut, "/api/rounds/"+round3ID, map[string]any{
			"round_order":      2,
			"round_type":       "System Design",
			"title":            "Distributed Ledger Architecture",
			"scheduled_at":     t0.Add(96 * time.Hour).Format(time.RFC3339),
			"duration_minutes": 60,
			"status":           "Completed",
			"prep_notes":       "Whiteboard multi-region settlement",
			"contact_ids":      []string{alice.ID, bob.ID},
			"story_ids":        []string{storyLedger.ID, storyConflict.ID},
		})
		if updR3.Code != http.StatusOK {
			t.Fatalf("expected 200 updating round 3 order, got %d: %s", updR3.Code, updR3.Body.String())
		}

		listRec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+oppStripe.ID+"/rounds", nil)
		if listRec.Code != http.StatusOK {
			t.Fatalf("expected 200 listing rounds, got %d", listRec.Code)
		}
		var rounds []httpadapter.InterviewRoundDTO
		_ = json.Unmarshal(listRec.Body.Bytes(), &rounds)
		if len(rounds) != 3 {
			t.Fatalf("expected 3 rounds, got %d", len(rounds))
		}
		// Ordered by RoundOrder: round1 (1), round3 (2), round2 (3)
		if rounds[0].ID != round1ID || rounds[1].ID != round3ID || rounds[2].ID != round2ID {
			t.Fatalf("expected rounds ordered [round1, round3, round2], got [%s, %s, %s]", rounds[0].ID, rounds[1].ID, rounds[2].ID)
		}
		// Now round3 is order 2 (no prior active rounds) and round2 is order 3 (warns that storyLedger was used in round3!)
		if len(rounds[1].Stories[0].UsedInEarlierRounds) != 0 {
			t.Errorf("expected round3 (now order 2) to have 0 warnings, got %+v", rounds[1].Stories[0].UsedInEarlierRounds)
		}
		if len(rounds[2].Stories[0].UsedInEarlierRounds) != 1 || rounds[2].Stories[0].UsedInEarlierRounds[0].RoundID != round3ID {
			t.Errorf("expected round2 (now order 3) to warn about round3 (order 2), got %+v", rounds[2].Stories[0].UsedInEarlierRounds)
		}
	})

	t.Run("records 1:1 Debrief on InterviewRounds and aggregates structured questions in /api/questions", func(t *testing.T) {
		clock.now = clock.now.Add(2 * time.Hour)

		// Validation: confidence score out of range -> 400
		badDebrief := doJSON(t, srv, http.MethodPut, "/api/rounds/"+round3ID+"/debrief", map[string]any{
			"confidence_score": 6,
			"went_well":        "Great",
		})
		if badDebrief.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid confidence_score=6, got %d: %s", badDebrief.Code, badDebrief.Body.String())
		}

		// Save Debrief on Round 3 (System Design)
		debriefR3Rec := doJSON(t, srv, http.MethodPut, "/api/rounds/"+round3ID+"/debrief", map[string]any{
			"questions_asked": []map[string]any{
				{
					"question":               "How do you guarantee idempotency across multi-region ledger writes?",
					"category":               "System Design",
					"answered_with_story_id": storyLedger.ID,
					"notes":                  "Drew deterministic fencing token diagram",
				},
				{
					"question":               "Tell me about a time you resolved a cross-team architectural disagreement.",
					"category":               "Behavioral",
					"answered_with_story_id": storyConflict.ID,
					"notes":                  "Shared Rust protobuf benchmark story",
				},
			},
			"went_well":        "Deep dive on fencing tokens and linearizability landed well",
			"to_improve":       "Could have budgeted 5 more minutes for observability metrics",
			"signals":          "Team is actively splitting monolithic ledger into regional cells",
			"confidence_score": 5,
			"follow_up_notes":  "Send paper reference in thank-you email",
		})
		if debriefR3Rec.Code != http.StatusOK {
			t.Fatalf("expected 200 saving debrief, got %d: %s", debriefR3Rec.Code, debriefR3Rec.Body.String())
		}
		var updatedRound httpadapter.InterviewRoundDTO
		_ = json.Unmarshal(debriefR3Rec.Body.Bytes(), &updatedRound)
		if updatedRound.Debrief == nil || updatedRound.Debrief.ConfidenceScore != 5 || len(updatedRound.Debrief.QuestionsAsked) != 2 {
			t.Fatalf("unexpected debrief on round DTO: %+v", updatedRound.Debrief)
		}
		if updatedRound.Debrief.QuestionsAsked[0].AnsweredWithStoryTitle != "Zero-Downtime Ledger Sharding" {
			t.Errorf("expected AnsweredWithStoryTitle 'Zero-Downtime Ledger Sharding', got %q", updatedRound.Debrief.QuestionsAsked[0].AnsweredWithStoryTitle)
		}

		// Query aggregated question bank: GET /api/questions
		qAllRec := doJSON(t, srv, http.MethodGet, "/api/questions", nil)
		if qAllRec.Code != http.StatusOK {
			t.Fatalf("expected 200 on GET /api/questions, got %d: %s", qAllRec.Code, qAllRec.Body.String())
		}
		var allQuestions []httpadapter.QuestionBankEntryDTO
		_ = json.Unmarshal(qAllRec.Body.Bytes(), &allQuestions)
		if len(allQuestions) != 2 {
			t.Fatalf("expected 2 aggregated questions, got %d: %+v", len(allQuestions), allQuestions)
		}
		if allQuestions[0].CompanyName != "Stripe" || allQuestions[0].OpportunityRoleTitle != "Staff Core Payments Engineer" || allQuestions[0].RoundType != "System Design" {
			t.Errorf("unexpected question bank metadata: %+v", allQuestions[0])
		}

		// Filter question bank by query & category
		qFilterRec := doJSON(t, srv, http.MethodGet, "/api/questions?q=idempotency&category=System+Design", nil)
		var filteredQuestions []httpadapter.QuestionBankEntryDTO
		_ = json.Unmarshal(qFilterRec.Body.Bytes(), &filteredQuestions)
		if len(filteredQuestions) != 1 || filteredQuestions[0].AnsweredWithStoryTitle != "Zero-Downtime Ledger Sharding" {
			t.Errorf("unexpected filtered question bank results: %+v", filteredQuestions)
		}
	})
}
