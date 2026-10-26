package domain

import (
	"fmt"
	"strings"
	"time"
)

type AIProvider string

const (
	AIProviderOpenRouter AIProvider = "openrouter"
	AIProviderOpenAI     AIProvider = "openai"
	AIProviderAnthropic  AIProvider = "anthropic"
	AIProviderGemini     AIProvider = "gemini"
	AIProviderOllama     AIProvider = "ollama"
)

func AllAIProviders() []AIProvider {
	return []AIProvider{
		AIProviderOpenRouter,
		AIProviderOpenAI,
		AIProviderAnthropic,
		AIProviderGemini,
		AIProviderOllama,
	}
}

func (p AIProvider) IsValid() bool {
	switch p {
	case AIProviderOpenRouter, AIProviderOpenAI, AIProviderAnthropic, AIProviderGemini, AIProviderOllama:
		return true
	default:
		return false
	}
}

type AISettings struct {
	Provider                AIProvider
	BaseURL                 string
	APIKey                  string
	FastExtractionModel     string
	ReasoningTailoringModel string
	BookmarkedModels        []string
	UpdatedAt               time.Time
}

type AISettingsInput struct {
	Provider                AIProvider
	BaseURL                 string
	APIKey                  string
	FastExtractionModel     string
	ReasoningTailoringModel string
	BookmarkedModels        []string
}

func DefaultAISettings(now time.Time) AISettings {
	return AISettings{
		Provider:         AIProviderOpenRouter,
		BookmarkedModels: []string{},
		UpdatedAt:        now.UTC(),
	}
}

func NewAISettings(in AISettingsInput, now time.Time) (AISettings, error) {
	provider := AIProvider(strings.ToLower(strings.TrimSpace(string(in.Provider))))
	if provider == "" {
		provider = AIProviderOpenRouter
	}
	if !provider.IsValid() {
		return AISettings{}, fmt.Errorf("%w: invalid ai provider %q", ErrValidation, in.Provider)
	}

	bookmarks := make([]string, 0, len(in.BookmarkedModels))
	seen := make(map[string]struct{}, len(in.BookmarkedModels))
	for _, b := range in.BookmarkedModels {
		trimmed := strings.TrimSpace(b)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		bookmarks = append(bookmarks, trimmed)
	}

	return AISettings{
		Provider:                provider,
		BaseURL:                 strings.TrimSpace(in.BaseURL),
		APIKey:                  strings.TrimSpace(in.APIKey),
		FastExtractionModel:     strings.TrimSpace(in.FastExtractionModel),
		ReasoningTailoringModel: strings.TrimSpace(in.ReasoningTailoringModel),
		BookmarkedModels:        bookmarks,
		UpdatedAt:               now.UTC(),
	}, nil
}

func (s AISettings) IsConfigured() bool {
	if strings.TrimSpace(s.APIKey) != "" {
		return true
	}
	if s.Provider == AIProviderOllama && strings.TrimSpace(s.BaseURL) != "" {
		return true
	}
	return false
}

func (s AISettings) ResolveFastExtractionModel(override string) string {
	if trimmed := strings.TrimSpace(override); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(s.FastExtractionModel)
}

func (s AISettings) ResolveReasoningModel(override string) string {
	if trimmed := strings.TrimSpace(override); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(s.ReasoningTailoringModel)
}

type OpenRouterModel struct {
	ID                string
	Name              string
	Description       string
	ContextLength     int
	PromptPricing     string
	CompletionPricing string
}

func FilterOpenRouterModels(models []OpenRouterModel, query string) []OpenRouterModel {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		out := make([]OpenRouterModel, len(models))
		copy(out, models)
		return out
	}
	out := make([]OpenRouterModel, 0)
	for _, m := range models {
		hay := strings.ToLower(m.ID + " " + m.Name + " " + m.Description)
		if strings.Contains(hay, q) {
			out = append(out, m)
		}
	}
	return out
}
