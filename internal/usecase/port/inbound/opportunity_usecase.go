package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type OpportunityUseCase interface {
	CreateOpportunity(ctx context.Context, in domain.OpportunityInput) (domain.Opportunity, error)
	GetOpportunity(ctx context.Context, id string) (domain.Opportunity, []domain.StageTransition, error)
	ListOpportunities(ctx context.Context, filter domain.OpportunityFilter) ([]domain.Opportunity, error)
	UpdateOpportunityDetails(ctx context.Context, id string, in domain.OpportunityDetailsInput) (domain.Opportunity, error)
	TransitionOpportunity(ctx context.Context, id string, toStage domain.Stage, toOutcome domain.Outcome, note string) (domain.Opportunity, error)
	ListStageTransitions(ctx context.Context, opportunityID string) ([]domain.StageTransition, error)
}
