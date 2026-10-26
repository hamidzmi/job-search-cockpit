package httpadapter

import (
	"strings"
	"time"

	"job-search-cockpit/internal/domain"
)

type AISettingsDTO struct {
	Provider                string    `json:"provider"`
	BaseURL                 string    `json:"base_url"`
	APIKey                  string    `json:"api_key"`
	HasAPIKey               bool      `json:"has_api_key"`
	IsConfigured            bool      `json:"is_configured"`
	FastExtractionModel     string    `json:"fast_extraction_model"`
	ReasoningTailoringModel string    `json:"reasoning_tailoring_model"`
	BookmarkedModels        []string  `json:"bookmarked_models"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type updateAISettingsRequest struct {
	Provider                string   `json:"provider"`
	BaseURL                 string   `json:"base_url"`
	APIKey                  string   `json:"api_key"`
	FastExtractionModel     string   `json:"fast_extraction_model"`
	ReasoningTailoringModel string   `json:"reasoning_tailoring_model"`
	BookmarkedModels        []string `json:"bookmarked_models"`
}

func (r updateAISettingsRequest) toDomainInput() domain.AISettingsInput {
	return domain.AISettingsInput{
		Provider:                domain.AIProvider(r.Provider),
		BaseURL:                 r.BaseURL,
		APIKey:                  r.APIKey,
		FastExtractionModel:     r.FastExtractionModel,
		ReasoningTailoringModel: r.ReasoningTailoringModel,
		BookmarkedModels:        r.BookmarkedModels,
	}
}

func toAISettingsDTO(s domain.AISettings) AISettingsDTO {
	bookmarks := s.BookmarkedModels
	if bookmarks == nil {
		bookmarks = []string{}
	}
	return AISettingsDTO{
		Provider:                string(s.Provider),
		BaseURL:                 s.BaseURL,
		APIKey:                  s.APIKey,
		HasAPIKey:               strings.TrimSpace(s.APIKey) != "",
		IsConfigured:            s.IsConfigured(),
		FastExtractionModel:     s.FastExtractionModel,
		ReasoningTailoringModel: s.ReasoningTailoringModel,
		BookmarkedModels:        bookmarks,
		UpdatedAt:               s.UpdatedAt,
	}
}

type OpenRouterModelDTO struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ContextLength     int    `json:"context_length"`
	PromptPricing     string `json:"prompt_pricing"`
	CompletionPricing string `json:"completion_pricing"`
}

func toOpenRouterModelDTO(m domain.OpenRouterModel) OpenRouterModelDTO {
	return OpenRouterModelDTO{
		ID:                m.ID,
		Name:              m.Name,
		Description:       m.Description,
		ContextLength:     m.ContextLength,
		PromptPricing:     m.PromptPricing,
		CompletionPricing: m.CompletionPricing,
	}
}
