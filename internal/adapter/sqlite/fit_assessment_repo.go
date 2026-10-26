package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"job-search-cockpit/internal/domain"
)

func (s *Store) SaveFitAssessment(ctx context.Context, fa domain.FitAssessment) error {
	strengthsJSON, err := json.Marshal(nonNilSlice(fa.MatchedStrengths))
	if err != nil {
		return err
	}
	gapsJSON, err := json.Marshal(nonNilSlice(fa.GapsAndRisks))
	if err != nil {
		return err
	}
	storyIDsJSON, err := json.Marshal(nonNilSlice(fa.RecommendedStoryIDs))
	if err != nil {
		return err
	}

	_, err = s.exec(ctx).ExecContext(ctx, `
		INSERT INTO fit_assessments (
			opportunity_id, overall_score, overall_summary,
			matched_strengths_json, gaps_and_risks_json,
			compensation_alignment, recommended_story_ids_json,
			model_used, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(opportunity_id) DO UPDATE SET
			overall_score = excluded.overall_score,
			overall_summary = excluded.overall_summary,
			matched_strengths_json = excluded.matched_strengths_json,
			gaps_and_risks_json = excluded.gaps_and_risks_json,
			compensation_alignment = excluded.compensation_alignment,
			recommended_story_ids_json = excluded.recommended_story_ids_json,
			model_used = excluded.model_used,
			updated_at = excluded.updated_at
	`,
		fa.OpportunityID,
		fa.OverallScore,
		fa.OverallSummary,
		string(strengthsJSON),
		string(gapsJSON),
		fa.CompensationAlignment,
		string(storyIDsJSON),
		fa.ModelUsed,
		fa.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetFitAssessment(ctx context.Context, opportunityID string) (domain.FitAssessment, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			opportunity_id, overall_score, overall_summary,
			matched_strengths_json, gaps_and_risks_json,
			compensation_alignment, recommended_story_ids_json,
			model_used, updated_at
		FROM fit_assessments
		WHERE opportunity_id = ?
	`, opportunityID)

	var fa domain.FitAssessment
	var strengthsJSON, gapsJSON, storyIDsJSON, updatedStr string
	err := row.Scan(
		&fa.OpportunityID,
		&fa.OverallScore,
		&fa.OverallSummary,
		&strengthsJSON,
		&gapsJSON,
		&fa.CompensationAlignment,
		&storyIDsJSON,
		&fa.ModelUsed,
		&updatedStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.FitAssessment{}, fmt.Errorf("%w: fit assessment for opportunity %s", domain.ErrNotFound, opportunityID)
	}
	if err != nil {
		return domain.FitAssessment{}, err
	}

	_ = json.Unmarshal([]byte(strengthsJSON), &fa.MatchedStrengths)
	_ = json.Unmarshal([]byte(gapsJSON), &fa.GapsAndRisks)
	_ = json.Unmarshal([]byte(storyIDsJSON), &fa.RecommendedStoryIDs)
	if fa.MatchedStrengths == nil {
		fa.MatchedStrengths = []string{}
	}
	if fa.GapsAndRisks == nil {
		fa.GapsAndRisks = []string{}
	}
	if fa.RecommendedStoryIDs == nil {
		fa.RecommendedStoryIDs = []string{}
	}
	fa.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return fa, nil
}
