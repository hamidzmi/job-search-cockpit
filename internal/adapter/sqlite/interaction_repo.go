package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"job-search-cockpit/internal/domain"
)

func (s *Store) SaveInteraction(ctx context.Context, interaction domain.Interaction) error {
	_, err := s.exec(ctx).ExecContext(ctx, `
		INSERT INTO interactions (
			id, contact_id, opportunity_id, kind, direction, occurred_at, summary, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			contact_id = excluded.contact_id,
			opportunity_id = excluded.opportunity_id,
			kind = excluded.kind,
			direction = excluded.direction,
			occurred_at = excluded.occurred_at,
			summary = excluded.summary
	`,
		interaction.ID,
		interaction.ContactID,
		nullableID(interaction.OpportunityID),
		string(interaction.Kind),
		string(interaction.Direction),
		interaction.OccurredAt.UTC().Format(time.RFC3339Nano),
		interaction.Summary,
		interaction.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetInteractionView(ctx context.Context, id string) (domain.InteractionView, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			i.id, i.contact_id, ct.name,
			COALESCE(i.opportunity_id, ''), COALESCE(o.role_title, ''), COALESCE(co.name, ''),
			i.kind, i.direction, i.occurred_at, i.summary, i.created_at
		FROM interactions i
		JOIN contacts ct ON ct.id = i.contact_id
		LEFT JOIN opportunities o ON o.id = i.opportunity_id
		LEFT JOIN companies co ON co.id = COALESCE(o.company_id, ct.company_id)
		WHERE i.id = ?
	`, id)
	view, err := scanInteractionView(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.InteractionView{}, fmt.Errorf("%w: interaction %s", domain.ErrNotFound, id)
	}
	return view, err
}

func (s *Store) ListInteractions(ctx context.Context, filter domain.InteractionFilter) ([]domain.InteractionView, error) {
	var clauses []string
	var args []any

	if contactID := strings.TrimSpace(filter.ContactID); contactID != "" {
		clauses = append(clauses, "i.contact_id = ?")
		args = append(args, contactID)
	}
	if oppID := strings.TrimSpace(filter.OpportunityID); oppID != "" {
		clauses = append(clauses, "i.opportunity_id = ?")
		args = append(args, oppID)
	}

	query := `
		SELECT
			i.id, i.contact_id, ct.name,
			COALESCE(i.opportunity_id, ''), COALESCE(o.role_title, ''), COALESCE(co.name, ''),
			i.kind, i.direction, i.occurred_at, i.summary, i.created_at
		FROM interactions i
		JOIN contacts ct ON ct.id = i.contact_id
		LEFT JOIN opportunities o ON o.id = i.opportunity_id
		LEFT JOIN companies co ON co.id = COALESCE(o.company_id, ct.company_id)
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY i.occurred_at DESC, i.created_at DESC"

	rows, err := s.exec(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.InteractionView
	for rows.Next() {
		view, err := scanInteractionView(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, view)
	}
	if result == nil {
		result = []domain.InteractionView{}
	}
	return result, rows.Err()
}

func scanInteractionView(scan func(dest ...any) error) (domain.InteractionView, error) {
	var i domain.Interaction
	var contactName, oppRoleTitle, companyName, kindStr, dirStr, occurredStr, createdStr string
	if err := scan(
		&i.ID, &i.ContactID, &contactName,
		&i.OpportunityID, &oppRoleTitle, &companyName,
		&kindStr, &dirStr, &occurredStr, &i.Summary, &createdStr,
	); err != nil {
		return domain.InteractionView{}, err
	}
	i.Kind = domain.InteractionKind(kindStr)
	i.Direction = domain.InteractionDirection(dirStr)
	i.OccurredAt, _ = time.Parse(time.RFC3339Nano, occurredStr)
	i.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	return domain.InteractionView{
		Interaction:          i,
		ContactName:          contactName,
		OpportunityRoleTitle: oppRoleTitle,
		CompanyName:          companyName,
	}, nil
}
