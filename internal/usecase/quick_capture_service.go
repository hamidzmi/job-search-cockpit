package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type QuickCaptureService struct {
	tx            outbound.TxManager
	companies     outbound.CompanyRepository
	opportunities outbound.OpportunityRepository
	transitions   outbound.StageTransitionRepository
	aiSettings    outbound.AISettingsRepository
	parser        outbound.JobPageParser
	assistant     outbound.AssistantGateway
	clock         outbound.Clock
}

func NewQuickCaptureService(
	tx outbound.TxManager,
	companies outbound.CompanyRepository,
	opportunities outbound.OpportunityRepository,
	transitions outbound.StageTransitionRepository,
	aiSettings outbound.AISettingsRepository,
	parser outbound.JobPageParser,
	assistant outbound.AssistantGateway,
	clock outbound.Clock,
) *QuickCaptureService {
	return &QuickCaptureService{
		tx:            tx,
		companies:     companies,
		opportunities: opportunities,
		transitions:   transitions,
		aiSettings:    aiSettings,
		parser:        parser,
		assistant:     assistant,
		clock:         clock,
	}
}

func (s *QuickCaptureService) ExtractJobPosting(ctx context.Context, in domain.QuickCaptureExtractInput) (domain.ExtractedJobPostingDraft, error) {
	sourceURL := strings.TrimSpace(in.SourceURL)
	userRawText := strings.TrimSpace(in.RawText)
	if sourceURL == "" && userRawText == "" {
		return domain.ExtractedJobPostingDraft{}, fmt.Errorf("%w: either source_url or raw_text is required", domain.ErrValidation)
	}

	var (
		parsedPage   domain.ParsedJobPage
		fetchWarning string
	)
	if sourceURL != "" {
		page, err := s.parser.ParseJobPage(ctx, sourceURL)
		if err != nil {
			fetchWarning = fmt.Sprintf("Could not fetch %s (%v); falling back to pasted text / manual fields.", sourceURL, err)
			parsedPage = domain.ParsedJobPage{SourceURL: sourceURL}
		} else {
			parsedPage = page
		}
	}

	combinedText := userRawText
	if parsedPage.CleanedText != "" {
		if combinedText == "" {
			combinedText = parsedPage.CleanedText
		} else if !strings.Contains(combinedText, parsedPage.CleanedText) {
			combinedText = combinedText + "\n\n" + parsedPage.CleanedText
		}
	}

	settings, err := s.aiSettings.GetAISettings(ctx)
	if err != nil {
		return domain.ExtractedJobPostingDraft{}, err
	}

	var (
		llmDraft         *domain.ExtractedJobPostingDraft
		assistantWarning string
	)
	if settings.IsConfigured() {
		resolvedModel := settings.ResolveFastExtractionModel(in.ModelOverride)
		extracted, llmErr := s.assistant.ExtractJobPosting(ctx, settings, domain.AssistantExtractInput{
			SourceURL:         sourceURL,
			RawText:           combinedText,
			DeterministicHint: parsedPage,
			Model:             resolvedModel,
		})
		if llmErr != nil {
			assistantWarning = fmt.Sprintf("AI extraction unavailable (%v); populated deterministic / fallback fields.", llmErr)
		} else {
			if extracted.ModelUsed == "" {
				extracted.ModelUsed = resolvedModel
			}
			llmDraft = &extracted
		}
	} else {
		assistantWarning = "No AI API key configured; using deterministic parser and manual entry."
	}

	draft := domain.MergeCaptureDraft(sourceURL, userRawText, parsedPage, llmDraft)
	draft.FetchWarning = fetchWarning
	draft.AssistantWarning = assistantWarning
	return draft, nil
}

func (s *QuickCaptureService) CaptureOpportunity(ctx context.Context, in domain.QuickCaptureSaveInput) (domain.OpportunityView, error) {
	now := s.clock.Now()
	var view domain.OpportunityView

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		companyID, err := s.resolveOrCreateCompany(txCtx, in, now)
		if err != nil {
			return err
		}

		opp, tr, err := domain.NewOpportunity(uuid.NewString(), uuid.NewString(), domain.OpportunityInput{
			OpportunityDetailsInput: domain.OpportunityDetailsInput{
				CompanyID:    companyID,
				RoleTitle:    in.RoleTitle,
				Seniority:    in.Seniority,
				Location:     in.Location,
				RemotePolicy: in.RemotePolicy,
				Priority:     in.Priority,
				Notes:        in.Notes,
				JobPosting:   in.JobPosting,
			},
			Stage:          in.Stage,
			TransitionNote: in.TransitionNote,
		}, now)
		if err != nil {
			return err
		}

		if err := s.opportunities.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		if err := s.transitions.AppendStageTransition(txCtx, tr); err != nil {
			return err
		}

		loaded, err := s.opportunities.GetOpportunityView(txCtx, opp.ID)
		if err != nil {
			return err
		}
		view = loaded
		return nil
	})
	if err != nil {
		return domain.OpportunityView{}, err
	}
	return view, nil
}

func (s *QuickCaptureService) resolveOrCreateCompany(ctx context.Context, in domain.QuickCaptureSaveInput, now time.Time) (string, error) {
	companyID := strings.TrimSpace(in.CompanyID)
	if companyID != "" {
		c, err := s.companies.GetCompany(ctx, companyID)
		if err != nil {
			return "", err
		}
		return c.ID, nil
	}

	companyName := strings.TrimSpace(in.CompanyName)
	if companyName == "" {
		return "", fmt.Errorf("%w: company_id or company_name is required", domain.ErrValidation)
	}

	existing, err := s.companies.ListCompanies(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range existing {
		if strings.EqualFold(strings.TrimSpace(c.Name), companyName) {
			return c.ID, nil
		}
	}

	created, err := domain.NewCompany(uuid.NewString(), domain.CompanyInput{
		Name:         companyName,
		Domain:       strings.TrimSpace(in.CompanyDomain),
		Headquarters: strings.TrimSpace(in.Location),
		RemotePolicy: string(in.RemotePolicy),
	}, now)
	if err != nil {
		return "", err
	}
	if err := s.companies.SaveCompany(ctx, created); err != nil {
		return "", err
	}
	return created.ID, nil
}
