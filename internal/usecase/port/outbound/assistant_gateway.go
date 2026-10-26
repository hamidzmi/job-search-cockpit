package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type AssistantGateway interface {
	ExtractJobPosting(ctx context.Context, settings domain.AISettings, in domain.AssistantExtractInput) (domain.ExtractedJobPostingDraft, error)
	ListOpenRouterModels(ctx context.Context, settings domain.AISettings) ([]domain.OpenRouterModel, error)
}
