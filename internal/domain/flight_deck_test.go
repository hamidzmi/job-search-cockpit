package domain_test

import (
	"testing"
	"time"

	"job-search-cockpit/internal/domain"
)

func TestEvaluateFlightDeck_StalenessRules(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		input         domain.FlightDeckInput
		wantStaleIDs  []string
		wantReasons   map[string]string
	}{
		{
			name: "Discovered stale when exceeding 3 days with no activity",
			input: domain.FlightDeckInput{
				Now: now,
				Opportunities: []domain.OpportunityView{
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-disc-stale",
							RoleTitle:      "Staff Engineer",
							Stage:          domain.StageDiscovered,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-73 * time.Hour), // 3 days + 1 hr
							CreatedAt:      now.Add(-80 * time.Hour),
						},
						CompanyName: "Acme Corp",
					},
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-disc-fresh",
							RoleTitle:      "Senior Engineer",
							Stage:          domain.StageDiscovered,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-48 * time.Hour), // 2 days
							CreatedAt:      now.Add(-48 * time.Hour),
						},
						CompanyName: "Beta Co",
					},
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-disc-inactive",
							RoleTitle:      "Lead Engineer",
							Stage:          domain.StageDiscovered,
							Outcome:        domain.OutcomeRejected, // Not Active
							LastActivityAt: now.Add(-100 * time.Hour),
							CreatedAt:      now.Add(-100 * time.Hour),
						},
						CompanyName: "Gamma Inc",
					},
				},
				StageTransitions: []domain.StageTransition{
					{
						OpportunityID:  "opp-disc-stale",
						ToStage:        domain.StageDiscovered,
						TransitionedAt: now.Add(-80 * time.Hour),
					},
				},
			},
			wantStaleIDs: []string{"opp-disc-stale"},
		},
		{
			name: "Discovered staleness resets when LastActivityAt is newer than stage transition",
			input: domain.FlightDeckInput{
				Now: now,
				Opportunities: []domain.OpportunityView{
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-disc-recent-act",
							RoleTitle:      "Staff Engineer",
							Stage:          domain.StageDiscovered,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-24 * time.Hour), // 1 day ago activity
							CreatedAt:      now.Add(-120 * time.Hour),
						},
						CompanyName: "Acme Corp",
					},
				},
				StageTransitions: []domain.StageTransition{
					{
						OpportunityID:  "opp-disc-recent-act",
						ToStage:        domain.StageDiscovered,
						TransitionedAt: now.Add(-120 * time.Hour), // 5 days ago transition
					},
				},
			},
			wantStaleIDs: nil,
		},
		{
			name: "Outreach stale when exceeding 4 days with no interaction",
			input: domain.FlightDeckInput{
				Now: now,
				Opportunities: []domain.OpportunityView{
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-outreach-stale",
							RoleTitle:      "Principal Engineer",
							Stage:          domain.StageOutreach,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-100 * time.Hour),
							CreatedAt:      now.Add(-120 * time.Hour),
						},
						CompanyName: "Delta Labs",
					},
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-outreach-fresh",
							RoleTitle:      "Tech Lead",
							Stage:          domain.StageOutreach,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-24 * time.Hour),
							CreatedAt:      now.Add(-120 * time.Hour),
						},
						CompanyName: "Echo AI",
					},
				},
				StageTransitions: []domain.StageTransition{
					{
						OpportunityID:  "opp-outreach-stale",
						ToStage:        domain.StageOutreach,
						TransitionedAt: now.Add(-120 * time.Hour),
					},
					{
						OpportunityID:  "opp-outreach-fresh",
						ToStage:        domain.StageOutreach,
						TransitionedAt: now.Add(-120 * time.Hour),
					},
				},
				Interactions: []domain.Interaction{
					{
						OpportunityID: "opp-outreach-stale",
						OccurredAt:    now.Add(-97 * time.Hour), // 4 days + 1 hour ago
					},
					{
						OpportunityID: "opp-outreach-fresh",
						OccurredAt:    now.Add(-24 * time.Hour), // 1 day ago
					},
				},
			},
			wantStaleIDs: []string{"opp-outreach-stale"},
		},
		{
			name: "Applied stale when exceeding 10 days with no activity",
			input: domain.FlightDeckInput{
				Now: now,
				Opportunities: []domain.OpportunityView{
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-applied-stale",
							RoleTitle:      "Software Architect",
							Stage:          domain.StageApplied,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-241 * time.Hour), // > 10 days
							CreatedAt:      now.Add(-250 * time.Hour),
						},
						CompanyName: "Foxtrot Systems",
					},
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-applied-fresh",
							RoleTitle:      "Senior Backend Engineer",
							Stage:          domain.StageApplied,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-72 * time.Hour), // 3 days ago
							CreatedAt:      now.Add(-250 * time.Hour),
						},
						CompanyName: "Golf Tech",
					},
				},
				StageTransitions: []domain.StageTransition{
					{
						OpportunityID:  "opp-applied-stale",
						ToStage:        domain.StageApplied,
						TransitionedAt: now.Add(-250 * time.Hour),
					},
					{
						OpportunityID:  "opp-applied-fresh",
						ToStage:        domain.StageApplied,
						TransitionedAt: now.Add(-250 * time.Hour),
					},
				},
			},
			wantStaleIDs: []string{"opp-applied-stale"},
		},
		{
			name: "Screening/Interviewing stale when exceeding 5 days after completed round with no upcoming round scheduled",
			input: domain.FlightDeckInput{
				Now: now,
				Opportunities: []domain.OpportunityView{
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-screening-stale",
							RoleTitle:      "Platform Engineer",
							Stage:          domain.StageScreening,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-144 * time.Hour),
							CreatedAt:      now.Add(-200 * time.Hour),
						},
						CompanyName: "Hotel Corp",
					},
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-interviewing-with-upcoming",
							RoleTitle:      "Infrastructure Lead",
							Stage:          domain.StageInterviewing,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-144 * time.Hour),
							CreatedAt:      now.Add(-200 * time.Hour),
						},
						CompanyName: "India Inc",
					},
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-interviewing-recent-completed",
							RoleTitle:      "Site Reliability Engineer",
							Stage:          domain.StageInterviewing,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-48 * time.Hour),
							CreatedAt:      now.Add(-200 * time.Hour),
						},
						CompanyName: "Juliet Co",
					},
					{
						Opportunity: domain.Opportunity{
							ID:             "opp-screening-no-rounds",
							RoleTitle:      "Frontend Architect",
							Stage:          domain.StageScreening,
							Outcome:        domain.OutcomeActive,
							LastActivityAt: now.Add(-48 * time.Hour),
							CreatedAt:      now.Add(-100 * time.Hour),
						},
						CompanyName: "Kilo LLC",
					},
				},
				InterviewRounds: []domain.InterviewRound{
					// opp-screening-stale: completed 6 days ago, no upcoming round
					{
						ID:            "round-1",
						OpportunityID: "opp-screening-stale",
						Title:         "Recruiter Screen",
						RoundType:     domain.InterviewRoundTypeRecruiterScreen,
						Status:        domain.InterviewRoundStatusCompleted,
						ScheduledAt:   now.Add(-144 * time.Hour), // 6 days ago
					},
					// opp-interviewing-with-upcoming: completed 6 days ago, BUT upcoming in 24 hours
					{
						ID:            "round-2",
						OpportunityID: "opp-interviewing-with-upcoming",
						Title:         "Technical Screen",
						RoundType:     domain.InterviewRoundTypeTechnical,
						Status:        domain.InterviewRoundStatusCompleted,
						ScheduledAt:   now.Add(-144 * time.Hour),
					},
					{
						ID:            "round-3",
						OpportunityID: "opp-interviewing-with-upcoming",
						Title:         "System Design",
						RoundType:     domain.InterviewRoundTypeSystemDesign,
						Status:        domain.InterviewRoundStatusScheduled,
						ScheduledAt:   now.Add(24 * time.Hour), // upcoming
					},
					// opp-interviewing-recent-completed: completed 2 days ago
					{
						ID:            "round-4",
						OpportunityID: "opp-interviewing-recent-completed",
						Title:         "Hiring Manager",
						RoundType:     domain.InterviewRoundTypeHiringManager,
						Status:        domain.InterviewRoundStatusCompleted,
						ScheduledAt:   now.Add(-48 * time.Hour),
					},
				},
			},
			wantStaleIDs: []string{"opp-screening-stale"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fd := domain.EvaluateFlightDeck(tc.input)

			var gotStaleIDs []string
			for _, s := range fd.StaleOpportunities {
				gotStaleIDs = append(gotStaleIDs, s.OpportunityID)
			}

			if len(gotStaleIDs) != len(tc.wantStaleIDs) {
				t.Fatalf("expected %d stale opportunities, got %d (%v vs %v)",
					len(tc.wantStaleIDs), len(gotStaleIDs), tc.wantStaleIDs, gotStaleIDs)
			}
			for i, wantID := range tc.wantStaleIDs {
				if gotStaleIDs[i] != wantID {
					t.Errorf("stale[%d]: expected ID %s, got %s", i, wantID, gotStaleIDs[i])
				}
			}
		})
	}
}

