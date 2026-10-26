import React, { createContext, useContext } from 'react';
import { type ActionItemApi, defaultActionItemApi } from './actionItemApi';
import { type CompanyApi, defaultCompanyApi } from './companyApi';
import { type ContactApi, defaultContactApi } from './contactApi';
import { type InteractionApi, defaultInteractionApi } from './interactionApi';
import { type OpportunityApi, defaultOpportunityApi } from './opportunityApi';
import { type ProfileApi, defaultProfileApi } from './profileApi';
import { type StoryApi, defaultStoryApi } from './storyApi';

export interface CockpitApi {
  companies: CompanyApi;
  opportunities: OpportunityApi;
  profile: ProfileApi;
  stories: StoryApi;
  contacts: ContactApi;
  interactions: InteractionApi;
  actionItems: ActionItemApi;
}

const defaultCockpitApi: CockpitApi = {
  companies: defaultCompanyApi,
  opportunities: defaultOpportunityApi,
  profile: defaultProfileApi,
  stories: defaultStoryApi,
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
