import { requestJSON } from './httpClient';
import type {
  DebriefUpsertInput,
  InterviewRound,
  InterviewRoundUpsertInput,
  QuestionBankEntry,
  QuestionBankFilter,
} from './types';

export interface InterviewRoundApi {
  listInterviewRounds(opportunityId: string): Promise<InterviewRound[]>;
  createInterviewRound(
    opportunityId: string,
    input: InterviewRoundUpsertInput,
  ): Promise<InterviewRound>;
  updateInterviewRound(
    roundId: string,
    input: InterviewRoundUpsertInput,
  ): Promise<InterviewRound>;
  deleteInterviewRound(roundId: string): Promise<void>;
  recordDebrief(roundId: string, input: DebriefUpsertInput): Promise<InterviewRound>;
  listQuestionBank(filter?: QuestionBankFilter): Promise<QuestionBankEntry[]>;
}

export const defaultInterviewRoundApi: InterviewRoundApi = {
  listInterviewRounds(opportunityId: string) {
    return requestJSON<InterviewRound[]>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/rounds`,
    );
  },
  createInterviewRound(opportunityId: string, input: InterviewRoundUpsertInput) {
    return requestJSON<InterviewRound>(
      `/api/opportunities/${encodeURIComponent(opportunityId)}/rounds`,
      {
        method: 'POST',
        body: JSON.stringify(input),
      },
    );
  },
  updateInterviewRound(roundId: string, input: InterviewRoundUpsertInput) {
    return requestJSON<InterviewRound>(`/api/rounds/${encodeURIComponent(roundId)}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    });
  },
  deleteInterviewRound(roundId: string) {
    return requestJSON<void>(`/api/rounds/${encodeURIComponent(roundId)}`, {
      method: 'DELETE',
    });
  },
  recordDebrief(roundId: string, input: DebriefUpsertInput) {
    return requestJSON<InterviewRound>(
      `/api/rounds/${encodeURIComponent(roundId)}/debrief`,
      {
        method: 'PUT',
        body: JSON.stringify(input),
      },
    );
  },
  listQuestionBank(filter?: QuestionBankFilter) {
    const params = new URLSearchParams();
    if (filter?.q) params.set('q', filter.q);
    if (filter?.category) params.set('category', filter.category);
    if (filter?.company_id) params.set('company_id', filter.company_id);
    if (filter?.round_type) params.set('round_type', filter.round_type);
    if (filter?.story_id) params.set('story_id', filter.story_id);
    const qs = params.toString();
    return requestJSON<QuestionBankEntry[]>(`/api/questions${qs ? `?${qs}` : ''}`);
  },
};
