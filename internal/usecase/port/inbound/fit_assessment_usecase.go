package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type FitAssessmentUseCase interface {
	GetFitAssessment(ctx context.Context, opportunityID string) (domain.FitAssessmentView, error)
	UpsertFitAssessment(ctx context.Context, opportunityID string, in domain.FitAssessmentInput) (domain.FitAssessmentView, error)
	GenerateFitAssessment(ctx context.Context, opportunityID string, modelOverride string) (domain.FitAssessmentView, error)
}
