package domain

import (
	"fmt"
	"strings"
	"time"
)

type OfferStatus string

const (
	OfferStatusPending     OfferStatus = "Pending"
	OfferStatusNegotiating OfferStatus = "Negotiating"
	OfferStatusAccepted    OfferStatus = "Accepted"
	OfferStatusDeclined    OfferStatus = "Declined"
	OfferStatusExpired     OfferStatus = "Expired"
)

func AllOfferStatuses() []OfferStatus {
	return []OfferStatus{
		OfferStatusPending,
		OfferStatusNegotiating,
		OfferStatusAccepted,
		OfferStatusDeclined,
		OfferStatusExpired,
	}
}

func (s OfferStatus) IsValid() bool {
	switch s {
	case OfferStatusPending,
		OfferStatusNegotiating,
		OfferStatusAccepted,
		OfferStatusDeclined,
		OfferStatusExpired:
		return true
	default:
		return false
	}
}

type Offer struct {
	ID                 string
	OpportunityID      string
	BaseSalary         int
	BonusTargetPct     int
	BonusAmount        int
	EquityType         string
	EquityTotalValue   int
	EquityVestingYears int
	SignOnBonus        int
	Currency           string
	DeadlineAt         time.Time
	BenefitsNotes      string
	NegotiationNotes   string
	Status             OfferStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type OfferInput struct {
	OpportunityID      string
	BaseSalary         int
	BonusTargetPct     int
	BonusAmount        int
	EquityType         string
	EquityTotalValue   int
	EquityVestingYears int
	SignOnBonus        int
	Currency           string
	DeadlineAt         time.Time
	BenefitsNotes      string
	NegotiationNotes   string
	Status             OfferStatus
}

func NewOffer(id string, in OfferInput, now time.Time) (Offer, error) {
	if strings.TrimSpace(id) == "" {
		return Offer{}, fmt.Errorf("%w: offer id is required", ErrValidation)
	}
	oppID := strings.TrimSpace(in.OpportunityID)
	if oppID == "" {
		return Offer{}, fmt.Errorf("%w: opportunity id is required", ErrValidation)
	}
	ts := now.UTC()
	offer := Offer{
		ID:            strings.TrimSpace(id),
		OpportunityID: oppID,
		CreatedAt:     ts,
	}
	if err := offer.Update(in, ts); err != nil {
		return Offer{}, err
	}
	return offer, nil
}

func (o *Offer) Update(in OfferInput, now time.Time) error {
	if oppID := strings.TrimSpace(in.OpportunityID); oppID != "" {
		o.OpportunityID = oppID
	}
	if o.OpportunityID == "" {
		return fmt.Errorf("%w: opportunity id is required", ErrValidation)
	}
	if in.BaseSalary <= 0 {
		return fmt.Errorf("%w: base salary must be greater than zero", ErrValidation)
	}
	if in.BonusTargetPct < 0 || in.BonusAmount < 0 || in.EquityTotalValue < 0 || in.EquityVestingYears < 0 || in.SignOnBonus < 0 {
		return fmt.Errorf("%w: offer compensation values cannot be negative", ErrValidation)
	}

	status := OfferStatus(strings.TrimSpace(string(in.Status)))
	if status == "" {
		if o.Status != "" {
			status = o.Status
		} else {
			status = OfferStatusPending
		}
	}
	if !status.IsValid() {
		return fmt.Errorf("%w: invalid offer status %q", ErrValidation, status)
	}

	vestingYears := in.EquityVestingYears
	if vestingYears == 0 {
		vestingYears = 4
	}

	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "USD"
	}

	equityType := strings.TrimSpace(in.EquityType)
	if equityType == "" && in.EquityTotalValue > 0 {
		equityType = "RSU"
	}

	var deadlineAt time.Time
	if !in.DeadlineAt.IsZero() {
		deadlineAt = in.DeadlineAt.UTC()
	}

	o.BaseSalary = in.BaseSalary
	o.BonusTargetPct = in.BonusTargetPct
	o.BonusAmount = in.BonusAmount
	o.EquityType = equityType
	o.EquityTotalValue = in.EquityTotalValue
	o.EquityVestingYears = vestingYears
	o.SignOnBonus = in.SignOnBonus
	o.Currency = currency
	o.DeadlineAt = deadlineAt
	o.BenefitsNotes = strings.TrimSpace(in.BenefitsNotes)
	o.NegotiationNotes = strings.TrimSpace(in.NegotiationNotes)
	o.Status = status
	o.UpdatedAt = now.UTC()
	return nil
}

func (o Offer) AnnualBonus() int {
	if o.BonusAmount > 0 {
		return o.BonusAmount
	}
	if o.BaseSalary > 0 && o.BonusTargetPct > 0 {
		return (o.BaseSalary * o.BonusTargetPct) / 100
	}
	return 0
}

func (o Offer) AnnualEquity() int {
	if o.EquityTotalValue <= 0 {
		return 0
	}
	years := o.EquityVestingYears
	if years <= 0 {
		years = 4
	}
	return o.EquityTotalValue / years
}

func (o Offer) RecurringAnnualTotalComp() int {
	return o.BaseSalary + o.AnnualBonus() + o.AnnualEquity()
}

func (o Offer) Year1TotalComp() int {
	return o.RecurringAnnualTotalComp() + o.SignOnBonus
}

func (o Offer) DeltaVsCompFloor(compFloor int) int {
	return o.RecurringAnnualTotalComp() - compFloor
}

type OfferView struct {
	Offer                Offer
	CompanyID            string
	CompanyName          string
	OpportunityRoleTitle string
	OpportunityStage     Stage
	OpportunityOutcome   Outcome
	CompensationFloor    int
}

type OfferComparisonView struct {
	CompensationFloor  int
	TargetCompensation int
	ProfileCurrency    string
	Offers             []OfferView
}
