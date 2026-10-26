package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"job-search-cockpit/internal/domain"
)

func (s *Store) SaveOpportunity(ctx context.Context, opp domain.Opportunity) error {
	reqsJSON, err := json.Marshal(nonNilSlice(opp.JobPosting.Requirements))
	if err != nil {
		return err
	}
	tagsJSON, err := json.Marshal(nonNilSlice(opp.JobPosting.TechStackTags))
	if err != nil {
		return err
	}

	_, err = s.exec(ctx).ExecContext(ctx, `
		INSERT INTO opportunities (
			id, company_id, role_title, seniority, location, remote_policy, priority,
			stage, outcome, notes, source_url, raw_text, requirements_json, tech_stack_tags_json,
			stated_min_comp, stated_max_comp, currency, last_activity_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			company_id = excluded.company_id,
			role_title = excluded.role_title,
			seniority = excluded.seniority,
			location = excluded.location,
			remote_policy = excluded.remote_policy,
			priority = excluded.priority,
			stage = excluded.stage,
			outcome = excluded.outcome,
			notes = excluded.notes,
			source_url = excluded.source_url,
			raw_text = excluded.raw_text,
			requirements_json = excluded.requirements_json,
			tech_stack_tags_json = excluded.tech_stack_tags_json,
			stated_min_comp = excluded.stated_min_comp,
			stated_max_comp = excluded.stated_max_comp,
			currency = excluded.currency,
			last_activity_at = excluded.last_activity_at,
			updated_at = excluded.updated_at
	`,
		opp.ID, opp.CompanyID, opp.RoleTitle, opp.Seniority, opp.Location, opp.RemotePolicy, opp.Priority,
		string(opp.Stage), string(opp.Outcome), opp.Notes,
		opp.JobPosting.SourceURL, opp.JobPosting.RawText, string(reqsJSON), string(tagsJSON),
		opp.JobPosting.StatedMinComp, opp.JobPosting.StatedMaxComp, opp.JobPosting.Currency,
		opp.LastActivityAt.UTC().Format(time.RFC3339Nano),
		opp.CreatedAt.UTC().Format(time.RFC3339Nano),
		opp.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetOpportunity(ctx context.Context, id string) (domain.Opportunity, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			o.id, o.company_id, c.name, o.role_title, o.seniority, o.location, o.remote_policy, o.priority,
			o.stage, o.outcome, o.notes, o.source_url, o.raw_text, o.requirements_json, o.tech_stack_tags_json,
			o.stated_min_comp, o.stated_max_comp, o.currency, o.last_activity_at, o.created_at, o.updated_at
		FROM opportunities o
		JOIN companies c ON c.id = o.company_id
		WHERE o.id = ?
	`, id)
	opp, err := scanOpportunity(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Opportunity{}, fmt.Errorf("%w: opportunity %s", domain.ErrNotFound, id)
	}
	return opp, err
}

func (s *Store) ListOpportunities(ctx context.Context, filter domain.OpportunityFilter) ([]domain.Opportunity, error) {
	var clauses []string
	var args []any

	if filter.Stage != "" {
		clauses = append(clauses, "o.stage = ?")
		args = append(args, string(filter.Stage))
	}
	if filter.Outcome != "" {
		clauses = append(clauses, "o.outcome = ?")
		args = append(args, string(filter.Outcome))
	}
	if strings.TrimSpace(filter.Priority) != "" {
		clauses = append(clauses, "LOWER(o.priority) = LOWER(?)")
		args = append(args, strings.TrimSpace(filter.Priority))
	}
	if strings.TrimSpace(filter.RemotePolicy) != "" {
		clauses = append(clauses, "LOWER(o.remote_policy) = LOWER(?)")
		args = append(args, strings.TrimSpace(filter.RemotePolicy))
	}
	if strings.TrimSpace(filter.CompanyID) != "" {
		clauses = append(clauses, "o.company_id = ?")
		args = append(args, strings.TrimSpace(filter.CompanyID))
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		clauses = append(clauses, "(LOWER(o.role_title) LIKE ? OR LOWER(c.name) LIKE ? OR LOWER(o.notes) LIKE ? OR LOWER(o.tech_stack_tags_json) LIKE ?)")
		args = append(args, like, like, like, like)
	}

	query := `
		SELECT
			o.id, o.company_id, c.name, o.role_title, o.seniority, o.location, o.remote_policy, o.priority,
			o.stage, o.outcome, o.notes, o.source_url, o.raw_text, o.requirements_json, o.tech_stack_tags_json,
			o.stated_min_comp, o.stated_max_comp, o.currency, o.last_activity_at, o.created_at, o.updated_at
		FROM opportunities o
		JOIN companies c ON c.id = o.company_id
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY o.last_activity_at DESC, o.created_at DESC"

	rows, err := s.exec(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Opportunity
	for rows.Next() {
		opp, err := scanOpportunity(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, opp)
	}
	if result == nil {
		result = []domain.Opportunity{}
	}
	return result, rows.Err()
}

func scanOpportunity(scan func(dest ...any) error) (domain.Opportunity, error) {
	var opp domain.Opportunity
	var stageStr, outcomeStr, reqsJSON, tagsJSON, lastActStr, createdStr, updatedStr string
	err := scan(
		&opp.ID, &opp.CompanyID, &opp.CompanyName, &opp.RoleTitle, &opp.Seniority, &opp.Location, &opp.RemotePolicy, &opp.Priority,
		&stageStr, &outcomeStr, &opp.Notes,
		&opp.JobPosting.SourceURL, &opp.JobPosting.RawText, &reqsJSON, &tagsJSON,
		&opp.JobPosting.StatedMinComp, &opp.JobPosting.StatedMaxComp, &opp.JobPosting.Currency,
		&lastActStr, &createdStr, &updatedStr,
	)
	if err != nil {
		return domain.Opportunity{}, err
	}
	opp.Stage = domain.Stage(stageStr)
	opp.Outcome = domain.Outcome(outcomeStr)
	_ = json.Unmarshal([]byte(reqsJSON), &opp.JobPosting.Requirements)
	_ = json.Unmarshal([]byte(tagsJSON), &opp.JobPosting.TechStackTags)
	if opp.JobPosting.Requirements == nil {
		opp.JobPosting.Requirements = []string{}
	}
	if opp.JobPosting.TechStackTags == nil {
		opp.JobPosting.TechStackTags = []string{}
	}
	opp.LastActivityAt, _ = time.Parse(time.RFC3339Nano, lastActStr)
	opp.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	opp.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return opp, nil
}

func nonNilSlice(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}
