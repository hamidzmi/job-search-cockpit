package domain

import (
	"fmt"
	"strings"
	"time"
)

type Story struct {
	ID             string
	Title          string
	Hook           string
	Situation      string
	Task           string
	Action         string
	Result         string
	CompetencyTags []string
	SkillTags      []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StoryInput struct {
	Title          string
	Hook           string
	Situation      string
	Task           string
	Action         string
	Result         string
	CompetencyTags []string
	SkillTags      []string
}

type StoryFilter struct {
	Query         string
	CompetencyTag string
	SkillTag      string
}

func NewStory(id string, in StoryInput, now time.Time) (Story, error) {
	if strings.TrimSpace(id) == "" {
		return Story{}, fmt.Errorf("%w: story id is required", ErrValidation)
	}
	ts := now.UTC()
	story := Story{
		ID:        strings.TrimSpace(id),
		CreatedAt: ts,
	}
	if err := story.Update(in, ts); err != nil {
		return Story{}, err
	}
	return story, nil
}

func (s *Story) Update(in StoryInput, now time.Time) error {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return fmt.Errorf("%w: story title is required", ErrValidation)
	}
	s.Title = title
	s.Hook = strings.TrimSpace(in.Hook)
	s.Situation = strings.TrimSpace(in.Situation)
	s.Task = strings.TrimSpace(in.Task)
	s.Action = strings.TrimSpace(in.Action)
	s.Result = strings.TrimSpace(in.Result)
	s.CompetencyTags = normalizeUniqueStrings(in.CompetencyTags)
	s.SkillTags = normalizeUniqueStrings(in.SkillTags)
	s.UpdatedAt = now.UTC()
	return nil
}

func (s Story) MatchesFilter(f StoryFilter) bool {
	if wantComp := strings.TrimSpace(f.CompetencyTag); wantComp != "" {
		if !containsTagFold(s.CompetencyTags, wantComp) {
			return false
		}
	}
	if wantSkill := strings.TrimSpace(f.SkillTag); wantSkill != "" {
		if !containsTagFold(s.SkillTags, wantSkill) {
			return false
		}
	}
	if q := strings.ToLower(strings.TrimSpace(f.Query)); q != "" {
		haystack := strings.ToLower(strings.Join([]string{
			s.Title,
			s.Hook,
			s.Situation,
			s.Task,
			s.Action,
			s.Result,
			strings.Join(s.CompetencyTags, " "),
			strings.Join(s.SkillTags, " "),
		}, " "))
		if !strings.Contains(haystack, q) {
			return false
		}
	}
	return true
}

func containsTagFold(tags []string, target string) bool {
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), target) {
			return true
		}
	}
	return false
}
