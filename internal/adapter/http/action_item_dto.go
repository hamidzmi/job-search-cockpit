package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type ActionItemDTO struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title"`
	OpportunityID        string    `json:"opportunity_id"`
	OpportunityRoleTitle string    `json:"opportunity_role_title"`
	ContactID            string    `json:"contact_id"`
	ContactName          string    `json:"contact_name"`
	CompanyName          string    `json:"company_name"`
	DueAt                time.Time `json:"due_at"`
	Status               string    `json:"status"`
	Notes                string    `json:"notes"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ActionItemCreateRequest struct {
	Title         string `json:"title"`
	OpportunityID string `json:"opportunity_id"`
	ContactID     string `json:"contact_id"`
	DueAt         string `json:"due_at"`
	Notes         string `json:"notes"`
}

func (r ActionItemCreateRequest) toDomainInput() (domain.ActionItemInput, error) {
	dueAt, err := parseFlexibleTimestamp(r.DueAt)
	if err != nil {
		return domain.ActionItemInput{}, err
	}
	return domain.ActionItemInput{
		Title:         r.Title,
		OpportunityID: r.OpportunityID,
		ContactID:     r.ContactID,
		DueAt:         dueAt,
		Notes:         r.Notes,
	}, nil
}

type ActionItemUpdateRequest struct {
	Title         string `json:"title"`
	OpportunityID string `json:"opportunity_id"`
	ContactID     string `json:"contact_id"`
	DueAt         string `json:"due_at"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}

func (r ActionItemUpdateRequest) toDomainInput() (domain.ActionItemUpdateInput, error) {
	dueAt, err := parseFlexibleTimestamp(r.DueAt)
	if err != nil {
		return domain.ActionItemUpdateInput{}, err
	}
	return domain.ActionItemUpdateInput{
		Title:         r.Title,
		OpportunityID: r.OpportunityID,
		ContactID:     r.ContactID,
		DueAt:         dueAt,
		Status:        domain.ActionItemStatus(r.Status),
		Notes:         r.Notes,
	}, nil
}

func toActionItemDTO(v domain.ActionItemView) ActionItemDTO {
	a := v.ActionItem
	return ActionItemDTO{
		ID:                   a.ID,
		Title:                a.Title,
		OpportunityID:        a.OpportunityID,
		OpportunityRoleTitle: v.OpportunityRoleTitle,
		ContactID:            a.ContactID,
		ContactName:          v.ContactName,
		CompanyName:          v.CompanyName,
		DueAt:                a.DueAt,
		Status:               string(a.Status),
		Notes:                a.Notes,
		CreatedAt:            a.CreatedAt,
		UpdatedAt:            a.UpdatedAt,
	}
}
