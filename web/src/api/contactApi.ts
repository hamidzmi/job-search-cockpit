import { requestJSON } from './httpClient';
import type {
  Contact,
  ContactFilter,
  ContactUpsertInput,
  OpportunityContact,
  OpportunityContactLinkInput,
} from './types';

export interface ContactApi {
  listContacts(filter?: ContactFilter): Promise<Contact[]>;
  getContact(id: string): Promise<Contact>;
  createContact(input: ContactUpsertInput): Promise<Contact>;
  updateContact(id: string, input: ContactUpsertInput): Promise<Contact>;
  listOpportunityContacts(opportunityId: string): Promise<OpportunityContact[]>;
  linkContactToOpportunity(
    opportunityId: string,
    input: OpportunityContactLinkInput,
  ): Promise<OpportunityContact>;
  unlinkContactFromOpportunity(opportunityId: string, contactId: string): Promise<void>;
}

export const defaultContactApi: ContactApi = {
  listContacts: (filter = {}) => {
    const params = new URLSearchParams();
    if (filter.company_id) params.set('company_id', filter.company_id);
    if (filter.warmth) params.set('warmth', filter.warmth);
    if (filter.q) params.set('q', filter.q);
    const qs = params.toString();
    return requestJSON<Contact[]>(`/api/contacts${qs ? `?${qs}` : ''}`);
  },
  getContact: (id) => requestJSON<Contact>(`/api/contacts/${encodeURIComponent(id)}`),
  createContact: (input) =>
    requestJSON<Contact>('/api/contacts', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  updateContact: (id, input) =>
    requestJSON<Contact>(`/api/contacts/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
  listOpportunityContacts: (opportunityId) =>
    requestJSON<OpportunityContact[]>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/contacts`,
    ),
  linkContactToOpportunity: (opportunityId, input) =>
    requestJSON<OpportunityContact>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/contacts`,
      {
        method: 'POST',
        body: JSON.stringify(input),
      },
    ),
  unlinkContactFromOpportunity: async (opportunityId, contactId) => {
    await requestJSON<void>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/contacts/${encodeURIComponent(contactId)}`,
      {
        method: 'DELETE',
      },
    );
  },
};
