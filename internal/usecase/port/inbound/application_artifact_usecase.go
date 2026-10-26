package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type ApplicationArtifactUseCase interface {
	ListApplicationArtifacts(ctx context.Context, opportunityID string) ([]domain.ApplicationArtifact, error)
	GetApplicationArtifact(ctx context.Context, opportunityID, artifactID string) (domain.ApplicationArtifact, error)
	CreateApplicationArtifact(ctx context.Context, opportunityID string, in domain.ApplicationArtifactInput) (domain.ApplicationArtifact, error)
	UpdateApplicationArtifact(ctx context.Context, opportunityID, artifactID string, in domain.ApplicationArtifactInput) (domain.ApplicationArtifact, error)
	DeleteApplicationArtifact(ctx context.Context, opportunityID, artifactID string) error
	DraftApplicationArtifact(ctx context.Context, opportunityID string, in domain.ArtifactDraftRequestInput) (domain.AssistantArtifactDraft, error)
	AttachArtifactPDF(ctx context.Context, opportunityID, artifactID, fileName string, content []byte) (domain.ApplicationArtifact, error)
	DownloadArtifactPDF(ctx context.Context, opportunityID, artifactID string) (domain.ArtifactFilePayload, error)
}
