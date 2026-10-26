package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type CandidateProfileHandler struct {
	profiles inbound.CandidateProfileUseCase
}

func NewCandidateProfileHandler(profiles inbound.CandidateProfileUseCase) *CandidateProfileHandler {
	return &CandidateProfileHandler{profiles: profiles}
}

func (h *CandidateProfileHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/profile", h.handleGetProfile)
	mux.HandleFunc("PUT /api/profile", h.handleSaveProfile)
}

func (h *CandidateProfileHandler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.profiles.GetCandidateProfile(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCandidateProfileDTO(profile))
}

func (h *CandidateProfileHandler) handleSaveProfile(w http.ResponseWriter, r *http.Request) {
	var req CandidateProfileUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	profile, err := h.profiles.SaveCandidateProfile(r.Context(), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCandidateProfileDTO(profile))
}
