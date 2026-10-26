package domain

import (
	"fmt"
	"strings"
	"time"
)

type ActionItemStatus string

const (
	ActionItemStatusOpen      ActionItemStatus = "open"
	ActionItemStatusCompleted ActionItemStatus = "completed"
	ActionItemStatusDismissed ActionItemStatus = "dismissed"
)

func AllActionItemStatuses() []ActionItemStatus {
	return []ActionItemStatus{
		ActionItemStatusOpen,
		ActionItemStatusCompleted,
		ActionItemStatusDismissed,
	}
}

func (s ActionItemStatus) IsValid() bool {
	switch s {
	case ActionItemStatusOpen, ActionItemStatusCompleted, ActionItemStatusDismissed:
		return true
	default:
		return false
	}
}

type ActionItem struct {
	ID            string
	Title         string
	OpportunityID string
	ContactID     string
	DueAt         time.Time
	Status        ActionItemStatus
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ActionItemInput struct {
	Title         string
	OpportunityID string
	ContactID     string
	DueAt         time.Time
	Notes         string
}

type ActionItemUpdateInput struct {
	Title         string
	OpportunityID string
	ContactID     string
	DueAt         time.Time
	Status        ActionItemStatus
	Notes         string
}

func NewActionItem(id string, in ActionItemInput, now time.Time) (ActionItem, error) {
	if strings.TrimSpace(id) == "" {
		return ActionItem{}, fmt.Errorf("%w: action item id is required", ErrValidation)
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return ActionItem{}, fmt.Errorf("%w: action item title is required", ErrValidation)
	}
	if in.DueAt.IsZero() {
		return ActionItem{}, fmt.Errorf("%w: action item due date is required", ErrValidation)
	}
	ts := now.UTC()
	return ActionItem{
		ID:            strings.TrimSpace(id),
		Title:         title,
		OpportunityID: strings.TrimSpace(in.OpportunityID),
		ContactID:     strings.TrimSpace(in.ContactID),
		DueAt:         in.DueAt.UTC(),
		Status:        ActionItemStatusOpen,
		Notes:         strings.TrimSpace(in.Notes),
		CreatedAt:     ts,
		UpdatedAt:     ts,
	}, nil
}

func (a *ActionItem) Update(in ActionItemUpdateInput, now time.Time) error {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return fmt.Errorf("%w: action item title is required", ErrValidation)
	}
	if in.DueAt.IsZero() {
		return fmt.Errorf("%w: action item due date is required", ErrValidation)
	}
	status := ActionItemStatus(strings.TrimSpace(string(in.Status)))
	if status == "" {
		status = a.Status
	}
	if !status.IsValid() {
		return fmt.Errorf("%w: invalid action item status %q", ErrValidation, status)
	}

	a.Title = title
	a.OpportunityID = strings.TrimSpace(in.OpportunityID)
	a.ContactID = strings.TrimSpace(in.ContactID)
	a.DueAt = in.DueAt.UTC()
	a.Status = status
	a.Notes = strings.TrimSpace(in.Notes)
	a.UpdatedAt = now.UTC()
	return nil
}

type ActionItemFilter struct {
	Status        ActionItemStatus
	OpportunityID string
	ContactID     string
}

func (f ActionItemFilter) Validate() error {
	if f.Status != "" && !f.Status.IsValid() {
		return fmt.Errorf("%w: invalid action item status filter %q", ErrValidation, f.Status)
	}
	return nil
}

type ActionItemView struct {
	ActionItem           ActionItem
	ContactName          string
	OpportunityRoleTitle string
	CompanyName          string
}
