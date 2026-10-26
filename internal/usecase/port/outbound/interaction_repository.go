package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type InteractionRepository interface {
	SaveInteraction(ctx context.Context, interaction domain.Interaction) error
	GetInteractionView(ctx context.Context, id string) (domain.InteractionView, error)
	ListInteractions(ctx context.Context, filter domain.InteractionFilter) ([]domain.InteractionView, error)
}
