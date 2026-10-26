package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type ApplicationArtifactService struct {
	tx             outbound.TxManager
	artifacts      outbound.ApplicationArtifactRepository
	opportunities  outbound.OpportunityRepository
	profiles       outbound.CandidateProfileRepository
	stories        outbound.StoryRepository
	fitAssessments outbound.FitAssessmentRepository
	aiSettings     outbound.AISettingsRepository
	assistant      outbound.AssistantGateway
	blobs          outbound.ArtifactBlobStore
	clock          outbound.Clock
}

func NewApplicationArtifactService(
	tx outbound.TxManager,
	artifacts outbound.ApplicationArtifactRepository,
	opportunities outbound.OpportunityRepository,
	profiles outbound.CandidateProfileRepository,
	stories outbound.StoryRepository,
	fitAssessments outbound.FitAssessmentRepository,
	aiSettings outbound.AISettingsRepository,
	assistant outbound.AssistantGateway,
	blobs outbound.ArtifactBlobStore,
	clock outbound.Clock,
) *ApplicationArtifactService {
	return &ApplicationArtifactService{
		tx:             tx,
		artifacts:      artifacts,
		opportunities:  opportunities,
		profiles:       profiles,
		stories:        stories,
		fitAssessments: fitAssessments,
		aiSettings:     aiSettings,
		assistant:      assistant,
		blobs:          blobs,
		clock:          clock,
	}
}

func (s *ApplicationArtifactService) ListApplicationArtifacts(ctx context.Context, opportunityID string) ([]domain.ApplicationArtifact, error) {
	trimmedOppID := strings.TrimSpace(opportunityID)
	if _, err := s.opportunities.GetOpportunity(ctx, trimmedOppID); err != nil {
		return nil, err
	}
	return s.artifacts.ListApplicationArtifacts(ctx, trimmedOppID)
}

func (s *ApplicationArtifactService) GetApplicationArtifact(ctx context.Context, opportunityID, artifactID string) (domain.ApplicationArtifact, error) {
	art, err := s.artifacts.GetApplicationArtifact(ctx, strings.TrimSpace(artifactID))
	if err != nil {
		return domain.ApplicationArtifact{}, err
	}
	if art.OpportunityID != strings.TrimSpace(opportunityID) {
		return domain.ApplicationArtifact{}, fmt.Errorf("%w: application artifact %s", domain.ErrNotFound, artifactID)
	}
	return art, nil
}

func (s *ApplicationArtifactService) CreateApplicationArtifact(ctx context.Context, opportunityID string, in domain.ApplicationArtifactInput) (domain.ApplicationArtifact, error) {
	trimmedOppID := strings.TrimSpace(opportunityID)
	now := s.clock.Now().UTC()

	var created domain.ApplicationArtifact
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		opp, err := s.opportunities.GetOpportunity(txCtx, trimmedOppID)
		if err != nil {
			return err
		}
		art, err := domain.NewApplicationArtifact(uuid.NewString(), opp.ID, in, now)
		if err != nil {
			return err
		}
		if err := s.artifacts.SaveApplicationArtifact(txCtx, art); err != nil {
			return err
		}
		opp.TouchUpdatedAt(now)
		if err := s.opportunities.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		created = art
		return nil
	})
	if err != nil {
		return domain.ApplicationArtifact{}, err
	}
	return created, nil
}

func (s *ApplicationArtifactService) UpdateApplicationArtifact(ctx context.Context, opportunityID, artifactID string, in domain.ApplicationArtifactInput) (domain.ApplicationArtifact, error) {
	trimmedOppID := strings.TrimSpace(opportunityID)
	trimmedArtID := strings.TrimSpace(artifactID)
	now := s.clock.Now().UTC()

	var updated domain.ApplicationArtifact
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		opp, err := s.opportunities.GetOpportunity(txCtx, trimmedOppID)
		if err != nil {
			return err
		}
		art, err := s.artifacts.GetApplicationArtifact(txCtx, trimmedArtID)
		if err != nil {
			return err
		}
		if art.OpportunityID != opp.ID {
			return fmt.Errorf("%w: application artifact %s", domain.ErrNotFound, trimmedArtID)
		}
		if err := art.Update(in, now); err != nil {
			return err
		}
		if err := s.artifacts.SaveApplicationArtifact(txCtx, art); err != nil {
			return err
		}
		opp.TouchUpdatedAt(now)
		if err := s.opportunities.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		updated = art
		return nil
	})
	if err != nil {
		return domain.ApplicationArtifact{}, err
	}
	return updated, nil
}

func (s *ApplicationArtifactService) DeleteApplicationArtifact(ctx context.Context, opportunityID, artifactID string) error {
	trimmedOppID := strings.TrimSpace(opportunityID)
	trimmedArtID := strings.TrimSpace(artifactID)
	now := s.clock.Now().UTC()

	var filePath string
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		opp, err := s.opportunities.GetOpportunity(txCtx, trimmedOppID)
		if err != nil {
			return err
		}
		art, err := s.artifacts.GetApplicationArtifact(txCtx, trimmedArtID)
		if err != nil {
			return err
		}
		if art.OpportunityID != opp.ID {
			return fmt.Errorf("%w: application artifact %s", domain.ErrNotFound, trimmedArtID)
		}
		filePath = art.FilePath
		if err := s.artifacts.DeleteApplicationArtifact(txCtx, art.ID); err != nil {
			return err
		}
		opp.TouchUpdatedAt(now)
		return s.opportunities.SaveOpportunity(txCtx, opp)
	})
	if err != nil {
		return err
	}
	if filePath != "" && s.blobs != nil {
		_ = s.blobs.DeleteArtifactBlob(ctx, filePath)
	}
	return nil
}

