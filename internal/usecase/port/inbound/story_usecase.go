package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type StoryUseCase interface {
	CreateStory(ctx context.Context, in domain.StoryInput) (domain.Story, error)
	GetStory(ctx context.Context, id string) (domain.Story, error)
	ListStories(ctx context.Context, filter domain.StoryFilter) ([]domain.Story, error)
	UpdateStory(ctx context.Context, id string, in domain.StoryInput) (domain.Story, error)
	DeleteStory(ctx context.Context, id string) error
}
