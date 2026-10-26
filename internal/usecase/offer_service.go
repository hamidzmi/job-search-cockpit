package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type OfferService struct {
	tx          outbound.TxManager
	offerRepo   outbound.OfferRepository
	oppRepo     outbound.OpportunityRepository
	profileRepo outbound.CandidateProfileRepository
	clock       outbound.Clock
}

func NewOfferService(
	tx outbound.TxManager,
	offerRepo outbound.OfferRepository,
	oppRepo outbound.OpportunityRepository,
	profileRepo outbound.CandidateProfileRepository,
	clock outbound.Clock,
) *OfferService {
	return &OfferService{
		tx:          tx,
		offerRepo:   offerRepo,
		oppRepo:     oppRepo,
		profileRepo: profileRepo,
		clock:       clock,
	}
}

var _ inbound.OfferUseCase = (*OfferService)(nil)

func (s *OfferService) GetOffer(ctx context.Context, opportunityID string) (domain.OfferView, error) {
	oppView, err := s.oppRepo.GetOpportunityView(ctx, opportunityID)
	if err != nil {
		return domain.OfferView{}, err
	}

	offer, err := s.offerRepo.GetOfferByOpportunityID(ctx, opportunityID)
	if err != nil {
		return domain.OfferView{}, err
	}

	profile, _ := s.profileRepo.GetCandidateProfile(ctx)

	return domain.OfferView{
		Offer:                offer,
		CompanyID:            oppView.Opportunity.CompanyID,
		CompanyName:          oppView.CompanyName,
		OpportunityRoleTitle: oppView.Opportunity.RoleTitle,
		OpportunityStage:     oppView.Opportunity.Stage,
		OpportunityOutcome:   oppView.Opportunity.Outcome,
		CompensationFloor:    profile.CompensationFloor,
	}, nil
}

func (s *OfferService) UpsertOffer(ctx context.Context, opportunityID string, in domain.OfferInput) (domain.OfferView, error) {
	now := s.clock.Now().UTC()
	var finalOffer domain.Offer
	var finalOppView domain.OpportunityView
	var finalCompFloor int

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		oppView, err := s.oppRepo.GetOpportunityView(txCtx, opportunityID)
		if err != nil {
			return err
		}
		finalOppView = oppView

		in.OpportunityID = opportunityID
		existing, err := s.offerRepo.GetOfferByOpportunityID(txCtx, opportunityID)
		if err == nil {
			if err := existing.Update(in, now); err != nil {
				return err
			}
			if err := s.offerRepo.SaveOffer(txCtx, existing); err != nil {
				return err
			}
			finalOffer = existing
		} else if errors.Is(err, domain.ErrNotFound) {
			newOffer, err := domain.NewOffer(uuid.NewString(), in, now)
			if err != nil {
				return err
			}
			if err := s.offerRepo.SaveOffer(txCtx, newOffer); err != nil {
				return err
			}
			finalOffer = newOffer
		} else {
			return err
		}

		profile, _ := s.profileRepo.GetCandidateProfile(txCtx)
		finalCompFloor = profile.CompensationFloor
		return nil
	})
	if err != nil {
		return domain.OfferView{}, err
	}

	return domain.OfferView{
		Offer:                finalOffer,
		CompanyID:            finalOppView.Opportunity.CompanyID,
		CompanyName:          finalOppView.CompanyName,
		OpportunityRoleTitle: finalOppView.Opportunity.RoleTitle,
		OpportunityStage:     finalOppView.Opportunity.Stage,
		OpportunityOutcome:   finalOppView.Opportunity.Outcome,
		CompensationFloor:    finalCompFloor,
	}, nil
}

func (s *OfferService) DeleteOffer(ctx context.Context, opportunityID string) error {
	return s.offerRepo.DeleteOfferByOpportunityID(ctx, opportunityID)
}

func (s *OfferService) GetOfferComparison(ctx context.Context) (domain.OfferComparisonView, error) {
	profile, _ := s.profileRepo.GetCandidateProfile(ctx)
	offers, err := s.offerRepo.ListOffers(ctx)
	if err != nil {
		return domain.OfferComparisonView{}, err
	}

	views := make([]domain.OfferView, 0, len(offers))
	for _, o := range offers {
		oppView, err := s.oppRepo.GetOpportunityView(ctx, o.OpportunityID)
		if err != nil {
			continue
		}
		views = append(views, domain.OfferView{
			Offer:                o,
			CompanyID:            oppView.Opportunity.CompanyID,
			CompanyName:          oppView.CompanyName,
			OpportunityRoleTitle: oppView.Opportunity.RoleTitle,
			OpportunityStage:     oppView.Opportunity.Stage,
			OpportunityOutcome:   oppView.Opportunity.Outcome,
			CompensationFloor:    profile.CompensationFloor,
		})
	}

	currency := profile.Currency
	if currency == "" {
		currency = "USD"
	}

	return domain.OfferComparisonView{
		CompensationFloor:  profile.CompensationFloor,
		TargetCompensation: profile.TargetCompensation,
		ProfileCurrency:    currency,
		Offers:             views,
	}, nil
}
