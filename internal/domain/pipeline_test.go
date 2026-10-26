package domain_test

import (
	"errors"
	"testing"
	"time"

	"job-search-cockpit/internal/domain"
)

func TestNewCompany(t *testing.T) {
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		id       string
		input    domain.CompanyInput
		wantName string
		wantErr  error
	}{
		{
			name:    "rejects blank company name",
			id:      "comp-1",
			input:   domain.CompanyInput{Name: "   "},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "rejects blank company id",
			id:      "  ",
			input:   domain.CompanyInput{Name: "Acme"},
			wantErr: domain.ErrValidation,
		},
		{
			name: "trims fields and sets timestamps on valid company",
			id:   "comp-1",
			input: domain.CompanyInput{
				Name:         "  Acme Robotics  ",
				Domain:       " acmerobotics.com ",
				Headquarters: "San Francisco, CA",
				RemotePolicy: "Hybrid",
				SizeStage:    "Series B (150)",
				Notes:        "Strong systems team",
			},
			wantName: "Acme Robotics",
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.NewCompany(tc.id, tc.input, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tc.wantName {
				t.Errorf("expected Name %q, got %q", tc.wantName, got.Name)
			}
			if !got.CreatedAt.Equal(now) || !got.UpdatedAt.Equal(now) {
				t.Errorf("expected CreatedAt and UpdatedAt %v, got %v / %v", now, got.CreatedAt, got.UpdatedAt)
			}
		})
	}
}

