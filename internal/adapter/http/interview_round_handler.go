package httpadapter

import (
	"encoding/json"
	"net/http"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
)

type InterviewRoundHandler struct {
	rounds inbound.InterviewRoundUseCase
}

func NewInterviewRoundHandler(rounds inbound.InterviewRoundUseCase) *InterviewRoundHandler {
	return &InterviewRoundHandler{rounds: rounds}
}

func (h *InterviewRoundHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/opportunities/{id}/rounds", h.handleListInterviewRounds)
	mux.HandleFunc("POST /api/opportunities/{id}/rounds", h.handleCreateInterviewRound)
	mux.HandleFunc("PUT /api/rounds/{id}", h.handleUpdateInterviewRound)
	mux.HandleFunc("DELETE /api/rounds/{id}", h.handleDeleteInterviewRound)
	mux.HandleFunc("PUT /api/rounds/{id}/debrief", h.handleRecordDebrief)
	mux.HandleFunc("GET /api/questions", h.handleListQuestionBank)
}

func (h *InterviewRoundHandler) handleListInterviewRounds(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	views, err := h.rounds.ListInterviewRounds(r.Context(), oppID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]InterviewRoundDTO, 0, len(views))
	for _, v := range views {
		dtos = append(dtos, toInterviewRoundDTO(v))
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (h *InterviewRoundHandler) handleCreateInterviewRound(w http.ResponseWriter, r *http.Request) {
	oppID := r.PathValue("id")
	var req InterviewRoundUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	in, err := req.toDomainInput()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	created, err := h.rounds.CreateInterviewRound(r.Context(), oppID, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toInterviewRoundDTO(created))
}

func (h *InterviewRoundHandler) handleUpdateInterviewRound(w http.ResponseWriter, r *http.Request) {
	roundID := r.PathValue("id")
	var req InterviewRoundUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	in, err := req.toDomainInput()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	updated, err := h.rounds.UpdateInterviewRound(r.Context(), roundID, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toInterviewRoundDTO(updated))
}

func (h *InterviewRoundHandler) handleDeleteInterviewRound(w http.ResponseWriter, r *http.Request) {
	roundID := r.PathValue("id")
	if err := h.rounds.DeleteInterviewRound(r.Context(), roundID); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *InterviewRoundHandler) handleRecordDebrief(w http.ResponseWriter, r *http.Request) {
	roundID := r.PathValue("id")
	var req DebriefUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDomainError(w, domain.ErrValidation)
		return
	}
	updated, err := h.rounds.RecordDebrief(r.Context(), roundID, req.toDomainInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toInterviewRoundDTO(updated))
}

func (h *InterviewRoundHandler) handleListQuestionBank(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := domain.QuestionBankFilter{
		Query:     q.Get("q"),
		Category:  q.Get("category"),
		CompanyID: q.Get("company_id"),
		RoundType: domain.InterviewRoundType(q.Get("round_type")),
		StoryID:   q.Get("story_id"),
	}
	views, err := h.rounds.ListQuestionBank(r.Context(), filter)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	dtos := make([]QuestionBankEntryDTO, 0, len(views))
	for _, v := range views {
		dtos = append(dtos, toQuestionBankEntryDTO(v))
	}
	writeJSON(w, http.StatusOK, dtos)
}
