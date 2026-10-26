package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type AnalyticsUseCase interface {
	GetFunnelAnalytics(ctx context.Context) (domain.FunnelAnalytics, error)
}