func TestNewOpportunity(t *testing.T) {
	now := time.Date(2026, 10, 25, 13, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		id             string
		transitionID   string
		input          domain.OpportunityInput
		wantStage      domain.Stage
		wantOutcome    domain.Outcome
		wantHasPosting bool
		wantErr        error
	}{
		{
			name:         "unposted pursuit defaults to Discovered and Active and emits initial StageTransition",
			id:           "opp-1",
			transitionID: "tr-1",
			input: domain.OpportunityInput{
				OpportunityDetailsInput: domain.OpportunityDetailsInput{
					CompanyID: "comp-1",
					RoleTitle: "  Staff Backend Engineer (Exploratory) ",
					Priority:  domain.PriorityHigh,
				},
				TransitionNote: "Warm intro from former colleague",
			},
			wantStage:      domain.StageDiscovered,
			wantOutcome:    domain.OutcomeActive,
			wantHasPosting: false,
		},
		{
			name:         "posted pursuit preserves explicit stage and structured JobPosting",
			id:           "opp-2",
			transitionID: "tr-2",
			input: domain.OpportunityInput{
				OpportunityDetailsInput: domain.OpportunityDetailsInput{
					CompanyID:    "comp-1",
					RoleTitle:    "Principal Distributed Systems Engineer",
					Seniority:    "Principal",
					Location:     "Remote (US)",
					RemotePolicy: domain.RemotePolicyRemote,
					Priority:     domain.PriorityHigh,
					JobPosting: domain.JobPosting{
						SourceURL:     "https://jobs.example.com/123",
						RawText:       "Build distributed storage engines in Go.",
						Requirements:  []string{"10+ years systems experience", "Go and Raft"},
						TechStackTags: []string{"Go", "SQLite", "Raft"},
						StatedMinComp: 220000,
						StatedMaxComp: 260000,
						Currency:      "USD",
					},
				},
				Stage: domain.StageApplied,
			},
			wantStage:      domain.StageApplied,
			wantOutcome:    domain.OutcomeActive,
			wantHasPosting: true,
		},
		{
			name:         "rejects missing company id",
			id:           "opp-3",
			transitionID: "tr-3",
			input: domain.OpportunityInput{
				OpportunityDetailsInput: domain.OpportunityDetailsInput{
					RoleTitle: "Engineer",
				},
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:         "rejects missing role title",
			id:           "opp-4",
			transitionID: "tr-4",
			input: domain.OpportunityInput{
				OpportunityDetailsInput: domain.OpportunityDetailsInput{
					CompanyID: "comp-1",
					RoleTitle: "   ",
				},
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:         "rejects non-canonical stage",
			id:           "opp-5",
			transitionID: "tr-5",
			input: domain.OpportunityInput{
				OpportunityDetailsInput: domain.OpportunityDetailsInput{
					CompanyID: "comp-1",
					RoleTitle: "Engineer",
				},
				Stage: domain.Stage("PhoneScreen"),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:         "rejects inverted stated compensation range",
			id:           "opp-7",
			transitionID: "tr-7",
			input: domain.OpportunityInput{
				OpportunityDetailsInput: domain.OpportunityDetailsInput{
					CompanyID: "comp-1",
					RoleTitle: "Engineer",
					JobPosting: domain.JobPosting{
						StatedMinComp: 300000,
						StatedMaxComp: 200000,
					},
				},
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:         "rejects non-canonical priority",
			id:           "opp-8",
			transitionID: "tr-8",
			input: domain.OpportunityInput{
				OpportunityDetailsInput: domain.OpportunityDetailsInput{
					CompanyID: "comp-1",
					RoleTitle: "Engineer",
					Priority:  domain.Priority("Urgent"),
				},
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:         "rejects non-canonical remote policy",
			id:           "opp-9",
			transitionID: "tr-9",
			input: domain.OpportunityInput{
				OpportunityDetailsInput: domain.OpportunityDetailsInput{
					CompanyID:    "comp-1",
					RoleTitle:    "Engineer",
					RemotePolicy: domain.RemotePolicy("Anywhere"),
				},
			},
			wantErr: domain.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opp, tr, err := domain.NewOpportunity(tc.id, tc.transitionID, tc.input, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opp.Stage != tc.wantStage {
				t.Errorf("expected Stage %q, got %q", tc.wantStage, opp.Stage)
			}
			if opp.Outcome != tc.wantOutcome {
				t.Errorf("expected Outcome %q, got %q", tc.wantOutcome, opp.Outcome)
			}
			if opp.JobPosting.HasContent() != tc.wantHasPosting {
				t.Errorf("expected HasContent()=%v, got %v", tc.wantHasPosting, opp.JobPosting.HasContent())
			}
			if !opp.LastActivityAt.Equal(now) || !opp.UpdatedAt.Equal(now) {
				t.Errorf("expected LastActivityAt and UpdatedAt %v, got %v / %v", now, opp.LastActivityAt, opp.UpdatedAt)
			}
			if tr.ID != tc.transitionID || tr.OpportunityID != tc.id || tr.FromStage != "" || tr.ToStage != tc.wantStage || tr.FromOutcome != "" || tr.ToOutcome != tc.wantOutcome || !tr.TransitionedAt.Equal(now) {
				t.Errorf("unexpected initial StageTransition: %+v", tr)
			}
		})
	}
}

func TestOpportunity_Transition(t *testing.T) {
	createdAt := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	transitionAt := time.Date(2026, 10, 25, 16, 30, 0, 0, time.UTC)

	tests := []struct {
		name            string
		initialStage    domain.Stage
		initialOutcome  domain.Outcome
		toStage         domain.Stage
		toOutcome       domain.Outcome
		note            string
		wantStage       domain.Stage
		wantOutcome     domain.Outcome
		wantFromStage   domain.Stage
		wantFromOutcome domain.Outcome
		wantErr         error
	}{
		{
			name:            "advances stage while preserving Active outcome when outcome omitted",
			initialStage:    domain.StageDiscovered,
			initialOutcome:  domain.OutcomeActive,
			toStage:         domain.StageApplied,
			toOutcome:       "",
			note:            "Submitted tailored resume via Greenhouse",
			wantStage:       domain.StageApplied,
			wantOutcome:     domain.OutcomeActive,
			wantFromStage:   domain.StageDiscovered,
			wantFromOutcome: domain.OutcomeActive,
		},
		{
			name:            "updates outcome to Rejected while preserving Interviewing stage (orthogonal stage/outcome)",
			initialStage:    domain.StageInterviewing,
			initialOutcome:  domain.OutcomeActive,
			toStage:         "",
			toOutcome:       domain.OutcomeRejected,
			note:            "Stopped after system design round",
			wantStage:       domain.StageInterviewing,
			wantOutcome:     domain.OutcomeRejected,
			wantFromStage:   domain.StageInterviewing,
			wantFromOutcome: domain.OutcomeActive,
		},
		{
			name:            "updates both stage and outcome simultaneously",
			initialStage:    domain.StageInterviewing,
			initialOutcome:  domain.OutcomeActive,
			toStage:         domain.StageOffer,
			toOutcome:       domain.OutcomeWon,
			note:            "Signed offer letter",
			wantStage:       domain.StageOffer,
			wantOutcome:     domain.OutcomeWon,
			wantFromStage:   domain.StageInterviewing,
			wantFromOutcome: domain.OutcomeActive,
		},
		{
			name:           "rejects no-op transition when stage and outcome are unchanged",
			initialStage:   domain.StageApplied,
			initialOutcome: domain.OutcomeActive,
			toStage:        domain.StageApplied,
			toOutcome:      domain.OutcomeActive,
			wantErr:        domain.ErrInvalidTransition,
		},
		{
			name:           "rejects invalid target stage",
			initialStage:   domain.StageApplied,
			initialOutcome: domain.OutcomeActive,
			toStage:        domain.Stage("Onsite"),
			wantErr:        domain.ErrValidation,
		},
		{
			name:           "rejects invalid target outcome",
			initialStage:   domain.StageApplied,
			initialOutcome: domain.OutcomeActive,
			toOutcome:      domain.Outcome("Closed"),
			wantErr:        domain.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opp, _, err := domain.NewOpportunity("opp-1", "tr-init", domain.OpportunityInput{
				OpportunityDetailsInput: domain.OpportunityDetailsInput{
					CompanyID: "comp-1",
					RoleTitle: "Staff Engineer",
				},
				Stage: tc.initialStage,
			}, createdAt)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			tr, err := opp.Transition("tr-next", tc.toStage, tc.toOutcome, tc.note, transitionAt)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opp.Stage != tc.wantStage {
				t.Errorf("expected opp.Stage=%q, got %q", tc.wantStage, opp.Stage)
			}
			if opp.Outcome != tc.wantOutcome {
				t.Errorf("expected opp.Outcome=%q, got %q", tc.wantOutcome, opp.Outcome)
			}
			if !opp.LastActivityAt.Equal(transitionAt) || !opp.UpdatedAt.Equal(transitionAt) {
				t.Errorf("expected LastActivityAt and UpdatedAt=%v, got %v / %v", transitionAt, opp.LastActivityAt, opp.UpdatedAt)
			}
			if tr.FromStage != tc.wantFromStage || tr.ToStage != tc.wantStage || tr.FromOutcome != tc.wantFromOutcome || tr.ToOutcome != tc.wantOutcome {
				t.Errorf("unexpected StageTransition state: %+v", tr)
			}
			if !tr.TransitionedAt.Equal(transitionAt) {
				t.Errorf("expected TransitionedAt=%v, got %v", transitionAt, tr.TransitionedAt)
			}
		})
	}
}

func TestOpportunity_UpdateDetailsPreservesLastActivityAt(t *testing.T) {
	createdAt := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	editedAt := time.Date(2026, 10, 24, 15, 0, 0, 0, time.UTC)

	opp, _, err := domain.NewOpportunity("opp-1", "tr-1", domain.OpportunityInput{
		OpportunityDetailsInput: domain.OpportunityDetailsInput{
			CompanyID: "comp-1",
			RoleTitle: "Exploratory Backend Role",
		},
	}, createdAt)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = opp.UpdateDetails(domain.OpportunityDetailsInput{
		CompanyID:    "comp-1",
		RoleTitle:    "Staff Backend Engineer",
		Seniority:    "Staff",
		Location:     "New York, NY",
		RemotePolicy: domain.RemotePolicyHybrid,
		Priority:     domain.PriorityHigh,
		Notes:        "Attached formal job posting after chat",
		JobPosting: domain.JobPosting{
			SourceURL:     "https://boards.greenhouse.io/acme/jobs/999",
			RawText:       "Staff Backend Engineer role description",
			StatedMinComp: 210000,
			StatedMaxComp: 250000,
			Currency:      "USD",
		},
	}, editedAt)
	if err != nil {
		t.Fatalf("unexpected error updating details: %v", err)
	}

	if !opp.JobPosting.HasContent() {
		t.Errorf("expected JobPosting to be attached after UpdateDetails")
	}
	if !opp.UpdatedAt.Equal(editedAt) {
		t.Errorf("expected UpdatedAt to advance to %v, got %v", editedAt, opp.UpdatedAt)
	}
	if !opp.LastActivityAt.Equal(createdAt) {
		t.Errorf("expected LastActivityAt to remain %v on detail edit, got %v", createdAt, opp.LastActivityAt)
	}
}

func TestOpportunityFilter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filter  domain.OpportunityFilter
		wantErr error
	}{
		{
			name:    "allows empty filter",
			filter:  domain.OpportunityFilter{},
			wantErr: nil,
		},
		{
			name: "allows valid stage and outcome filters",
			filter: domain.OpportunityFilter{
				Stage:   domain.StageInterviewing,
				Outcome: domain.OutcomeActive,
			},
			wantErr: nil,
		},
		{
			name: "rejects invalid stage filter",
			filter: domain.OpportunityFilter{
				Stage: domain.Stage("InvalidStage"),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "rejects invalid outcome filter",
			filter: domain.OpportunityFilter{
				Outcome: domain.Outcome("InvalidOutcome"),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "rejects invalid priority filter",
			filter: domain.OpportunityFilter{
				Priority: domain.Priority("InvalidPriority"),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "rejects invalid remote policy filter",
			filter: domain.OpportunityFilter{
				RemotePolicy: domain.RemotePolicy("InvalidRemotePolicy"),
			},
			wantErr: domain.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.filter.Validate()
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}


