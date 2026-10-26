package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type OfferHandler struct {
	offers inbound.OfferUseCase
}

func NewOfferHandler(offers inbound.OfferUseCase) *OfferHandler {
	return &OfferHandler{offers: offers}
}

func (h *OfferHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/offers", h.handleGetOfferComparison)
	mux.HandleFunc("GET /api/opportunities/{id}/offer", h.handleGetOffer)
	mux.HandleFunc("PUT /api/opportunities/{id}/offer", h.handleUpsertOffer)
	mux.HandleFunc("DELETE /api/opportunities/{id}/offer", h.handleDeleteOffer)
}

func (h *OfferHandler) handleGetOfferComparison(w http.ResponseWriter, r *http.Request) {
	view, err := h.offers.GetOfferComparison(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toOfferComparisonDTO(view))
}

func (h *OfferHandler) handleGetOffer(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	view, err := h.offers.GetOffer(r.Context(), oppID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toOfferViewDTO(view))
}

func (h *OfferHandler) handleUpsertOffer(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	var req OfferUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	in, err := req.toDomainInput()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	view, err := h.offers.UpsertOffer(r.Context(), oppID, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toOfferViewDTO(view))
}

func (h *OfferHandler) handleDeleteOffer(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	if err := h.offers.DeleteOffer(r.Context(), oppID); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
