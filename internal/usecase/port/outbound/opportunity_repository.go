package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type OpportunityRepository interface {
	SaveOpportunity(ctx context.Context, opp domain.Opportunity) error
	GetOpportunity(ctx context.Context, id string) (domain.Opportunity, error)
	GetOpportunityView(ctx context.Context, id string) (domain.OpportunityView, error)
	ListOpportunities(ctx context.Context, filter domain.OpportunityFilter) ([]domain.OpportunityView, error)
}
