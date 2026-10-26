package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type JobPostingDTO struct {
	SourceURL     string   `json:"source_url"`
	RawText       string   `json:"raw_text"`
	Requirements  []string `json:"requirements"`
	TechStackTags []string `json:"tech_stack_tags"`
	StatedMinComp int      `json:"stated_min_comp"`
	StatedMaxComp int      `json:"stated_max_comp"`
	Currency      string   `json:"currency"`
	HasContent    bool     `json:"has_content"`
}

type StageTransitionDTO struct {
	ID             string    `json:"id"`
	OpportunityID  string    `json:"opportunity_id"`
	FromStage      string    `json:"from_stage"`
	ToStage        string    `json:"to_stage"`
	FromOutcome    string    `json:"from_outcome"`
	ToOutcome      string    `json:"to_outcome"`
	Note           string    `json:"note"`
	TransitionedAt time.Time `json:"transitioned_at"`
}

type OpportunityDTO struct {
	ID               string               `json:"id"`
	CompanyID        string               `json:"company_id"`
	CompanyName      string               `json:"company_name"`
	RoleTitle        string               `json:"role_title"`
	Seniority        string               `json:"seniority"`
	Location         string               `json:"location"`
	RemotePolicy     string               `json:"remote_policy"`
	Priority         string               `json:"priority"`
	Stage            string               `json:"stage"`
	Outcome          string               `json:"outcome"`
	Notes            string               `json:"notes"`
	JobPosting       JobPostingDTO        `json:"job_posting"`
	StageTransitions []StageTransitionDTO `json:"stage_transitions,omitempty"`
	LastActivityAt   time.Time            `json:"last_activity_at"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

type JobPostingRequest struct {
	SourceURL     string   `json:"source_url"`
	RawText       string   `json:"raw_text"`
	Requirements  []string `json:"requirements"`
	TechStackTags []string `json:"tech_stack_tags"`
	StatedMinComp int      `json:"stated_min_comp"`
	StatedMaxComp int      `json:"stated_max_comp"`
	Currency      string   `json:"currency"`
}

func (r JobPostingRequest) toDomain() domain.JobPosting {
	return domain.JobPosting{
		SourceURL:     r.SourceURL,
		RawText:       r.RawText,
		Requirements:  r.Requirements,
		TechStackTags: r.TechStackTags,
		StatedMinComp: r.StatedMinComp,
		StatedMaxComp: r.StatedMaxComp,
		Currency:      r.Currency,
	}
}

type OpportunityCreateRequest struct {
	CompanyID    string            `json:"company_id"`
	RoleTitle    string            `json:"role_title"`
	Seniority    string            `json:"seniority"`
	Location     string            `json:"location"`
	RemotePolicy string            `json:"remote_policy"`
	Priority     string            `json:"priority"`
	Stage        string            `json:"stage"`
	Outcome      string            `json:"outcome"`
	Notes        string            `json:"notes"`
	Note         string            `json:"note"`
	JobPosting   JobPostingRequest `json:"job_posting"`
}

func (r OpportunityCreateRequest) toDomainInput() domain.OpportunityInput {
	return domain.OpportunityInput{
		CompanyID:    r.CompanyID,
		RoleTitle:    r.RoleTitle,
		Seniority:    r.Seniority,
		Location:     r.Location,
		RemotePolicy: r.RemotePolicy,
		Priority:     r.Priority,
		Stage:        domain.Stage(r.Stage),
		Outcome:      domain.Outcome(r.Outcome),
		Notes:        r.Notes,
		Note:         r.Note,
		JobPosting:   r.JobPosting.toDomain(),
	}
}

type OpportunityUpdateRequest struct {
	CompanyID    string            `json:"company_id"`
	RoleTitle    string            `json:"role_title"`
	Seniority    string            `json:"seniority"`
	Location     string            `json:"location"`
	RemotePolicy string            `json:"remote_policy"`
	Priority     string            `json:"priority"`
	Notes        string            `json:"notes"`
	JobPosting   JobPostingRequest `json:"job_posting"`
}

func (r OpportunityUpdateRequest) toDomainInput() domain.OpportunityDetailsInput {
	return domain.OpportunityDetailsInput{
		CompanyID:    r.CompanyID,
		RoleTitle:    r.RoleTitle,
		Seniority:    r.Seniority,
		Location:     r.Location,
		RemotePolicy: r.RemotePolicy,
		Priority:     r.Priority,
		Notes:        r.Notes,
		JobPosting:   r.JobPosting.toDomain(),
	}
}

type OpportunityTransitionRequest struct {
	Stage   string `json:"stage"`
	Outcome string `json:"outcome"`
	Note    string `json:"note"`
}

func toStageTransitionDTO(tr domain.StageTransition) StageTransitionDTO {
	return StageTransitionDTO{
		ID:             tr.ID,
		OpportunityID:  tr.OpportunityID,
		FromStage:      string(tr.FromStage),
		ToStage:        string(tr.ToStage),
		FromOutcome:    string(tr.FromOutcome),
		ToOutcome:      string(tr.ToOutcome),
		Note:           tr.Note,
		TransitionedAt: tr.TransitionedAt,
	}
}

func toOpportunityDTO(o domain.Opportunity, transitions []domain.StageTransition) OpportunityDTO {
	reqs := o.JobPosting.Requirements
	if reqs == nil {
		reqs = []string{}
	}
	tags := o.JobPosting.TechStackTags
	if tags == nil {
		tags = []string{}
	}

	var trDTOs []StageTransitionDTO
	if transitions != nil {
		trDTOs = make([]StageTransitionDTO, 0, len(transitions))
		for _, tr := range transitions {
			trDTOs = append(trDTOs, toStageTransitionDTO(tr))
		}
	}

	return OpportunityDTO{
		ID:           o.ID,
		CompanyID:    o.CompanyID,
		CompanyName:  o.CompanyName,
		RoleTitle:    o.RoleTitle,
		Seniority:    o.Seniority,
		Location:     o.Location,
		RemotePolicy: o.RemotePolicy,
		Priority:     o.Priority,
		Stage:        string(o.Stage),
		Outcome:      string(o.Outcome),
		Notes:        o.Notes,
		JobPosting: JobPostingDTO{
			SourceURL:     o.JobPosting.SourceURL,
			RawText:       o.JobPosting.RawText,
			Requirements:  reqs,
			TechStackTags: tags,
			StatedMinComp: o.JobPosting.StatedMinComp,
			StatedMaxComp: o.JobPosting.StatedMaxComp,
			Currency:      o.JobPosting.Currency,
			HasContent:    o.JobPosting.HasContent(),
		},
		StageTransitions: trDTOs,
		LastActivityAt:   o.LastActivityAt,
		CreatedAt:        o.CreatedAt,
		UpdatedAt:        o.UpdatedAt,
	}
}