func (s *ApplicationArtifactService) DraftApplicationArtifact(ctx context.Context, opportunityID string, in domain.ArtifactDraftRequestInput) (domain.AssistantArtifactDraft, error) {
	kind := domain.ArtifactKind(strings.TrimSpace(string(in.Kind)))
	if !kind.IsValid() {
		return domain.AssistantArtifactDraft{}, fmt.Errorf("%w: invalid application artifact kind %q", domain.ErrValidation, in.Kind)
	}

	oppView, err := s.opportunities.GetOpportunityView(ctx, strings.TrimSpace(opportunityID))
	if err != nil {
		return domain.AssistantArtifactDraft{}, err
	}
	profile, err := loadCandidateProfileOrDefault(ctx, s.profiles)
	if err != nil {
		return domain.AssistantArtifactDraft{}, err
	}
	allStories, err := s.stories.ListStories(ctx, domain.StoryFilter{})
	if err != nil {
		return domain.AssistantArtifactDraft{}, err
	}
	settings, err := s.aiSettings.GetAISettings(ctx)
	if err != nil {
		return domain.AssistantArtifactDraft{}, err
	}

	var fitPtr *domain.FitAssessment
	if fa, err := s.fitAssessments.GetFitAssessment(ctx, oppView.Opportunity.ID); err == nil {
		fitPtr = &fa
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.AssistantArtifactDraft{}, err
	}

	return s.assistant.DraftArtifact(ctx, settings, domain.AssistantDraftArtifactInput{
		Opportunity:     oppView,
		Profile:         profile,
		Stories:         allStories,
		FitAssessment:   fitPtr,
		Kind:            kind,
		Title:           strings.TrimSpace(in.Title),
		ExistingContent: strings.TrimSpace(in.ExistingContent),
		Instructions:    strings.TrimSpace(in.Instructions),
		ModelOverride:   strings.TrimSpace(in.ModelOverride),
	})
}

func (s *ApplicationArtifactService) AttachArtifactPDF(ctx context.Context, opportunityID, artifactID, fileName string, content []byte) (domain.ApplicationArtifact, error) {
	if len(content) < 5 || string(content[:5]) != "%PDF-" {
		return domain.ApplicationArtifact{}, fmt.Errorf("%w: uploaded file must be a valid PDF document", domain.ErrValidation)
	}
	if s.blobs == nil {
		return domain.ApplicationArtifact{}, fmt.Errorf("artifact blob store is not configured")
	}

	trimmedOppID := strings.TrimSpace(opportunityID)
	trimmedArtID := strings.TrimSpace(artifactID)

	artCheck, err := s.GetApplicationArtifact(ctx, trimmedOppID, trimmedArtID)
	if err != nil {
		return domain.ApplicationArtifact{}, err
	}

	storageKey, err := s.blobs.SaveArtifactBlob(ctx, trimmedOppID, artCheck.ID, fileName, content)
	if err != nil {
		return domain.ApplicationArtifact{}, err
	}

	now := s.clock.Now().UTC()
	var updated domain.ApplicationArtifact
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		opp, err := s.opportunities.GetOpportunity(txCtx, trimmedOppID)
		if err != nil {
			return err
		}
		art, err := s.artifacts.GetApplicationArtifact(txCtx, trimmedArtID)
		if err != nil {
			return err
		}
		if art.OpportunityID != opp.ID {
			return fmt.Errorf("%w: application artifact %s", domain.ErrNotFound, trimmedArtID)
		}
		if err := art.AttachFile(storageKey, fileName, "application/pdf", int64(len(content)), now); err != nil {
			return err
		}
		if err := s.artifacts.SaveApplicationArtifact(txCtx, art); err != nil {
			return err
		}
		opp.TouchUpdatedAt(now)
		if err := s.opportunities.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		updated = art
		return nil
	})
	if err != nil {
		_ = s.blobs.DeleteArtifactBlob(ctx, storageKey)
		return domain.ApplicationArtifact{}, err
	}
	return updated, nil
}

func (s *ApplicationArtifactService) DownloadArtifactPDF(ctx context.Context, opportunityID, artifactID string) (domain.ArtifactFilePayload, error) {
	art, err := s.GetApplicationArtifact(ctx, opportunityID, artifactID)
	if err != nil {
		return domain.ArtifactFilePayload{}, err
	}
	if !art.HasAttachedFile() || s.blobs == nil {
		return domain.ArtifactFilePayload{}, fmt.Errorf("%w: no PDF file attached to artifact %s", domain.ErrNotFound, artifactID)
	}
	data, err := s.blobs.ReadArtifactBlob(ctx, art.FilePath)
	if err != nil {
		return domain.ArtifactFilePayload{}, err
	}
	return domain.ArtifactFilePayload{
		FileName: art.FileName,
		MimeType: art.MimeType,
		Content:  data,
	}, nil
}
