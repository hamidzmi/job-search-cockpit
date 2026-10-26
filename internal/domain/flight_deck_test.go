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

	tests := []struct {
		name                string
		input               domain.FlightDeckInput
		wantUpcomingIDs     []string
		wantMissingDebriefs []string
		checkFirstUpcoming  func(t *testing.T, round domain.UpcomingInterviewRoundView)
	}{
		{
			name: "filters upcoming rounds to next 48h and identifies past rounds missing debriefs",
			input: domain.FlightDeckInput{
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
			},
			wantUpcomingIDs:     []string{"round-upcoming-12h", "round-upcoming-36h"},
			wantMissingDebriefs: []string{"round-missing-debrief-1", "round-missing-debrief-2"},
			checkFirstUpcoming: func(t *testing.T, round domain.UpcomingInterviewRoundView) {
				if round.CompanyName != "Stripe" {
					t.Errorf("expected company Stripe, got %s", round.CompanyName)
				}
				if len(round.InterviewerNames) != 1 || round.InterviewerNames[0] != "Alice Walker" {
					t.Errorf("expected interviewer Alice Walker, got %v", round.InterviewerNames)
				}
				if len(round.StoryTitles) != 1 || round.StoryTitles[0] != "Scaling Payment Pipeline to 10k RPS" {
					t.Errorf("expected story title, got %v", round.StoryTitles)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fd := domain.EvaluateFlightDeck(tc.input)

			if len(fd.UpcomingRounds) != len(tc.wantUpcomingIDs) {
				t.Fatalf("expected %d upcoming rounds, got %d", len(tc.wantUpcomingIDs), len(fd.UpcomingRounds))
			}
			for i, wantID := range tc.wantUpcomingIDs {
				if fd.UpcomingRounds[i].Round.ID != wantID {
					t.Errorf("upcoming[%d]: expected %s, got %s", i, wantID, fd.UpcomingRounds[i].Round.ID)
				}
			}
			if tc.checkFirstUpcoming != nil && len(fd.UpcomingRounds) > 0 {
				tc.checkFirstUpcoming(t, fd.UpcomingRounds[0])
			}

			if len(fd.MissingDebriefs) != len(tc.wantMissingDebriefs) {
				t.Fatalf("expected %d missing debriefs, got %d", len(tc.wantMissingDebriefs), len(fd.MissingDebriefs))
			}
			for i, wantID := range tc.wantMissingDebriefs {
				if fd.MissingDebriefs[i].Round.ID != wantID {
					t.Errorf("missing debrief[%d]: expected %s, got %s", i, wantID, fd.MissingDebriefs[i].Round.ID)
				}
			}
		})
	}
}

func TestEvaluateFlightDeck_ActionItems(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

	type expectedItem struct {
		id         string
		isOverdue  bool
		isDueToday bool
	}

	tests := []struct {
		name      string
		input     domain.FlightDeckInput
		wantItems []expectedItem
	}{
		{
			name: "orders open action items by due date with mutually exclusive is_overdue and is_due_today flags",
			input: domain.FlightDeckInput{
				Now: now,
				ActionItems: []domain.ActionItemView{
					{
						ActionItem: domain.ActionItem{
							ID:     "item-due-future",
							Title:  "Prep notes for tomorrow",
							DueAt:  now.Add(24 * time.Hour), // Tomorrow (2026-10-27)
							Status: domain.ActionItemStatusOpen,
						},
						CompanyName: "Acme",
					},
					{
						ActionItem: domain.ActionItem{
							ID:     "item-overdue",
							Title:  "Send thank you note",
							DueAt:  now.Add(-36 * time.Hour), // Strictly before start of today (2026-10-25 00:00)
							Status: domain.ActionItemStatusOpen,
						},
						CompanyName: "Beta",
					},
					{
						ActionItem: domain.ActionItem{
							ID:     "item-due-today",
							Title:  "Review system design doc",
							DueAt:  time.Date(2026, 10, 26, 18, 0, 0, 0, time.UTC), // Today at 18:00
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
			},
			wantItems: []expectedItem{
				{id: "item-overdue", isOverdue: true, isDueToday: false},
				{id: "item-due-today", isOverdue: false, isDueToday: true},
				{id: "item-due-future", isOverdue: false, isDueToday: false},
			},
		},
		{
			name: "same-day past timestamp is marked is_due_today and never is_overdue",
			input: domain.FlightDeckInput{
				Now: now, // 2026-10-26 12:00:00 UTC
				ActionItems: []domain.ActionItemView{
					{
						ActionItem: domain.ActionItem{
							ID:     "item-earlier-today",
							Title:  "Morning standup notes",
							DueAt:  time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC), // earlier same calendar day
							Status: domain.ActionItemStatusOpen,
						},
					},
				},
			},
			wantItems: []expectedItem{
				{id: "item-earlier-today", isOverdue: false, isDueToday: true},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fd := domain.EvaluateFlightDeck(tc.input)

			if len(fd.ActionItems) != len(tc.wantItems) {
				t.Fatalf("expected %d open action items, got %d", len(tc.wantItems), len(fd.ActionItems))
			}
			for i, want := range tc.wantItems {
				got := fd.ActionItems[i]
				if got.ActionItem.ID != want.id {
					t.Errorf("item[%d]: expected ID %s, got %s", i, want.id, got.ActionItem.ID)
				}
				if got.IsOverdue != want.isOverdue {
					t.Errorf("item[%s]: IsOverdue = %v, want %v", want.id, got.IsOverdue, want.isOverdue)
				}
				if got.IsDueToday != want.isDueToday {
					t.Errorf("item[%s]: IsDueToday = %v, want %v", want.id, got.IsDueToday, want.isDueToday)
				}
				if got.IsOverdue && got.IsDueToday {
					t.Errorf("item[%s]: IsOverdue and IsDueToday must be mutually exclusive", want.id)
				}
			}
		})
	}
}
