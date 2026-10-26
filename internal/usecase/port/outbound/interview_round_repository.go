package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type InterviewRoundRepository interface {
	SaveInterviewRound(ctx context.Context, round domain.InterviewRound) error
	GetInterviewRound(ctx context.Context, id string) (domain.InterviewRound, error)
	ListInterviewRoundsByOpportunity(ctx context.Context, opportunityID string) ([]domain.InterviewRound, error)
	ListAllInterviewRounds(ctx context.Context) ([]domain.InterviewRound, error)
	DeleteInterviewRound(ctx context.Context, id string) error
	ListQuestionBankEntries(ctx context.Context, filter domain.QuestionBankFilter) ([]domain.QuestionBankEntryView, error)
}
