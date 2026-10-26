package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type AISettingsHandler struct {
	svc inbound.AISettingsUseCase
}

func NewAISettingsHandler(svc inbound.AISettingsUseCase) *AISettingsHandler {
	return &AISettingsHandler{svc: svc}
}

func (h *AISettingsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/settings/ai", h.handleGetAISettings)
	mux.HandleFunc("PUT /api/settings/ai", h.handleUpdateAISettings)
	mux.HandleFunc("GET /api/settings/ai/openrouter-models", h.handleListOpenRouterModels)
}

func (h *AISettingsHandler) handleGetAISettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.svc.GetAISettings(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAISettingsDTO(settings))
}

func (h *AISettingsHandler) handleUpdateAISettings(w http.ResponseWriter, r *http.Request) {
	var req updateAISettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	updated, err := h.svc.UpdateAISettings(r.Context(), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAISettingsDTO(updated))
}

func (h *AISettingsHandler) handleListOpenRouterModels(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	models, err := h.svc.ListOpenRouterModels(r.Context(), q)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out := make([]OpenRouterModelDTO, 0, len(models))
	for _, m := range models {
		out = append(out, toOpenRouterModelDTO(m))
	}
	writeJSON(w, http.StatusOK, out)
}
