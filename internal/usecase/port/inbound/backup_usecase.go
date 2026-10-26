package inbound

import (
	"context"
)

type BackupUseCase interface {
	ExportJSONBackup(ctx context.Context) ([]byte, error)
	ExportSQLiteBackup(ctx context.Context) ([]byte, error)
}
