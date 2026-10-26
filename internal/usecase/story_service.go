package usecase

import (
	"context"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type StoryService struct {
	tx      outbound.TxManager
	stories outbound.StoryRepository
	clock   outbound.Clock
}

func NewStoryService(
	tx outbound.TxManager,
	stories outbound.StoryRepository,
	clock outbound.Clock,
) *StoryService {
	return &StoryService{
		tx:      tx,
		stories: stories,
		clock:   clock,
	}
}

func (s *StoryService) CreateStory(ctx context.Context, in domain.StoryInput) (domain.Story, error) {
	story, err := domain.NewStory(uuid.NewString(), in, s.clock.Now())
	if err != nil {
		return domain.Story{}, err
	}
	if err := s.stories.SaveStory(ctx, story); err != nil {
		return domain.Story{}, err
	}
	return story, nil
}

func (s *StoryService) GetStory(ctx context.Context, id string) (domain.Story, error) {
	return s.stories.GetStory(ctx, id)
}

func (s *StoryService) ListStories(ctx context.Context, filter domain.StoryFilter) ([]domain.Story, error) {
	return s.stories.ListStories(ctx, filter)
}

func (s *StoryService) UpdateStory(ctx context.Context, id string, in domain.StoryInput) (domain.Story, error) {
	var updated domain.Story
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		existing, err := s.stories.GetStory(txCtx, id)
		if err != nil {
			return err
		}
		if err := existing.Update(in, s.clock.Now()); err != nil {
			return err
		}
		if err := s.stories.SaveStory(txCtx, existing); err != nil {
			return err
		}
		updated = existing
		return nil
	})
	if err != nil {
		return domain.Story{}, err
	}
	return updated, nil
}

func (s *StoryService) DeleteStory(ctx context.Context, id string) error {
	return s.stories.DeleteStory(ctx, id)
}
