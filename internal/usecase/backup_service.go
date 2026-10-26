package usecase

import (
	"context"

	"job-search-cockpit/internal/usecase/port/inbound"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type BackupService struct {
	backupPort outbound.BackupPort
}

func NewBackupService(backupPort outbound.BackupPort) *BackupService {
	return &BackupService{backupPort: backupPort}
}

var _ inbound.BackupUseCase = (*BackupService)(nil)

func (s *BackupService) ExportJSONBackup(ctx context.Context) ([]byte, error) {
	return s.backupPort.ExportJSONSnapshot(ctx)
}

func (s *BackupService) ExportSQLiteBackup(ctx context.Context) ([]byte, error) {
	return s.backupPort.ExportSQLiteSnapshot(ctx)
}
