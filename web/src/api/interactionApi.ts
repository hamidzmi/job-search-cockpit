import { requestJSON } from './httpClient';
import type {
  Interaction,
  InteractionCreateInput,
  InteractionFilter,
  LogInteractionResponse,
} from './types';

export interface InteractionApi {
  listInteractions(filter?: InteractionFilter): Promise<Interaction[]>;
  logInteraction(input: InteractionCreateInput): Promise<LogInteractionResponse>;
}

export const defaultInteractionApi: InteractionApi = {
  listInteractions: (filter = {}) => {
    const params = new URLSearchParams();
    if (filter.contact_id) params.set('contact_id', filter.contact_id);
    if (filter.opportunity_id) params.set('opportunity_id', filter.opportunity_id);
    const qs = params.toString();
    return requestJSON<Interaction[]>(`/api/interactions${qs ? `?${qs}` : ''}`);
  },
  logInteraction: (input) =>
    requestJSON<LogInteractionResponse>('/api/interactions', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
};
