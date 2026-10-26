package httpadapter

import (
	"net/http"

	"job-search-cockpit/internal/usecase/port/inbound"
)

type SeedHandler struct {
	seed inbound.SeedUseCase
}

func NewSeedHandler(seed inbound.SeedUseCase) *SeedHandler {
	return &SeedHandler{seed: seed}
}

func (h *SeedHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/seed", h.handleSeed)
}

func (h *SeedHandler) handleSeed(w http.ResponseWriter, r *http.Request) {
	if err := h.seed.SeedSampleData(r.Context()); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}
