package usecase

import (
	"context"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type OpportunityService struct {
	tx          outbound.TxManager
	companies   outbound.CompanyRepository
	opps        outbound.OpportunityRepository
	transitions outbound.StageTransitionRepository
	clock       outbound.Clock
}

func NewOpportunityService(
	tx outbound.TxManager,
	companies outbound.CompanyRepository,
	opps outbound.OpportunityRepository,
	transitions outbound.StageTransitionRepository,
	clock outbound.Clock,
) *OpportunityService {
	if clock == nil {
		clock = outbound.SystemClock{}
	}
	return &OpportunityService{
		tx:          tx,
		companies:   companies,
		opps:        opps,
		transitions: transitions,
		clock:       clock,
	}
}

func (s *OpportunityService) CreateOpportunity(ctx context.Context, in domain.OpportunityInput) (domain.Opportunity, error) {
	var created domain.Opportunity
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		company, err := s.companies.GetCompany(txCtx, in.CompanyID)
		if err != nil {
			return err
		}
		opp, tr, err := domain.NewOpportunity(uuid.NewString(), uuid.NewString(), in, s.clock.Now())
		if err != nil {
			return err
		}
		opp.CompanyName = company.Name
		if err := s.opps.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		if err := s.transitions.AppendStageTransition(txCtx, tr); err != nil {
			return err
		}
		created = opp
		return nil
	})
	if err != nil {
		return domain.Opportunity{}, err
	}
	return created, nil
}

func (s *OpportunityService) GetOpportunity(ctx context.Context, id string) (domain.Opportunity, []domain.StageTransition, error) {
	opp, err := s.opps.GetOpportunity(ctx, id)
	if err != nil {
		return domain.Opportunity{}, nil, err
	}
	history, err := s.transitions.ListStageTransitions(ctx, id)
	if err != nil {
		return domain.Opportunity{}, nil, err
	}
	return opp, history, nil
}

func (s *OpportunityService) ListOpportunities(ctx context.Context, filter domain.OpportunityFilter) ([]domain.Opportunity, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	return s.opps.ListOpportunities(ctx, filter)
}

func (s *OpportunityService) UpdateOpportunityDetails(ctx context.Context, id string, in domain.OpportunityDetailsInput) (domain.Opportunity, error) {
	var updated domain.Opportunity
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		opp, err := s.opps.GetOpportunity(txCtx, id)
		if err != nil {
			return err
		}
		if err := opp.UpdateDetails(in, s.clock.Now()); err != nil {
			return err
		}
		company, err := s.companies.GetCompany(txCtx, opp.CompanyID)
		if err != nil {
			return err
		}
		opp.CompanyName = company.Name
		if err := s.opps.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		updated = opp
		return nil
	})
	if err != nil {
		return domain.Opportunity{}, err
	}
	return updated, nil
}

func (s *OpportunityService) TransitionOpportunity(ctx context.Context, id string, toStage domain.Stage, toOutcome domain.Outcome, note string) (domain.Opportunity, error) {
	var updated domain.Opportunity
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		opp, err := s.opps.GetOpportunity(txCtx, id)
		if err != nil {
			return err
		}
		tr, err := opp.Transition(uuid.NewString(), toStage, toOutcome, note, s.clock.Now())
		if err != nil {
			return err
		}
		if err := s.opps.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		if err := s.transitions.AppendStageTransition(txCtx, tr); err != nil {
			return err
		}
		updated = opp
		return nil
	})
	if err != nil {
		return domain.Opportunity{}, err
	}
	return updated, nil
}

func (s *OpportunityService) ListStageTransitions(ctx context.Context, opportunityID string) ([]domain.StageTransition, error) {
	if _, err := s.opps.GetOpportunity(ctx, opportunityID); err != nil {
		return nil, err
	}
	return s.transitions.ListStageTransitions(ctx, opportunityID)
}
