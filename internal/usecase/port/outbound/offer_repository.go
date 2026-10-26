package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type OfferRepository interface {
	SaveOffer(ctx context.Context, offer domain.Offer) error
	GetOfferByOpportunityID(ctx context.Context, opportunityID string) (domain.Offer, error)
	DeleteOfferByOpportunityID(ctx context.Context, opportunityID string) error
	ListOffers(ctx context.Context) ([]domain.Offer, error)
}
