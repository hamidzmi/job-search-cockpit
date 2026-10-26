package httpadapter

import (
	"job-search-cockpit/internal/domain"
)

type StageFunnelMetricsDTO struct {
	Stage                     string         `json:"stage"`
	ReachedCount              int            `json:"reached_count"`
	ConversionFromPreviousPct float64        `json:"conversion_from_previous_pct"`
	DropOffCount              int            `json:"drop_off_count"`
	OutcomeBreakdown          map[string]int `json:"outcome_breakdown"`
	AverageDaysInStage        float64        `json:"average_days_in_stage"`
}

type FunnelAnalyticsDTO struct {
	TotalOpportunities  int                     `json:"total_opportunities"`
	ActiveOpportunities int                     `json:"active_opportunities"`
	Stages              []StageFunnelMetricsDTO `json:"stages"`
}

func toFunnelAnalyticsDTO(f domain.FunnelAnalytics) FunnelAnalyticsDTO {
	stages := make([]StageFunnelMetricsDTO, 0, len(f.Stages))
	for _, s := range f.Stages {
		ob := make(map[string]int, len(s.OutcomeBreakdown))
		for k, v := range s.OutcomeBreakdown {
			ob[string(k)] = v
		}
		stages = append(stages, StageFunnelMetricsDTO{
			Stage:                     string(s.Stage),
			ReachedCount:              s.ReachedCount,
			ConversionFromPreviousPct: s.ConversionFromPreviousPct,
			DropOffCount:              s.DropOffCount,
			OutcomeBreakdown:          ob,
			AverageDaysInStage:        s.AverageDaysInStage,
		})
	}
	return FunnelAnalyticsDTO{
		TotalOpportunities:  f.TotalOpportunities,
		ActiveOpportunities: f.ActiveOpportunities,
		Stages:              stages,
	}
}
