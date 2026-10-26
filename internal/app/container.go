package app

import (
	"net/http"

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
		clock = outbound.SystemClock{}
	}

	companySvc := usecase.NewCompanyService(store, store, clock)
	opportunitySvc := usecase.NewOpportunityService(store, store, store, store, clock)

	return httpadapter.NewRouter(
		httpadapter.NewCompanyHandler(companySvc),
		httpadapter.NewOpportunityHandler(opportunitySvc),
	)
}
