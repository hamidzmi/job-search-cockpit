import { requestJSON } from './httpClient';
import type { AISettings, AISettingsUpsertInput, OpenRouterModel } from './types';

export interface AISettingsApi {
  getSettings(): Promise<AISettings>;
  updateSettings(input: AISettingsUpsertInput): Promise<AISettings>;
  listOpenRouterModels(query?: string): Promise<OpenRouterModel[]>;
}

export const defaultAISettingsApi: AISettingsApi = {
  getSettings() {
    return requestJSON<AISettings>('/api/settings/ai');
  },
  updateSettings(input) {
    return requestJSON<AISettings>('/api/settings/ai', {
      method: 'PUT',
      body: JSON.stringify(input),
    });
  },
  listOpenRouterModels(query = '') {
    const params = new URLSearchParams();
    if (query.trim()) {
      params.set('q', query.trim());
    }
    const qs = params.toString();
    return requestJSON<OpenRouterModel[]>(
      `/api/settings/ai/openrouter-models${qs ? `?${qs}` : ''}`,
    );
  },
};
