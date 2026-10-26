package domain

import (
	"fmt"
	"strings"
	"time"
)

type ContactWarmth string

const (
	ContactWarmthCold     ContactWarmth = "Cold"
	ContactWarmthWarm     ContactWarmth = "Warm"
	ContactWarmthAdvocate ContactWarmth = "Advocate"
)

func AllContactWarmths() []ContactWarmth {
	return []ContactWarmth{
		ContactWarmthCold,
		ContactWarmthWarm,
		ContactWarmthAdvocate,
	}
}

func (w ContactWarmth) IsValid() bool {
	switch w {
	case ContactWarmthCold, ContactWarmthWarm, ContactWarmthAdvocate:
		return true
	default:
		return false
	}
}

type StakeholderRole string

const (
	StakeholderRoleReferrer      StakeholderRole = "Referrer"
	StakeholderRoleRecruiter     StakeholderRole = "Recruiter"
	StakeholderRoleHiringManager StakeholderRole = "Hiring Manager"
	StakeholderRoleInterviewer   StakeholderRole = "Interviewer"
	StakeholderRoleChampion      StakeholderRole = "Champion"
	StakeholderRoleOther         StakeholderRole = "Other"
)

func AllStakeholderRoles() []StakeholderRole {
	return []StakeholderRole{
		StakeholderRoleReferrer,
		StakeholderRoleRecruiter,
		StakeholderRoleHiringManager,
		StakeholderRoleInterviewer,
		StakeholderRoleChampion,
		StakeholderRoleOther,
	}
}

func (r StakeholderRole) IsValid() bool {
	switch r {
	case StakeholderRoleReferrer,
		StakeholderRoleRecruiter,
		StakeholderRoleHiringManager,
		StakeholderRoleInterviewer,
		StakeholderRoleChampion,
		StakeholderRoleOther:
		return true
	default:
		return false
	}
}

type Contact struct {
	ID          string
	CompanyID   string
	Name        string
	Title       string
	Email       string
	LinkedInURL string
	Warmth      ContactWarmth
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ContactInput struct {
	CompanyID   string
	Name        string
	Title       string
	Email       string
	LinkedInURL string
	Warmth      ContactWarmth
	Notes       string
}

func NewContact(id string, in ContactInput, now time.Time) (Contact, error) {
	if strings.TrimSpace(id) == "" {
		return Contact{}, fmt.Errorf("%w: contact id is required", ErrValidation)
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return Contact{}, fmt.Errorf("%w: contact name is required", ErrValidation)
	}
	warmth := ContactWarmth(strings.TrimSpace(string(in.Warmth)))
	if warmth == "" {
		warmth = ContactWarmthWarm
	}
	if !warmth.IsValid() {
		return Contact{}, fmt.Errorf("%w: invalid contact warmth %q", ErrValidation, warmth)
	}
	ts := now.UTC()
	return Contact{
		ID:          strings.TrimSpace(id),
		CompanyID:   strings.TrimSpace(in.CompanyID),
		Name:        name,
		Title:       strings.TrimSpace(in.Title),
		Email:       strings.TrimSpace(in.Email),
		LinkedInURL: strings.TrimSpace(in.LinkedInURL),
		Warmth:      warmth,
		Notes:       strings.TrimSpace(in.Notes),
		CreatedAt:   ts,
		UpdatedAt:   ts,
	}, nil
}

func (c *Contact) Update(in ContactInput, now time.Time) error {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return fmt.Errorf("%w: contact name is required", ErrValidation)
	}
	warmth := ContactWarmth(strings.TrimSpace(string(in.Warmth)))
	if warmth == "" {
		warmth = c.Warmth
	}
	if !warmth.IsValid() {
		return fmt.Errorf("%w: invalid contact warmth %q", ErrValidation, warmth)
	}
	c.CompanyID = strings.TrimSpace(in.CompanyID)
	c.Name = name
	c.Title = strings.TrimSpace(in.Title)
	c.Email = strings.TrimSpace(in.Email)
	c.LinkedInURL = strings.TrimSpace(in.LinkedInURL)
	c.Warmth = warmth
	c.Notes = strings.TrimSpace(in.Notes)
	c.UpdatedAt = now.UTC()
	return nil
}

type ContactFilter struct {
	CompanyID string
	Warmth    ContactWarmth
	Query     string
}

func (f ContactFilter) Validate() error {
	if f.Warmth != "" && !f.Warmth.IsValid() {
		return fmt.Errorf("%w: invalid contact warmth filter %q", ErrValidation, f.Warmth)
	}
	return nil
}

type OpportunityContact struct {
	OpportunityID string
	ContactID     string
	Role          StakeholderRole
	Notes         string
	LinkedAt      time.Time
}

type OpportunityContactInput struct {
	OpportunityID string
	ContactID     string
	Role          StakeholderRole
	Notes         string
}

func NewOpportunityContact(in OpportunityContactInput, now time.Time) (OpportunityContact, error) {
	oppID := strings.TrimSpace(in.OpportunityID)
	if oppID == "" {
		return OpportunityContact{}, fmt.Errorf("%w: opportunity id is required", ErrValidation)
	}
	cntID := strings.TrimSpace(in.ContactID)
	if cntID == "" {
		return OpportunityContact{}, fmt.Errorf("%w: contact id is required", ErrValidation)
	}
	role := StakeholderRole(strings.TrimSpace(string(in.Role)))
	if !role.IsValid() {
		return OpportunityContact{}, fmt.Errorf("%w: invalid stakeholder role %q", ErrValidation, role)
	}
	return OpportunityContact{
		OpportunityID: oppID,
		ContactID:     cntID,
		Role:          role,
		Notes:         strings.TrimSpace(in.Notes),
		LinkedAt:      now.UTC(),
	}, nil
}

type ContactOpportunityLinkView struct {
	OpportunityID string
	CompanyID     string
	CompanyName   string
	RoleTitle     string
	Stage         Stage
	Outcome       Outcome
	Role          StakeholderRole
	Notes         string
	LinkedAt      time.Time
}

type ContactView struct {
	Contact             Contact
	CompanyName         string
	LinkedOpportunities []ContactOpportunityLinkView
}

type OpportunityContactView struct {
	OpportunityContact OpportunityContact
	Contact            Contact
	CompanyName        string
}
