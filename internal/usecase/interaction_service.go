package usecase

import (
	"context"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type InteractionService struct {
	tx           outbound.TxManager
	contacts     outbound.ContactRepository
	opps         outbound.OpportunityRepository
	interactions outbound.InteractionRepository
	actionItems  outbound.ActionItemRepository
	clock        outbound.Clock
}

func NewInteractionService(
	tx outbound.TxManager,
	contacts outbound.ContactRepository,
	opps outbound.OpportunityRepository,
	interactions outbound.InteractionRepository,
	actionItems outbound.ActionItemRepository,
	clock outbound.Clock,
) *InteractionService {
	return &InteractionService{
		tx:           tx,
		contacts:     contacts,
		opps:         opps,
		interactions: interactions,
		actionItems:  actionItems,
		clock:        clock,
	}
}

func (s *InteractionService) LogInteraction(ctx context.Context, in domain.InteractionInput) (domain.InteractionView, *domain.ActionItemView, error) {
	var createdInteraction domain.InteractionView
	var createdFollowUp *domain.ActionItemView

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		now := s.clock.Now()
		interaction, err := domain.NewInteraction(uuid.NewString(), in, now)
		if err != nil {
			return err
		}
		if _, err := s.contacts.GetContact(txCtx, interaction.ContactID); err != nil {
			return err
		}
		if interaction.OpportunityID != "" {
			opp, err := s.opps.GetOpportunity(txCtx, interaction.OpportunityID)
			if err != nil {
				return err
			}
			opp.RecordActivity(interaction.OccurredAt, now)
			if err := s.opps.SaveOpportunity(txCtx, opp); err != nil {
				return err
			}
		}
		if err := s.interactions.SaveInteraction(txCtx, interaction); err != nil {
			return err
		}
		intView, err := s.interactions.GetInteractionView(txCtx, interaction.ID)
		if err != nil {
			return err
		}
		createdInteraction = intView

		if in.FollowUp != nil {
			item, err := domain.NewActionItem(uuid.NewString(), domain.ActionItemInput{
				Title:         in.FollowUp.Title,
				OpportunityID: interaction.OpportunityID,
				ContactID:     interaction.ContactID,
				DueAt:         in.FollowUp.DueAt,
				Notes:         in.FollowUp.Notes,
			}, now)
			if err != nil {
				return err
			}
			if err := s.actionItems.SaveActionItem(txCtx, item); err != nil {
				return err
			}
			itemView, err := s.actionItems.GetActionItemView(txCtx, item.ID)
			if err != nil {
				return err
			}
			createdFollowUp = &itemView
		}

		return nil
	})
	if err != nil {
		return domain.InteractionView{}, nil, err
	}
	return createdInteraction, createdFollowUp, nil
}

func (s *InteractionService) ListInteractions(ctx context.Context, filter domain.InteractionFilter) ([]domain.InteractionView, error) {
	return s.interactions.ListInteractions(ctx, filter)
}
