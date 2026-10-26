package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type CandidateProfileRepository interface {
	GetCandidateProfile(ctx context.Context) (domain.CandidateProfile, error)
	SaveCandidateProfile(ctx context.Context, profile domain.CandidateProfile) error
}
