package httpadapter

import (
	"encoding/json"
	"io"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type FitAssessmentHandler struct {
	svc inbound.FitAssessmentUseCase
}

func NewFitAssessmentHandler(svc inbound.FitAssessmentUseCase) *FitAssessmentHandler {
	return &FitAssessmentHandler{svc: svc}
}

func (h *FitAssessmentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/opportunities/{id}/fit-assessment", h.handleGetFitAssessment)
	mux.HandleFunc("PUT /api/opportunities/{id}/fit-assessment", h.handleUpsertFitAssessment)
	mux.HandleFunc("POST /api/opportunities/{id}/fit-assessment/generate", h.handleGenerateFitAssessment)
}

func (h *FitAssessmentHandler) handleGetFitAssessment(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	view, err := h.svc.GetFitAssessment(r.Context(), oppID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toFitAssessmentDTO(view))
}

func (h *FitAssessmentHandler) handleUpsertFitAssessment(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	var req FitAssessmentUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	view, err := h.svc.UpsertFitAssessment(r.Context(), oppID, req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toFitAssessmentDTO(view))
}

func (h *FitAssessmentHandler) handleGenerateFitAssessment(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	var req FitAssessmentGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	view, err := h.svc.GenerateFitAssessment(r.Context(), oppID, req.ModelOverride)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toFitAssessmentDTO(view))
}
