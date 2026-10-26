package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"job-search-cockpit/internal/domain"
)

func (s *Store) SaveApplicationArtifact(ctx context.Context, art domain.ApplicationArtifact) error {
	isSubmittedInt := 0
	if art.IsSubmitted {
		isSubmittedInt = 1
	}

	_, err := s.exec(ctx).ExecContext(ctx, `
		INSERT INTO application_artifacts (
			id, opportunity_id, kind, title, content_markdown,
			file_path, file_name, mime_type, file_size_bytes,
			is_submitted, model_used, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			opportunity_id = excluded.opportunity_id,
			kind = excluded.kind,
			title = excluded.title,
			content_markdown = excluded.content_markdown,
			file_path = excluded.file_path,
			file_name = excluded.file_name,
			mime_type = excluded.mime_type,
			file_size_bytes = excluded.file_size_bytes,
			is_submitted = excluded.is_submitted,
			model_used = excluded.model_used,
			updated_at = excluded.updated_at
	`,
		art.ID,
		art.OpportunityID,
		string(art.Kind),
		art.Title,
		art.ContentMarkdown,
		art.FilePath,
		art.FileName,
		art.MimeType,
		art.FileSizeBytes,
		isSubmittedInt,
		art.ModelUsed,
		art.CreatedAt.UTC().Format(time.RFC3339Nano),
		art.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetApplicationArtifact(ctx context.Context, id string) (domain.ApplicationArtifact, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			id, opportunity_id, kind, title, content_markdown,
			file_path, file_name, mime_type, file_size_bytes,
			is_submitted, model_used, created_at, updated_at
		FROM application_artifacts
		WHERE id = ?
	`, id)
	art, err := scanApplicationArtifact(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ApplicationArtifact{}, fmt.Errorf("%w: application artifact %s", domain.ErrNotFound, id)
	}
	return art, err
}

func (s *Store) ListApplicationArtifacts(ctx context.Context, opportunityID string) ([]domain.ApplicationArtifact, error) {
	rows, err := s.exec(ctx).QueryContext(ctx, `
		SELECT
			id, opportunity_id, kind, title, content_markdown,
			file_path, file_name, mime_type, file_size_bytes,
			is_submitted, model_used, created_at, updated_at
		FROM application_artifacts
		WHERE opportunity_id = ?
		ORDER BY is_submitted DESC, updated_at DESC, created_at DESC
	`, opportunityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.ApplicationArtifact{}
	for rows.Next() {
		art, err := scanApplicationArtifact(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, art)
	}
	return out, rows.Err()
}

func (s *Store) DeleteApplicationArtifact(ctx context.Context, id string) error {
	res, err := s.exec(ctx).ExecContext(ctx, `DELETE FROM application_artifacts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%w: application artifact %s", domain.ErrNotFound, id)
	}
	return nil
}

func scanApplicationArtifact(scan func(dest ...any) error) (domain.ApplicationArtifact, error) {
	var art domain.ApplicationArtifact
	var kindStr, createdStr, updatedStr string
	var isSubmittedInt int
	err := scan(
		&art.ID,
		&art.OpportunityID,
		&kindStr,
		&art.Title,
		&art.ContentMarkdown,
		&art.FilePath,
		&art.FileName,
		&art.MimeType,
		&art.FileSizeBytes,
		&isSubmittedInt,
		&art.ModelUsed,
		&createdStr,
		&updatedStr,
	)
	if err != nil {
		return domain.ApplicationArtifact{}, err
	}
	art.Kind = domain.ArtifactKind(kindStr)
	art.IsSubmitted = isSubmittedInt != 0
	art.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	art.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return art, nil
}
