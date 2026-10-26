package domain

import (
	"math"
	"sort"
	"time"
)

type StageFunnelMetrics struct {
	Stage                     Stage
	ReachedCount              int
	ConversionFromPreviousPct float64
	DropOffCount              int
	OutcomeBreakdown          map[Outcome]int
	AverageDaysInStage        float64
}

type FunnelAnalytics struct {
	TotalOpportunities  int
	ActiveOpportunities int
	Stages              []StageFunnelMetrics
}

func ComputeFunnelAnalytics(opportunities []Opportunity, transitions []StageTransition, now time.Time) FunnelAnalytics {
	allStages := AllStages()
	stagesCount := len(allStages)
	stageIndices := make(map[Stage]int, stagesCount)
	for i, st := range allStages {
		stageIndices[st] = i
	}

	result := FunnelAnalytics{
		TotalOpportunities: len(opportunities),
		Stages:             make([]StageFunnelMetrics, stagesCount),
	}

	for i, st := range allStages {
		result.Stages[i] = StageFunnelMetrics{
			Stage:            st,
			OutcomeBreakdown: make(map[Outcome]int),
		}
	}

	transitionsByOpp := make(map[string][]StageTransition)
	for _, tr := range transitions {
		transitionsByOpp[tr.OpportunityID] = append(transitionsByOpp[tr.OpportunityID], tr)
	}

	totalDaysByStage := make(map[Stage]float64)

	for _, opp := range opportunities {
		if opp.Outcome == OutcomeActive {
			result.ActiveOpportunities++
		}

		reached := make(map[Stage]bool)
		oppTransitions := transitionsByOpp[opp.ID]
		if len(oppTransitions) > 0 {
			sort.Slice(oppTransitions, func(i, j int) bool {
				return oppTransitions[i].TransitionedAt.Before(oppTransitions[j].TransitionedAt)
			})

			for _, tr := range oppTransitions {
				if tr.FromStage.IsValid() {
					reached[tr.FromStage] = true
				}
				if tr.ToStage.IsValid() {
					reached[tr.ToStage] = true
				}
			}
		}
		if opp.Stage.IsValid() {
			reached[opp.Stage] = true
		}

		for st := range reached {
			if idx, ok := stageIndices[st]; ok {
				result.Stages[idx].ReachedCount++
			}
		}

		if opp.Stage.IsValid() {
			if idx, ok := stageIndices[opp.Stage]; ok {
				result.Stages[idx].OutcomeBreakdown[opp.Outcome]++
				if opp.Outcome != OutcomeActive && opp.Outcome != OutcomeWon {
					result.Stages[idx].DropOffCount++
				}
			}
		}

		// Stage duration calculation
		if len(oppTransitions) > 0 {
			var currStage Stage
			var stageStartTime time.Time
			var isClosed bool

			for _, tr := range oppTransitions {
				if currStage == "" {
					currStage = tr.ToStage
					stageStartTime = tr.TransitionedAt
					if tr.ToOutcome != "" && tr.ToOutcome != OutcomeActive {
						isClosed = true
					}
					continue
				}

				if tr.ToStage != currStage {
					dur := tr.TransitionedAt.Sub(stageStartTime).Hours() / 24.0
					if dur > 0 {
						totalDaysByStage[currStage] += dur
					}
					currStage = tr.ToStage
					stageStartTime = tr.TransitionedAt
					if tr.ToOutcome != "" && tr.ToOutcome != OutcomeActive {
						isClosed = true
					} else {
						isClosed = false
					}
				} else if tr.ToOutcome != "" && tr.ToOutcome != OutcomeActive {
					dur := tr.TransitionedAt.Sub(stageStartTime).Hours() / 24.0
					if dur > 0 {
						totalDaysByStage[currStage] += dur
					}
					isClosed = true
					stageStartTime = tr.TransitionedAt
				}
			}

			if currStage.IsValid() && !isClosed && !stageStartTime.IsZero() {
				dur := now.Sub(stageStartTime).Hours() / 24.0
				if dur > 0 {
					totalDaysByStage[currStage] += dur
				}
			}
		}
	}

	for i, st := range allStages {
		reached := result.Stages[i].ReachedCount
		if i == 0 {
			if result.TotalOpportunities > 0 {
				conv := (float64(reached) / float64(result.TotalOpportunities)) * 100.0
				result.Stages[i].ConversionFromPreviousPct = math.Round(conv*10) / 10
			}
		} else {
			prevReached := result.Stages[i-1].ReachedCount
			if prevReached > 0 {
				conv := (float64(reached) / float64(prevReached)) * 100.0
				result.Stages[i].ConversionFromPreviousPct = math.Round(conv*10) / 10
			}
		}

		if reached > 0 {
			avg := totalDaysByStage[st] / float64(reached)
			result.Stages[i].AverageDaysInStage = math.Round(avg*10) / 10
		}
	}

	return result
}
