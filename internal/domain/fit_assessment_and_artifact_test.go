package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"job-search-cockpit/internal/domain"
)

func TestFitAssessmentInvariantsAndCompensationAlignment(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.FixedZone("EST", -5*3600))

	t.Run("creates and updates FitAssessment with normalized lists and score bounds", func(t *testing.T) {
		fa, err := domain.NewFitAssessment("opp-1", domain.FitAssessmentInput{
			OverallScore:          86,
			OverallSummary:        "  Strong match on Go distributed systems and observability.  ",
			MatchedStrengths:      []string{" Go microservices ", "Go microservices", "Kubernetes"},
			GapsAndRisks:          []string{"  No production Rust experience  ", ""},
			CompensationAlignment: "  Stated $210k-$245k exceeds $200k USD floor. ",
			RecommendedStoryIDs:   []string{" story-1 ", "story-2", "story-1"},
			ModelUsed:             " anthropic/claude-3.7-sonnet ",
		}, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if fa.OpportunityID != "opp-1" {
			t.Fatalf("expected OpportunityID opp-1, got %q", fa.OpportunityID)
		}
		if fa.OverallScore != 86 {
			t.Fatalf("expected OverallScore 86, got %d", fa.OverallScore)
		}
		if fa.OverallSummary != "Strong match on Go distributed systems and observability." {
			t.Fatalf("unexpected OverallSummary: %q", fa.OverallSummary)
		}
		if len(fa.MatchedStrengths) != 2 || fa.MatchedStrengths[0] != "Go microservices" || fa.MatchedStrengths[1] != "Kubernetes" {
			t.Fatalf("unexpected MatchedStrengths: %#v", fa.MatchedStrengths)
		}
		if len(fa.GapsAndRisks) != 1 || fa.GapsAndRisks[0] != "No production Rust experience" {
			t.Fatalf("unexpected GapsAndRisks: %#v", fa.GapsAndRisks)
		}
		if len(fa.RecommendedStoryIDs) != 2 || fa.RecommendedStoryIDs[0] != "story-1" || fa.RecommendedStoryIDs[1] != "story-2" {
			t.Fatalf("unexpected RecommendedStoryIDs: %#v", fa.RecommendedStoryIDs)
		}
		if fa.ModelUsed != "anthropic/claude-3.7-sonnet" {
			t.Fatalf("unexpected ModelUsed: %q", fa.ModelUsed)
		}
		if fa.UpdatedAt.Location() != time.UTC {
			t.Fatalf("expected UpdatedAt in UTC, got %v", fa.UpdatedAt.Location())
		}
	})

	t.Run("rejects missing opportunity id or out-of-range score", func(t *testing.T) {
		cases := []struct {
			name string
			opp  string
			in   domain.FitAssessmentInput
		}{
			{
				name: "missing opportunity id",
				opp:  "   ",
				in:   domain.FitAssessmentInput{OverallScore: 75, OverallSummary: "Good fit"},
			},
			{
				name: "negative score",
				opp:  "opp-1",
				in:   domain.FitAssessmentInput{OverallScore: -1, OverallSummary: "Good fit"},
			},
			{
				name: "score above 100",
				opp:  "opp-1",
				in:   domain.FitAssessmentInput{OverallScore: 101, OverallSummary: "Good fit"},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := domain.NewFitAssessment(tc.opp, tc.in, now)
				if !errors.Is(err, domain.ErrValidation) {
					t.Fatalf("expected ErrValidation, got %v", err)
				}
			})
		}
	})

	t.Run("evaluates compensation alignment against CandidateProfile floor", func(t *testing.T) {
		profileWithFloor := domain.CandidateProfile{
			CompensationFloor:  200000,
			TargetCompensation: 230000,
			Currency:           "USD",
		}

		cases := []struct {
			name    string
			jp      domain.JobPosting
			profile domain.CandidateProfile
			wantSub string
		}{
			{
				name:    "no floor configured",
				jp:      domain.JobPosting{StatedMinComp: 180000, StatedMaxComp: 210000, Currency: "USD"},
				profile: domain.CandidateProfile{CompensationFloor: 0, Currency: "USD"},
				wantSub: "No compensation floor configured",
			},
			{
				name:    "unstated job posting compensation",
				jp:      domain.JobPosting{},
				profile: profileWithFloor,
				wantSub: "Unstated in Job Posting",
			},
			{
				name:    "below candidate floor",
				jp:      domain.JobPosting{StatedMinComp: 150000, StatedMaxComp: 185000, Currency: "USD"},
				profile: profileWithFloor,
				wantSub: "Below floor",
			},
			{
				name:    "overlaps candidate floor",
				jp:      domain.JobPosting{StatedMinComp: 185000, StatedMaxComp: 220000, Currency: "USD"},
				profile: profileWithFloor,
				wantSub: "Overlaps floor",
			},
			{
				name:    "meets or exceeds candidate floor",
				jp:      domain.JobPosting{StatedMinComp: 210000, StatedMaxComp: 250000, Currency: "USD"},
				profile: profileWithFloor,
				wantSub: "Meets floor",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got := domain.EvaluateCompensationAlignment(tc.jp, tc.profile)
				if !strings.Contains(got, tc.wantSub) {
					t.Fatalf("expected %q to contain %q", got, tc.wantSub)
				}
			})
		}
	})
}

