package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type CompanyUseCase interface {
	CreateCompany(ctx context.Context, in domain.CompanyInput) (domain.Company, error)
	GetCompany(ctx context.Context, id string) (domain.Company, error)
	ListCompanies(ctx context.Context) ([]domain.Company, error)
	UpdateCompany(ctx context.Context, id string, in domain.CompanyInput) (domain.Company, error)
}
