package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type StaleOpportunityView struct {
	OpportunityID  string
	CompanyID      string
	CompanyName    string
	RoleTitle      string
	Stage          Stage
	Outcome        Outcome
	Priority       Priority
	StaleReason    string
	DaysStale      int
	AnchorDate     time.Time
	LastActivityAt time.Time
}

type UpcomingInterviewRoundView struct {
	Round                InterviewRound
	CompanyName          string
	OpportunityRoleTitle string
	InterviewerNames     []string
	StoryTitles          []string
}

type MissingDebriefRoundView struct {
	Round                InterviewRound
	CompanyName          string
	OpportunityRoleTitle string
}

type FlightDeckActionItemView struct {
	ActionItemView
	IsOverdue  bool
	IsDueToday bool
}

type FlightDeck struct {
	UpcomingRounds     []UpcomingInterviewRoundView
	MissingDebriefs    []MissingDebriefRoundView
	ActionItems        []FlightDeckActionItemView
	StaleOpportunities []StaleOpportunityView
}

type FlightDeckInput struct {
	Now                time.Time
	Opportunities      []OpportunityView
	StageTransitions   []StageTransition
	Interactions       []Interaction
	LatestInteractions map[string]time.Time
	InterviewRounds    []InterviewRound
	ActionItems        []ActionItemView
	ContactNames       map[string]string
	StoryTitles        map[string]string
}

