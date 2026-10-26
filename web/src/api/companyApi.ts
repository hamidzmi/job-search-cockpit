import { requestJSON } from './httpClient';
import type { Company, CompanyUpsertInput } from './types';

export interface CompanyApi {
  listCompanies(): Promise<Company[]>;
  getCompany(id: string): Promise<Company>;
  createCompany(input: CompanyUpsertInput): Promise<Company>;
  updateCompany(id: string, input: CompanyUpsertInput): Promise<Company>;
}

export const defaultCompanyApi: CompanyApi = {
  listCompanies: () => requestJSON<Company[]>('/api/companies'),
  getCompany: (id) => requestJSON<Company>(`/api/companies/${encodeURIComponent(id)}`),
  createCompany: (input) =>
    requestJSON<Company>('/api/companies', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  updateCompany: (id, input) =>
    requestJSON<Company>(`/api/companies/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
};
