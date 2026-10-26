package domain

import (
	"fmt"
	"strings"
	"time"
)

type InteractionKind string

const (
	InteractionKindEmail           InteractionKind = "Email"
	InteractionKindDM              InteractionKind = "DM"
	InteractionKindReferralRequest InteractionKind = "Referral Request"
	InteractionKindCoffeeChat      InteractionKind = "Coffee Chat"
	InteractionKindPhoneCall       InteractionKind = "Phone Call"
	InteractionKindOther           InteractionKind = "Other"
)

func AllInteractionKinds() []InteractionKind {
	return []InteractionKind{
		InteractionKindEmail,
		InteractionKindDM,
		InteractionKindReferralRequest,
		InteractionKindCoffeeChat,
		InteractionKindPhoneCall,
		InteractionKindOther,
	}
}

func (k InteractionKind) IsValid() bool {
	switch k {
	case InteractionKindEmail,
		InteractionKindDM,
		InteractionKindReferralRequest,
		InteractionKindCoffeeChat,
		InteractionKindPhoneCall,
		InteractionKindOther:
		return true
	default:
		return false
	}
}

type InteractionDirection string

const (
	InteractionDirectionOutbound InteractionDirection = "Outbound"
	InteractionDirectionInbound  InteractionDirection = "Inbound"
)

func AllInteractionDirections() []InteractionDirection {
	return []InteractionDirection{
		InteractionDirectionOutbound,
		InteractionDirectionInbound,
	}
}

func (d InteractionDirection) IsValid() bool {
	switch d {
	case InteractionDirectionOutbound, InteractionDirectionInbound:
		return true
	default:
		return false
	}
}

type Interaction struct {
	ID            string
	ContactID     string
	OpportunityID string
	Kind          InteractionKind
	Direction     InteractionDirection
	OccurredAt    time.Time
	Summary       string
	CreatedAt     time.Time
}

type FollowUpActionItemInput struct {
	Title string
	DueAt time.Time
	Notes string
}

type InteractionInput struct {
	ContactID     string
	OpportunityID string
	Kind          InteractionKind
	Direction     InteractionDirection
	OccurredAt    time.Time
	Summary       string
	FollowUp      *FollowUpActionItemInput
}

func NewInteraction(id string, in InteractionInput, now time.Time) (Interaction, error) {
	if strings.TrimSpace(id) == "" {
		return Interaction{}, fmt.Errorf("%w: interaction id is required", ErrValidation)
	}
	contactID := strings.TrimSpace(in.ContactID)
	if contactID == "" {
		return Interaction{}, fmt.Errorf("%w: contact id is required", ErrValidation)
	}
	kind := InteractionKind(strings.TrimSpace(string(in.Kind)))
	if kind == "" {
		kind = InteractionKindEmail
	}
	if !kind.IsValid() {
		return Interaction{}, fmt.Errorf("%w: invalid interaction kind %q", ErrValidation, kind)
	}
	direction := InteractionDirection(strings.TrimSpace(string(in.Direction)))
	if direction == "" {
		direction = InteractionDirectionOutbound
	}
	if !direction.IsValid() {
		return Interaction{}, fmt.Errorf("%w: invalid interaction direction %q", ErrValidation, direction)
	}
	summary := strings.TrimSpace(in.Summary)
	if summary == "" {
		return Interaction{}, fmt.Errorf("%w: interaction summary is required", ErrValidation)
	}

	ts := now.UTC()
	occurredAt := ts
	if !in.OccurredAt.IsZero() {
		occurredAt = in.OccurredAt.UTC()
	}

	return Interaction{
		ID:            strings.TrimSpace(id),
		ContactID:     contactID,
		OpportunityID: strings.TrimSpace(in.OpportunityID),
		Kind:          kind,
		Direction:     direction,
		OccurredAt:    occurredAt,
		Summary:       summary,
		CreatedAt:     ts,
	}, nil
}

type InteractionFilter struct {
	ContactID     string
	OpportunityID string
}

type InteractionView struct {
	Interaction          Interaction
	ContactName          string
	OpportunityRoleTitle string
	CompanyName          string
}
