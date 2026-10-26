package inbound

import (
	"context"
)

type SeedUseCase interface {
	SeedSampleData(ctx context.Context) error
}
