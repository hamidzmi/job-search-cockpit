import React, { createContext, useContext } from 'react';
import { type AISettingsApi, defaultAISettingsApi } from './aiSettingsApi';
import { type CompanyApi, defaultCompanyApi } from './companyApi';
import { type OpportunityApi, defaultOpportunityApi } from './opportunityApi';
import { type QuickCaptureApi, defaultQuickCaptureApi } from './quickCaptureApi';

export interface CockpitApi {
  companies: CompanyApi;
  opportunities: OpportunityApi;
  aiSettings: AISettingsApi;
  quickCapture: QuickCaptureApi;
}

const defaultCockpitApi: CockpitApi = {
  companies: defaultCompanyApi,
  opportunities: defaultOpportunityApi,
  aiSettings: defaultAISettingsApi,
  quickCapture: defaultQuickCaptureApi,
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
