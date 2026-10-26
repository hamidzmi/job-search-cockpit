package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type AISettingsRepository interface {
	GetAISettings(ctx context.Context) (domain.AISettings, error)
	SaveAISettings(ctx context.Context, settings domain.AISettings) error
}
