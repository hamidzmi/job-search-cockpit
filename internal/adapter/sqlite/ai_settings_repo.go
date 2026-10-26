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

func (s *Store) GetAISettings(ctx context.Context) (domain.AISettings, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT provider, base_url, api_key, fast_extraction_model, reasoning_tailoring_model, bookmarked_models_json, updated_at
		FROM ai_settings
		WHERE id = 1
	`)

	var (
		providerStr    string
		baseURL        string
		apiKey         string
		fastModel      string
		reasoningModel string
		bookmarksJSON  string
		updatedAtStr   string
	)
	if err := row.Scan(
		&providerStr,
		&baseURL,
		&apiKey,
		&fastModel,
		&reasoningModel,
		&bookmarksJSON,
		&updatedAtStr,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DefaultAISettings(time.Time{}), nil
		}
		return domain.AISettings{}, fmt.Errorf("scan ai_settings: %w", err)
	}

	var bookmarks []string
	if bookmarksJSON != "" {
		if err := json.Unmarshal([]byte(bookmarksJSON), &bookmarks); err != nil {
			return domain.AISettings{}, fmt.Errorf("unmarshal bookmarked_models_json: %w", err)
		}
	}
	if bookmarks == nil {
		bookmarks = []string{}
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
	if err != nil {
		return domain.AISettings{}, fmt.Errorf("parse ai_settings.updated_at: %w", err)
	}

	return domain.AISettings{
		Provider:                domain.AIProvider(providerStr),
		BaseURL:                 baseURL,
		APIKey:                  apiKey,
		FastExtractionModel:     fastModel,
		ReasoningTailoringModel: reasoningModel,
		BookmarkedModels:        bookmarks,
		UpdatedAt:               updatedAt.UTC(),
	}, nil
}

func (s *Store) SaveAISettings(ctx context.Context, settings domain.AISettings) error {
	bookmarksBytes, err := json.Marshal(nonNilSlice(settings.BookmarkedModels))
	if err != nil {
		return fmt.Errorf("marshal bookmarked_models: %w", err)
	}

	_, err = s.exec(ctx).ExecContext(ctx, `
		INSERT INTO ai_settings (
			id, provider, base_url, api_key, fast_extraction_model, reasoning_tailoring_model, bookmarked_models_json, updated_at
		) VALUES (1, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			provider = excluded.provider,
			base_url = excluded.base_url,
			api_key = excluded.api_key,
			fast_extraction_model = excluded.fast_extraction_model,
			reasoning_tailoring_model = excluded.reasoning_tailoring_model,
			bookmarked_models_json = excluded.bookmarked_models_json,
			updated_at = excluded.updated_at
	`,
		string(settings.Provider),
		settings.BaseURL,
		settings.APIKey,
		settings.FastExtractionModel,
		settings.ReasoningTailoringModel,
		string(bookmarksBytes),
		settings.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("upsert ai_settings: %w", err)
	}
	return nil
}
