import { requestJSON } from './httpClient';
import type { CandidateProfile, CandidateProfileUpsertInput } from './types';

export interface ProfileApi {
  getProfile(): Promise<CandidateProfile>;
  saveProfile(input: CandidateProfileUpsertInput): Promise<CandidateProfile>;
}

export const defaultProfileApi: ProfileApi = {
  getProfile: () => requestJSON<CandidateProfile>('/api/profile'),
  saveProfile: (input) =>
    requestJSON<CandidateProfile>('/api/profile', {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
};
