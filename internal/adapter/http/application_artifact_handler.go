package httpadapter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

const maxArtifactUploadBytes = 25 << 20 // 25 MiB

type ApplicationArtifactHandler struct {
	svc inbound.ApplicationArtifactUseCase
}

func NewApplicationArtifactHandler(svc inbound.ApplicationArtifactUseCase) *ApplicationArtifactHandler {
	return &ApplicationArtifactHandler{svc: svc}
}

func (h *ApplicationArtifactHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/opportunities/{id}/artifacts", h.handleListArtifacts)
	mux.HandleFunc("POST /api/opportunities/{id}/artifacts", h.handleCreateArtifact)
	mux.HandleFunc("POST /api/opportunities/{id}/artifacts/draft", h.handleDraftArtifact)
	mux.HandleFunc("PUT /api/opportunities/{id}/artifacts/{artifactId}", h.handleUpdateArtifact)
	mux.HandleFunc("DELETE /api/opportunities/{id}/artifacts/{artifactId}", h.handleDeleteArtifact)
	mux.HandleFunc("POST /api/opportunities/{id}/artifacts/{artifactId}/file", h.handleUploadArtifactPDF)
	mux.HandleFunc("GET /api/opportunities/{id}/artifacts/{artifactId}/file", h.handleDownloadArtifactPDF)
}

func (h *ApplicationArtifactHandler) handleListArtifacts(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	items, err := h.svc.ListApplicationArtifacts(r.Context(), oppID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]ApplicationArtifactDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, toApplicationArtifactDTO(item))
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *ApplicationArtifactHandler) handleCreateArtifact(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	var req ApplicationArtifactUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	created, err := h.svc.CreateApplicationArtifact(r.Context(), oppID, req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toApplicationArtifactDTO(created))
}

func (h *ApplicationArtifactHandler) handleUpdateArtifact(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	artID := r.PathValue("artifactId")
	var req ApplicationArtifactUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	updated, err := h.svc.UpdateApplicationArtifact(r.Context(), oppID, artID, req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toApplicationArtifactDTO(updated))
}

func (h *ApplicationArtifactHandler) handleDeleteArtifact(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	artID := r.PathValue("artifactId")
	if err := h.svc.DeleteApplicationArtifact(r.Context(), oppID, artID); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ApplicationArtifactHandler) handleDraftArtifact(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	var req ArtifactDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	draft, err := h.svc.DraftApplicationArtifact(r.Context(), oppID, req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ArtifactDraftDTO{
		Kind:            string(draft.Kind),
		Title:           draft.Title,
		ContentMarkdown: draft.ContentMarkdown,
		ModelUsed:       draft.ModelUsed,
	})
}

func (h *ApplicationArtifactHandler) handleUploadArtifactPDF(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	artID := r.PathValue("artifactId")

	r.Body = http.MaxBytesReader(w, r.Body, maxArtifactUploadBytes)
	if err := r.ParseMultipartForm(maxArtifactUploadBytes); err != nil {
		writeDomainError(w, fmt.Errorf("%w: invalid multipart upload: %v", domain.ErrValidation, err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeDomainError(w, fmt.Errorf("%w: missing 'file' field in multipart upload", domain.ErrValidation))
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		writeDomainError(w, fmt.Errorf("%w: failed to read uploaded file", domain.ErrValidation))
		return
	}

	fileName := "artifact.pdf"
	if header != nil && strings.TrimSpace(header.Filename) != "" {
		fileName = strings.TrimSpace(header.Filename)
	}

	updated, err := h.svc.AttachArtifactPDF(r.Context(), oppID, artID, fileName, content)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toApplicationArtifactDTO(updated))
}

func (h *ApplicationArtifactHandler) handleDownloadArtifactPDF(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	artID := r.PathValue("artifactId")

	payload, err := h.svc.DownloadArtifactPDF(r.Context(), oppID, artID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	safeName := strings.ReplaceAll(payload.FileName, `"`, "")
	if safeName == "" {
		safeName = "artifact.pdf"
	}
	w.Header().Set("Content-Type", payload.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, safeName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload.Content)
}
