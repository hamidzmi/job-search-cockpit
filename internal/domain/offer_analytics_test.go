package domain_test

import (
	"errors"
	"testing"
	"time"

	"job-search-cockpit/internal/domain"
)

func TestOfferCompensationMathAndValidation(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, 11, 5, 23, 59, 0, 0, time.UTC)

	tests := []struct {
		name               string
		id                 string
		in                 domain.OfferInput
		compFloor          int
		wantErr            error
		wantVestingYears   int
		wantCurrency       string
		wantStatus         domain.OfferStatus
		wantAnnualBonus    int
		wantAnnualEquity   int
		wantRecurringTotal int
		wantYear1Total     int
		wantDeltaVsFloor   int
	}{
		{
			name: "target bonus pct and default 4-year equity vesting",
			id:   "offer-1",
			in: domain.OfferInput{
				OpportunityID:    "opp-1",
				BaseSalary:       200000,
				BonusTargetPct:   15,
				EquityType:       "RSU",
				EquityTotalValue: 160000,
				SignOnBonus:      25000,
				Currency:         "usd",
				DeadlineAt:       deadline,
				BenefitsNotes:    "100% medical, 401k match",
				NegotiationNotes: "Asked for +20k equity",
			},
			compFloor:          240000,
			wantVestingYears:   4,
			wantCurrency:       "USD",
			wantStatus:         domain.OfferStatusPending,
			wantAnnualBonus:    30000,
			wantAnnualEquity:   40000,
			wantRecurringTotal: 270000,
			wantYear1Total:     295000,
			wantDeltaVsFloor:   30000,
		},
		{
			name: "explicit bonus amount overrides bonus target pct and custom vesting years",
			id:   "offer-2",
			in: domain.OfferInput{
				OpportunityID:      "opp-2",
				BaseSalary:         190000,
				BonusTargetPct:     10,
				BonusAmount:        25000,
				EquityType:         "Options",
				EquityTotalValue:   90000,
				EquityVestingYears: 3,
				SignOnBonus:        10000,
				Currency:           "USD",
				Status:             domain.OfferStatusNegotiating,
			},
			compFloor:          250000,
			wantVestingYears:   3,
			wantCurrency:       "USD",
			wantStatus:         domain.OfferStatusNegotiating,
			wantAnnualBonus:    25000,
			wantAnnualEquity:   30000,
			wantRecurringTotal: 245000,
			wantYear1Total:     255000,
			wantDeltaVsFloor:   -5000,
		},
		{
			name: "cash-only offer without bonus or equity",
			id:   "offer-3",
			in: domain.OfferInput{
				OpportunityID: "opp-3",
				BaseSalary:    210000,
				Status:        domain.OfferStatusAccepted,
			},
			compFloor:          200000,
			wantVestingYears:   4,
			wantCurrency:       "USD",
			wantStatus:         domain.OfferStatusAccepted,
			wantAnnualBonus:    0,
			wantAnnualEquity:   0,
			wantRecurringTotal: 210000,
			wantYear1Total:     210000,
			wantDeltaVsFloor:   10000,
		},
		{
			name: "rejects missing opportunity id",
			id:   "offer-err-1",
			in: domain.OfferInput{
				BaseSalary: 200000,
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "rejects non-positive base salary",
			id:   "offer-err-2",
			in: domain.OfferInput{
				OpportunityID: "opp-1",
				BaseSalary:    0,
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "rejects invalid offer status",
			id:   "offer-err-3",
			in: domain.OfferInput{
				OpportunityID: "opp-1",
				BaseSalary:    180000,
				Status:        domain.OfferStatus("Unknown"),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "rejects negative equity or bonus values",
			id:   "offer-err-4",
			in: domain.OfferInput{
				OpportunityID:    "opp-1",
				BaseSalary:       180000,
				EquityTotalValue: -1000,
			},
			wantErr: domain.ErrValidation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			offer, err := domain.NewOffer(tc.id, tc.in, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if offer.EquityVestingYears != tc.wantVestingYears {
				t.Errorf("EquityVestingYears = %d, want %d", offer.EquityVestingYears, tc.wantVestingYears)
			}
			if offer.Currency != tc.wantCurrency {
				t.Errorf("Currency = %q, want %q", offer.Currency, tc.wantCurrency)
			}
			if offer.Status != tc.wantStatus {
				t.Errorf("Status = %q, want %q", offer.Status, tc.wantStatus)
			}
			if got := offer.AnnualBonus(); got != tc.wantAnnualBonus {
				t.Errorf("AnnualBonus() = %d, want %d", got, tc.wantAnnualBonus)
			}
			if got := offer.AnnualEquity(); got != tc.wantAnnualEquity {
				t.Errorf("AnnualEquity() = %d, want %d", got, tc.wantAnnualEquity)
			}
			if got := offer.RecurringAnnualTotalComp(); got != tc.wantRecurringTotal {
				t.Errorf("RecurringAnnualTotalComp() = %d, want %d", got, tc.wantRecurringTotal)
			}
			if got := offer.Year1TotalComp(); got != tc.wantYear1Total {
				t.Errorf("Year1TotalComp() = %d, want %d", got, tc.wantYear1Total)
			}
			if got := offer.DeltaVsCompFloor(tc.compFloor); got != tc.wantDeltaVsFloor {
				t.Errorf("DeltaVsCompFloor(%d) = %d, want %d", tc.compFloor, got, tc.wantDeltaVsFloor)
			}
		})
	}
}

func TestComputeFunnelAnalytics(t *testing.T) {
	day := func(d int) time.Time {
		return time.Date(2026, 10, 1+d, 12, 0, 0, 0, time.UTC)
	}
	now := day(30)

	tests := []struct {
		name                 string
		now                  time.Time
		opportunities        []domain.Opportunity
		transitions          []domain.StageTransition
		wantTotal            int
		wantActive           int
		wantReachedByStage   map[domain.Stage]int
		wantConvFromPrevPct  map[domain.Stage]float64
		wantDropOffByStage   map[domain.Stage]int
		wantOutcomeAtStage   map[domain.Stage]map[domain.Outcome]int
		wantAvgDaysByStage   map[domain.Stage]float64
	}{
		{
			name:               "empty pipeline returns zeroed metrics for all stages",
			opportunities:      nil,
			transitions:        nil,
			wantTotal:          0,
			wantActive:         0,
			wantReachedByStage: map[domain.Stage]int{
				domain.StageDiscovered:   0,
				domain.StageOutreach:     0,
				domain.StageApplied:      0,
				domain.StageScreening:    0,
				domain.StageInterviewing: 0,
				domain.StageOffer:        0,
			},
			wantConvFromPrevPct: map[domain.Stage]float64{
				domain.StageDiscovered:   0,
				domain.StageOutreach:     0,
				domain.StageApplied:      0,
				domain.StageScreening:    0,
				domain.StageInterviewing: 0,
				domain.StageOffer:        0,
			},
			wantDropOffByStage: map[domain.Stage]int{
				domain.StageDiscovered:   0,
				domain.StageOutreach:     0,
				domain.StageApplied:      0,
				domain.StageScreening:    0,
				domain.StageInterviewing: 0,
				domain.StageOffer:        0,
			},
			wantAvgDaysByStage: map[domain.Stage]float64{
				domain.StageDiscovered:   0,
				domain.StageOutreach:     0,
				domain.StageApplied:      0,
				domain.StageScreening:    0,
				domain.StageInterviewing: 0,
				domain.StageOffer:        0,
			},
		},
		{
			name: "computes stage reach counts, conversion percentages, outcome drop-offs, and average days per stage",
			opportunities: []domain.Opportunity{
				{
					ID:        "opp-1",
					CompanyID: "comp-1",
					RoleTitle: "Staff Engineer",
					Stage:     domain.StageOffer,
					Outcome:   domain.OutcomeActive,
				},
				{
					ID:        "opp-2",
					CompanyID: "comp-2",
					RoleTitle: "Principal Engineer",
					Stage:     domain.StageInterviewing,
					Outcome:   domain.OutcomeRejected,
				},
				{
					ID:        "opp-3",
					CompanyID: "comp-3",
					RoleTitle: "Senior Backend Engineer",
					Stage:     domain.StageApplied,
					Outcome:   domain.OutcomeGhosted,
				},
				{
					ID:        "opp-4",
					CompanyID: "comp-4",
					RoleTitle: "Platform Lead",
					Stage:     domain.StageOutreach,
					Outcome:   domain.OutcomeWithdrawn,
				},
			},
			transitions: []domain.StageTransition{
				// opp-1: Discovered (2d) -> Outreach (4d) -> Applied (6d) -> Screening (4d) -> Interviewing (8d) -> Offer (6d until now=day(30))
				{ID: "tr-1a", OpportunityID: "opp-1", FromStage: "", ToStage: domain.StageDiscovered, FromOutcome: "", ToOutcome: domain.OutcomeActive, TransitionedAt: day(0)},
				{ID: "tr-1b", OpportunityID: "opp-1", FromStage: domain.StageDiscovered, ToStage: domain.StageOutreach, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(2)},
				{ID: "tr-1c", OpportunityID: "opp-1", FromStage: domain.StageOutreach, ToStage: domain.StageApplied, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(6)},
				{ID: "tr-1d", OpportunityID: "opp-1", FromStage: domain.StageApplied, ToStage: domain.StageScreening, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(12)},
				{ID: "tr-1e", OpportunityID: "opp-1", FromStage: domain.StageScreening, ToStage: domain.StageInterviewing, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(16)},
				{ID: "tr-1f", OpportunityID: "opp-1", FromStage: domain.StageInterviewing, ToStage: domain.StageOffer, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(24)},

				// opp-2: Discovered (4d) -> Outreach (2d) -> Applied (8d) -> Screening (6d) -> Interviewing (4d, then Rejected at day 24)
				{ID: "tr-2a", OpportunityID: "opp-2", FromStage: "", ToStage: domain.StageDiscovered, FromOutcome: "", ToOutcome: domain.OutcomeActive, TransitionedAt: day(0)},
				{ID: "tr-2b", OpportunityID: "opp-2", FromStage: domain.StageDiscovered, ToStage: domain.StageOutreach, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(4)},
				{ID: "tr-2c", OpportunityID: "opp-2", FromStage: domain.StageOutreach, ToStage: domain.StageApplied, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(6)},
				{ID: "tr-2d", OpportunityID: "opp-2", FromStage: domain.StageApplied, ToStage: domain.StageScreening, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(14)},
				{ID: "tr-2e", OpportunityID: "opp-2", FromStage: domain.StageScreening, ToStage: domain.StageInterviewing, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(20)},
				{ID: "tr-2f", OpportunityID: "opp-2", FromStage: domain.StageInterviewing, ToStage: domain.StageInterviewing, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeRejected, TransitionedAt: day(24)},

				// opp-3: Discovered (3d) -> Outreach (3d) -> Applied (10d, then Ghosted at day 16)
				{ID: "tr-3a", OpportunityID: "opp-3", FromStage: "", ToStage: domain.StageDiscovered, FromOutcome: "", ToOutcome: domain.OutcomeActive, TransitionedAt: day(0)},
				{ID: "tr-3b", OpportunityID: "opp-3", FromStage: domain.StageDiscovered, ToStage: domain.StageOutreach, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(3)},
				{ID: "tr-3c", OpportunityID: "opp-3", FromStage: domain.StageOutreach, ToStage: domain.StageApplied, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(6)},
				{ID: "tr-3d", OpportunityID: "opp-3", FromStage: domain.StageApplied, ToStage: domain.StageApplied, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeGhosted, TransitionedAt: day(16)},

				// opp-4: Discovered (3d) -> Outreach (5d, then Withdrawn at day 8)
				{ID: "tr-4a", OpportunityID: "opp-4", FromStage: "", ToStage: domain.StageDiscovered, FromOutcome: "", ToOutcome: domain.OutcomeActive, TransitionedAt: day(0)},
				{ID: "tr-4b", OpportunityID: "opp-4", FromStage: domain.StageDiscovered, ToStage: domain.StageOutreach, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeActive, TransitionedAt: day(3)},
				{ID: "tr-4c", OpportunityID: "opp-4", FromStage: domain.StageOutreach, ToStage: domain.StageOutreach, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeWithdrawn, TransitionedAt: day(8)},
			},
			wantTotal:  4,
			wantActive: 1,
			wantReachedByStage: map[domain.Stage]int{
				domain.StageDiscovered:   4,
				domain.StageOutreach:     4,
				domain.StageApplied:      3,
				domain.StageScreening:    2,
				domain.StageInterviewing: 2,
				domain.StageOffer:        1,
			},
			wantConvFromPrevPct: map[domain.Stage]float64{
				domain.StageDiscovered:   100.0,
				domain.StageOutreach:     100.0,
				domain.StageApplied:      75.0,
				domain.StageScreening:    66.7,
				domain.StageInterviewing: 100.0,
				domain.StageOffer:        50.0,
			},
			wantDropOffByStage: map[domain.Stage]int{
				domain.StageDiscovered:   0,
				domain.StageOutreach:     1,
				domain.StageApplied:      1,
				domain.StageScreening:    0,
				domain.StageInterviewing: 1,
				domain.StageOffer:        0,
			},
			wantOutcomeAtStage: map[domain.Stage]map[domain.Outcome]int{
				domain.StageOutreach: {
					domain.OutcomeWithdrawn: 1,
				},
				domain.StageApplied: {
					domain.OutcomeGhosted: 1,
				},
				domain.StageInterviewing: {
					domain.OutcomeRejected: 1,
				},
				domain.StageOffer: {
					domain.OutcomeActive: 1,
				},
			},
			wantAvgDaysByStage: map[domain.Stage]float64{
				domain.StageDiscovered:   3.0, // (2 + 4 + 3 + 3) / 4 = 3.0
				domain.StageOutreach:     3.5, // (4 + 2 + 3 + 5) / 4 = 3.5
				domain.StageApplied:      8.0, // (6 + 8 + 10) / 3 = 8.0
				domain.StageScreening:    5.0, // (4 + 6) / 2 = 5.0
				domain.StageInterviewing: 6.0, // (8 + 4) / 2 = 6.0
				domain.StageOffer:        6.0, // 6 / 1 = 6.0
			},
		},
		{
			name: "does not count On Hold as drop-off and correctly resumes duration tracking upon reactivation",
			now:  day(10),
			opportunities: []domain.Opportunity{
				{
					ID:        "opp-onhold-reactivated",
					CompanyID: "comp-1",
					RoleTitle: "Staff Engineer",
					Stage:     domain.StageScreening,
					Outcome:   domain.OutcomeActive,
				},
				{
					ID:        "opp-staying-onhold",
					CompanyID: "comp-2",
					RoleTitle: "Lead Engineer",
					Stage:     domain.StageDiscovered,
					Outcome:   domain.OutcomeOnHold,
				},
			},
			transitions: []domain.StageTransition{
				// opp-onhold-reactivated:
				// Day 0: Discovered -> Screening (Active)
				{ID: "tr-r1", OpportunityID: "opp-onhold-reactivated", FromStage: "", ToStage: domain.StageScreening, FromOutcome: "", ToOutcome: domain.OutcomeActive, TransitionedAt: day(0)},
				// Day 2: Screening goes On Hold (duration: 2 days)
				{ID: "tr-r2", OpportunityID: "opp-onhold-reactivated", FromStage: domain.StageScreening, ToStage: domain.StageScreening, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeOnHold, TransitionedAt: day(2)},
				// Day 6: Screening reactivated to Active (on hold for 4 days, not counted)
				{ID: "tr-r3", OpportunityID: "opp-onhold-reactivated", FromStage: domain.StageScreening, ToStage: domain.StageScreening, FromOutcome: domain.OutcomeOnHold, ToOutcome: domain.OutcomeActive, TransitionedAt: day(6)},
				// Evaluated at now = day(10): duration from day 6 to day 10 = 4 days. Total days = 2 + 4 = 6 days.

				// opp-staying-onhold:
				// Day 0: Discovered (Active)
				{ID: "tr-h1", OpportunityID: "opp-staying-onhold", FromStage: "", ToStage: domain.StageDiscovered, FromOutcome: "", ToOutcome: domain.OutcomeActive, TransitionedAt: day(0)},
				// Day 3: Discovered goes On Hold (duration: 3 days)
				{ID: "tr-h2", OpportunityID: "opp-staying-onhold", FromStage: domain.StageDiscovered, ToStage: domain.StageDiscovered, FromOutcome: domain.OutcomeActive, ToOutcome: domain.OutcomeOnHold, TransitionedAt: day(3)},
			},
			wantTotal:  2,
			wantActive: 1, // only opp-onhold-reactivated
			wantReachedByStage: map[domain.Stage]int{
				domain.StageDiscovered: 1,
				domain.StageScreening:  1,
			},
			wantDropOffByStage: map[domain.Stage]int{
				domain.StageDiscovered: 0, // On Hold is NOT a drop-off
				domain.StageScreening:  0, // Active
			},
			wantOutcomeAtStage: map[domain.Stage]map[domain.Outcome]int{
				domain.StageDiscovered: {
					domain.OutcomeOnHold: 1,
				},
				domain.StageScreening: {
					domain.OutcomeActive: 1,
				},
			},
			wantAvgDaysByStage: map[domain.Stage]float64{
				domain.StageDiscovered: 3.0, // 3 / 1 = 3.0
				domain.StageScreening:  6.0, // (2 + 4) / 1 = 6.0 (reactivation resumed calculation!)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			evalNow := tc.now
			if evalNow.IsZero() {
				evalNow = now
			}
			analytics := domain.ComputeFunnelAnalytics(tc.opportunities, tc.transitions, evalNow)
			if analytics.TotalOpportunities != tc.wantTotal {
				t.Errorf("TotalOpportunities = %d, want %d", analytics.TotalOpportunities, tc.wantTotal)
			}
			if analytics.ActiveOpportunities != tc.wantActive {
				t.Errorf("ActiveOpportunities = %d, want %d", analytics.ActiveOpportunities, tc.wantActive)
			}
			if len(analytics.Stages) != len(domain.AllStages()) {
				t.Fatalf("len(Stages) = %d, want %d", len(analytics.Stages), len(domain.AllStages()))
			}
			for _, st := range analytics.Stages {
				if wantReached, ok := tc.wantReachedByStage[st.Stage]; ok && st.ReachedCount != wantReached {
					t.Errorf("stage %s ReachedCount = %d, want %d", st.Stage, st.ReachedCount, wantReached)
				}
				if wantConv, ok := tc.wantConvFromPrevPct[st.Stage]; ok && st.ConversionFromPreviousPct != wantConv {
					t.Errorf("stage %s ConversionFromPreviousPct = %.1f, want %.1f", st.Stage, st.ConversionFromPreviousPct, wantConv)
				}
				if wantDrop, ok := tc.wantDropOffByStage[st.Stage]; ok && st.DropOffCount != wantDrop {
					t.Errorf("stage %s DropOffCount = %d, want %d", st.Stage, st.DropOffCount, wantDrop)
				}
				if wantOutcomes, ok := tc.wantOutcomeAtStage[st.Stage]; ok {
					for outcome, wantCount := range wantOutcomes {
						if gotCount := st.OutcomeBreakdown[outcome]; gotCount != wantCount {
							t.Errorf("stage %s OutcomeBreakdown[%s] = %d, want %d", st.Stage, outcome, gotCount, wantCount)
						}
					}
				}
				if wantAvg, ok := tc.wantAvgDaysByStage[st.Stage]; ok && st.AverageDaysInStage != wantAvg {
					t.Errorf("stage %s AverageDaysInStage = %.1f, want %.1f", st.Stage, st.AverageDaysInStage, wantAvg)
				}
			}
		})
	}
}