// EvaluateFlightDeck is a pure domain evaluator that evaluates Flight Deck
// triage: imminent interview rounds (next 48h), past/completed rounds missing debriefs,
// open action items ordered by due date, and stage-specific staleness rules for Active opportunities.
func EvaluateFlightDeck(in FlightDeckInput) FlightDeck {
	now := in.Now.UTC()

	// 1. Index opportunities and companies
	oppByID := make(map[string]OpportunityView, len(in.Opportunities))
	for _, ov := range in.Opportunities {
		oppByID[ov.Opportunity.ID] = ov
	}

	// 2. Compute latest stage transition timestamp per opportunity
	latestTransitionAt := make(map[string]time.Time)
	for _, tr := range in.StageTransitions {
		ts := tr.TransitionedAt.UTC()
		if ts.After(latestTransitionAt[tr.OpportunityID]) {
			latestTransitionAt[tr.OpportunityID] = ts
		}
	}

	// 3. Compute latest interaction timestamp per opportunity
	latestInteractionAt := make(map[string]time.Time)
	for k, v := range in.LatestInteractions {
		latestInteractionAt[k] = v.UTC()
	}
	for _, ix := range in.Interactions {
		if ix.OpportunityID != "" {
			ts := ix.OccurredAt.UTC()
			if ts.After(latestInteractionAt[ix.OpportunityID]) {
				latestInteractionAt[ix.OpportunityID] = ts
			}
		}
	}

	// 4. Index interview rounds by opportunity ID
	roundsByOppID := make(map[string][]InterviewRound)
	for _, r := range in.InterviewRounds {
		roundsByOppID[r.OpportunityID] = append(roundsByOppID[r.OpportunityID], r)
	}

	// 5. Evaluate Staleness for Active Opportunities
	staleOpps := []StaleOpportunityView{}
	for _, ov := range in.Opportunities {
		opp := ov.Opportunity
		if opp.Outcome != OutcomeActive {
			continue
		}

		var isStale bool
		var staleReason string
		var anchorDate time.Time

		switch opp.Stage {
		case StageDiscovered:
			// Discovered > 3 days measured from max(LatestStageTransitionAt, LastActivityAt)
			transAt := latestTransitionAt[opp.ID]
			actAt := opp.LastActivityAt.UTC()
			anchorDate = maxTime(transAt, actAt)
			if anchorDate.IsZero() {
				anchorDate = opp.CreatedAt.UTC()
			}
			diff := now.Sub(anchorDate)
			if diff > 3*24*time.Hour {
				isStale = true
				days := int(diff / (24 * time.Hour))
				staleReason = fmt.Sprintf("In Discovered for %d days with no activity", days)
			}

		case StageOutreach:
			// Outreach > 4 days measured from max(LatestStageTransitionAt, LatestInteractionAt)
			transAt := latestTransitionAt[opp.ID]
			ixAt := latestInteractionAt[opp.ID]
			anchorDate = maxTime(transAt, ixAt)
			if anchorDate.IsZero() {
				anchorDate = opp.LastActivityAt.UTC()
			}
			if anchorDate.IsZero() {
				anchorDate = opp.CreatedAt.UTC()
			}
			diff := now.Sub(anchorDate)
			if diff > 4*24*time.Hour {
				isStale = true
				days := int(diff / (24 * time.Hour))
				staleReason = fmt.Sprintf("In Outreach for %d days with no interaction", days)
			}

		case StageApplied:
			// Applied > 10 days measured from max(LatestStageTransitionAt, LastActivityAt)
			transAt := latestTransitionAt[opp.ID]
			actAt := opp.LastActivityAt.UTC()
			anchorDate = maxTime(transAt, actAt)
			if anchorDate.IsZero() {
				anchorDate = opp.CreatedAt.UTC()
			}
			diff := now.Sub(anchorDate)
			if diff > 10*24*time.Hour {
				isStale = true
				days := int(diff / (24 * time.Hour))
				staleReason = fmt.Sprintf("In Applied for %d days with no activity", days)
			}

		case StageScreening, StageInterviewing:
			// Screening or Interviewing > 5 days after latest completed Interview Round (ScheduledAt)
			// with no upcoming Interview Round scheduled
			rounds := roundsByOppID[opp.ID]
			hasUpcomingRound := false
			var latestCompletedRound *InterviewRound

			for i := range rounds {
				r := &rounds[i]
				if r.Status == InterviewRoundStatusScheduled && !r.ScheduledAt.UTC().Before(now) {
					hasUpcomingRound = true
					break
				}
				if r.Status == InterviewRoundStatusCompleted {
					if latestCompletedRound == nil || r.ScheduledAt.UTC().After(latestCompletedRound.ScheduledAt.UTC()) {
						latestCompletedRound = r
					}
				}
			}

			if !hasUpcomingRound && latestCompletedRound != nil {
				anchorDate = latestCompletedRound.ScheduledAt.UTC()
				diff := now.Sub(anchorDate)
				if diff > 5*24*time.Hour {
					isStale = true
					days := int(diff / (24 * time.Hour))
					roundName := latestCompletedRound.Title
					if roundName == "" {
						roundName = string(latestCompletedRound.RoundType)
					}
					staleReason = fmt.Sprintf("In %s with no upcoming round scheduled %d days after %s", opp.Stage, days, roundName)
				}
			}
		}

		if isStale {
			daysStale := int(now.Sub(anchorDate) / (24 * time.Hour))
			staleOpps = append(staleOpps, StaleOpportunityView{
				OpportunityID:  opp.ID,
				CompanyID:      opp.CompanyID,
				CompanyName:    ov.CompanyName,
				RoleTitle:      opp.RoleTitle,
				Stage:          opp.Stage,
				Outcome:        opp.Outcome,
				Priority:       opp.Priority,
				StaleReason:    staleReason,
				DaysStale:      daysStale,
				AnchorDate:     anchorDate,
				LastActivityAt: opp.LastActivityAt.UTC(),
			})
		}
	}

	// Sort stale opportunities: highest DaysStale first, then priority
	sort.SliceStable(staleOpps, func(i, j int) bool {
		if staleOpps[i].DaysStale != staleOpps[j].DaysStale {
			return staleOpps[i].DaysStale > staleOpps[j].DaysStale
		}
		return staleOpps[i].AnchorDate.Before(staleOpps[j].AnchorDate)
	})

	// 6. Upcoming Interview Rounds scheduled within the next 48 hours
	upcomingRounds := []UpcomingInterviewRoundView{}
	for _, r := range in.InterviewRounds {
		if r.Status != InterviewRoundStatusScheduled {
			continue
		}
		schedUTC := r.ScheduledAt.UTC()
		// scheduled between now and now + 48h
		if !schedUTC.Before(now) && !schedUTC.After(now.Add(48*time.Hour)) {
			ov := oppByID[r.OpportunityID]
			interviewerNames := make([]string, 0, len(r.ContactIDs))
			for _, cid := range r.ContactIDs {
				if name, ok := in.ContactNames[cid]; ok && strings.TrimSpace(name) != "" {
					interviewerNames = append(interviewerNames, name)
				}
			}
			storyTitles := make([]string, 0, len(r.StoryIDs))
			for _, sid := range r.StoryIDs {
				if title, ok := in.StoryTitles[sid]; ok && strings.TrimSpace(title) != "" {
					storyTitles = append(storyTitles, title)
				}
			}

			upcomingRounds = append(upcomingRounds, UpcomingInterviewRoundView{
				Round:                r,
				CompanyName:          ov.CompanyName,
				OpportunityRoleTitle: ov.Opportunity.RoleTitle,
				InterviewerNames:     interviewerNames,
				StoryTitles:          storyTitles,
			})
		}
	}
	sort.SliceStable(upcomingRounds, func(i, j int) bool {
		return upcomingRounds[i].Round.ScheduledAt.UTC().Before(upcomingRounds[j].Round.ScheduledAt.UTC())
	})

	// 7. Missing Debriefs on past/completed interview rounds
	missingDebriefs := []MissingDebriefRoundView{}
	for _, r := range in.InterviewRounds {
		if r.Status == InterviewRoundStatusCancelled {
			continue
		}
		isPastOrCompleted := r.Status == InterviewRoundStatusCompleted ||
			(r.Status == InterviewRoundStatusScheduled && r.ScheduledAt.UTC().Before(now))

		if isPastOrCompleted && isDebriefMissing(r.Debrief) {
			ov := oppByID[r.OpportunityID]
			missingDebriefs = append(missingDebriefs, MissingDebriefRoundView{
				Round:                r,
				CompanyName:          ov.CompanyName,
				OpportunityRoleTitle: ov.Opportunity.RoleTitle,
			})
		}
	}
	sort.SliceStable(missingDebriefs, func(i, j int) bool {
		return missingDebriefs[i].Round.ScheduledAt.UTC().After(missingDebriefs[j].Round.ScheduledAt.UTC())
	})

	// 8. Open Action Items ordered by due date (highlighting overdue and due-today items)
	todayStartUTC := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	actionItems := []FlightDeckActionItemView{}
	for _, av := range in.ActionItems {
		if av.ActionItem.Status != ActionItemStatusOpen {
			continue
		}
		dueUTC := av.ActionItem.DueAt.UTC()
		var isOverdue, isDueToday bool
		if isSameDay(dueUTC, now) {
			isDueToday = true
		} else if dueUTC.Before(todayStartUTC) {
			isOverdue = true
		}

		actionItems = append(actionItems, FlightDeckActionItemView{
			ActionItemView: av,
			IsOverdue:      isOverdue,
			IsDueToday:     isDueToday,
		})
	}
	sort.SliceStable(actionItems, func(i, j int) bool {
		if !actionItems[i].ActionItem.DueAt.Equal(actionItems[j].ActionItem.DueAt) {
			return actionItems[i].ActionItem.DueAt.Before(actionItems[j].ActionItem.DueAt)
		}
		return actionItems[i].ActionItem.CreatedAt.Before(actionItems[j].ActionItem.CreatedAt)
	})

	return FlightDeck{
		UpcomingRounds:     upcomingRounds,
		MissingDebriefs:    missingDebriefs,
		ActionItems:        actionItems,
		StaleOpportunities: staleOpps,
	}
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func isSameDay(a, b time.Time) bool {
	y1, m1, d1 := a.UTC().Date()
	y2, m2, d2 := b.UTC().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func isDebriefMissing(d *Debrief) bool {
	if d == nil {
		return true
	}
	return len(d.QuestionsAsked) == 0 &&
		strings.TrimSpace(d.WentWell) == "" &&
		strings.TrimSpace(d.ToImprove) == "" &&
		strings.TrimSpace(d.Signals) == "" &&
		d.ConfidenceScore == 0 &&
		strings.TrimSpace(d.FollowUpNotes) == ""
}
