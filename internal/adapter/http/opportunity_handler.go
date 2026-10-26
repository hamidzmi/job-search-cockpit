package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type OpportunityHandler struct {
	opportunities inbound.OpportunityUseCase
}

func NewOpportunityHandler(opportunities inbound.OpportunityUseCase) *OpportunityHandler {
	return &OpportunityHandler{opportunities: opportunities}
}

func (h *OpportunityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/opportunities", h.handleListOpportunities)
	mux.HandleFunc("POST /api/opportunities", h.handleCreateOpportunity)
	mux.HandleFunc("GET /api/opportunities/{id}", h.handleGetOpportunity)
	mux.HandleFunc("PUT /api/opportunities/{id}", h.handleUpdateOpportunity)
	mux.HandleFunc("POST /api/opportunities/{id}/transition", h.handleTransitionOpportunity)
	mux.HandleFunc("GET /api/opportunities/{id}/transitions", h.handleListStageTransitions)
}

func (h *OpportunityHandler) handleListOpportunities(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.OpportunityFilter{
		Stage:        domain.Stage(q.Get("stage")),
		Outcome:      domain.Outcome(q.Get("outcome")),
		Priority:     q.Get("priority"),
		RemotePolicy: q.Get("remote_policy"),
		CompanyID:    q.Get("company_id"),
		Query:        q.Get("q"),
	}
	opps, err := h.opportunities.ListOpportunities(r.Context(), filter)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]OpportunityDTO, 0, len(opps))
	for _, o := range opps {
		dtos = append(dtos, toOpportunityDTO(o, nil))
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *OpportunityHandler) handleCreateOpportunity(w http.ResponseWriter, r *http.Request) {
	var req OpportunityCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	opp, err := h.opportunities.CreateOpportunity(r.Context(), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toOpportunityDTO(opp, nil))
}

func (h *OpportunityHandler) handleGetOpportunity(w http.ResponseWriter, r *http.Request) {
	opp, transitions, err := h.opportunities.GetOpportunity(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toOpportunityDTO(opp, transitions))
}

func (h *OpportunityHandler) handleUpdateOpportunity(w http.ResponseWriter, r *http.Request) {
	var req OpportunityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	opp, err := h.opportunities.UpdateOpportunityDetails(r.Context(), r.PathValue("id"), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toOpportunityDTO(opp, nil))
}

func (h *OpportunityHandler) handleTransitionOpportunity(w http.ResponseWriter, r *http.Request) {
	var req OpportunityTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	opp, err := h.opportunities.TransitionOpportunity(
		r.Context(),
		r.PathValue("id"),
		domain.Stage(req.Stage),
		domain.Outcome(req.Outcome),
		req.Note,
	)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toOpportunityDTO(opp, nil))
}

func (h *OpportunityHandler) handleListStageTransitions(w http.ResponseWriter, r *http.Request) {
	transitions, err := h.opportunities.ListStageTransitions(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]StageTransitionDTO, 0, len(transitions))
	for _, tr := range transitions {
		dtos = append(dtos, toStageTransitionDTO(tr))
	}
	writeJSON(w, http.StatusOK, dtos)
}
