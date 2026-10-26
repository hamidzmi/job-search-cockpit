package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type StoryDTO struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Hook           string    `json:"hook"`
	Situation      string    `json:"situation"`
	Task           string    `json:"task"`
	Action         string    `json:"action"`
	Result         string    `json:"result"`
	CompetencyTags []string  `json:"competency_tags"`
	SkillTags      []string  `json:"skill_tags"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type StoryUpsertRequest struct {
	Title          string   `json:"title"`
	Hook           string   `json:"hook"`
	Situation      string   `json:"situation"`
	Task           string   `json:"task"`
	Action         string   `json:"action"`
	Result         string   `json:"result"`
	CompetencyTags []string `json:"competency_tags"`
	SkillTags      []string `json:"skill_tags"`
}

func (r StoryUpsertRequest) toDomainInput() domain.StoryInput {
	return domain.StoryInput{
		Title:          r.Title,
		Hook:           r.Hook,
		Situation:      r.Situation,
		Task:           r.Task,
		Action:         r.Action,
		Result:         r.Result,
		CompetencyTags: r.CompetencyTags,
		SkillTags:      r.SkillTags,
	}
}

func toStoryDTO(s domain.Story) StoryDTO {
	compTags := s.CompetencyTags
	if compTags == nil {
		compTags = []string{}
	}
	skillTags := s.SkillTags
	if skillTags == nil {
		skillTags = []string{}
	}
	return StoryDTO{
		ID:             s.ID,
		Title:          s.Title,
		Hook:           s.Hook,
		Situation:      s.Situation,
		Task:           s.Task,
		Action:         s.Action,
		Result:         s.Result,
		CompetencyTags: compTags,
		SkillTags:      skillTags,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}
