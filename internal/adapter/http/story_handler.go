package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type StoryHandler struct {
	stories inbound.StoryUseCase
}

func NewStoryHandler(stories inbound.StoryUseCase) *StoryHandler {
	return &StoryHandler{stories: stories}
}

func (h *StoryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stories", h.handleListStories)
	mux.HandleFunc("POST /api/stories", h.handleCreateStory)
	mux.HandleFunc("GET /api/stories/{id}", h.handleGetStory)
	mux.HandleFunc("PUT /api/stories/{id}", h.handleUpdateStory)
	mux.HandleFunc("DELETE /api/stories/{id}", h.handleDeleteStory)
}

func (h *StoryHandler) handleListStories(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	competencyTag := q.Get("competency_tag")
	if competencyTag == "" {
		competencyTag = q.Get("competency")
	}
	skillTag := q.Get("skill_tag")
	if skillTag == "" {
		skillTag = q.Get("skill")
	}
	filter := domain.StoryFilter{
		Query:         q.Get("q"),
		CompetencyTag: competencyTag,
		SkillTag:      skillTag,
	}

	stories, err := h.stories.ListStories(r.Context(), filter)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]StoryDTO, 0, len(stories))
	for _, s := range stories {
		dtos = append(dtos, toStoryDTO(s))
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *StoryHandler) handleCreateStory(w http.ResponseWriter, r *http.Request) {
	var req StoryUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	story, err := h.stories.CreateStory(r.Context(), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toStoryDTO(story))
}

func (h *StoryHandler) handleGetStory(w http.ResponseWriter, r *http.Request) {
	story, err := h.stories.GetStory(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toStoryDTO(story))
}

func (h *StoryHandler) handleUpdateStory(w http.ResponseWriter, r *http.Request) {
	var req StoryUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	story, err := h.stories.UpdateStory(r.Context(), r.PathValue("id"), req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toStoryDTO(story))
}

func (h *StoryHandler) handleDeleteStory(w http.ResponseWriter, r *http.Request) {
	if err := h.stories.DeleteStory(r.Context(), r.PathValue("id")); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
