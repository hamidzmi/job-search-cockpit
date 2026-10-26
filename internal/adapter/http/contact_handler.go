package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type ContactHandler struct {
	contacts inbound.ContactUseCase
}

func NewContactHandler(contacts inbound.ContactUseCase) *ContactHandler {
	return &ContactHandler{contacts: contacts}
}

func (h *ContactHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/contacts", h.handleListContacts)
	mux.HandleFunc("POST /api/contacts", h.handleCreateContact)
	mux.HandleFunc("GET /api/contacts/{id}", h.handleGetContact)
	mux.HandleFunc("PUT /api/contacts/{id}", h.handleUpdateContact)
	mux.HandleFunc("GET /api/opportunities/{id}/contacts", h.handleListOpportunityContacts)
	mux.HandleFunc("POST /api/opportunities/{id}/contacts", h.handleLinkContactToOpportunity)
	mux.HandleFunc("DELETE /api/opportunities/{id}/contacts/{contact_id}", h.handleUnlinkContactFromOpportunity)
}

func (h *ContactHandler) handleListContacts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.ContactFilter{
		CompanyID: q.Get("company_id"),
		Warmth:    domain.ContactWarmth(q.Get("warmth")),
		Query:     q.Get("q"),
	}
	views, err := h.contacts.ListContacts(r.Context(), filter)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]ContactDTO, 0, len(views))
	for _, v := range views {
		dtos = append(dtos, toContactDTO(v))
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *ContactHandler) handleCreateContact(w http.ResponseWriter, r *http.Request) {
	var req ContactUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	created, err := h.contacts.CreateContact(r.Context(), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toContactDTO(created))
}

func (h *ContactHandler) handleGetContact(w http.ResponseWriter, r *http.Request) {
	view, err := h.contacts.GetContact(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toContactDTO(view))
}

func (h *ContactHandler) handleUpdateContact(w http.ResponseWriter, r *http.Request) {
	var req ContactUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	updated, err := h.contacts.UpdateContact(r.Context(), r.PathValue("id"), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toContactDTO(updated))
}

func (h *ContactHandler) handleListOpportunityContacts(w http.ResponseWriter, r *http.Request) {
	views, err := h.contacts.ListOpportunityContacts(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]OpportunityContactDTO, 0, len(views))
	for _, v := range views {
		dtos = append(dtos, toOpportunityContactDTO(v))
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *ContactHandler) handleLinkContactToOpportunity(w http.ResponseWriter, r *http.Request) {
	var req OpportunityContactLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	linked, err := h.contacts.LinkContactToOpportunity(r.Context(), domain.OpportunityContactInput{
		OpportunityID: r.PathValue("id"),
		ContactID:     req.ContactID,
		Role:          domain.StakeholderRole(req.Role),
		Notes:         req.Notes,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toOpportunityContactDTO(linked))
}

func (h *ContactHandler) handleUnlinkContactFromOpportunity(w http.ResponseWriter, r *http.Request) {
	if err := h.contacts.UnlinkContactFromOpportunity(r.Context(), r.PathValue("id"), r.PathValue("contact_id")); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
