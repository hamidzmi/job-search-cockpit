package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type FitAssessmentRepository interface {
	SaveFitAssessment(ctx context.Context, fa domain.FitAssessment) error
	GetFitAssessment(ctx context.Context, opportunityID string) (domain.FitAssessment, error)
}
