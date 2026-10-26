package usecase

import (
	"context"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type FlightDeckService struct {
	opps        outbound.OpportunityRepository
	transitions outbound.StageTransitionRepository
	rounds      outbound.InterviewRoundRepository
	actions     outbound.ActionItemRepository
	ixs         outbound.InteractionRepository
	contacts    outbound.ContactRepository
	stories     outbound.StoryRepository
	clock       outbound.Clock
}

func NewFlightDeckService(
	opps outbound.OpportunityRepository,
	transitions outbound.StageTransitionRepository,
	rounds outbound.InterviewRoundRepository,
	actions outbound.ActionItemRepository,
	ixs outbound.InteractionRepository,
	contacts outbound.ContactRepository,
	stories outbound.StoryRepository,
	clock outbound.Clock,
) *FlightDeckService {
	return &FlightDeckService{
		opps:        opps,
		transitions: transitions,
		rounds:      rounds,
		actions:     actions,
		ixs:         ixs,
		contacts:    contacts,
		stories:     stories,
		clock:       clock,
	}
}

func (s *FlightDeckService) GetFlightDeck(ctx context.Context) (domain.FlightDeck, error) {
	oppViews, err := s.opps.ListOpportunities(ctx, domain.OpportunityFilter{})
	if err != nil {
		return domain.FlightDeck{}, err
	}

	transitions, err := s.transitions.ListAllStageTransitions(ctx)
	if err != nil {
		return domain.FlightDeck{}, err
	}

	rounds, err := s.rounds.ListAllInterviewRounds(ctx)
	if err != nil {
		return domain.FlightDeck{}, err
	}

	actionItems, err := s.actions.ListActionItems(ctx, domain.ActionItemFilter{Status: domain.ActionItemStatusOpen})
	if err != nil {
		return domain.FlightDeck{}, err
	}

	ixViews, err := s.ixs.ListInteractions(ctx, domain.InteractionFilter{})
	if err != nil {
		return domain.FlightDeck{}, err
	}

	contactViews, err := s.contacts.ListContacts(ctx, domain.ContactFilter{})
	if err != nil {
		return domain.FlightDeck{}, err
	}
	contactNames := make(map[string]string, len(contactViews))
	for _, cv := range contactViews {
		contactNames[cv.Contact.ID] = cv.Contact.Name
	}

	stories, err := s.stories.ListStories(ctx, domain.StoryFilter{})
	if err != nil {
		return domain.FlightDeck{}, err
	}
	storyTitles := make(map[string]string, len(stories))
	for _, st := range stories {
		storyTitles[st.ID] = st.Title
	}

	interactions := make([]domain.Interaction, 0, len(ixViews))
	for _, iv := range ixViews {
		interactions = append(interactions, iv.Interaction)
	}

	input := domain.FlightDeckInput{
		Now:              s.clock.Now(),
		Opportunities:    oppViews,
		StageTransitions: transitions,
		Interactions:     interactions,
		InterviewRounds:  rounds,
		ActionItems:      actionItems,
		ContactNames:     contactNames,
		StoryTitles:      storyTitles,
	}

	return domain.EvaluateFlightDeck(input), nil
}
