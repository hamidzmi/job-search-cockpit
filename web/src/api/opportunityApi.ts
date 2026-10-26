import { requestJSON } from './httpClient';
import type {
  Opportunity,
  OpportunityCreateInput,
  OpportunityFilter,
  OpportunityTransitionInput,
  OpportunityUpdateInput,
} from './types';

export interface OpportunityApi {
  listOpportunities(filter?: OpportunityFilter): Promise<Opportunity[]>;
  getOpportunity(id: string): Promise<Opportunity>;
  createOpportunity(input: OpportunityCreateInput): Promise<Opportunity>;
  updateOpportunity(id: string, input: OpportunityUpdateInput): Promise<Opportunity>;
  transitionOpportunity(id: string, input: OpportunityTransitionInput): Promise<Opportunity>;
}

export const defaultOpportunityApi: OpportunityApi = {
  listOpportunities: (filter = {}) => {
    const params = new URLSearchParams();
    if (filter.stage) params.set('stage', filter.stage);
    if (filter.outcome) params.set('outcome', filter.outcome);
    if (filter.priority) params.set('priority', filter.priority);
    if (filter.remote_policy) params.set('remote_policy', filter.remote_policy);
    if (filter.company_id) params.set('company_id', filter.company_id);
    if (filter.q) params.set('q', filter.q);
    const qs = params.toString();
    return requestJSON<Opportunity[]>(`/api/opportunities${qs ? `?${qs}` : ''}`);
  },
  getOpportunity: (id) => requestJSON<Opportunity>(`/api/opportunities/${encodeURIComponent(id)}`),
  createOpportunity: (input) =>
    requestJSON<Opportunity>('/api/opportunities', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  updateOpportunity: (id, input) =>
    requestJSON<Opportunity>(`/api/opportunities/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
  transitionOpportunity: (id, input) =>
    requestJSON<Opportunity>(`/api/opportunities/${encodeURIComponent(id)}/transition`, {
      method: 'POST',
      body: JSON.stringify(input),
    }),
};
