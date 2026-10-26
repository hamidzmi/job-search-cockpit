import { requestJSON } from './httpClient';
import type { FunnelAnalytics } from './types';

export interface AnalyticsApi {
  getFunnelAnalytics(): Promise<FunnelAnalytics>;
}

export const defaultAnalyticsApi: AnalyticsApi = {
  getFunnelAnalytics: () => requestJSON<FunnelAnalytics>('/api/analytics/funnel'),
};
