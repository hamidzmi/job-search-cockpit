package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type OfferDTO struct {
	ID                 string `json:"id"`
	OpportunityID      string `json:"opportunity_id"`
	BaseSalary         int    `json:"base_salary"`
	BonusTargetPct     int    `json:"bonus_target_pct"`
	BonusAmount        int    `json:"bonus_amount"`
	EquityType         string `json:"equity_type"`
	EquityTotalValue   int    `json:"equity_total_value"`
	EquityVestingYears int    `json:"equity_vesting_years"`
	SignOnBonus        int    `json:"sign_on_bonus"`
	Currency           string `json:"currency"`
	DeadlineAt         string `json:"deadline_at,omitempty"`
	BenefitsNotes      string `json:"benefits_notes"`
	NegotiationNotes   string `json:"negotiation_notes"`
	Status             string `json:"status"`
	AnnualBonus        int    `json:"annual_bonus"`
	AnnualEquity       int    `json:"annual_equity"`
	RecurringTotalComp int    `json:"recurring_total_comp"`
	Year1TotalComp     int    `json:"year1_total_comp"`
	DeltaVsCompFloor   int    `json:"delta_vs_comp_floor"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type OfferViewDTO struct {
	Offer                OfferDTO `json:"offer"`
	CompanyID            string   `json:"company_id"`
	CompanyName          string   `json:"company_name"`
	OpportunityRoleTitle string   `json:"opportunity_role_title"`
	OpportunityStage     string   `json:"opportunity_stage"`
	OpportunityOutcome   string   `json:"opportunity_outcome"`
	CompensationFloor    int      `json:"compensation_floor"`
}

type OfferComparisonDTO struct {
	CompensationFloor  int            `json:"compensation_floor"`
	TargetCompensation int            `json:"target_compensation"`
	ProfileCurrency    string         `json:"profile_currency"`
	Offers             []OfferViewDTO `json:"offers"`
}

type OfferUpsertRequest struct {
	BaseSalary         int    `json:"base_salary"`
	BonusTargetPct     int    `json:"bonus_target_pct"`
	BonusAmount        int    `json:"bonus_amount"`
	EquityType         string `json:"equity_type"`
	EquityTotalValue   int    `json:"equity_total_value"`
	EquityVestingYears int    `json:"equity_vesting_years"`
	SignOnBonus        int    `json:"sign_on_bonus"`
	Currency           string `json:"currency"`
	DeadlineAt         string `json:"deadline_at"`
	BenefitsNotes      string `json:"benefits_notes"`
	NegotiationNotes   string `json:"negotiation_notes"`
	Status             string `json:"status"`
}

func (r OfferUpsertRequest) toDomainInput() (domain.OfferInput, error) {
	var deadline time.Time
	if r.DeadlineAt != "" {
		parsed, err := time.Parse(time.RFC3339, r.DeadlineAt)
		if err != nil {
			parsedNano, errNano := time.Parse(time.RFC3339Nano, r.DeadlineAt)
			if errNano != nil {
				return domain.OfferInput{}, domain.ErrValidation
			}
			deadline = parsedNano
		} else {
			deadline = parsed
		}
	}

	return domain.OfferInput{
		BaseSalary:         r.BaseSalary,
		BonusTargetPct:     r.BonusTargetPct,
		BonusAmount:        r.BonusAmount,
		EquityType:         r.EquityType,
		EquityTotalValue:   r.EquityTotalValue,
		EquityVestingYears: r.EquityVestingYears,
		SignOnBonus:        r.SignOnBonus,
		Currency:           r.Currency,
		DeadlineAt:         deadline,
		BenefitsNotes:      r.BenefitsNotes,
		NegotiationNotes:   r.NegotiationNotes,
		Status:             domain.OfferStatus(r.Status),
	}, nil
}

func toOfferDTO(o domain.Offer, compFloor int) OfferDTO {
	var deadlineStr string
	if !o.DeadlineAt.IsZero() {
		deadlineStr = o.DeadlineAt.UTC().Format(time.RFC3339)
	}

	return OfferDTO{
		ID:                 o.ID,
		OpportunityID:      o.OpportunityID,
		BaseSalary:         o.BaseSalary,
		BonusTargetPct:     o.BonusTargetPct,
		BonusAmount:        o.BonusAmount,
		EquityType:         o.EquityType,
		EquityTotalValue:   o.EquityTotalValue,
		EquityVestingYears: o.EquityVestingYears,
		SignOnBonus:        o.SignOnBonus,
		Currency:           o.Currency,
		DeadlineAt:         deadlineStr,
		BenefitsNotes:      o.BenefitsNotes,
		NegotiationNotes:   o.NegotiationNotes,
		Status:             string(o.Status),
		AnnualBonus:        o.AnnualBonus(),
		AnnualEquity:       o.AnnualEquity(),
		RecurringTotalComp: o.RecurringAnnualTotalComp(),
		Year1TotalComp:     o.Year1TotalComp(),
		DeltaVsCompFloor:   o.DeltaVsCompFloor(compFloor),
		CreatedAt:          o.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:          o.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toOfferViewDTO(v domain.OfferView) OfferViewDTO {
	return OfferViewDTO{
		Offer:                toOfferDTO(v.Offer, v.CompensationFloor),
		CompanyID:            v.CompanyID,
		CompanyName:          v.CompanyName,
		OpportunityRoleTitle: v.OpportunityRoleTitle,
		OpportunityStage:     string(v.OpportunityStage),
		OpportunityOutcome:   string(v.OpportunityOutcome),
		CompensationFloor:    v.CompensationFloor,
	}
}

func toOfferComparisonDTO(c domain.OfferComparisonView) OfferComparisonDTO {
	offers := make([]OfferViewDTO, 0, len(c.Offers))
	for _, o := range c.Offers {
		offers = append(offers, toOfferViewDTO(o))
	}
	return OfferComparisonDTO{
		CompensationFloor:  c.CompensationFloor,
		TargetCompensation: c.TargetCompensation,
		ProfileCurrency:    c.ProfileCurrency,
		Offers:             offers,
	}
}
