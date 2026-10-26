package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type ApplicationArtifactRepository interface {
	SaveApplicationArtifact(ctx context.Context, artifact domain.ApplicationArtifact) error
	GetApplicationArtifact(ctx context.Context, id string) (domain.ApplicationArtifact, error)
	ListApplicationArtifacts(ctx context.Context, opportunityID string) ([]domain.ApplicationArtifact, error)
	DeleteApplicationArtifact(ctx context.Context, id string) error
}
