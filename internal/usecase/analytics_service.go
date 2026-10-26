package usecase

import (
	"context"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type AnalyticsService struct {
	oppRepo        outbound.OpportunityRepository
	transitionRepo outbound.StageTransitionRepository
	clock          outbound.Clock
}

func NewAnalyticsService(
	oppRepo outbound.OpportunityRepository,
	transitionRepo outbound.StageTransitionRepository,
	clock outbound.Clock,
) *AnalyticsService {
	return &AnalyticsService{
		oppRepo:        oppRepo,
		transitionRepo: transitionRepo,
		clock:          clock,
	}
}

var _ inbound.AnalyticsUseCase = (*AnalyticsService)(nil)

func (s *AnalyticsService) GetFunnelAnalytics(ctx context.Context) (domain.FunnelAnalytics, error) {
	views, err := s.oppRepo.ListOpportunities(ctx, domain.OpportunityFilter{})
	if err != nil {
		return domain.FunnelAnalytics{}, err
	}

	opps := make([]domain.Opportunity, 0, len(views))
	for _, v := range views {
		opps = append(opps, v.Opportunity)
	}

	transitions, err := s.transitionRepo.ListAllStageTransitions(ctx)
	if err != nil {
		return domain.FunnelAnalytics{}, err
	}

	return domain.ComputeFunnelAnalytics(opps, transitions, s.clock.Now().UTC()), nil
}
