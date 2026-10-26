package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type InteractionHandler struct {
	interactions inbound.InteractionUseCase
}

func NewInteractionHandler(interactions inbound.InteractionUseCase) *InteractionHandler {
	return &InteractionHandler{interactions: interactions}
}

func (h *InteractionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/interactions", h.handleListInteractions)
	mux.HandleFunc("POST /api/interactions", h.handleLogInteraction)
}

func (h *InteractionHandler) handleListInteractions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.InteractionFilter{
		ContactID:     q.Get("contact_id"),
		OpportunityID: q.Get("opportunity_id"),
	}
	views, err := h.interactions.ListInteractions(r.Context(), filter)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]InteractionDTO, 0, len(views))
	for _, v := range views {
		dtos = append(dtos, toInteractionDTO(v))
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *InteractionHandler) handleLogInteraction(w http.ResponseWriter, r *http.Request) {
	var req InteractionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	in, err := req.toDomainInput()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	intView, followUpView, err := h.interactions.LogInteraction(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	var followUpDTO *ActionItemDTO
	if followUpView != nil {
		dto := toActionItemDTO(*followUpView)
		followUpDTO = &dto
	}
	writeJSON(w, http.StatusCreated, LogInteractionResponseDTO{
		Interaction:        toInteractionDTO(intView),
		FollowUpActionItem: followUpDTO,
	})
}
