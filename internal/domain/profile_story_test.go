package domain_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"job-search-cockpit/internal/domain"
)

func TestNewCandidateProfile(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		input     domain.CandidateProfileInput
		wantFloor int
		wantCurr  string
		wantErr   error
	}{
		{
			name: "normalizes work history, impact bullets, skills, preferences, and compensation floor",
			input: domain.CandidateProfileInput{
				FullName: "  Alex Rivera  ",
				Headline: "  Staff Distributed Systems Engineer ",
				Summary:  "  12 years building low-latency payment and storage systems. ",
				WorkHistory: []domain.WorkHistoryEntry{
					{
						Company:   "  Stripe ",
						Title:     " Staff Engineer ",
						StartDate: " 2022-01 ",
						EndDate:   " Present ",
						Summary:   " Led global ledger reliability. ",
						ImpactBullets: []string{
							" Reduced p99 settlement latency by 42% across 3 regions ",
							"   ",
							" Mentored 8 senior engineers through promotion ",
						},
					},
				},
				Skills:              []string{" Go ", "Kafka", "go", "  ", "PostgreSQL"},
				TargetRoleTitles:    []string{" Staff Backend Engineer ", "Principal Systems Engineer"},
				LocationPreferences: []string{" San Francisco, CA ", "Remote (US)"},
				RemotePreference:    " Remote ",
				CompensationFloor:   240000,
				TargetCompensation:  285000,
				Currency:            "",
			},
			wantFloor: 240000,
			wantCurr:  "USD",
			wantErr:   nil,
		},
		{
			name: "rejects negative compensation floor",
			input: domain.CandidateProfileInput{
				CompensationFloor: -100,
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "rejects negative target compensation",
			input: domain.CandidateProfileInput{
				TargetCompensation: -100,
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "rejects target compensation below compensation floor when both are set",
			input: domain.CandidateProfileInput{
				CompensationFloor:  250000,
				TargetCompensation: 200000,
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "rejects work history entry missing company or title",
			input: domain.CandidateProfileInput{
				WorkHistory: []domain.WorkHistoryEntry{
					{
						Company: "Stripe",
						Title:   "   ",
					},
				},
			},
			wantErr: domain.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.NewCandidateProfile(tc.input, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.FullName != "Alex Rivera" {
				t.Errorf("expected FullName 'Alex Rivera', got %q", got.FullName)
			}
			if got.CompensationFloor != tc.wantFloor {
				t.Errorf("expected CompensationFloor %d, got %d", tc.wantFloor, got.CompensationFloor)
			}
			if got.Currency != tc.wantCurr {
				t.Errorf("expected Currency %q, got %q", tc.wantCurr, got.Currency)
			}
			wantSkills := []string{"Go", "Kafka", "PostgreSQL"}
			if !reflect.DeepEqual(got.Skills, wantSkills) {
				t.Errorf("expected Skills %v, got %v", wantSkills, got.Skills)
			}
			if len(got.WorkHistory) != 1 || len(got.WorkHistory[0].ImpactBullets) != 2 {
				t.Errorf("expected 1 work history entry with 2 cleaned impact bullets, got %+v", got.WorkHistory)
			}
			if !got.UpdatedAt.Equal(now) {
				t.Errorf("expected UpdatedAt %v, got %v", now, got.UpdatedAt)
			}
		})
	}
}

func TestStory_LifecycleAndFiltering(t *testing.T) {
	createdAt := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 10, 22, 15, 30, 0, 0, time.UTC)

	createCases := []struct {
		name    string
		id      string
		input   domain.StoryInput
		wantErr error
	}{
		{
			name:    "rejects blank story id",
			id:      "   ",
			input:   domain.StoryInput{Title: "Ledger Migration"},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects blank story title",
			id:      "story-1",
			input:   domain.StoryInput{Title: "   "},
			wantErr: domain.ErrValidation,
		},
		{
			name: "creates story with STAR fields and deduplicated competency and skill tags",
			id:   "story-1",
			input: domain.StoryInput{
				Title:          "  Zero-Downtime Ledger Sharding  ",
				Hook:           " Migrated $40B/yr ledger across 16 shards with zero customer downtime ",
				Situation:      " Single primary PostgreSQL cluster hit 85% write saturation prior to Black Friday. ",
				Task:           " Design and execute online horizontal sharding within 10 weeks. ",
				Action:         " Built dual-write shadow verifier in Go and coordinated phased cutover. ",
				Result:         " Cut write latency by 58% and handled 3.2x peak volume with zero incidents. ",
				CompetencyTags: []string{" System Design ", "Leadership", "system design", " "},
				SkillTags:      []string{" Go ", "PostgreSQL", "go", "Distributed Systems"},
			},
			wantErr: nil,
		},
	}

	var story domain.Story
	for _, tc := range createCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.NewStory(tc.id, tc.input, createdAt)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			story = got
		})
	}

	if story.Title != "Zero-Downtime Ledger Sharding" {
		t.Fatalf("expected trimmed title, got %q", story.Title)
	}
	if !reflect.DeepEqual(story.CompetencyTags, []string{"System Design", "Leadership"}) {
		t.Errorf("unexpected CompetencyTags: %v", story.CompetencyTags)
	}
	if !reflect.DeepEqual(story.SkillTags, []string{"Go", "PostgreSQL", "Distributed Systems"}) {
		t.Errorf("unexpected SkillTags: %v", story.SkillTags)
	}

	t.Run("updates story fields while preserving CreatedAt and advancing UpdatedAt", func(t *testing.T) {
		err := story.Update(domain.StoryInput{
			Title:          "Zero-Downtime Multi-Region Ledger Sharding",
			Hook:           story.Hook,
			Situation:      story.Situation,
			Task:           story.Task,
			Action:         story.Action,
			Result:         "Cut write latency by 62% and handled 4x peak volume.",
			CompetencyTags: []string{"System Design", "Ownership"},
			SkillTags:      []string{"Go", "PostgreSQL", "Raft"},
		}, updatedAt)
		if err != nil {
			t.Fatalf("unexpected update error: %v", err)
		}
		if !story.CreatedAt.Equal(createdAt) {
			t.Errorf("expected CreatedAt to remain %v, got %v", createdAt, story.CreatedAt)
		}
		if !story.UpdatedAt.Equal(updatedAt) {
			t.Errorf("expected UpdatedAt to advance to %v, got %v", updatedAt, story.UpdatedAt)
		}
	})

	filterCases := []struct {
		name      string
		filter    domain.StoryFilter
		wantMatch bool
	}{
		{
			name:      "matches empty filter",
			filter:    domain.StoryFilter{},
			wantMatch: true,
		},
		{
			name:      "matches keyword in Result field case-insensitively",
			filter:    domain.StoryFilter{Query: "62%"},
			wantMatch: true,
		},
		{
			name:      "matches keyword in Action field case-insensitively",
			filter:    domain.StoryFilter{Query: "shadow verifier"},
			wantMatch: true,
		},
		{
			name:      "matches exact CompetencyTag case-insensitively",
			filter:    domain.StoryFilter{CompetencyTag: "ownership"},
			wantMatch: true,
		},
		{
			name:      "does not match partial CompetencyTag",
			filter:    domain.StoryFilter{CompetencyTag: "System"},
			wantMatch: false,
		},
		{
			name:      "matches exact SkillTag case-insensitively",
			filter:    domain.StoryFilter{SkillTag: "go"},
			wantMatch: true,
		},
		{
			name:      "matches combined Query, CompetencyTag, and SkillTag",
			filter:    domain.StoryFilter{Query: "sharding", CompetencyTag: "System Design", SkillTag: "Raft"},
			wantMatch: true,
		},
		{
			name:      "fails combined filter when SkillTag does not match",
			filter:    domain.StoryFilter{Query: "sharding", CompetencyTag: "System Design", SkillTag: "Rust"},
			wantMatch: false,
		},
	}

	for _, tc := range filterCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := story.MatchesFilter(tc.filter); got != tc.wantMatch {
				t.Errorf("MatchesFilter(%+v) = %v, want %v", tc.filter, got, tc.wantMatch)
			}
		})
	}
}
