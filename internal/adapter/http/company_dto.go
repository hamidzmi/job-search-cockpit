package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type CompanyDTO struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Domain       string    `json:"domain"`
	Headquarters string    `json:"headquarters"`
	RemotePolicy string    `json:"remote_policy"`
	SizeStage    string    `json:"size_stage"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CompanyUpsertRequest struct {
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	Headquarters string `json:"headquarters"`
	RemotePolicy string `json:"remote_policy"`
	SizeStage    string `json:"size_stage"`
	Notes        string `json:"notes"`
}

func (r CompanyUpsertRequest) toDomainInput() domain.CompanyInput {
	return domain.CompanyInput{
		Name:         r.Name,
		Domain:       r.Domain,
		Headquarters: r.Headquarters,
		RemotePolicy: r.RemotePolicy,
		SizeStage:    r.SizeStage,
		Notes:        r.Notes,
	}
}

func toCompanyDTO(c domain.Company) CompanyDTO {
	return CompanyDTO{
		ID:           c.ID,
		Name:         c.Name,
		Domain:       c.Domain,
		Headquarters: c.Headquarters,
		RemotePolicy: c.RemotePolicy,
		SizeStage:    c.SizeStage,
		Notes:        c.Notes,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}
