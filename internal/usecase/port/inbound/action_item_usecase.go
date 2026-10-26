package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type ActionItemUseCase interface {
	CreateActionItem(ctx context.Context, in domain.ActionItemInput) (domain.ActionItemView, error)
	UpdateActionItem(ctx context.Context, id string, in domain.ActionItemUpdateInput) (domain.ActionItemView, error)
	ListActionItems(ctx context.Context, filter domain.ActionItemFilter) ([]domain.ActionItemView, error)
}
