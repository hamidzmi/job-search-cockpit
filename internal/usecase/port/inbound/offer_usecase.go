package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type OfferUseCase interface {
	GetOffer(ctx context.Context, opportunityID string) (domain.OfferView, error)
	UpsertOffer(ctx context.Context, opportunityID string, in domain.OfferInput) (domain.OfferView, error)
	DeleteOffer(ctx context.Context, opportunityID string) error
	GetOfferComparison(ctx context.Context) (domain.OfferComparisonView, error)
}
