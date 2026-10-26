package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type InterviewRoundUseCase interface {
	CreateInterviewRound(ctx context.Context, opportunityID string, in domain.InterviewRoundInput) (domain.InterviewRoundView, error)
	UpdateInterviewRound(ctx context.Context, id string, in domain.InterviewRoundInput) (domain.InterviewRoundView, error)
	DeleteInterviewRound(ctx context.Context, id string) error
	ListInterviewRounds(ctx context.Context, opportunityID string) ([]domain.InterviewRoundView, error)
	RecordDebrief(ctx context.Context, roundID string, in domain.DebriefInput) (domain.InterviewRoundView, error)
	ListQuestionBank(ctx context.Context, filter domain.QuestionBankFilter) ([]domain.QuestionBankEntryView, error)
}
