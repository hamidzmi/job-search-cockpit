package httpadapter

import (
	"net/http"

	"job-search-cockpit/internal/usecase/port/inbound"
)

type AnalyticsHandler struct {
	analytics inbound.AnalyticsUseCase
}

func NewAnalyticsHandler(analytics inbound.AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/analytics/funnel", h.handleGetFunnelAnalytics)
}

func (h *AnalyticsHandler) handleGetFunnelAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics, err := h.analytics.GetFunnelAnalytics(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toFunnelAnalyticsDTO(analytics))
}
