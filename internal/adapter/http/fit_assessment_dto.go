package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type RecommendedStorySummaryDTO struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Hook           string   `json:"hook"`
	CompetencyTags []string `json:"competency_tags"`
	SkillTags      []string `json:"skill_tags"`
}

type FitAssessmentDTO struct {
	OpportunityID         string                       `json:"opportunity_id"`
	OverallScore          int                          `json:"overall_score"`
	OverallSummary        string                       `json:"overall_summary"`
	MatchedStrengths      []string                     `json:"matched_strengths"`
	GapsAndRisks          []string                     `json:"gaps_and_risks"`
	CompensationAlignment string                       `json:"compensation_alignment"`
	RecommendedStoryIDs   []string                     `json:"recommended_story_ids"`
	RecommendedStories    []RecommendedStorySummaryDTO `json:"recommended_stories"`
	ModelUsed             string                       `json:"model_used"`
	UpdatedAt             string                       `json:"updated_at"`
}

type FitAssessmentUpsertRequest struct {
	OverallScore          int      `json:"overall_score"`
	OverallSummary        string   `json:"overall_summary"`
	MatchedStrengths      []string `json:"matched_strengths"`
	GapsAndRisks          []string `json:"gaps_and_risks"`
	CompensationAlignment string   `json:"compensation_alignment"`
	RecommendedStoryIDs   []string `json:"recommended_story_ids"`
	ModelUsed             string   `json:"model_used"`
}

type FitAssessmentGenerateRequest struct {
	ModelOverride string `json:"model_override"`
}

func (r FitAssessmentUpsertRequest) toDomainInput() domain.FitAssessmentInput {
	return domain.FitAssessmentInput{
		OverallScore:          r.OverallScore,
		OverallSummary:        r.OverallSummary,
		MatchedStrengths:      r.MatchedStrengths,
		GapsAndRisks:          r.GapsAndRisks,
		CompensationAlignment: r.CompensationAlignment,
		RecommendedStoryIDs:   r.RecommendedStoryIDs,
		ModelUsed:             r.ModelUsed,
	}
}

func toFitAssessmentDTO(view domain.FitAssessmentView) FitAssessmentDTO {
	fa := view.FitAssessment
	strengths := fa.MatchedStrengths
	if strengths == nil {
		strengths = []string{}
	}
	gaps := fa.GapsAndRisks
	if gaps == nil {
		gaps = []string{}
	}
	storyIDs := fa.RecommendedStoryIDs
	if storyIDs == nil {
		storyIDs = []string{}
	}
	stories := make([]RecommendedStorySummaryDTO, 0, len(view.RecommendedStories))
	for _, st := range view.RecommendedStories {
		comp := st.CompetencyTags
		if comp == nil {
			comp = []string{}
		}
		skills := st.SkillTags
		if skills == nil {
			skills = []string{}
		}
		stories = append(stories, RecommendedStorySummaryDTO{
			ID:             st.ID,
			Title:          st.Title,
			Hook:           st.Hook,
			CompetencyTags: comp,
			SkillTags:      skills,
		})
	}
	updatedAt := ""
	if !fa.UpdatedAt.IsZero() {
		updatedAt = fa.UpdatedAt.UTC().Format(time.RFC3339)
	}
	return FitAssessmentDTO{
		OpportunityID:         fa.OpportunityID,
		OverallScore:          fa.OverallScore,
		OverallSummary:        fa.OverallSummary,
		MatchedStrengths:      strengths,
		GapsAndRisks:          gaps,
		CompensationAlignment: fa.CompensationAlignment,
		RecommendedStoryIDs:   storyIDs,
		RecommendedStories:    stories,
		ModelUsed:             fa.ModelUsed,
		UpdatedAt:             updatedAt,
	}
}
