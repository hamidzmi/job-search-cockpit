package usecase

import (
	"context"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type CompanyService struct {
	tx        outbound.TxManager
	companies outbound.CompanyRepository
	clock     outbound.Clock
}

func NewCompanyService(tx outbound.TxManager, companies outbound.CompanyRepository, clock outbound.Clock) *CompanyService {
	if clock == nil {
		clock = outbound.SystemClock{}
	}
	return &CompanyService{
		tx:        tx,
		companies: companies,
		clock:     clock,
	}
}

func (s *CompanyService) CreateCompany(ctx context.Context, in domain.CompanyInput) (domain.Company, error) {
	company, err := domain.NewCompany(uuid.NewString(), in, s.clock.Now())
	if err != nil {
		return domain.Company{}, err
	}
	if err := s.companies.SaveCompany(ctx, company); err != nil {
		return domain.Company{}, err
	}
	return company, nil
}

func (s *CompanyService) GetCompany(ctx context.Context, id string) (domain.Company, error) {
	return s.companies.GetCompany(ctx, id)
}

func (s *CompanyService) ListCompanies(ctx context.Context) ([]domain.Company, error) {
	return s.companies.ListCompanies(ctx)
}

func (s *CompanyService) UpdateCompany(ctx context.Context, id string, in domain.CompanyInput) (domain.Company, error) {
	var updated domain.Company
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		company, err := s.companies.GetCompany(txCtx, id)
		if err != nil {
			return err
		}
		if err := company.Update(in, s.clock.Now()); err != nil {
			return err
		}
		if err := s.companies.SaveCompany(txCtx, company); err != nil {
			return err
		}
		updated = company
		return nil
	})
	if err != nil {
		return domain.Company{}, err
	}
	return updated, nil
}
