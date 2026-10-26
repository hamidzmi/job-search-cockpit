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

func (s *Store) SaveStory(ctx context.Context, story domain.Story) error {
	compJSON, err := json.Marshal(nonNilSlice(story.CompetencyTags))
	if err != nil {
		return err
	}
	skillJSON, err := json.Marshal(nonNilSlice(story.SkillTags))
	if err != nil {
		return err
	}

	_, err = s.exec(ctx).ExecContext(ctx, `
		INSERT INTO stories (
			id, title, hook, situation, task, action, result,
			competency_tags_json, skill_tags_json, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			hook = excluded.hook,
			situation = excluded.situation,
			task = excluded.task,
			action = excluded.action,
			result = excluded.result,
			competency_tags_json = excluded.competency_tags_json,
			skill_tags_json = excluded.skill_tags_json,
			updated_at = excluded.updated_at
	`,
		story.ID,
		story.Title,
		story.Hook,
		story.Situation,
		story.Task,
		story.Action,
		story.Result,
		string(compJSON),
		string(skillJSON),
		story.CreatedAt.UTC().Format(time.RFC3339Nano),
		story.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetStory(ctx context.Context, id string) (domain.Story, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			id, title, hook, situation, task, action, result,
			competency_tags_json, skill_tags_json, created_at, updated_at
		FROM stories
		WHERE id = ?
	`, id)
	story, err := scanStory(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Story{}, fmt.Errorf("%w: story %s", domain.ErrNotFound, id)
	}
	return story, err
}

func (s *Store) ListStories(ctx context.Context, filter domain.StoryFilter) ([]domain.Story, error) {
	rows, err := s.exec(ctx).QueryContext(ctx, `
		SELECT
			id, title, hook, situation, task, action, result,
			competency_tags_json, skill_tags_json, created_at, updated_at
		FROM stories
		ORDER BY updated_at DESC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []domain.Story{}
	for rows.Next() {
		story, err := scanStory(rows.Scan)
		if err != nil {
			return nil, err
		}
		if story.MatchesFilter(filter) {
			result = append(result, story)
		}
	}
	return result, rows.Err()
}

func (s *Store) DeleteStory(ctx context.Context, id string) error {
	res, err := s.exec(ctx).ExecContext(ctx, `DELETE FROM stories WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%w: story %s", domain.ErrNotFound, id)
	}
	return nil
}

func scanStory(scan func(dest ...any) error) (domain.Story, error) {
	var story domain.Story
	var compJSON, skillJSON, createdStr, updatedStr string
	err := scan(
		&story.ID,
		&story.Title,
		&story.Hook,
		&story.Situation,
		&story.Task,
		&story.Action,
		&story.Result,
		&compJSON,
		&skillJSON,
		&createdStr,
		&updatedStr,
	)
	if err != nil {
		return domain.Story{}, err
	}
	_ = json.Unmarshal([]byte(compJSON), &story.CompetencyTags)
	_ = json.Unmarshal([]byte(skillJSON), &story.SkillTags)
	if story.CompetencyTags == nil {
		story.CompetencyTags = []string{}
	}
	if story.SkillTags == nil {
		story.SkillTags = []string{}
	}
	story.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	story.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return story, nil
}
