package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type ActionItemHandler struct {
	actionItems inbound.ActionItemUseCase
}

func NewActionItemHandler(actionItems inbound.ActionItemUseCase) *ActionItemHandler {
	return &ActionItemHandler{actionItems: actionItems}
}

func (h *ActionItemHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/action-items", h.handleListActionItems)
	mux.HandleFunc("POST /api/action-items", h.handleCreateActionItem)
	mux.HandleFunc("PUT /api/action-items/{id}", h.handleUpdateActionItem)
}

func (h *ActionItemHandler) handleListActionItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.ActionItemFilter{
		Status:        domain.ActionItemStatus(q.Get("status")),
		OpportunityID: q.Get("opportunity_id"),
		ContactID:     q.Get("contact_id"),
	}
	views, err := h.actionItems.ListActionItems(r.Context(), filter)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]ActionItemDTO, 0, len(views))
	for _, v := range views {
		dtos = append(dtos, toActionItemDTO(v))
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *ActionItemHandler) handleCreateActionItem(w http.ResponseWriter, r *http.Request) {
	var req ActionItemCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	in, err := req.toDomainInput()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	created, err := h.actionItems.CreateActionItem(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toActionItemDTO(created))
}

func (h *ActionItemHandler) handleUpdateActionItem(w http.ResponseWriter, r *http.Request) {
	var req ActionItemUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	in, err := req.toDomainInput()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	updated, err := h.actionItems.UpdateActionItem(r.Context(), r.PathValue("id"), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toActionItemDTO(updated))
}
