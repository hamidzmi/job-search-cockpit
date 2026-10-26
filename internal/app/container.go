package app

import (
	"net/http"

	clockadapter "job-search-cockpit/internal/adapter/clock"
	httpadapter "job-search-cockpit/internal/adapter/http"
	"job-search-cockpit/internal/adapter/sqlite"
	"job-search-cockpit/internal/usecase"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type Dependencies struct {
	Clock outbound.Clock
}

func NewHTTPHandler(store *sqlite.Store, deps Dependencies) http.Handler {
	clock := deps.Clock
	if clock == nil {
		clock = clockadapter.SystemClock{}
	}

	companySvc := usecase.NewCompanyService(store, store, clock)
	opportunitySvc := usecase.NewOpportunityService(store, store, store, store, clock)
	contactSvc := usecase.NewContactService(store, store, store, store, clock)
	interactionSvc := usecase.NewInteractionService(store, store, store, store, store, clock)
	actionItemSvc := usecase.NewActionItemService(store, store, store, store, clock)

	return httpadapter.NewRouter(
		httpadapter.NewCompanyHandler(companySvc),
		httpadapter.NewOpportunityHandler(opportunitySvc),
		httpadapter.NewContactHandler(contactSvc),
		httpadapter.NewInteractionHandler(interactionSvc),
		httpadapter.NewActionItemHandler(actionItemSvc),
	)
}
