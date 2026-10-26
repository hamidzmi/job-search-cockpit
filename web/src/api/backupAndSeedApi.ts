import { requestJSON } from './httpClient';

export interface BackupAndSeedApi {
  seedSampleData(): Promise<{ success: boolean }>;
  getBackupExportUrl(format: 'json' | 'sqlite'): string;
}

export const defaultBackupAndSeedApi: BackupAndSeedApi = {
  seedSampleData: () =>
    requestJSON<{ success: boolean }>('/api/seed', {
      method: 'POST',
      body: JSON.stringify({}),
    }),
  getBackupExportUrl: (format) => `/api/backup/export?format=${format}`,
};
