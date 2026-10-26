import { requestJSON } from './httpClient';
import type {
  ExtractedJobPostingDraft,
  Opportunity,
  QuickCaptureExtractInput,
  QuickCaptureSaveInput,
} from './types';

export interface QuickCaptureApi {
  extractJobPosting(input: QuickCaptureExtractInput): Promise<ExtractedJobPostingDraft>;
  captureOpportunity(input: QuickCaptureSaveInput): Promise<Opportunity>;
}

export const defaultQuickCaptureApi: QuickCaptureApi = {
  extractJobPosting(input) {
    return requestJSON<ExtractedJobPostingDraft>('/api/quick-capture/extract', {
      method: 'POST',
      body: JSON.stringify(input),
    });
  },
  captureOpportunity(input) {
    return requestJSON<Opportunity>('/api/quick-capture', {
      method: 'POST',
      body: JSON.stringify(input),
    });
  },
};
