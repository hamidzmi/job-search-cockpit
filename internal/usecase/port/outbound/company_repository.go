package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type CompanyRepository interface {
	SaveCompany(ctx context.Context, c domain.Company) error
	GetCompany(ctx context.Context, id string) (domain.Company, error)
	ListCompanies(ctx context.Context) ([]domain.Company, error)
}
