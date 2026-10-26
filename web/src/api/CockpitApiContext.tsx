import React, { createContext, useContext } from 'react';
import { type ActionItemApi, defaultActionItemApi } from './actionItemApi';
import { type AISettingsApi, defaultAISettingsApi } from './aiSettingsApi';
import { type CompanyApi, defaultCompanyApi } from './companyApi';
import { type ContactApi, defaultContactApi } from './contactApi';
import { type InteractionApi, defaultInteractionApi } from './interactionApi';
import { type OpportunityApi, defaultOpportunityApi } from './opportunityApi';
import { type ProfileApi, defaultProfileApi } from './profileApi';
import { type QuickCaptureApi, defaultQuickCaptureApi } from './quickCaptureApi';
import { type StoryApi, defaultStoryApi } from './storyApi';
import { type TailoringApi, defaultTailoringApi } from './tailoringApi';

export interface CockpitApi {
  companies: CompanyApi;
  opportunities: OpportunityApi;
  aiSettings: AISettingsApi;
  quickCapture: QuickCaptureApi;
  profile: ProfileApi;
  stories: StoryApi;
  contacts: ContactApi;
  interactions: InteractionApi;
  actionItems: ActionItemApi;
  tailoring: TailoringApi;
}

const defaultCockpitApi: CockpitApi = {
  companies: defaultCompanyApi,
  opportunities: defaultOpportunityApi,
  aiSettings: defaultAISettingsApi,
  quickCapture: defaultQuickCaptureApi,
  profile: defaultProfileApi,
  stories: defaultStoryApi,
  contacts: defaultContactApi,
  interactions: defaultInteractionApi,
  actionItems: defaultActionItemApi,
  tailoring: defaultTailoringApi,
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
