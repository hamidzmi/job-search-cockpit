package usecase

import (
	"context"
	"errors"
	"strings"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type FitAssessmentService struct {
	tx             outbound.TxManager
	fitAssessments outbound.FitAssessmentRepository
	opportunities  outbound.OpportunityRepository
	profiles       outbound.CandidateProfileRepository
	stories        outbound.StoryRepository
	aiSettings     outbound.AISettingsRepository
	assistant      outbound.AssistantGateway
	clock          outbound.Clock
}

func NewFitAssessmentService(
	tx outbound.TxManager,
	fitAssessments outbound.FitAssessmentRepository,
	opportunities outbound.OpportunityRepository,
	profiles outbound.CandidateProfileRepository,
	stories outbound.StoryRepository,
	aiSettings outbound.AISettingsRepository,
	assistant outbound.AssistantGateway,
	clock outbound.Clock,
) *FitAssessmentService {
	return &FitAssessmentService{
		tx:             tx,
		fitAssessments: fitAssessments,
		opportunities:  opportunities,
		profiles:       profiles,
		stories:        stories,
		aiSettings:     aiSettings,
		assistant:      assistant,
		clock:          clock,
	}
}

func (s *FitAssessmentService) GetFitAssessment(ctx context.Context, opportunityID string) (domain.FitAssessmentView, error) {
	oppView, err := s.opportunities.GetOpportunityView(ctx, strings.TrimSpace(opportunityID))
	if err != nil {
		return domain.FitAssessmentView{}, err
	}
	profile, err := loadCandidateProfileOrDefault(ctx, s.profiles)
	if err != nil {
		return domain.FitAssessmentView{}, err
	}

	fa, err := s.fitAssessments.GetFitAssessment(ctx, oppView.Opportunity.ID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.FitAssessmentView{
			FitAssessment: domain.FitAssessment{
				OpportunityID:         oppView.Opportunity.ID,
				MatchedStrengths:      []string{},
				GapsAndRisks:          []string{},
				CompensationAlignment: domain.EvaluateCompensationAlignment(oppView.Opportunity.JobPosting, profile),
				RecommendedStoryIDs:   []string{},
			},
			RecommendedStories: []domain.RecommendedStorySummary{},
		}, nil
	}
	if err != nil {
		return domain.FitAssessmentView{}, err
	}

	if strings.TrimSpace(fa.CompensationAlignment) == "" {
		fa.CompensationAlignment = domain.EvaluateCompensationAlignment(oppView.Opportunity.JobPosting, profile)
	}
	return s.hydrateView(ctx, fa)
}

func (s *FitAssessmentService) UpsertFitAssessment(ctx context.Context, opportunityID string, in domain.FitAssessmentInput) (domain.FitAssessmentView, error) {
	trimmedOppID := strings.TrimSpace(opportunityID)
	now := s.clock.Now().UTC()

	var saved domain.FitAssessment
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		opp, err := s.opportunities.GetOpportunity(txCtx, trimmedOppID)
		if err != nil {
			return err
		}
		profile, err := loadCandidateProfileOrDefault(txCtx, s.profiles)
		if err != nil {
			return err
		}
		if strings.TrimSpace(in.CompensationAlignment) == "" {
			in.CompensationAlignment = domain.EvaluateCompensationAlignment(opp.JobPosting, profile)
		}

		fa, err := domain.NewFitAssessment(opp.ID, in, now)
		if err != nil {
			return err
		}
		if err := s.fitAssessments.SaveFitAssessment(txCtx, fa); err != nil {
			return err
		}

		opp.TouchUpdatedAt(now)
		if err := s.opportunities.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		saved = fa
		return nil
	})
	if err != nil {
		return domain.FitAssessmentView{}, err
	}

	return s.hydrateView(ctx, saved)
}

func (s *FitAssessmentService) GenerateFitAssessment(ctx context.Context, opportunityID string, modelOverride string) (domain.FitAssessmentView, error) {
	trimmedOppID := strings.TrimSpace(opportunityID)
	oppView, err := s.opportunities.GetOpportunityView(ctx, trimmedOppID)
	if err != nil {
		return domain.FitAssessmentView{}, err
	}
	profile, err := loadCandidateProfileOrDefault(ctx, s.profiles)
	if err != nil {
		return domain.FitAssessmentView{}, err
	}
	allStories, err := s.stories.ListStories(ctx, domain.StoryFilter{})
	if err != nil {
		return domain.FitAssessmentView{}, err
	}
	settings, err := s.aiSettings.GetAISettings(ctx)
	if err != nil {
		return domain.FitAssessmentView{}, err
	}

	draft, err := s.assistant.AssessFit(ctx, settings, domain.AssistantAssessFitInput{
		Opportunity:   oppView,
		Profile:       profile,
		Stories:       allStories,
		ModelOverride: strings.TrimSpace(modelOverride),
	})
	if err != nil {
		return domain.FitAssessmentView{}, err
	}

	return s.UpsertFitAssessment(ctx, oppView.Opportunity.ID, domain.FitAssessmentInput{
		OverallScore:          draft.OverallScore,
		OverallSummary:        draft.OverallSummary,
		MatchedStrengths:      draft.MatchedStrengths,
		GapsAndRisks:          draft.GapsAndRisks,
		CompensationAlignment: draft.CompensationAlignment,
		RecommendedStoryIDs:   draft.RecommendedStoryIDs,
		ModelUsed:             draft.ModelUsed,
	})
}

func (s *FitAssessmentService) hydrateView(ctx context.Context, fa domain.FitAssessment) (domain.FitAssessmentView, error) {
	allStories, err := s.stories.ListStories(ctx, domain.StoryFilter{})
	if err != nil {
		return domain.FitAssessmentView{}, err
	}
	byID := make(map[string]domain.Story, len(allStories))
	for _, st := range allStories {
		byID[st.ID] = st
	}

	rec := make([]domain.RecommendedStorySummary, 0, len(fa.RecommendedStoryIDs))
	for _, id := range fa.RecommendedStoryIDs {
		if st, ok := byID[id]; ok {
			rec = append(rec, domain.RecommendedStorySummary{
				ID:             st.ID,
				Title:          st.Title,
				Hook:           st.Hook,
				CompetencyTags: st.CompetencyTags,
				SkillTags:      st.SkillTags,
			})
		}
	}

	return domain.FitAssessmentView{
		FitAssessment:      fa,
		RecommendedStories: rec,
	}, nil
}

func loadCandidateProfileOrDefault(ctx context.Context, profiles outbound.CandidateProfileRepository) (domain.CandidateProfile, error) {
	profile, err := profiles.GetCandidateProfile(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.DefaultCandidateProfile(), nil
	}
	return profile, err
}

