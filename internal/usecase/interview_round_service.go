package usecase

import (
	"context"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type InterviewRoundService struct {
	tx       outbound.TxManager
	rounds   outbound.InterviewRoundRepository
	opps     outbound.OpportunityRepository
	contacts outbound.ContactRepository
	stories  outbound.StoryRepository
	clock    outbound.Clock
}

func NewInterviewRoundService(
	tx outbound.TxManager,
	rounds outbound.InterviewRoundRepository,
	opps outbound.OpportunityRepository,
	contacts outbound.ContactRepository,
	stories outbound.StoryRepository,
	clock outbound.Clock,
) *InterviewRoundService {
	return &InterviewRoundService{
		tx:       tx,
		rounds:   rounds,
		opps:     opps,
		contacts: contacts,
		stories:  stories,
		clock:    clock,
	}
}

func (s *InterviewRoundService) CreateInterviewRound(ctx context.Context, opportunityID string, in domain.InterviewRoundInput) (domain.InterviewRoundView, error) {
	var createdView domain.InterviewRoundView

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		now := s.clock.Now()
		in.OpportunityID = opportunityID
		round, err := domain.NewInterviewRound(uuid.NewString(), in, now)
		if err != nil {
			return err
		}
		opp, err := s.opps.GetOpportunity(txCtx, round.OpportunityID)
		if err != nil {
			return err
		}
		if err := s.validateLinkedIDs(txCtx, round.ContactIDs, round.StoryIDs); err != nil {
			return err
		}
		if err := s.rounds.SaveInterviewRound(txCtx, round); err != nil {
			return err
		}
		if round.Status != domain.InterviewRoundStatusCancelled {
			opp.RecordActivity(now, now)
			if err := s.opps.SaveOpportunity(txCtx, opp); err != nil {
				return err
			}
		}
		views, err := s.listViewsForOpportunity(txCtx, round.OpportunityID)
		if err != nil {
			return err
		}
		for _, v := range views {
			if v.InterviewRound.ID == round.ID {
				createdView = v
				break
			}
		}
		return nil
	})
	if err != nil {
		return domain.InterviewRoundView{}, err
	}
	return createdView, nil
}

func (s *InterviewRoundService) UpdateInterviewRound(ctx context.Context, id string, in domain.InterviewRoundInput) (domain.InterviewRoundView, error) {
	var updatedView domain.InterviewRoundView

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		now := s.clock.Now()
		existing, err := s.rounds.GetInterviewRound(txCtx, id)
		if err != nil {
			return err
		}
		prevStatus := existing.Status
		prevScheduled := existing.ScheduledAt

		in.OpportunityID = existing.OpportunityID
		if err := existing.Update(in, now); err != nil {
			return err
		}
		if err := s.validateLinkedIDs(txCtx, existing.ContactIDs, existing.StoryIDs); err != nil {
			return err
		}
		if err := s.rounds.SaveInterviewRound(txCtx, existing); err != nil {
			return err
		}

		if existing.Status != domain.InterviewRoundStatusCancelled && (existing.Status != prevStatus || !existing.ScheduledAt.Equal(prevScheduled)) {
			opp, err := s.opps.GetOpportunity(txCtx, existing.OpportunityID)
			if err != nil {
				return err
			}
			opp.RecordActivity(now, now)
			if err := s.opps.SaveOpportunity(txCtx, opp); err != nil {
				return err
			}
		}

		views, err := s.listViewsForOpportunity(txCtx, existing.OpportunityID)
		if err != nil {
			return err
		}
		for _, v := range views {
			if v.InterviewRound.ID == existing.ID {
				updatedView = v
				break
			}
		}
		return nil
	})
	if err != nil {
		return domain.InterviewRoundView{}, err
	}
	return updatedView, nil
}

func (s *InterviewRoundService) DeleteInterviewRound(ctx context.Context, id string) error {
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.rounds.DeleteInterviewRound(txCtx, id)
	})
}

func (s *InterviewRoundService) ListInterviewRounds(ctx context.Context, opportunityID string) ([]domain.InterviewRoundView, error) {
	if _, err := s.opps.GetOpportunity(ctx, opportunityID); err != nil {
		return nil, err
	}
	return s.listViewsForOpportunity(ctx, opportunityID)
}

func (s *InterviewRoundService) RecordDebrief(ctx context.Context, roundID string, in domain.DebriefInput) (domain.InterviewRoundView, error) {
	var updatedView domain.InterviewRoundView

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		now := s.clock.Now()
		existing, err := s.rounds.GetInterviewRound(txCtx, roundID)
		if err != nil {
			return err
		}
		if err := existing.RecordDebrief(in, now); err != nil {
			return err
		}
		for _, q := range existing.Debrief.QuestionsAsked {
			if q.AnsweredWithStoryID != "" {
				if _, err := s.stories.GetStory(txCtx, q.AnsweredWithStoryID); err != nil {
					return err
				}
			}
		}
		if err := s.rounds.SaveInterviewRound(txCtx, existing); err != nil {
			return err
		}
		views, err := s.listViewsForOpportunity(txCtx, existing.OpportunityID)
		if err != nil {
			return err
		}
		for _, v := range views {
			if v.InterviewRound.ID == existing.ID {
				updatedView = v
				break
			}
		}
		return nil
	})
	if err != nil {
		return domain.InterviewRoundView{}, err
	}
	return updatedView, nil
}

