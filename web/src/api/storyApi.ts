import { requestJSON } from './httpClient';
import type { Story, StoryFilter, StoryUpsertInput } from './types';

export interface StoryApi {
  listStories(filter?: StoryFilter): Promise<Story[]>;
  getStory(id: string): Promise<Story>;
  createStory(input: StoryUpsertInput): Promise<Story>;
  updateStory(id: string, input: StoryUpsertInput): Promise<Story>;
  deleteStory(id: string): Promise<void>;
}

export const defaultStoryApi: StoryApi = {
  listStories: (filter = {}) => {
    const params = new URLSearchParams();
    if (filter.q) params.set('q', filter.q);
    if (filter.competency_tag) params.set('competency_tag', filter.competency_tag);
    if (filter.skill_tag) params.set('skill_tag', filter.skill_tag);
    const qs = params.toString();
    return requestJSON<Story[]>(`/api/stories${qs ? `?${qs}` : ''}`);
  },
  getStory: (id) => requestJSON<Story>(`/api/stories/${encodeURIComponent(id)}`),
  createStory: (input) =>
    requestJSON<Story>('/api/stories', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  updateStory: (id, input) =>
    requestJSON<Story>(`/api/stories/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),
  deleteStory: (id) =>
    requestJSON<void>(`/api/stories/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),
};
