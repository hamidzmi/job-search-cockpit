package domain

import (
	"fmt"
	"strings"
	"time"
)

type FitAssessment struct {
	OpportunityID         string
	OverallScore          int
	OverallSummary        string
	MatchedStrengths      []string
	GapsAndRisks          []string
	CompensationAlignment string
	RecommendedStoryIDs   []string
	ModelUsed             string
	UpdatedAt             time.Time
}

type FitAssessmentInput struct {
	OverallScore          int
	OverallSummary        string
	MatchedStrengths      []string
	GapsAndRisks          []string
	CompensationAlignment string
	RecommendedStoryIDs   []string
	ModelUsed             string
}

type RecommendedStorySummary struct {
	ID             string
	Title          string
	Hook           string
	CompetencyTags []string
	SkillTags      []string
}

type FitAssessmentView struct {
	FitAssessment      FitAssessment
	RecommendedStories []RecommendedStorySummary
}

type AssistantAssessFitInput struct {
	Opportunity   OpportunityView
	Profile       CandidateProfile
	Stories       []Story
	ModelOverride string
}

type AssistantFitAssessmentDraft struct {
	OverallScore          int
	OverallSummary        string
	MatchedStrengths      []string
	GapsAndRisks          []string
	CompensationAlignment string
	RecommendedStoryIDs   []string
	ModelUsed             string
}

func NewFitAssessment(opportunityID string, in FitAssessmentInput, now time.Time) (FitAssessment, error) {
	oppID := strings.TrimSpace(opportunityID)
	if oppID == "" {
		return FitAssessment{}, fmt.Errorf("%w: opportunity id is required", ErrValidation)
	}
	fa := FitAssessment{
		OpportunityID: oppID,
	}
	if err := fa.Update(in, now); err != nil {
		return FitAssessment{}, err
	}
	return fa, nil
}

func (f *FitAssessment) Update(in FitAssessmentInput, now time.Time) error {
	if in.OverallScore < 0 || in.OverallScore > 100 {
		return fmt.Errorf("%w: fit assessment score must be between 0 and 100", ErrValidation)
	}
	f.OverallScore = in.OverallScore
	f.OverallSummary = strings.TrimSpace(in.OverallSummary)
	f.MatchedStrengths = normalizeUniqueStrings(in.MatchedStrengths)
	f.GapsAndRisks = normalizeUniqueStrings(in.GapsAndRisks)
	f.CompensationAlignment = strings.TrimSpace(in.CompensationAlignment)
	f.RecommendedStoryIDs = normalizeUniqueStrings(in.RecommendedStoryIDs)
	f.ModelUsed = strings.TrimSpace(in.ModelUsed)
	f.UpdatedAt = now.UTC()
	return nil
}

func EvaluateCompensationAlignment(jp JobPosting, profile CandidateProfile) string {
	currency := strings.ToUpper(strings.TrimSpace(profile.Currency))
	if currency == "" {
		currency = strings.ToUpper(strings.TrimSpace(jp.Currency))
	}
	if currency == "" {
		currency = "USD"
	}

	if profile.CompensationFloor <= 0 {
		if jp.StatedMinComp > 0 || jp.StatedMaxComp > 0 {
			return fmt.Sprintf(
				"No compensation floor configured on Candidate Profile (Job Posting states %d-%d %s).",
				jp.StatedMinComp,
				jp.StatedMaxComp,
				firstNonEmpty(jp.Currency, currency),
			)
		}
		return "No compensation floor configured on Candidate Profile."
	}

	if jp.StatedMinComp <= 0 && jp.StatedMaxComp <= 0 {
		return fmt.Sprintf(
			"Unstated in Job Posting (Candidate floor: %d %s).",
			profile.CompensationFloor,
			currency,
		)
	}

	effectiveMin := jp.StatedMinComp
	effectiveMax := jp.StatedMaxComp
	if effectiveMin <= 0 {
		effectiveMin = effectiveMax
	}
	if effectiveMax <= 0 {
		effectiveMax = effectiveMin
	}
	jpCurrency := firstNonEmpty(strings.ToUpper(strings.TrimSpace(jp.Currency)), currency)

	if effectiveMax < profile.CompensationFloor {
		return fmt.Sprintf(
			"Below floor: stated %d-%d %s is below candidate floor of %d %s.",
			effectiveMin,
			effectiveMax,
			jpCurrency,
			profile.CompensationFloor,
			currency,
		)
	}
	if effectiveMin >= profile.CompensationFloor {
		return fmt.Sprintf(
			"Meets floor: stated %d-%d %s meets or exceeds candidate floor of %d %s.",
			effectiveMin,
			effectiveMax,
			jpCurrency,
			profile.CompensationFloor,
			currency,
		)
	}
	return fmt.Sprintf(
		"Overlaps floor: stated %d-%d %s overlaps candidate floor of %d %s.",
		effectiveMin,
		effectiveMax,
		jpCurrency,
		profile.CompensationFloor,
		currency,
	)
}
