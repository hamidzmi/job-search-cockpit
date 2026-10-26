package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"

	"job-search-cockpit/internal/domain"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		writeJSON(w, http.StatusBadRequest, ErrorEnvelope{
			Error: ErrorBody{Code: "validation_error", Message: err.Error()},
		})
	case errors.Is(err, domain.ErrInvalidTransition):
		writeJSON(w, http.StatusUnprocessableEntity, ErrorEnvelope{
			Error: ErrorBody{Code: "invalid_transition", Message: err.Error()},
		})
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, ErrorEnvelope{
			Error: ErrorBody{Code: "not_found", Message: err.Error()},
		})
	default:
		writeJSON(w, http.StatusInternalServerError, ErrorEnvelope{
			Error: ErrorBody{Code: "internal_error", Message: err.Error()},
		})
	}
}
