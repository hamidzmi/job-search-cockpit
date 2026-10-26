import React, { createContext, useContext } from 'react';
import { type CompanyApi, defaultCompanyApi } from './companyApi';
import { type OpportunityApi, defaultOpportunityApi } from './opportunityApi';

export interface CockpitApi {
  companies: CompanyApi;
  opportunities: OpportunityApi;
}

const defaultCockpitApi: CockpitApi = {
  companies: defaultCompanyApi,
  opportunities: defaultOpportunityApi,
};

const CockpitApiContext = createContext<CockpitApi>(defaultCockpitApi);

export function CockpitApiProvider({
  api = defaultCockpitApi,
  children,
}: {
  api?: CockpitApi;
  children: React.ReactNode;
}) {
  return <CockpitApiContext.Provider value={api}>{children}</CockpitApiContext.Provider>;
}

export function useCockpitApi(): CockpitApi {
  return useContext(CockpitApiContext);
}
