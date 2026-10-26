package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type JobPageParser interface {
	ParseJobPage(ctx context.Context, rawURL string) (domain.ParsedJobPage, error)
}
