package httpadapter

import (
	"fmt"
	"strings"
	"time"

	"job-search-cockpit/internal/domain"
)

func parseFlexibleTimestamp(raw string) (time.Time, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("%w: invalid timestamp %q", domain.ErrValidation, raw)
}

type InteractionDTO struct {
	ID                   string    `json:"id"`
	ContactID            string    `json:"contact_id"`
	ContactName          string    `json:"contact_name"`
	OpportunityID        string    `json:"opportunity_id"`
	OpportunityRoleTitle string    `json:"opportunity_role_title"`
	CompanyName          string    `json:"company_name"`
	Kind                 string    `json:"kind"`
	Direction            string    `json:"direction"`
	OccurredAt           time.Time `json:"occurred_at"`
	Summary              string    `json:"summary"`
	CreatedAt            time.Time `json:"created_at"`
}

type FollowUpActionItemRequest struct {
	Title string `json:"title"`
	DueAt string `json:"due_at"`
	Notes string `json:"notes"`
}

type InteractionCreateRequest struct {
	ContactID     string                     `json:"contact_id"`
	OpportunityID string                     `json:"opportunity_id"`
	Kind          string                     `json:"kind"`
	Direction     string                     `json:"direction"`
	OccurredAt    string                     `json:"occurred_at"`
	Summary       string                     `json:"summary"`
	FollowUp      *FollowUpActionItemRequest `json:"follow_up,omitempty"`
}

func (r InteractionCreateRequest) toDomainInput() (domain.InteractionInput, error) {
	occurredAt, err := parseFlexibleTimestamp(r.OccurredAt)
	if err != nil {
		return domain.InteractionInput{}, err
	}
	var followUp *domain.FollowUpActionItemInput
	if r.FollowUp != nil && strings.TrimSpace(r.FollowUp.Title) != "" {
		dueAt, err := parseFlexibleTimestamp(r.FollowUp.DueAt)
		if err != nil {
			return domain.InteractionInput{}, err
		}
		followUp = &domain.FollowUpActionItemInput{
			Title: r.FollowUp.Title,
			DueAt: dueAt,
			Notes: r.FollowUp.Notes,
		}
	}
	return domain.InteractionInput{
		ContactID:     r.ContactID,
		OpportunityID: r.OpportunityID,
		Kind:          domain.InteractionKind(r.Kind),
		Direction:     domain.InteractionDirection(r.Direction),
		OccurredAt:    occurredAt,
		Summary:       r.Summary,
		FollowUp:      followUp,
	}, nil
}

type LogInteractionResponseDTO struct {
	Interaction        InteractionDTO `json:"interaction"`
	FollowUpActionItem *ActionItemDTO `json:"follow_up_action_item,omitempty"`
}

func toInteractionDTO(v domain.InteractionView) InteractionDTO {
	i := v.Interaction
	return InteractionDTO{
		ID:                   i.ID,
		ContactID:            i.ContactID,
		ContactName:          v.ContactName,
		OpportunityID:        i.OpportunityID,
		OpportunityRoleTitle: v.OpportunityRoleTitle,
		CompanyName:          v.CompanyName,
		Kind:                 string(i.Kind),
		Direction:            string(i.Direction),
		OccurredAt:           i.OccurredAt,
		Summary:              i.Summary,
		CreatedAt:            i.CreatedAt,
	}
}
