import { requestJSON } from './httpClient';
import type {
  ActionItem,
  ActionItemCreateInput,
  ActionItemFilter,
  ActionItemUpdateInput,
} from './types';

export interface ActionItemApi {
  listActionItems(filter?: ActionItemFilter): Promise<ActionItem[]>;
  createActionItem(input: ActionItemCreateInput): Promise<ActionItem>;
  updateActionItem(id: string, input: ActionItemUpdateInput): Promise<ActionItem>;
}

export const defaultActionItemApi: ActionItemApi = {
  listActionItems: (filter = {}) => {
    const params = new URLSearchParams();
    if (filter.status) params.set('status', filter.status);
    if (filter.opportunity_id) params.set('opportunity_id', filter.opportunity_id);
    if (filter.contact_id) params.set('contact_id', filter.contact_id);
    const qs = params.toString();
    return requestJSON<ActionItem[]>(`/api/action-items${qs ? `?${qs}` : ''}`);
  },
  createActionItem: (input) =>
    requestJSON<ActionItem>('/api/action-items', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  updateActionItem: (id, input) =>
    requestJSON<ActionItem>(`/api/action-items/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
};
