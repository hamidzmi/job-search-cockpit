package domain_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"job-search-cockpit/internal/domain"
)

func TestInterviewRound_LifecycleAndDebrief(t *testing.T) {
	now := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	scheduledAt := time.Date(2026, 10, 28, 15, 0, 0, 0, time.UTC)
	updatedAt := now.Add(2 * time.Hour)

	tests := []struct {
		name           string
		id             string
		input          domain.InterviewRoundInput
		wantOrder      int
		wantType       domain.InterviewRoundType
		wantTitle      string
		wantDuration   int
		wantStatus     domain.InterviewRoundStatus
		wantContactIDs []string
		wantStoryIDs   []string
		wantErr        error
	}{
		{
			name:    "rejects blank round id",
			id:      "   ",
			input:   domain.InterviewRoundInput{OpportunityID: "opp-1", RoundOrder: 1, RoundType: domain.InterviewRoundTypeRecruiterScreen},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects blank opportunity id",
			id:      "rnd-1",
			input:   domain.InterviewRoundInput{RoundOrder: 1, RoundType: domain.InterviewRoundTypeRecruiterScreen},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects non-positive round order",
			id:      "rnd-1",
			input:   domain.InterviewRoundInput{OpportunityID: "opp-1", RoundOrder: 0, RoundType: domain.InterviewRoundTypeRecruiterScreen},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects when both round type and title are blank",
			id:      "rnd-1",
			input:   domain.InterviewRoundInput{OpportunityID: "opp-1", RoundOrder: 1},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects negative duration",
			id:      "rnd-1",
			input:   domain.InterviewRoundInput{OpportunityID: "opp-1", RoundOrder: 1, RoundType: domain.InterviewRoundTypeTechnical, DurationMinutes: -15},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects invalid round status",
			id:      "rnd-1",
			input:   domain.InterviewRoundInput{OpportunityID: "opp-1", RoundOrder: 1, RoundType: domain.InterviewRoundTypeTechnical, Status: domain.InterviewRoundStatus("Pending")},
			wantErr: domain.ErrValidation,
		},
		{
			name: "creates canonical round with default title, Scheduled status, and deduplicated contact/story IDs",
			id:   " rnd-1 ",
			input: domain.InterviewRoundInput{
				OpportunityID:   " opp-1 ",
				RoundOrder:      1,
				RoundType:       domain.InterviewRoundTypeSystemDesign,
				ScheduledAt:     scheduledAt,
				DurationMinutes: 60,
				PrepNotes:       " Review consistent hashing and Raft ",
				ContactIDs:      []string{" cnt-1 ", "cnt-2", "cnt-1", " "},
				StoryIDs:        []string{" story-1 ", "story-2", "story-1"},
			},
			wantOrder:      1,
			wantType:       domain.InterviewRoundTypeSystemDesign,
			wantTitle:      "System Design",
			wantDuration:   60,
			wantStatus:     domain.InterviewRoundStatusScheduled,
			wantContactIDs: []string{"cnt-1", "cnt-2"},
			wantStoryIDs:   []string{"story-1", "story-2"},
		},
		{
			name: "supports custom round title when round type is custom or omitted",
			id:   "rnd-2",
			input: domain.InterviewRoundInput{
				OpportunityID:   "opp-1",
				RoundOrder:      2,
				Title:           "  Architecture Deep Dive with VP  ",
				DurationMinutes: 45,
				Status:          domain.InterviewRoundStatusCompleted,
			},
			wantOrder:      2,
			wantType:       domain.InterviewRoundType("Architecture Deep Dive with VP"),
			wantTitle:      "Architecture Deep Dive with VP",
			wantDuration:   45,
			wantStatus:     domain.InterviewRoundStatusCompleted,
			wantContactIDs: []string{},
			wantStoryIDs:   []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.NewInterviewRound(tc.id, tc.input, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.RoundOrder != tc.wantOrder || got.RoundType != tc.wantType || got.Title != tc.wantTitle || got.DurationMinutes != tc.wantDuration || got.Status != tc.wantStatus {
				t.Errorf("unexpected InterviewRound: %+v", got)
			}
			if !reflect.DeepEqual(got.ContactIDs, tc.wantContactIDs) {
				t.Errorf("expected ContactIDs %v, got %v", tc.wantContactIDs, got.ContactIDs)
			}
			if !reflect.DeepEqual(got.StoryIDs, tc.wantStoryIDs) {
				t.Errorf("expected StoryIDs %v, got %v", tc.wantStoryIDs, got.StoryIDs)
			}
			if !got.CreatedAt.Equal(now) || !got.UpdatedAt.Equal(now) {
				t.Errorf("expected CreatedAt and UpdatedAt %v, got %v / %v", now, got.CreatedAt, got.UpdatedAt)
			}
		})
	}

	t.Run("records and updates 1:1 Debrief with validation on confidence score and structured questions", func(t *testing.T) {
		rnd, err := domain.NewInterviewRound("rnd-1", domain.InterviewRoundInput{
			OpportunityID:   "opp-1",
			RoundOrder:      1,
			RoundType:       domain.InterviewRoundTypeSystemDesign,
			ScheduledAt:     scheduledAt,
			DurationMinutes: 60,
		}, now)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Invalid confidence scores (< 1 or > 5) rejected
		for _, badScore := range []int{0, -1, 6} {
			if err := rnd.RecordDebrief(domain.DebriefInput{
				ConfidenceScore: badScore,
				WentWell:        "Good",
			}, updatedAt); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("expected ErrValidation for confidence score %d, got %v", badScore, err)
			}
		}

		// Question entry with blank Question text rejected
		if err := rnd.RecordDebrief(domain.DebriefInput{
			ConfidenceScore: 4,
			QuestionsAsked: []domain.DebriefQuestion{
				{Question: "   ", Category: "System Design"},
			},
		}, updatedAt); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation for blank question text, got %v", err)
		}

		err = rnd.RecordDebrief(domain.DebriefInput{
			QuestionsAsked: []domain.DebriefQuestion{
				{
					Question:            "  Design a multi-region rate limiter for payment APIs  ",
					Category:            " System Design ",
					AnsweredWithStoryID: " story-1 ",
					Notes:               " Emphasized token bucket state sync ",
				},
			},
			WentWell:        "  Strong back-of-the-envelope math and failure mode analysis  ",
			ToImprove:       "  Spend more time on client SDK retry jitter  ",
			Signals:         "  Team is migrating from Redis to custom Rust proxy in Q1  ",
			ConfidenceScore: 4,
			FollowUpNotes:   "  Send thank-you note to Alice  ",
		}, updatedAt)
		if err != nil {
			t.Fatalf("unexpected RecordDebrief error: %v", err)
		}
		if rnd.Debrief == nil {
			t.Fatalf("expected Debrief to be recorded")
		}
		if rnd.Debrief.ConfidenceScore != 4 || rnd.Debrief.WentWell != "Strong back-of-the-envelope math and failure mode analysis" {
			t.Errorf("unexpected Debrief fields: %+v", rnd.Debrief)
		}
		if len(rnd.Debrief.QuestionsAsked) != 1 || rnd.Debrief.QuestionsAsked[0].Question != "Design a multi-region rate limiter for payment APIs" || rnd.Debrief.QuestionsAsked[0].AnsweredWithStoryID != "story-1" {
			t.Errorf("unexpected QuestionsAsked: %+v", rnd.Debrief.QuestionsAsked)
		}
		if !rnd.UpdatedAt.Equal(updatedAt) || !rnd.Debrief.UpdatedAt.Equal(updatedAt) {
			t.Errorf("expected UpdatedAt %v, got %v / %v", updatedAt, rnd.UpdatedAt, rnd.Debrief.UpdatedAt)
		}
	})
}

