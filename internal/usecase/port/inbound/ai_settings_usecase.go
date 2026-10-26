package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type AISettingsUseCase interface {
	GetAISettings(ctx context.Context) (domain.AISettings, error)
	UpdateAISettings(ctx context.Context, in domain.AISettingsInput) (domain.AISettings, error)
	ListOpenRouterModels(ctx context.Context, query string) ([]domain.OpenRouterModel, error)
}
