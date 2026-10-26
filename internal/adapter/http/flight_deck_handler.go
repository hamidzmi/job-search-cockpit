package httpadapter

import (
	"net/http"

	"job-search-cockpit/internal/usecase/port/inbound"
)

type FlightDeckHandler struct {
	flightDeck inbound.FlightDeckUseCase
}

func NewFlightDeckHandler(flightDeck inbound.FlightDeckUseCase) *FlightDeckHandler {
	return &FlightDeckHandler{flightDeck: flightDeck}
}

func (h *FlightDeckHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/flight-deck", h.handleGetFlightDeck)
}

func (h *FlightDeckHandler) handleGetFlightDeck(w http.ResponseWriter, r *http.Request) {
	fd, err := h.flightDeck.GetFlightDeck(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toFlightDeckDTO(fd))
}
