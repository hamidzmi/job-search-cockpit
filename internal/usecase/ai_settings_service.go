package usecase

import (
	"context"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type AISettingsService struct {
	tx        outbound.TxManager
	repo      outbound.AISettingsRepository
	assistant outbound.AssistantGateway
	clock     outbound.Clock
}

func NewAISettingsService(
	tx outbound.TxManager,
	repo outbound.AISettingsRepository,
	assistant outbound.AssistantGateway,
	clock outbound.Clock,
) *AISettingsService {
	return &AISettingsService{
		tx:        tx,
		repo:      repo,
		assistant: assistant,
		clock:     clock,
	}
}

func (s *AISettingsService) GetAISettings(ctx context.Context) (domain.AISettings, error) {
	settings, err := s.repo.GetAISettings(ctx)
	if err != nil {
		return domain.AISettings{}, err
	}
	if settings.UpdatedAt.IsZero() {
		settings.UpdatedAt = s.clock.Now().UTC()
	}
	return settings, nil
}

func (s *AISettingsService) UpdateAISettings(ctx context.Context, in domain.AISettingsInput) (domain.AISettings, error) {
	settings, err := domain.NewAISettings(in, s.clock.Now())
	if err != nil {
		return domain.AISettings{}, err
	}
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.repo.SaveAISettings(txCtx, settings)
	})
	if err != nil {
		return domain.AISettings{}, err
	}
	return settings, nil
}

func (s *AISettingsService) ListOpenRouterModels(ctx context.Context, query string) ([]domain.OpenRouterModel, error) {
	settings, err := s.repo.GetAISettings(ctx)
	if err != nil {
		return nil, err
	}
	models, err := s.assistant.ListOpenRouterModels(ctx, settings)
	if err != nil {
		return nil, err
	}
	return domain.FilterOpenRouterModels(models, query), nil
}
