package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type CandidateProfileUseCase interface {
	GetCandidateProfile(ctx context.Context) (domain.CandidateProfile, error)
	SaveCandidateProfile(ctx context.Context, in domain.CandidateProfileInput) (domain.CandidateProfile, error)
}
