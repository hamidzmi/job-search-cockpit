package sqlite

import (
	"context"
	"time"

	"job-search-cockpit/internal/domain"
)

func (s *Store) AppendStageTransition(ctx context.Context, tr domain.StageTransition) error {
	_, err := s.exec(ctx).ExecContext(ctx, `
		INSERT INTO stage_transitions (
			id, opportunity_id, from_stage, to_stage, from_outcome, to_outcome, note, transitioned_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		tr.ID, tr.OpportunityID, string(tr.FromStage), string(tr.ToStage),
		string(tr.FromOutcome), string(tr.ToOutcome), tr.Note,
		tr.TransitionedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) ListStageTransitions(ctx context.Context, opportunityID string) ([]domain.StageTransition, error) {
	rows, err := s.exec(ctx).QueryContext(ctx, `
		SELECT id, opportunity_id, from_stage, to_stage, from_outcome, to_outcome, note, transitioned_at
		FROM stage_transitions
		WHERE opportunity_id = ?
		ORDER BY transitioned_at ASC, rowid ASC
	`, opportunityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.StageTransition
	for rows.Next() {
		var tr domain.StageTransition
		var fromStage, toStage, fromOutcome, toOutcome, tsStr string
		if err := rows.Scan(&tr.ID, &tr.OpportunityID, &fromStage, &toStage, &fromOutcome, &toOutcome, &tr.Note, &tsStr); err != nil {
			return nil, err
		}
		tr.FromStage = domain.Stage(fromStage)
		tr.ToStage = domain.Stage(toStage)
		tr.FromOutcome = domain.Outcome(fromOutcome)
		tr.ToOutcome = domain.Outcome(toOutcome)
		tr.TransitionedAt, _ = time.Parse(time.RFC3339Nano, tsStr)
		result = append(result, tr)
	}
	if result == nil {
		result = []domain.StageTransition{}
	}
	return result, rows.Err()
}
