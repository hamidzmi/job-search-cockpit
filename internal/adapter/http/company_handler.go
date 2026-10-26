package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type CompanyHandler struct {
	companies inbound.CompanyUseCase
}

func NewCompanyHandler(companies inbound.CompanyUseCase) *CompanyHandler {
	return &CompanyHandler{companies: companies}
}

func (h *CompanyHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/companies", h.handleListCompanies)
	mux.HandleFunc("POST /api/companies", h.handleCreateCompany)
	mux.HandleFunc("GET /api/companies/{id}", h.handleGetCompany)
	mux.HandleFunc("PUT /api/companies/{id}", h.handleUpdateCompany)
}

func (h *CompanyHandler) handleListCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := h.companies.ListCompanies(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]CompanyDTO, 0, len(companies))
	for _, c := range companies {
		dtos = append(dtos, toCompanyDTO(c))
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *CompanyHandler) handleCreateCompany(w http.ResponseWriter, r *http.Request) {
	var req CompanyUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	c, err := h.companies.CreateCompany(r.Context(), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toCompanyDTO(c))
}

func (h *CompanyHandler) handleGetCompany(w http.ResponseWriter, r *http.Request) {
	c, err := h.companies.GetCompany(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCompanyDTO(c))
}

func (h *CompanyHandler) handleUpdateCompany(w http.ResponseWriter, r *http.Request) {
	var req CompanyUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	c, err := h.companies.UpdateCompany(r.Context(), r.PathValue("id"), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCompanyDTO(c))
}