func (s *InterviewRoundService) ListQuestionBank(ctx context.Context, filter domain.QuestionBankFilter) ([]domain.QuestionBankEntryView, error) {
	return s.rounds.ListQuestionBankEntries(ctx, filter)
}

func (s *InterviewRoundService) validateLinkedIDs(ctx context.Context, contactIDs, storyIDs []string) error {
	for _, cid := range contactIDs {
		if _, err := s.contacts.GetContact(ctx, cid); err != nil {
			return err
		}
	}
	for _, sid := range storyIDs {
		if _, err := s.stories.GetStory(ctx, sid); err != nil {
			return err
		}
	}
	return nil
}

func (s *InterviewRoundService) listViewsForOpportunity(ctx context.Context, opportunityID string) ([]domain.InterviewRoundView, error) {
	rounds, err := s.rounds.ListInterviewRoundsByOpportunity(ctx, opportunityID)
	if err != nil {
		return nil, err
	}
	repeatedWarnings := domain.DetectRepeatedStories(rounds)

	allContacts, err := s.contacts.ListContacts(ctx, domain.ContactFilter{})
	if err != nil {
		return nil, err
	}
	contactsByID := make(map[string]domain.ContactView, len(allContacts))
	for _, c := range allContacts {
		contactsByID[c.Contact.ID] = c
	}

	allStories, err := s.stories.ListStories(ctx, domain.StoryFilter{})
	if err != nil {
		return nil, err
	}
	storiesByID := make(map[string]domain.Story, len(allStories))
	for _, st := range allStories {
		storiesByID[st.ID] = st
	}

	views := make([]domain.InterviewRoundView, 0, len(rounds))
	for _, r := range rounds {
		contactViews := make([]domain.InterviewRoundContactView, 0, len(r.ContactIDs))
		for _, cid := range r.ContactIDs {
			cv, ok := contactsByID[cid]
			if !ok {
				continue
			}
			contactViews = append(contactViews, domain.InterviewRoundContactView{
				ContactID:   cv.Contact.ID,
				Name:        cv.Contact.Name,
				Title:       cv.Contact.Title,
				Email:       cv.Contact.Email,
				Warmth:      cv.Contact.Warmth,
				CompanyName: cv.CompanyName,
			})
		}

		storyViews := make([]domain.InterviewRoundStoryView, 0, len(r.StoryIDs))
		roundWarnings := repeatedWarnings[r.ID]
		for _, sid := range r.StoryIDs {
			st, ok := storiesByID[sid]
			if !ok {
				continue
			}
			priors := roundWarnings[sid]
			if priors == nil {
				priors = []domain.StoryPriorRoundRef{}
			}
			storyViews = append(storyViews, domain.InterviewRoundStoryView{
				StoryID:             st.ID,
				Title:               st.Title,
				Hook:                st.Hook,
				CompetencyTags:      st.CompetencyTags,
				SkillTags:           st.SkillTags,
				UsedInEarlierRounds: priors,
			})
		}

		var debriefView *domain.DebriefView
		if r.Debrief != nil {
			qViews := make([]domain.DebriefQuestionView, 0, len(r.Debrief.QuestionsAsked))
			for _, q := range r.Debrief.QuestionsAsked {
				var storyTitle string
				if q.AnsweredWithStoryID != "" {
					if st, ok := storiesByID[q.AnsweredWithStoryID]; ok {
						storyTitle = st.Title
					}
				}
				qViews = append(qViews, domain.DebriefQuestionView{
					Question:               q.Question,
					Category:               q.Category,
					AnsweredWithStoryID:    q.AnsweredWithStoryID,
					AnsweredWithStoryTitle: storyTitle,
					Notes:                  q.Notes,
				})
			}
			debriefView = &domain.DebriefView{
				QuestionsAsked:  qViews,
				WentWell:        r.Debrief.WentWell,
				ToImprove:       r.Debrief.ToImprove,
				Signals:         r.Debrief.Signals,
				ConfidenceScore: r.Debrief.ConfidenceScore,
				FollowUpNotes:   r.Debrief.FollowUpNotes,
				UpdatedAt:       r.Debrief.UpdatedAt,
			}
		}

		views = append(views, domain.InterviewRoundView{
			InterviewRound: r,
			Contacts:       contactViews,
			Stories:        storyViews,
			DebriefView:    debriefView,
		})
	}

	return views, nil
}
