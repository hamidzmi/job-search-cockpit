package usecase

import (
	"context"
	"errors"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type CandidateProfileService struct {
	tx       outbound.TxManager
	profiles outbound.CandidateProfileRepository
	clock    outbound.Clock
}

func NewCandidateProfileService(
	tx outbound.TxManager,
	profiles outbound.CandidateProfileRepository,
	clock outbound.Clock,
) *CandidateProfileService {
	return &CandidateProfileService{
		tx:       tx,
		profiles: profiles,
		clock:    clock,
	}
}

func (s *CandidateProfileService) GetCandidateProfile(ctx context.Context) (domain.CandidateProfile, error) {
	profile, err := s.profiles.GetCandidateProfile(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.DefaultCandidateProfile(), nil
	}
	return profile, err
}

func (s *CandidateProfileService) SaveCandidateProfile(ctx context.Context, in domain.CandidateProfileInput) (domain.CandidateProfile, error) {
	profile, err := domain.NewCandidateProfile(in, s.clock.Now())
	if err != nil {
		return domain.CandidateProfile{}, err
	}
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.profiles.SaveCandidateProfile(txCtx, profile)
	})
	if err != nil {
		return domain.CandidateProfile{}, err
	}
	return profile, nil
}
