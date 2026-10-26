import React, { createContext, useContext } from 'react';
import { type ActionItemApi, defaultActionItemApi } from './actionItemApi';
import { type CompanyApi, defaultCompanyApi } from './companyApi';
import { type ContactApi, defaultContactApi } from './contactApi';
import { type InteractionApi, defaultInteractionApi } from './interactionApi';
import { type OpportunityApi, defaultOpportunityApi } from './opportunityApi';

export interface CockpitApi {
  companies: CompanyApi;
  opportunities: OpportunityApi;
  contacts: ContactApi;
  interactions: InteractionApi;
  actionItems: ActionItemApi;
}

const defaultCockpitApi: CockpitApi = {
  companies: defaultCompanyApi,
  opportunities: defaultOpportunityApi,
  contacts: defaultContactApi,
  interactions: defaultInteractionApi,
  actionItems: defaultActionItemApi,
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
