package httpadapter

import (
	"job-search-cockpit/internal/domain"
)

type ExtractedJobPostingDraftDTO struct {
	CompanyName      string   `json:"company_name"`
	CompanyDomain    string   `json:"company_domain"`
	RoleTitle        string   `json:"role_title"`
	Seniority        string   `json:"seniority"`
	Location         string   `json:"location"`
	RemotePolicy     string   `json:"remote_policy"`
	StatedMinComp    int      `json:"stated_min_comp"`
	StatedMaxComp    int      `json:"stated_max_comp"`
	Currency         string   `json:"currency"`
	TechStackTags    []string `json:"tech_stack_tags"`
	Requirements     []string `json:"requirements"`
	SourceURL        string   `json:"source_url"`
	RawText          string   `json:"raw_text"`
	ExtractionSource string   `json:"extraction_source"`
	ModelUsed        string   `json:"model_used"`
	FetchWarning     string   `json:"fetch_warning"`
	AssistantWarning string   `json:"assistant_warning"`
}

type quickCaptureExtractRequest struct {
	SourceURL     string `json:"source_url"`
	RawText       string `json:"raw_text"`
	ModelOverride string `json:"model_override"`
}

func (r quickCaptureExtractRequest) toDomainInput() domain.QuickCaptureExtractInput {
	return domain.QuickCaptureExtractInput{
		SourceURL:     r.SourceURL,
		RawText:       r.RawText,
		ModelOverride: r.ModelOverride,
	}
}

func toExtractedJobPostingDraftDTO(d domain.ExtractedJobPostingDraft) ExtractedJobPostingDraftDTO {
	tags := d.TechStackTags
	if tags == nil {
		tags = []string{}
	}
	reqs := d.Requirements
	if reqs == nil {
		reqs = []string{}
	}
	return ExtractedJobPostingDraftDTO{
		CompanyName:      d.CompanyName,
		CompanyDomain:    d.CompanyDomain,
		RoleTitle:        d.RoleTitle,
		Seniority:        d.Seniority,
		Location:         d.Location,
		RemotePolicy:     string(d.RemotePolicy),
		StatedMinComp:    d.StatedMinComp,
		StatedMaxComp:    d.StatedMaxComp,
		Currency:         d.Currency,
		TechStackTags:    tags,
		Requirements:     reqs,
		SourceURL:        d.SourceURL,
		RawText:          d.RawText,
		ExtractionSource: d.ExtractionSource,
		ModelUsed:        d.ModelUsed,
		FetchWarning:     d.FetchWarning,
		AssistantWarning: d.AssistantWarning,
	}
}

type quickCaptureSaveRequest struct {
	OpportunityCreateRequest
	CompanyName   string `json:"company_name"`
	CompanyDomain string `json:"company_domain"`
}

func (r quickCaptureSaveRequest) toDomainInput() domain.QuickCaptureSaveInput {
	in := r.OpportunityCreateRequest.toDomainInput()
	return domain.QuickCaptureSaveInput{
		CompanyID:      in.CompanyID,
		CompanyName:    r.CompanyName,
		CompanyDomain:  r.CompanyDomain,
		RoleTitle:      in.RoleTitle,
		Seniority:      in.Seniority,
		Location:       in.Location,
		RemotePolicy:   in.RemotePolicy,
		Priority:       in.Priority,
		Stage:          in.Stage,
		TransitionNote: in.TransitionNote,
		Notes:          in.Notes,
		JobPosting:     in.JobPosting,
	}
}
