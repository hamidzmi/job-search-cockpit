package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"job-search-cockpit/internal/domain"
)

func (s *Store) SaveOffer(ctx context.Context, offer domain.Offer) error {
	var deadlineStr string
	if !offer.DeadlineAt.IsZero() {
		deadlineStr = offer.DeadlineAt.UTC().Format(time.RFC3339Nano)
	}

	_, err := s.exec(ctx).ExecContext(ctx, `
		INSERT INTO offers (
			id, opportunity_id, base_salary, bonus_target_pct, bonus_amount,
			equity_type, equity_total_value, equity_vesting_years, sign_on_bonus,
			currency, deadline_at, benefits_notes, negotiation_notes, status,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(opportunity_id) DO UPDATE SET
			base_salary = excluded.base_salary,
			bonus_target_pct = excluded.bonus_target_pct,
			bonus_amount = excluded.bonus_amount,
			equity_type = excluded.equity_type,
			equity_total_value = excluded.equity_total_value,
			equity_vesting_years = excluded.equity_vesting_years,
			sign_on_bonus = excluded.sign_on_bonus,
			currency = excluded.currency,
			deadline_at = excluded.deadline_at,
			benefits_notes = excluded.benefits_notes,
			negotiation_notes = excluded.negotiation_notes,
			status = excluded.status,
			updated_at = excluded.updated_at
	`,
		offer.ID, offer.OpportunityID, offer.BaseSalary, offer.BonusTargetPct, offer.BonusAmount,
		offer.EquityType, offer.EquityTotalValue, offer.EquityVestingYears, offer.SignOnBonus,
		offer.Currency, deadlineStr, offer.BenefitsNotes, offer.NegotiationNotes, string(offer.Status),
		offer.CreatedAt.UTC().Format(time.RFC3339Nano), offer.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetOfferByOpportunityID(ctx context.Context, opportunityID string) (domain.Offer, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT id, opportunity_id, base_salary, bonus_target_pct, bonus_amount,
			equity_type, equity_total_value, equity_vesting_years, sign_on_bonus,
			currency, deadline_at, benefits_notes, negotiation_notes, status,
			created_at, updated_at
		FROM offers
		WHERE opportunity_id = ?
	`, opportunityID)

	var offer domain.Offer
	var statusStr, deadlineStr, createdStr, updatedStr string
	err := row.Scan(
		&offer.ID, &offer.OpportunityID, &offer.BaseSalary, &offer.BonusTargetPct, &offer.BonusAmount,
		&offer.EquityType, &offer.EquityTotalValue, &offer.EquityVestingYears, &offer.SignOnBonus,
		&offer.Currency, &deadlineStr, &offer.BenefitsNotes, &offer.NegotiationNotes, &statusStr,
		&createdStr, &updatedStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Offer{}, fmt.Errorf("%w: offer for opportunity %s", domain.ErrNotFound, opportunityID)
	}
	if err != nil {
		return domain.Offer{}, err
	}

	offer.Status = domain.OfferStatus(statusStr)
	if deadlineStr != "" {
		offer.DeadlineAt, _ = time.Parse(time.RFC3339Nano, deadlineStr)
	}
	offer.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	offer.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return offer, nil
}

func (s *Store) DeleteOfferByOpportunityID(ctx context.Context, opportunityID string) error {
	_, err := s.exec(ctx).ExecContext(ctx, `
		DELETE FROM offers WHERE opportunity_id = ?
	`, opportunityID)
	return err
}

func (s *Store) ListOffers(ctx context.Context) ([]domain.Offer, error) {
	rows, err := s.exec(ctx).QueryContext(ctx, `
		SELECT id, opportunity_id, base_salary, bonus_target_pct, bonus_amount,
			equity_type, equity_total_value, equity_vesting_years, sign_on_bonus,
			currency, deadline_at, benefits_notes, negotiation_notes, status,
			created_at, updated_at
		FROM offers
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Offer
	for rows.Next() {
		var offer domain.Offer
		var statusStr, deadlineStr, createdStr, updatedStr string
		err := rows.Scan(
			&offer.ID, &offer.OpportunityID, &offer.BaseSalary, &offer.BonusTargetPct, &offer.BonusAmount,
			&offer.EquityType, &offer.EquityTotalValue, &offer.EquityVestingYears, &offer.SignOnBonus,
			&offer.Currency, &deadlineStr, &offer.BenefitsNotes, &offer.NegotiationNotes, &statusStr,
			&createdStr, &updatedStr,
		)
		if err != nil {
			return nil, err
		}
		offer.Status = domain.OfferStatus(statusStr)
		if deadlineStr != "" {
			offer.DeadlineAt, _ = time.Parse(time.RFC3339Nano, deadlineStr)
		}
		offer.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
		offer.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
		result = append(result, offer)
	}
	if result == nil {
		result = []domain.Offer{}
	}
	return result, rows.Err()
}
