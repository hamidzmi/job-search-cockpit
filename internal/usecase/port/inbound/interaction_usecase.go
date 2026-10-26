package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type InteractionUseCase interface {
	LogInteraction(ctx context.Context, in domain.InteractionInput) (domain.InteractionView, *domain.ActionItemView, error)
	ListInteractions(ctx context.Context, filter domain.InteractionFilter) ([]domain.InteractionView, error)
}
