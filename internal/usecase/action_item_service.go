package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type ActionItemService struct {
	tx          outbound.TxManager
	contacts    outbound.ContactRepository
	opps        outbound.OpportunityRepository
	actionItems outbound.ActionItemRepository
	clock       outbound.Clock
}

func NewActionItemService(
	tx outbound.TxManager,
	contacts outbound.ContactRepository,
	opps outbound.OpportunityRepository,
	actionItems outbound.ActionItemRepository,
	clock outbound.Clock,
) *ActionItemService {
	return &ActionItemService{
		tx:          tx,
		contacts:    contacts,
		opps:        opps,
		actionItems: actionItems,
		clock:       clock,
	}
}

func (s *ActionItemService) CreateActionItem(ctx context.Context, in domain.ActionItemInput) (domain.ActionItemView, error) {
	var created domain.ActionItemView
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if oppID := strings.TrimSpace(in.OpportunityID); oppID != "" {
			if _, err := s.opps.GetOpportunity(txCtx, oppID); err != nil {
				return err
			}
		}
		if cntID := strings.TrimSpace(in.ContactID); cntID != "" {
			if _, err := s.contacts.GetContact(txCtx, cntID); err != nil {
				return err
			}
		}
		item, err := domain.NewActionItem(uuid.NewString(), in, s.clock.Now())
		if err != nil {
			return err
		}
		if err := s.actionItems.SaveActionItem(txCtx, item); err != nil {
			return err
		}
		view, err := s.actionItems.GetActionItemView(txCtx, item.ID)
		if err != nil {
			return err
		}
		created = view
		return nil
	})
	if err != nil {
		return domain.ActionItemView{}, err
	}
	return created, nil
}

func (s *ActionItemService) UpdateActionItem(ctx context.Context, id string, in domain.ActionItemUpdateInput) (domain.ActionItemView, error) {
	var updated domain.ActionItemView
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		item, err := s.actionItems.GetActionItem(txCtx, id)
		if err != nil {
			return err
		}
		if oppID := strings.TrimSpace(in.OpportunityID); oppID != "" {
			if _, err := s.opps.GetOpportunity(txCtx, oppID); err != nil {
				return err
			}
		}
		if cntID := strings.TrimSpace(in.ContactID); cntID != "" {
			if _, err := s.contacts.GetContact(txCtx, cntID); err != nil {
				return err
			}
		}
		if err := item.Update(in, s.clock.Now()); err != nil {
			return err
		}
		if err := s.actionItems.SaveActionItem(txCtx, item); err != nil {
			return err
		}
		view, err := s.actionItems.GetActionItemView(txCtx, item.ID)
		if err != nil {
			return err
		}
		updated = view
		return nil
	})
	if err != nil {
		return domain.ActionItemView{}, err
	}
	return updated, nil
}

func (s *ActionItemService) ListActionItems(ctx context.Context, filter domain.ActionItemFilter) ([]domain.ActionItemView, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	return s.actionItems.ListActionItems(ctx, filter)
}
