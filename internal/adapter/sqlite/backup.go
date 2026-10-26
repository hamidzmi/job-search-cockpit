package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func (s *Store) ExportJSONSnapshot(ctx context.Context) ([]byte, error) {
	snapshot := make(map[string]any)
	snapshot["exported_at"] = time.Now().UTC().Format(time.RFC3339Nano)

	// List of tables to dump
	tables := []string{
		"companies",
		"opportunities",
		"stage_transitions",
		"candidate_profile",
		"stories",
		"contacts",
		"opportunity_contacts",
		"interactions",
		"action_items",
		"interview_rounds",
		"debriefs",
		"fit_assessments",
		"application_artifacts",
		"offers",
		"ai_settings",
	}

	for _, tbl := range tables {
		rows, err := s.exec(ctx).QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s", tbl))
		if err != nil {
			// If table doesn't exist yet, continue
			continue
		}
		cols, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			return nil, err
		}

		var tableRows []map[string]any
		for rows.Next() {
			columnPointers := make([]any, len(cols))
			columnValues := make([]any, len(cols))
			for i := range columnValues {
				columnPointers[i] = &columnValues[i]
			}

			if err := rows.Scan(columnPointers...); err != nil {
				_ = rows.Close()
				return nil, err
			}

			rowMap := make(map[string]any)
			for i, colName := range cols {
				val := columnValues[i]
				if b, ok := val.([]byte); ok {
					rowMap[colName] = string(b)
				} else {
					rowMap[colName] = val
				}
			}
			tableRows = append(tableRows, rowMap)
		}
		_ = rows.Close()
		if tableRows == nil {
			tableRows = []map[string]any{}
		}
		snapshot[tbl] = tableRows
	}

	return json.MarshalIndent(snapshot, "", "  ")
}

func (s *Store) ExportSQLiteSnapshot(ctx context.Context) ([]byte, error) {
	tempFile, err := os.CreateTemp("", "cockpit-backup-*.db")
	if err != nil {
		return nil, fmt.Errorf("create temp backup file: %w", err)
	}
	tempPath := tempFile.Name()
	_ = tempFile.Close()
	_ = os.Remove(tempPath) // SQLite VACUUM INTO requires target file to not exist

	defer func() {
		_ = os.Remove(tempPath)
	}()

	query := fmt.Sprintf("VACUUM INTO '%s'", strings.ReplaceAll(tempPath, "'", "''"))
	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return nil, fmt.Errorf("vacuum into: %w", err)
	}

	data, err := os.ReadFile(tempPath)
	if err != nil {
		return nil, fmt.Errorf("read backup db: %w", err)
	}

	return data, nil
}
