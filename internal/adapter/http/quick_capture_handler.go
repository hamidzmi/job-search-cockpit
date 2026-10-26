package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type QuickCaptureHandler struct {
	svc inbound.QuickCaptureUseCase
}

func NewQuickCaptureHandler(svc inbound.QuickCaptureUseCase) *QuickCaptureHandler {
	return &QuickCaptureHandler{svc: svc}
}

func (h *QuickCaptureHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/quick-capture/extract", h.handleExtractJobPosting)
	mux.HandleFunc("POST /api/quick-capture", h.handleCaptureOpportunity)
	mux.HandleFunc("POST /api/quick-capture/save", h.handleCaptureOpportunity)
}

func (h *QuickCaptureHandler) handleExtractJobPosting(w http.ResponseWriter, r *http.Request) {
	var req quickCaptureExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	draft, err := h.svc.ExtractJobPosting(r.Context(), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toExtractedJobPostingDraftDTO(draft))
}

func (h *QuickCaptureHandler) handleCaptureOpportunity(w http.ResponseWriter, r *http.Request) {
	var req quickCaptureSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	created, err := h.svc.CaptureOpportunity(r.Context(), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toOpportunityDTO(created, nil))
}
