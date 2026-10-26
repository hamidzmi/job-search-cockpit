package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type StoryRepository interface {
	SaveStory(ctx context.Context, story domain.Story) error
	GetStory(ctx context.Context, id string) (domain.Story, error)
	ListStories(ctx context.Context, filter domain.StoryFilter) ([]domain.Story, error)
	DeleteStory(ctx context.Context, id string) error
}
