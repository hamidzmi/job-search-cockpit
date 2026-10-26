package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type WorkHistoryEntryDTO struct {
	Company       string   `json:"company"`
	Title         string   `json:"title"`
	StartDate     string   `json:"start_date"`
	EndDate       string   `json:"end_date"`
	Summary       string   `json:"summary"`
	ImpactBullets []string `json:"impact_bullets"`
}

type CandidateProfileDTO struct {
	FullName            string                `json:"full_name"`
	Headline            string                `json:"headline"`
	Summary             string                `json:"summary"`
	WorkHistory         []WorkHistoryEntryDTO `json:"work_history"`
	Skills              []string              `json:"skills"`
	TargetRoleTitles    []string              `json:"target_role_titles"`
	LocationPreferences []string              `json:"location_preferences"`
	RemotePreference    string                `json:"remote_preference"`
	CompensationFloor   int                   `json:"compensation_floor"`
	TargetCompensation  int                   `json:"target_compensation"`
	Currency            string                `json:"currency"`
	UpdatedAt           time.Time             `json:"updated_at"`
}

type CandidateProfileUpsertRequest struct {
	FullName            string                `json:"full_name"`
	Headline            string                `json:"headline"`
	Summary             string                `json:"summary"`
	WorkHistory         []WorkHistoryEntryDTO `json:"work_history"`
	Skills              []string              `json:"skills"`
	TargetRoleTitles    []string              `json:"target_role_titles"`
	LocationPreferences []string              `json:"location_preferences"`
	RemotePreference    string                `json:"remote_preference"`
	CompensationFloor   int                   `json:"compensation_floor"`
	TargetCompensation  int                   `json:"target_compensation"`
	Currency            string                `json:"currency"`
}

func (r CandidateProfileUpsertRequest) toDomainInput() domain.CandidateProfileInput {
	history := make([]domain.WorkHistoryEntry, 0, len(r.WorkHistory))
	for _, entry := range r.WorkHistory {
		history = append(history, domain.WorkHistoryEntry{
			Company:       entry.Company,
			Title:         entry.Title,
			StartDate:     entry.StartDate,
			EndDate:       entry.EndDate,
			Summary:       entry.Summary,
			ImpactBullets: entry.ImpactBullets,
		})
	}
	return domain.CandidateProfileInput{
		FullName:            r.FullName,
		Headline:            r.Headline,
		Summary:             r.Summary,
		WorkHistory:         history,
		Skills:              r.Skills,
		TargetRoleTitles:    r.TargetRoleTitles,
		LocationPreferences: r.LocationPreferences,
		RemotePreference:    r.RemotePreference,
		CompensationFloor:   r.CompensationFloor,
		TargetCompensation:  r.TargetCompensation,
		Currency:            r.Currency,
	}
}

func toCandidateProfileDTO(p domain.CandidateProfile) CandidateProfileDTO {
	history := make([]WorkHistoryEntryDTO, 0, len(p.WorkHistory))
	for _, entry := range p.WorkHistory {
		bullets := entry.ImpactBullets
		if bullets == nil {
			bullets = []string{}
		}
		history = append(history, WorkHistoryEntryDTO{
			Company:       entry.Company,
			Title:         entry.Title,
			StartDate:     entry.StartDate,
			EndDate:       entry.EndDate,
			Summary:       entry.Summary,
			ImpactBullets: bullets,
		})
	}
	skills := p.Skills
	if skills == nil {
		skills = []string{}
	}
	titles := p.TargetRoleTitles
	if titles == nil {
		titles = []string{}
	}
	locations := p.LocationPreferences
	if locations == nil {
		locations = []string{}
	}
	return CandidateProfileDTO{
		FullName:            p.FullName,
		Headline:            p.Headline,
		Summary:             p.Summary,
		WorkHistory:         history,
		Skills:              skills,
		TargetRoleTitles:    titles,
		LocationPreferences: locations,
		RemotePreference:    p.RemotePreference,
		CompensationFloor:   p.CompensationFloor,
		TargetCompensation:  p.TargetCompensation,
		Currency:            p.Currency,
		UpdatedAt:           p.UpdatedAt,
	}
}
