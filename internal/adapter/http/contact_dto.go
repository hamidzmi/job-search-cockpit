package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type ContactOpportunityLinkDTO struct {
	OpportunityID string    `json:"opportunity_id"`
	CompanyID     string    `json:"company_id"`
	CompanyName   string    `json:"company_name"`
	RoleTitle     string    `json:"role_title"`
	Stage         string    `json:"stage"`
	Outcome       string    `json:"outcome"`
	Role          string    `json:"role"`
	Notes         string    `json:"notes"`
	LinkedAt      time.Time `json:"linked_at"`
}

type ContactDTO struct {
	ID                  string                      `json:"id"`
	CompanyID           string                      `json:"company_id"`
	CompanyName         string                      `json:"company_name"`
	Name                string                      `json:"name"`
	Title               string                      `json:"title"`
	Email               string                      `json:"email"`
	LinkedInURL         string                      `json:"linkedin_url"`
	Warmth              string                      `json:"warmth"`
	Notes               string                      `json:"notes"`
	LinkedOpportunities []ContactOpportunityLinkDTO `json:"linked_opportunities"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
}

type ContactUpsertRequest struct {
	CompanyID   string `json:"company_id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Email       string `json:"email"`
	LinkedInURL string `json:"linkedin_url"`
	Warmth      string `json:"warmth"`
	Notes       string `json:"notes"`
}

func (r ContactUpsertRequest) toDomainInput() domain.ContactInput {
	return domain.ContactInput{
		CompanyID:   r.CompanyID,
		Name:        r.Name,
		Title:       r.Title,
		Email:       r.Email,
		LinkedInURL: r.LinkedInURL,
		Warmth:      domain.ContactWarmth(r.Warmth),
		Notes:       r.Notes,
	}
}

type OpportunityContactDTO struct {
	OpportunityID string    `json:"opportunity_id"`
	ContactID     string    `json:"contact_id"`
	Role          string    `json:"role"`
	Notes         string    `json:"notes"`
	LinkedAt      time.Time `json:"linked_at"`
	CompanyID     string    `json:"company_id"`
	CompanyName   string    `json:"company_name"`
	Name          string    `json:"name"`
	Title         string    `json:"title"`
	Email         string    `json:"email"`
	LinkedInURL   string    `json:"linkedin_url"`
	Warmth        string    `json:"warmth"`
	ContactNotes  string    `json:"contact_notes"`
}

type OpportunityContactLinkRequest struct {
	ContactID string `json:"contact_id"`
	Role      string `json:"role"`
	Notes     string `json:"notes"`
}

func toContactDTO(v domain.ContactView) ContactDTO {
	c := v.Contact
	links := make([]ContactOpportunityLinkDTO, 0, len(v.LinkedOpportunities))
	for _, l := range v.LinkedOpportunities {
		links = append(links, ContactOpportunityLinkDTO{
			OpportunityID: l.OpportunityID,
			CompanyID:     l.CompanyID,
			CompanyName:   l.CompanyName,
			RoleTitle:     l.RoleTitle,
			Stage:         string(l.Stage),
			Outcome:       string(l.Outcome),
			Role:          string(l.Role),
			Notes:         l.Notes,
			LinkedAt:      l.LinkedAt,
		})
	}
	return ContactDTO{
		ID:                  c.ID,
		CompanyID:           c.CompanyID,
		CompanyName:         v.CompanyName,
		Name:                c.Name,
		Title:               c.Title,
		Email:               c.Email,
		LinkedInURL:         c.LinkedInURL,
		Warmth:              string(c.Warmth),
		Notes:               c.Notes,
		LinkedOpportunities: links,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
}

func toOpportunityContactDTO(v domain.OpportunityContactView) OpportunityContactDTO {
	oc := v.OpportunityContact
	c := v.Contact
	return OpportunityContactDTO{
		OpportunityID: oc.OpportunityID,
		ContactID:     oc.ContactID,
		Role:          string(oc.Role),
		Notes:         oc.Notes,
		LinkedAt:      oc.LinkedAt,
		CompanyID:     c.CompanyID,
		CompanyName:   v.CompanyName,
		Name:          c.Name,
		Title:         c.Title,
		Email:         c.Email,
		LinkedInURL:   c.LinkedInURL,
		Warmth:        string(c.Warmth),
		ContactNotes:  c.Notes,
	}
}