func TestApplicationArtifactInvariantsAndPDFAttachment(t *testing.T) {
	createdAt := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(30 * time.Minute)

	t.Run("creates, updates, and attaches PDF file to ApplicationArtifact", func(t *testing.T) {
		art, err := domain.NewApplicationArtifact("art-1", "opp-1", domain.ApplicationArtifactInput{
			Kind:            domain.ArtifactKindResume,
			Title:           "  Tailored Staff Backend Resume  ",
			ContentMarkdown: "# Jane Doe\n\n- Built distributed Go services",
			IsSubmitted:     false,
			ModelUsed:       "anthropic/claude-3.7-sonnet",
		}, createdAt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if art.Kind != domain.ArtifactKindResume || art.Title != "Tailored Staff Backend Resume" {
			t.Fatalf("unexpected artifact: %#v", art)
		}
		if art.HasAttachedFile() {
			t.Fatalf("expected HasAttachedFile false before attachment")
		}

		if err := art.AttachFile("opp-1/art-1.pdf", "jane-doe-resume.pdf", "application/pdf", 4096, updatedAt); err != nil {
			t.Fatalf("attach file failed: %v", err)
		}
		if !art.HasAttachedFile() {
			t.Fatalf("expected HasAttachedFile true after attachment")
		}
		if art.FilePath != "opp-1/art-1.pdf" || art.FileName != "jane-doe-resume.pdf" || art.MimeType != "application/pdf" || art.FileSizeBytes != 4096 {
			t.Fatalf("unexpected file attachment metadata: %#v", art)
		}
		if !art.UpdatedAt.Equal(updatedAt) {
			t.Fatalf("expected UpdatedAt %v, got %v", updatedAt, art.UpdatedAt)
		}
	})

	t.Run("validates kind, title, and pdf mime type", func(t *testing.T) {
		if _, err := domain.NewApplicationArtifact("art-1", "opp-1", domain.ApplicationArtifactInput{
			Kind:  domain.ArtifactKind("invalid"),
			Title: "Test",
		}, createdAt); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for invalid kind, got %v", err)
		}

		if _, err := domain.NewApplicationArtifact("art-1", "opp-1", domain.ApplicationArtifactInput{
			Kind:  domain.ArtifactKindCoverLetter,
			Title: "   ",
		}, createdAt); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for empty title, got %v", err)
		}

		art, err := domain.NewApplicationArtifact("art-1", "opp-1", domain.ApplicationArtifactInput{
			Kind:  domain.ArtifactKindResume,
			Title: "Resume",
		}, createdAt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := art.AttachFile("opp-1/art-1.txt", "resume.txt", "text/plain", 128, updatedAt); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for non-PDF mime type, got %v", err)
		}
	})
}

func TestOpportunityTouchUpdatedAtDoesNotAdvanceLastActivityAt(t *testing.T) {
	initialAt := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	opp, _, err := domain.NewOpportunity("opp-1", "tr-1", domain.OpportunityInput{
		OpportunityDetailsInput: domain.OpportunityDetailsInput{
			CompanyID: "comp-1",
			RoleTitle: "Staff Backend Engineer",
		},
		Stage: domain.StageApplied,
	}, initialAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	laterAt := initialAt.Add(72 * time.Hour)
	opp.TouchUpdatedAt(laterAt)

	if !opp.LastActivityAt.Equal(initialAt) {
		t.Fatalf("expected LastActivityAt to remain %v, got %v", initialAt, opp.LastActivityAt)
	}
	if !opp.UpdatedAt.Equal(laterAt) {
		t.Fatalf("expected UpdatedAt to advance to %v, got %v", laterAt, opp.UpdatedAt)
	}
}
