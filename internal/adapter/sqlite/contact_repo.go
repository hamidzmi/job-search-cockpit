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

func nullableID(id string) any {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func (s *Store) SaveContact(ctx context.Context, contact domain.Contact) error {
	_, err := s.exec(ctx).ExecContext(ctx, `
		INSERT INTO contacts (
			id, company_id, name, title, email, linkedin_url, warmth, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			company_id = excluded.company_id,
			name = excluded.name,
			title = excluded.title,
			email = excluded.email,
			linkedin_url = excluded.linkedin_url,
			warmth = excluded.warmth,
			notes = excluded.notes,
			updated_at = excluded.updated_at
	`,
		contact.ID,
		nullableID(contact.CompanyID),
		contact.Name,
		contact.Title,
		contact.Email,
		contact.LinkedInURL,
		string(contact.Warmth),
		contact.Notes,
		contact.CreatedAt.UTC().Format(time.RFC3339Nano),
		contact.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetContact(ctx context.Context, id string) (domain.Contact, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT id, COALESCE(company_id, ''), name, title, email, linkedin_url, warmth, notes, created_at, updated_at
		FROM contacts
		WHERE id = ?
	`, id)
	contact, err := scanContactOnly(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Contact{}, fmt.Errorf("%w: contact %s", domain.ErrNotFound, id)
	}
	return contact, err
}

func (s *Store) GetContactView(ctx context.Context, id string) (domain.ContactView, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			ct.id, COALESCE(ct.company_id, ''), COALESCE(co.name, ''),
			ct.name, ct.title, ct.email, ct.linkedin_url, ct.warmth, ct.notes, ct.created_at, ct.updated_at
		FROM contacts ct
		LEFT JOIN companies co ON co.id = ct.company_id
		WHERE ct.id = ?
	`, id)
	view, err := scanContactViewBase(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ContactView{}, fmt.Errorf("%w: contact %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return domain.ContactView{}, err
	}
	links, err := s.listContactOpportunityLinks(ctx, id)
	if err != nil {
		return domain.ContactView{}, err
	}
	view.LinkedOpportunities = links
	return view, nil
}

func (s *Store) ListContacts(ctx context.Context, filter domain.ContactFilter) ([]domain.ContactView, error) {
	var clauses []string
	var args []any

	if companyID := strings.TrimSpace(filter.CompanyID); companyID != "" {
		clauses = append(clauses, "ct.company_id = ?")
		args = append(args, companyID)
	}
	if filter.Warmth != "" {
		clauses = append(clauses, "ct.warmth = ?")
		args = append(args, string(filter.Warmth))
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		clauses = append(clauses, "(LOWER(ct.name) LIKE ? OR LOWER(ct.title) LIKE ? OR LOWER(ct.email) LIKE ? OR LOWER(ct.notes) LIKE ? OR LOWER(COALESCE(co.name, '')) LIKE ?)")
		args = append(args, like, like, like, like, like)
	}

	query := `
		SELECT
			ct.id, COALESCE(ct.company_id, ''), COALESCE(co.name, ''),
			ct.name, ct.title, ct.email, ct.linkedin_url, ct.warmth, ct.notes, ct.created_at, ct.updated_at
		FROM contacts ct
		LEFT JOIN companies co ON co.id = ct.company_id
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY ct.name ASC, ct.created_at DESC"

	rows, err := s.exec(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []domain.ContactView
	for rows.Next() {
		v, err := scanContactViewBase(rows.Scan)
		if err != nil {
			return nil, err
		}
		views = append(views, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if views == nil {
		return []domain.ContactView{}, nil
	}

	for i := range views {
		links, err := s.listContactOpportunityLinks(ctx, views[i].Contact.ID)
		if err != nil {
			return nil, err
		}
		views[i].LinkedOpportunities = links
	}
	return views, nil
}

func (s *Store) SaveOpportunityContact(ctx context.Context, link domain.OpportunityContact) error {
	_, err := s.exec(ctx).ExecContext(ctx, `
		INSERT INTO opportunity_contacts (
			opportunity_id, contact_id, role, notes, linked_at
		) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(opportunity_id, contact_id) DO UPDATE SET
			role = excluded.role,
			notes = excluded.notes,
			linked_at = excluded.linked_at
	`,
		link.OpportunityID,
		link.ContactID,
		string(link.Role),
		link.Notes,
		link.LinkedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) DeleteOpportunityContact(ctx context.Context, opportunityID, contactID string) error {
	res, err := s.exec(ctx).ExecContext(ctx, `
		DELETE FROM opportunity_contacts
		WHERE opportunity_id = ? AND contact_id = ?
	`, opportunityID, contactID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%w: opportunity contact link (%s, %s)", domain.ErrNotFound, opportunityID, contactID)
	}
	return nil
}

func (s *Store) ListOpportunityContacts(ctx context.Context, opportunityID string) ([]domain.OpportunityContactView, error) {
	rows, err := s.exec(ctx).QueryContext(ctx, `
		SELECT
			oc.opportunity_id, oc.contact_id, oc.role, oc.notes, oc.linked_at,
			ct.id, COALESCE(ct.company_id, ''), COALESCE(co.name, ''),
			ct.name, ct.title, ct.email, ct.linkedin_url, ct.warmth, ct.notes, ct.created_at, ct.updated_at
		FROM opportunity_contacts oc
		JOIN contacts ct ON ct.id = oc.contact_id
		LEFT JOIN companies co ON co.id = ct.company_id
		WHERE oc.opportunity_id = ?
		ORDER BY oc.linked_at ASC, ct.name ASC
	`, opportunityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.OpportunityContactView
	for rows.Next() {
		var oc domain.OpportunityContact
		var ct domain.Contact
		var companyName, roleStr, warmthStr, linkedAtStr, createdStr, updatedStr string
		if err := rows.Scan(
			&oc.OpportunityID, &oc.ContactID, &roleStr, &oc.Notes, &linkedAtStr,
			&ct.ID, &ct.CompanyID, &companyName,
			&ct.Name, &ct.Title, &ct.Email, &ct.LinkedInURL, &warmthStr, &ct.Notes, &createdStr, &updatedStr,
		); err != nil {
			return nil, err
		}
		oc.Role = domain.StakeholderRole(roleStr)
		oc.LinkedAt, _ = time.Parse(time.RFC3339Nano, linkedAtStr)
		ct.Warmth = domain.ContactWarmth(warmthStr)
		ct.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
		ct.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
		result = append(result, domain.OpportunityContactView{
			OpportunityContact: oc,
			Contact:            ct,
			CompanyName:        companyName,
		})
	}
	if result == nil {
		result = []domain.OpportunityContactView{}
	}
	return result, rows.Err()
}

func (s *Store) listContactOpportunityLinks(ctx context.Context, contactID string) ([]domain.ContactOpportunityLinkView, error) {
	rows, err := s.exec(ctx).QueryContext(ctx, `
		SELECT
			oc.opportunity_id, o.company_id, c.name, o.role_title, o.stage, o.outcome,
			oc.role, oc.notes, oc.linked_at
		FROM opportunity_contacts oc
		JOIN opportunities o ON o.id = oc.opportunity_id
		JOIN companies c ON c.id = o.company_id
		WHERE oc.contact_id = ?
		ORDER BY oc.linked_at DESC
	`, contactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.ContactOpportunityLinkView
	for rows.Next() {
		var link domain.ContactOpportunityLinkView
		var stageStr, outcomeStr, roleStr, linkedAtStr string
		if err := rows.Scan(
			&link.OpportunityID, &link.CompanyID, &link.CompanyName, &link.RoleTitle,
			&stageStr, &outcomeStr, &roleStr, &link.Notes, &linkedAtStr,
		); err != nil {
			return nil, err
		}
		link.Stage = domain.Stage(stageStr)
		link.Outcome = domain.Outcome(outcomeStr)
		link.Role = domain.StakeholderRole(roleStr)
		link.LinkedAt, _ = time.Parse(time.RFC3339Nano, linkedAtStr)
		links = append(links, link)
	}
	if links == nil {
		links = []domain.ContactOpportunityLinkView{}
	}
	return links, rows.Err()
}

func scanContactOnly(scan func(dest ...any) error) (domain.Contact, error) {
	var c domain.Contact
	var warmthStr, createdStr, updatedStr string
	if err := scan(
		&c.ID, &c.CompanyID, &c.Name, &c.Title, &c.Email, &c.LinkedInURL, &warmthStr, &c.Notes, &createdStr, &updatedStr,
	); err != nil {
		return domain.Contact{}, err
	}
	c.Warmth = domain.ContactWarmth(warmthStr)
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return c, nil
}

func scanContactViewBase(scan func(dest ...any) error) (domain.ContactView, error) {
	var c domain.Contact
	var companyName, warmthStr, createdStr, updatedStr string
	if err := scan(
		&c.ID, &c.CompanyID, &companyName, &c.Name, &c.Title, &c.Email, &c.LinkedInURL, &warmthStr, &c.Notes, &createdStr, &updatedStr,
	); err != nil {
		return domain.ContactView{}, err
	}
	c.Warmth = domain.ContactWarmth(warmthStr)
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return domain.ContactView{
		Contact:             c,
		CompanyName:         companyName,
		LinkedOpportunities: []domain.ContactOpportunityLinkView{},
	}, nil
}
