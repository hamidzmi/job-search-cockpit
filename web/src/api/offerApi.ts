import { requestJSON } from './httpClient';
import type { OfferComparison, OfferView, OfferUpsertInput } from './types';

export interface OfferApi {
  getOfferComparison(): Promise<OfferComparison>;
  getOffer(opportunityId: string): Promise<OfferView>;
  upsertOffer(opportunityId: string, input: OfferUpsertInput): Promise<OfferView>;
  deleteOffer(opportunityId: string): Promise<void>;
}

export const defaultOfferApi: OfferApi = {
  getOfferComparison: () => requestJSON<OfferComparison>('/api/offers'),
  getOffer: (opportunityId) =>
    requestJSON<OfferView>(`/api/opportunities/${opportunityId}/offer`),
  upsertOffer: (opportunityId, input) =>
    requestJSON<OfferView>(`/api/opportunities/${opportunityId}/offer`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
  deleteOffer: (opportunityId) =>
    requestJSON<void>(`/api/opportunities/${opportunityId}/offer`, {
      method: 'DELETE',
    }),
};
