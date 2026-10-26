package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"job-search-cockpit/internal/domain"
)

func (s *Store) SaveCompany(ctx context.Context, c domain.Company) error {
	_, err := s.exec(ctx).ExecContext(ctx, `
		INSERT INTO companies (id, name, domain, headquarters, remote_policy, size_stage, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			domain = excluded.domain,
			headquarters = excluded.headquarters,
			remote_policy = excluded.remote_policy,
			size_stage = excluded.size_stage,
			notes = excluded.notes,
			updated_at = excluded.updated_at
	`, c.ID, c.Name, c.Domain, c.Headquarters, c.RemotePolicy, c.SizeStage, c.Notes,
		c.CreatedAt.UTC().Format(time.RFC3339Nano),
		c.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetCompany(ctx context.Context, id string) (domain.Company, error) {
	var c domain.Company
	var createdStr, updatedStr string
	err := s.exec(ctx).QueryRowContext(ctx, `
		SELECT id, name, domain, headquarters, remote_policy, size_stage, notes, created_at, updated_at
		FROM companies
		WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.Domain, &c.Headquarters, &c.RemotePolicy, &c.SizeStage, &c.Notes, &createdStr, &updatedStr)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Company{}, fmt.Errorf("%w: company %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return domain.Company{}, err
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return c, nil
}

func (s *Store) ListCompanies(ctx context.Context) ([]domain.Company, error) {
	rows, err := s.exec(ctx).QueryContext(ctx, `
		SELECT id, name, domain, headquarters, remote_policy, size_stage, notes, created_at, updated_at
		FROM companies
		ORDER BY name COLLATE NOCASE ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Company
	for rows.Next() {
		var c domain.Company
		var createdStr, updatedStr string
		if err := rows.Scan(&c.ID, &c.Name, &c.Domain, &c.Headquarters, &c.RemotePolicy, &c.SizeStage, &c.Notes, &createdStr, &updatedStr); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
		c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
		result = append(result, c)
	}
	if result == nil {
		result = []domain.Company{}
	}
	return result, rows.Err()
}
