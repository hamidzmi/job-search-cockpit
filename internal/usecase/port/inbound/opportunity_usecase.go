package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type OpportunityUseCase interface {
	CreateOpportunity(ctx context.Context, in domain.OpportunityInput) (domain.OpportunityView, error)
	GetOpportunity(ctx context.Context, id string) (domain.OpportunityView, []domain.StageTransition, error)
	ListOpportunities(ctx context.Context, filter domain.OpportunityFilter) ([]domain.OpportunityView, error)
	UpdateOpportunityDetails(ctx context.Context, id string, in domain.OpportunityDetailsInput) (domain.OpportunityView, error)
	TransitionOpportunity(ctx context.Context, id string, toStage domain.Stage, toOutcome domain.Outcome, note string) (domain.OpportunityView, error)
	ListStageTransitions(ctx context.Context, opportunityID string) ([]domain.StageTransition, error)
}
