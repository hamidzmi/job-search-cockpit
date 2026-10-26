package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type StageTransitionRepository interface {
	AppendStageTransition(ctx context.Context, tr domain.StageTransition) error
	ListStageTransitions(ctx context.Context, opportunityID string) ([]domain.StageTransition, error)
}