func TestEvaluateFlightDeck_UpcomingAndMissingDebriefs(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

	input := domain.FlightDeckInput{
		Now: now,
		Opportunities: []domain.OpportunityView{
			{
				Opportunity: domain.Opportunity{
					ID:        "opp-1",
					RoleTitle: "Staff Engineer",
					Stage:     domain.StageInterviewing,
					Outcome:   domain.OutcomeActive,
				},
				CompanyName: "Stripe",
			},
		},
		InterviewRounds: []domain.InterviewRound{
			// Round 1: in 12 hours (within 48h) -> upcoming!
			{
				ID:              "round-upcoming-12h",
				OpportunityID:   "opp-1",
				Title:           "Architecture Deep Dive",
				RoundType:       domain.InterviewRoundTypeSystemDesign,
				Status:          domain.InterviewRoundStatusScheduled,
				ScheduledAt:     now.Add(12 * time.Hour),
				DurationMinutes: 60,
				ContactIDs:      []string{"contact-1"},
				StoryIDs:        []string{"story-1"},
			},
			// Round 2: in 36 hours (within 48h) -> upcoming!
			{
				ID:              "round-upcoming-36h",
				OpportunityID:   "opp-1",
				Title:           "Values & Leadership",
				RoundType:       domain.InterviewRoundTypeBehavioral,
				Status:          domain.InterviewRoundStatusScheduled,
				ScheduledAt:     now.Add(36 * time.Hour),
				DurationMinutes: 45,
			},
			// Round 3: in 50 hours (> 48h) -> excluded from upcoming
			{
				ID:              "round-future-50h",
				OpportunityID:   "opp-1",
				Title:           "Executive Chat",
				RoundType:       domain.InterviewRoundTypeExecutive,
				Status:          domain.InterviewRoundStatusScheduled,
				ScheduledAt:     now.Add(50 * time.Hour),
				DurationMinutes: 30,
			},
			// Round 4: completed 2 hours ago without debrief -> missing debrief!
			{
				ID:              "round-missing-debrief-1",
				OpportunityID:   "opp-1",
				Title:           "Coding Round",
				RoundType:       domain.InterviewRoundTypeTechnical,
				Status:          domain.InterviewRoundStatusCompleted,
				ScheduledAt:     now.Add(-2 * time.Hour),
				DurationMinutes: 60,
			},
			// Round 5: scheduled 5 hours ago (past), still Scheduled, without debrief -> missing debrief!
			{
				ID:              "round-missing-debrief-2",
				OpportunityID:   "opp-1",
				Title:           "Recruiter Screen",
				RoundType:       domain.InterviewRoundTypeRecruiterScreen,
				Status:          domain.InterviewRoundStatusScheduled,
				ScheduledAt:     now.Add(-5 * time.Hour),
				DurationMinutes: 30,
			},
			// Round 6: completed 10 hours ago WITH debrief -> excluded from missing debriefs
			{
				ID:              "round-with-debrief",
				OpportunityID:   "opp-1",
				Title:           "HM Chat",
				RoundType:       domain.InterviewRoundTypeHiringManager,
				Status:          domain.InterviewRoundStatusCompleted,
				ScheduledAt:     now.Add(-10 * time.Hour),
				DurationMinutes: 45,
				Debrief: &domain.Debrief{
					WentWell:        "Great rapport",
					Signals:         "Strong technical alignment",
					ConfidenceScore: 4,
				},
			},
			// Round 7: cancelled round in past without debrief -> excluded
			{
				ID:              "round-cancelled",
				OpportunityID:   "opp-1",
				Title:           "Cancelled Screen",
				RoundType:       domain.InterviewRoundTypeRecruiterScreen,
				Status:          domain.InterviewRoundStatusCancelled,
				ScheduledAt:     now.Add(-20 * time.Hour),
				DurationMinutes: 30,
			},
		},
		ContactNames: map[string]string{
			"contact-1": "Alice Walker",
		},
		StoryTitles: map[string]string{
			"story-1": "Scaling Payment Pipeline to 10k RPS",
		},
	}

	fd := domain.EvaluateFlightDeck(input)

	// Verify Upcoming Rounds
	if len(fd.UpcomingRounds) != 2 {
		t.Fatalf("expected 2 upcoming rounds within 48h, got %d", len(fd.UpcomingRounds))
	}
	if fd.UpcomingRounds[0].Round.ID != "round-upcoming-12h" {
		t.Errorf("expected first upcoming round round-upcoming-12h, got %s", fd.UpcomingRounds[0].Round.ID)
	}
	if fd.UpcomingRounds[0].CompanyName != "Stripe" {
		t.Errorf("expected company Stripe, got %s", fd.UpcomingRounds[0].CompanyName)
	}
	if len(fd.UpcomingRounds[0].InterviewerNames) != 1 || fd.UpcomingRounds[0].InterviewerNames[0] != "Alice Walker" {
		t.Errorf("expected interviewer Alice Walker, got %v", fd.UpcomingRounds[0].InterviewerNames)
	}
	if len(fd.UpcomingRounds[0].StoryTitles) != 1 || fd.UpcomingRounds[0].StoryTitles[0] != "Scaling Payment Pipeline to 10k RPS" {
		t.Errorf("expected story title, got %v", fd.UpcomingRounds[0].StoryTitles)
	}
	if fd.UpcomingRounds[1].Round.ID != "round-upcoming-36h" {
		t.Errorf("expected second upcoming round round-upcoming-36h, got %s", fd.UpcomingRounds[1].Round.ID)
	}

	// Verify Missing Debriefs
	if len(fd.MissingDebriefs) != 2 {
		t.Fatalf("expected 2 missing debrief rounds, got %d", len(fd.MissingDebriefs))
	}
	// Missing debriefs ordered by most recent first (-2h before -5h)
	if fd.MissingDebriefs[0].Round.ID != "round-missing-debrief-1" {
		t.Errorf("expected most recent missing debrief first, got %s", fd.MissingDebriefs[0].Round.ID)
	}
	if fd.MissingDebriefs[1].Round.ID != "round-missing-debrief-2" {
		t.Errorf("expected second missing debrief, got %s", fd.MissingDebriefs[1].Round.ID)
	}
}

