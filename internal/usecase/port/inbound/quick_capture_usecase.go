package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type QuickCaptureUseCase interface {
	ExtractJobPosting(ctx context.Context, in domain.QuickCaptureExtractInput) (domain.ExtractedJobPostingDraft, error)
	CaptureOpportunity(ctx context.Context, in domain.QuickCaptureSaveInput) (domain.OpportunityView, error)
}
