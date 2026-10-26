package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type FlightDeckUseCase interface {
	GetFlightDeck(ctx context.Context) (domain.FlightDeck, error)
}