func TestEvaluateFlightDeck_ActionItems(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

	input := domain.FlightDeckInput{
		Now: now,
		ActionItems: []domain.ActionItemView{
			{
				ActionItem: domain.ActionItem{
					ID:     "item-due-future",
					Title:  "Prep notes for tomorrow",
					DueAt:  now.Add(24 * time.Hour),
					Status: domain.ActionItemStatusOpen,
				},
				CompanyName: "Acme",
			},
			{
				ActionItem: domain.ActionItem{
					ID:     "item-overdue",
					Title:  "Send thank you note",
					DueAt:  now.Add(-2 * time.Hour), // overdue
					Status: domain.ActionItemStatusOpen,
				},
				CompanyName: "Beta",
			},
			{
				ActionItem: domain.ActionItem{
					ID:     "item-due-today",
					Title:  "Review system design doc",
					DueAt:  time.Date(2026, 10, 26, 18, 0, 0, 0, time.UTC), // today at 18:00
					Status: domain.ActionItemStatusOpen,
				},
				CompanyName: "Gamma",
			},
			{
				ActionItem: domain.ActionItem{
					ID:     "item-completed",
					Title:  "Already done item",
					DueAt:  now.Add(-48 * time.Hour),
					Status: domain.ActionItemStatusCompleted, // Not open
				},
			},
		},
	}

	fd := domain.EvaluateFlightDeck(input)

	if len(fd.ActionItems) != 3 {
		t.Fatalf("expected 3 open action items, got %d", len(fd.ActionItems))
	}

	// Should be ordered by DueAt ASC: overdue (-2h) < due today (18:00) < future (+24h)
	if fd.ActionItems[0].ActionItem.ID != "item-overdue" {
		t.Errorf("expected first item to be item-overdue, got %s", fd.ActionItems[0].ActionItem.ID)
	}
	if !fd.ActionItems[0].IsOverdue {
		t.Errorf("expected item-overdue to have IsOverdue = true")
	}

	if fd.ActionItems[1].ActionItem.ID != "item-due-today" {
		t.Errorf("expected second item to be item-due-today, got %s", fd.ActionItems[1].ActionItem.ID)
	}
	if !fd.ActionItems[1].IsDueToday {
		t.Errorf("expected item-due-today to have IsDueToday = true")
	}

	if fd.ActionItems[2].ActionItem.ID != "item-due-future" {
		t.Errorf("expected third item to be item-due-future, got %s", fd.ActionItems[2].ActionItem.ID)
	}
	if fd.ActionItems[2].IsOverdue || fd.ActionItems[2].IsDueToday {
		t.Errorf("expected future item to not be overdue or due today")
	}
}
