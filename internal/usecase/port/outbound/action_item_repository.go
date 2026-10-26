package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type ActionItemRepository interface {
	SaveActionItem(ctx context.Context, item domain.ActionItem) error
	GetActionItem(ctx context.Context, id string) (domain.ActionItem, error)
	GetActionItemView(ctx context.Context, id string) (domain.ActionItemView, error)
	ListActionItems(ctx context.Context, filter domain.ActionItemFilter) ([]domain.ActionItemView, error)
}
