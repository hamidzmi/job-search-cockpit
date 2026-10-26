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
	return &OpportunityService{
		tx:          tx,
		companies:   companies,
		opps:        opps,
		transitions: transitions,
		clock:       clock,
	}
}

func (s *OpportunityService) CreateOpportunity(ctx context.Context, in domain.OpportunityInput) (domain.OpportunityView, error) {
	var created domain.OpportunityView
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		company, err := s.companies.GetCompany(txCtx, in.CompanyID)
		if err != nil {
			return err
		}
		opp, tr, err := domain.NewOpportunity(uuid.NewString(), uuid.NewString(), in, s.clock.Now())
		if err != nil {
			return err
		}
		if err := s.opps.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		if err := s.transitions.AppendStageTransition(txCtx, tr); err != nil {
			return err
		}
		created = domain.OpportunityView{
			Opportunity: opp,
			CompanyName: company.Name,
		}
		return nil
	})
	if err != nil {
		return domain.OpportunityView{}, err
	}
	return created, nil
}

func (s *OpportunityService) GetOpportunity(ctx context.Context, id string) (domain.OpportunityView, []domain.StageTransition, error) {
	view, err := s.opps.GetOpportunityView(ctx, id)
	if err != nil {
		return domain.OpportunityView{}, nil, err
	}
	stageTransitions, err := s.transitions.ListStageTransitions(ctx, id)
	if err != nil {
		return domain.OpportunityView{}, nil, err
	}
	return view, stageTransitions, nil
}

func (s *OpportunityService) ListOpportunities(ctx context.Context, filter domain.OpportunityFilter) ([]domain.OpportunityView, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	return s.opps.ListOpportunities(ctx, filter)
}

func (s *OpportunityService) UpdateOpportunityDetails(ctx context.Context, id string, in domain.OpportunityDetailsInput) (domain.OpportunityView, error) {
	var updated domain.OpportunityView
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		opp, err := s.opps.GetOpportunity(txCtx, id)
		if err != nil {
			return err
		}
		if err := opp.UpdateDetails(in, s.clock.Now()); err != nil {
			return err
		}
		if err := s.opps.SaveOpportunity(txCtx, opp); err != nil {
			return err
		}
		view, err := s.opps.GetOpportunityView(txCtx, id)
		if err != nil {
			return err
		}
		updated = view
		return nil
	})
	if err != nil {
		return domain.OpportunityView{}, err
	}
	return updated, nil
}

func (s *OpportunityService) TransitionOpportunity(ctx context.Context, id string, toStage domain.Stage, toOutcome domain.Outcome, note string) (domain.OpportunityView, error) {
	var updated domain.OpportunityView
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
		view, err := s.opps.GetOpportunityView(txCtx, id)
		if err != nil {
			return err
		}
		updated = view
		return nil
	})
	if err != nil {
		return domain.OpportunityView{}, err
	}
	return updated, nil
}

func (s *OpportunityService) ListStageTransitions(ctx context.Context, opportunityID string) ([]domain.StageTransition, error) {
	if _, err := s.opps.GetOpportunity(ctx, opportunityID); err != nil {
		return nil, err
	}
	return s.transitions.ListStageTransitions(ctx, opportunityID)
}
