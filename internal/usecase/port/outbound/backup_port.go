package outbound

import (
	"context"
)

type BackupPort interface {
	ExportJSONSnapshot(ctx context.Context) ([]byte, error)
	ExportSQLiteSnapshot(ctx context.Context) ([]byte, error)
}