func TestDetectRepeatedStories(t *testing.T) {
	tests := []struct {
		name         string
		rounds       []domain.InterviewRound
		wantWarnings map[string]map[string][]domain.StoryPriorRoundRef
	}{
		{
			name: "flags Story repeated in a later Scheduled or Completed round of the same Opportunity ordered by RoundOrder",
			rounds: []domain.InterviewRound{
				{
					ID:            "rnd-3",
					OpportunityID: "opp-1",
					RoundOrder:    3,
					RoundType:     domain.InterviewRoundTypeExecutive,
					Title:         "Executive Chat",
					Status:        domain.InterviewRoundStatusScheduled,
					StoryIDs:      []string{"story-ledger", "story-conflict"},
				},
				{
					ID:            "rnd-1",
					OpportunityID: "opp-1",
					RoundOrder:    1,
					RoundType:     domain.InterviewRoundTypeHiringManager,
					Title:         "Hiring Manager Screen",
					Status:        domain.InterviewRoundStatusCompleted,
					StoryIDs:      []string{"story-ledger"},
				},
				{
					ID:            "rnd-2",
					OpportunityID: "opp-1",
					RoundOrder:    2,
					RoundType:     domain.InterviewRoundTypeSystemDesign,
					Title:         "System Design",
					Status:        domain.InterviewRoundStatusScheduled,
					StoryIDs:      []string{"story-ledger", "story-sharding"},
				},
			},
			wantWarnings: map[string]map[string][]domain.StoryPriorRoundRef{
				"rnd-2": {
					"story-ledger": {
						{
							RoundID:    "rnd-1",
							RoundOrder: 1,
							RoundType:  domain.InterviewRoundTypeHiringManager,
							Title:      "Hiring Manager Screen",
						},
					},
				},
				"rnd-3": {
					"story-ledger": {
						{
							RoundID:    "rnd-1",
							RoundOrder: 1,
							RoundType:  domain.InterviewRoundTypeHiringManager,
							Title:      "Hiring Manager Screen",
						},
						{
							RoundID:    "rnd-2",
							RoundOrder: 2,
							RoundType:  domain.InterviewRoundTypeSystemDesign,
							Title:      "System Design",
						},
					},
				},
			},
		},
		{
			name: "ignores Cancelled rounds both as sources of prior usage and as recipients of warnings",
			rounds: []domain.InterviewRound{
				{
					ID:            "rnd-1",
					OpportunityID: "opp-1",
					RoundOrder:    1,
					RoundType:     domain.InterviewRoundTypeRecruiterScreen,
					Title:         "Recruiter Screen (Cancelled)",
					Status:        domain.InterviewRoundStatusCancelled,
					StoryIDs:      []string{"story-ledger"},
				},
				{
					ID:            "rnd-2",
					OpportunityID: "opp-1",
					RoundOrder:    2,
					RoundType:     domain.InterviewRoundTypeHiringManager,
					Title:         "Hiring Manager",
					Status:        domain.InterviewRoundStatusCompleted,
					StoryIDs:      []string{"story-ledger"},
				},
				{
					ID:            "rnd-3",
					OpportunityID: "opp-1",
					RoundOrder:    3,
					RoundType:     domain.InterviewRoundTypeBehavioral,
					Title:         "Behavioral (Cancelled)",
					Status:        domain.InterviewRoundStatusCancelled,
					StoryIDs:      []string{"story-ledger"},
				},
				{
					ID:            "rnd-4",
					OpportunityID: "opp-1",
					RoundOrder:    4,
					RoundType:     domain.InterviewRoundTypeBehavioral,
					Title:         "Behavioral Rescheduled",
					Status:        domain.InterviewRoundStatusScheduled,
					StoryIDs:      []string{"story-ledger"},
				},
			},
			wantWarnings: map[string]map[string][]domain.StoryPriorRoundRef{
				"rnd-4": {
					"story-ledger": {
						{
							RoundID:    "rnd-2",
							RoundOrder: 2,
							RoundType:  domain.InterviewRoundTypeHiringManager,
							Title:      "Hiring Manager",
						},
					},
				},
			},
		},
		{
			name: "does not flag Story usage across different Opportunities",
			rounds: []domain.InterviewRound{
				{
					ID:            "rnd-opp1-1",
					OpportunityID: "opp-1",
					RoundOrder:    1,
					RoundType:     domain.InterviewRoundTypeHiringManager,
					Title:         "Stripe HM",
					Status:        domain.InterviewRoundStatusCompleted,
					StoryIDs:      []string{"story-ledger"},
				},
				{
					ID:            "rnd-opp2-1",
					OpportunityID: "opp-2",
					RoundOrder:    1,
					RoundType:     domain.InterviewRoundTypeBehavioral,
					Title:         "Datadog Behavioral",
					Status:        domain.InterviewRoundStatusScheduled,
					StoryIDs:      []string{"story-ledger"},
				},
			},
			wantWarnings: map[string]map[string][]domain.StoryPriorRoundRef{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.DetectRepeatedStories(tc.rounds)
			if !reflect.DeepEqual(got, tc.wantWarnings) {
				t.Errorf("DetectRepeatedStories() = %+v, want %+v", got, tc.wantWarnings)
			}
		})
	}
}
