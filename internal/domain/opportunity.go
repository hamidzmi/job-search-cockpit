package domain

import (
	"fmt"
	"strings"
	"time"
)

type Opportunity struct {
	ID             string
	CompanyID      string
	RoleTitle      string
	Seniority      string
	Location       string
	RemotePolicy   RemotePolicy
	Priority       Priority
	Stage          Stage
	Outcome        Outcome
	Notes          string
	JobPosting     JobPosting
	LastActivityAt time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type OpportunityDetailsInput struct {
	CompanyID    string
	RoleTitle    string
	Seniority    string
	Location     string
	RemotePolicy RemotePolicy
	Priority     Priority
	Notes        string
	JobPosting   JobPosting
}

type OpportunityInput struct {
	OpportunityDetailsInput
	Stage          Stage
	TransitionNote string
}

type StageTransition struct {
	ID             string
	OpportunityID  string
	FromStage      Stage
	ToStage        Stage
	FromOutcome    Outcome
	ToOutcome      Outcome
	Note           string
	TransitionedAt time.Time
}

func NewOpportunity(id, transitionID string, in OpportunityInput, now time.Time) (Opportunity, StageTransition, error) {
	if strings.TrimSpace(id) == "" {
		return Opportunity{}, StageTransition{}, fmt.Errorf("%w: opportunity id is required", ErrValidation)
	}
	companyID := strings.TrimSpace(in.CompanyID)
	if companyID == "" {
		return Opportunity{}, StageTransition{}, fmt.Errorf("%w: company id is required", ErrValidation)
	}
	roleTitle := strings.TrimSpace(in.RoleTitle)
	if roleTitle == "" {
		return Opportunity{}, StageTransition{}, fmt.Errorf("%w: role title is required", ErrValidation)
	}

	stage := in.Stage
	if stage == "" {
		stage = StageDiscovered
	}

	jp, err := normalizeJobPosting(in.JobPosting)
	if err != nil {
		return Opportunity{}, StageTransition{}, err
	}

	priority := Priority(strings.TrimSpace(string(in.Priority)))
	if priority == "" {
		priority = PriorityMedium
	}
	if !priority.IsValid() {
		return Opportunity{}, StageTransition{}, fmt.Errorf("%w: invalid priority %q", ErrValidation, priority)
	}

	remotePolicy := RemotePolicy(strings.TrimSpace(string(in.RemotePolicy)))
	if remotePolicy != "" && !remotePolicy.IsValid() {
		return Opportunity{}, StageTransition{}, fmt.Errorf("%w: invalid remote policy %q", ErrValidation, remotePolicy)
	}

	ts := now.UTC()
	opp := Opportunity{
		ID:           strings.TrimSpace(id),
		CompanyID:    companyID,
		RoleTitle:    roleTitle,
		Seniority:    strings.TrimSpace(in.Seniority),
		Location:     strings.TrimSpace(in.Location),
		RemotePolicy: remotePolicy,
		Priority:     priority,
		Notes:        strings.TrimSpace(in.Notes),
		JobPosting:   jp,
		CreatedAt:    ts,
	}

	tr, err := opp.Transition(transitionID, stage, OutcomeActive, in.TransitionNote, ts)
	if err != nil {
		return Opportunity{}, StageTransition{}, err
	}

	return opp, tr, nil
}

func (o *Opportunity) UpdateDetails(in OpportunityDetailsInput, now time.Time) error {
	companyID := strings.TrimSpace(in.CompanyID)
	if companyID == "" {
		companyID = o.CompanyID
	}
	roleTitle := strings.TrimSpace(in.RoleTitle)
	if roleTitle == "" {
		return fmt.Errorf("%w: role title is required", ErrValidation)
	}
	jp, err := normalizeJobPosting(in.JobPosting)
	if err != nil {
		return err
	}
	priority := Priority(strings.TrimSpace(string(in.Priority)))
	if priority == "" {
		priority = o.Priority
	}
	if !priority.IsValid() {
		return fmt.Errorf("%w: invalid priority %q", ErrValidation, priority)
	}
	remotePolicy := RemotePolicy(strings.TrimSpace(string(in.RemotePolicy)))
	if remotePolicy != "" && !remotePolicy.IsValid() {
		return fmt.Errorf("%w: invalid remote policy %q", ErrValidation, remotePolicy)
	}

	o.CompanyID = companyID
	o.RoleTitle = roleTitle
	o.Seniority = strings.TrimSpace(in.Seniority)
	o.Location = strings.TrimSpace(in.Location)
	o.RemotePolicy = remotePolicy
	o.Priority = priority
	o.Notes = strings.TrimSpace(in.Notes)
	o.JobPosting = jp
	o.UpdatedAt = now.UTC()
	return nil
}

func (o *Opportunity) Transition(transitionID string, toStage Stage, toOutcome Outcome, note string, now time.Time) (StageTransition, error) {
	if strings.TrimSpace(transitionID) == "" {
		return StageTransition{}, fmt.Errorf("%w: transition id is required", ErrValidation)
	}

	nextStage := o.Stage
	if toStage != "" {
		if !toStage.IsValid() {
			return StageTransition{}, fmt.Errorf("%w: invalid stage %q", ErrValidation, toStage)
		}
		nextStage = toStage
	}

	nextOutcome := o.Outcome
	if toOutcome != "" {
		if !toOutcome.IsValid() {
			return StageTransition{}, fmt.Errorf("%w: invalid outcome %q", ErrValidation, toOutcome)
		}
		nextOutcome = toOutcome
	}

	if nextStage == o.Stage && nextOutcome == o.Outcome {
		return StageTransition{}, fmt.Errorf("%w: stage (%s) and outcome (%s) are unchanged", ErrInvalidTransition, o.Stage, o.Outcome)
	}

	ts := now.UTC()
	tr := StageTransition{
		ID:             strings.TrimSpace(transitionID),
		OpportunityID:  o.ID,
		FromStage:      o.Stage,
		ToStage:        nextStage,
		FromOutcome:    o.Outcome,
		ToOutcome:      nextOutcome,
		Note:           strings.TrimSpace(note),
		TransitionedAt: ts,
	}

	o.Stage = nextStage
	o.Outcome = nextOutcome
	o.LastActivityAt = ts
	o.UpdatedAt = ts

	return tr, nil
}

func (o *Opportunity) RecordActivity(activityAt, now time.Time) {
	actUTC := activityAt.UTC()
	if actUTC.After(o.LastActivityAt) {
		o.LastActivityAt = actUTC
	}
	o.UpdatedAt = now.UTC()
}

