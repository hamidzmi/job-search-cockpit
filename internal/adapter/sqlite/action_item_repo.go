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

func (s *Store) SaveActionItem(ctx context.Context, item domain.ActionItem) error {
	_, err := s.exec(ctx).ExecContext(ctx, `
		INSERT INTO action_items (
			id, title, opportunity_id, contact_id, due_at, status, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			opportunity_id = excluded.opportunity_id,
			contact_id = excluded.contact_id,
			due_at = excluded.due_at,
			status = excluded.status,
			notes = excluded.notes,
			updated_at = excluded.updated_at
	`,
		item.ID,
		item.Title,
		nullableID(item.OpportunityID),
		nullableID(item.ContactID),
		item.DueAt.UTC().Format(time.RFC3339Nano),
		string(item.Status),
		item.Notes,
		item.CreatedAt.UTC().Format(time.RFC3339Nano),
		item.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetActionItem(ctx context.Context, id string) (domain.ActionItem, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			id, title, COALESCE(opportunity_id, ''), COALESCE(contact_id, ''),
			due_at, status, notes, created_at, updated_at
		FROM action_items
		WHERE id = ?
	`, id)
	var item domain.ActionItem
	var dueStr, statusStr, createdStr, updatedStr string
	err := row.Scan(
		&item.ID, &item.Title, &item.OpportunityID, &item.ContactID,
		&dueStr, &statusStr, &item.Notes, &createdStr, &updatedStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ActionItem{}, fmt.Errorf("%w: action item %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return domain.ActionItem{}, err
	}
	item.DueAt, _ = time.Parse(time.RFC3339Nano, dueStr)
	item.Status = domain.ActionItemStatus(statusStr)
	item.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return item, nil
}

func (s *Store) GetActionItemView(ctx context.Context, id string) (domain.ActionItemView, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			a.id, a.title,
			COALESCE(a.opportunity_id, ''), COALESCE(o.role_title, ''),
			COALESCE(a.contact_id, ''), COALESCE(ct.name, ''),
			COALESCE(co.name, ''),
			a.due_at, a.status, a.notes, a.created_at, a.updated_at
		FROM action_items a
		LEFT JOIN opportunities o ON o.id = a.opportunity_id
		LEFT JOIN contacts ct ON ct.id = a.contact_id
		LEFT JOIN companies co ON co.id = COALESCE(o.company_id, ct.company_id)
		WHERE a.id = ?
	`, id)
	view, err := scanActionItemView(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ActionItemView{}, fmt.Errorf("%w: action item %s", domain.ErrNotFound, id)
	}
	return view, err
}

func (s *Store) ListActionItems(ctx context.Context, filter domain.ActionItemFilter) ([]domain.ActionItemView, error) {
	var clauses []string
	var args []any

	if filter.Status != "" {
		clauses = append(clauses, "a.status = ?")
		args = append(args, string(filter.Status))
	}
	if oppID := strings.TrimSpace(filter.OpportunityID); oppID != "" {
		clauses = append(clauses, "a.opportunity_id = ?")
		args = append(args, oppID)
	}
	if cntID := strings.TrimSpace(filter.ContactID); cntID != "" {
		clauses = append(clauses, "a.contact_id = ?")
		args = append(args, cntID)
	}

	query := `
		SELECT
			a.id, a.title,
			COALESCE(a.opportunity_id, ''), COALESCE(o.role_title, ''),
			COALESCE(a.contact_id, ''), COALESCE(ct.name, ''),
			COALESCE(co.name, ''),
			a.due_at, a.status, a.notes, a.created_at, a.updated_at
		FROM action_items a
		LEFT JOIN opportunities o ON o.id = a.opportunity_id
		LEFT JOIN contacts ct ON ct.id = a.contact_id
		LEFT JOIN companies co ON co.id = COALESCE(o.company_id, ct.company_id)
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY a.due_at ASC, a.created_at ASC"

	rows, err := s.exec(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.ActionItemView
	for rows.Next() {
		view, err := scanActionItemView(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, view)
	}
	if result == nil {
		result = []domain.ActionItemView{}
	}
	return result, rows.Err()
}

func scanActionItemView(scan func(dest ...any) error) (domain.ActionItemView, error) {
	var a domain.ActionItem
	var oppRoleTitle, contactName, companyName, dueStr, statusStr, createdStr, updatedStr string
	if err := scan(
		&a.ID, &a.Title,
		&a.OpportunityID, &oppRoleTitle,
		&a.ContactID, &contactName,
		&companyName,
		&dueStr, &statusStr, &a.Notes, &createdStr, &updatedStr,
	); err != nil {
		return domain.ActionItemView{}, err
	}
	a.DueAt, _ = time.Parse(time.RFC3339Nano, dueStr)
	a.Status = domain.ActionItemStatus(statusStr)
	a.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	a.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return domain.ActionItemView{
		ActionItem:           a,
		ContactName:          contactName,
		OpportunityRoleTitle: oppRoleTitle,
		CompanyName:          companyName,
	}, nil
}
