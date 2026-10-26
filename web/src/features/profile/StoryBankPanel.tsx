import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useMemo, useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import type { Story, StoryFilter, StoryUpsertInput } from '../../api/types';
import { StoryFormModal } from './StoryFormModal';

const DEFAULT_COMPETENCY_TAGS = [
  'Leadership',
  'System Design',
  'Conflict',
  'Ownership',
];

export function StoryBankPanel() {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const [searchQuery, setSearchQuery] = useState('');
  const [competencyTag, setCompetencyTag] = useState('');
  const [skillTag, setSkillTag] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingStory, setEditingStory] = useState<Story | null>(null);

  const filter: StoryFilter = {
    q: searchQuery.trim() || undefined,
    competency_tag: competencyTag || undefined,
    skill_tag: skillTag || undefined,
  };

  const allStoriesQuery = useQuery({
    queryKey: ['stories', {}],
    queryFn: () => api.stories.listStories({}),
  });

  const filteredStoriesQuery = useQuery({
    queryKey: ['stories', filter],
    queryFn: () => api.stories.listStories(filter),
  });

  const competencyOptions = useMemo(() => {
    const set = new Set<string>(DEFAULT_COMPETENCY_TAGS);
    for (const s of allStoriesQuery.data ?? []) {
      for (const tag of s.competency_tags) {
        if (tag.trim()) set.add(tag.trim());
      }
    }
    return Array.from(set).sort((a, b) => a.localeCompare(b));
  }, [allStoriesQuery.data]);

  const skillOptions = useMemo(() => {
    const set = new Set<string>();
    for (const s of allStoriesQuery.data ?? []) {
      for (const tag of s.skill_tags) {
        if (tag.trim()) set.add(tag.trim());
      }
    }
    return Array.from(set).sort((a, b) => a.localeCompare(b));
  }, [allStoriesQuery.data]);

  const saveStoryMutation = useMutation({
    mutationFn: async (input: StoryUpsertInput) => {
      if (editingStory) {
        return api.stories.updateStory(editingStory.id, input);
      }
      return api.stories.createStory(input);
    },
    onSuccess: async () => {
      setEditingStory(null);
      setShowCreateModal(false);
      await queryClient.invalidateQueries({ queryKey: ['stories'] });
    },
  });

  const deleteStoryMutation = useMutation({
    mutationFn: (id: string) => api.stories.deleteStory(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['stories'] });
    },
  });

  const stories = filteredStoriesQuery.data ?? [];

  return (
    <section className="panel-card" aria-label="Story Bank">
      <div className="card-top-row">
        <div>
          <h3>Story Bank</h3>
          <p className="timestamp-text">
            Reusable STAR narratives tagged by Competency and Skill for interview loops and tailoring.
          </p>
        </div>
        <button
          type="button"
          className="btn-primary"
          onClick={() => setShowCreateModal(true)}
        >
          + New Story
        </button>
      </div>

      <div className="filter-bar">
        <label className="filter-control">
          Search Stories
          <input
            type="search"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Keyword in STAR fields or tags..."
          />
        </label>

        <label className="filter-control">
          Filter by Competency Tag
          <select
            value={competencyTag}
            onChange={(e) => setCompetencyTag(e.target.value)}
          >
            <option value="">All Competencies</option>
            {competencyOptions.map((tag) => (
              <option key={tag} value={tag}>
                {tag}
              </option>
            ))}
          </select>
        </label>

        <label className="filter-control">
          Filter by Skill Tag
          <select value={skillTag} onChange={(e) => setSkillTag(e.target.value)}>
            <option value="">All Skills</option>
            {skillOptions.map((tag) => (
              <option key={tag} value={tag}>
                {tag}
              </option>
            ))}
          </select>
        </label>
      </div>

      {filteredStoriesQuery.isLoading ? (
        <p>Loading Story Bank...</p>
      ) : stories.length === 0 ? (
        <p className="timestamp-text">No Stories match the current filters.</p>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem' }}>
          {stories.map((story) => (
            <article key={story.id} className="panel-card">
              <div className="card-top-row">
                <div>
                  <h4>{story.title}</h4>
                  {story.hook && (
                    <p style={{ margin: '0.25rem 0 0', fontWeight: 500, color: '#334155' }}>
                      {story.hook}
                    </p>
                  )}
                </div>
                <div className="workspace-actions">
                  <button
                    type="button"
                    className="btn-secondary"
                    onClick={() => setEditingStory(story)}
                  >
                    Edit Story
                  </button>
                  <button
                    type="button"
                    className="btn-ghost"
                    onClick={() => deleteStoryMutation.mutate(story.id)}
                  >
                    Delete Story
                  </button>
                </div>
              </div>

              <div className="form-grid" style={{ fontSize: '0.85rem' }}>
                {story.situation && (
                  <div>
                    <strong>Situation:</strong> {story.situation}
                  </div>
                )}
                {story.task && (
                  <div>
                    <strong>Task:</strong> {story.task}
                  </div>
                )}
                {story.action && (
                  <div>
                    <strong>Action:</strong> {story.action}
                  </div>
                )}
                {story.result && (
                  <div>
                    <strong>Result:</strong> {story.result}
                  </div>
                )}
              </div>

              <div className="tag-list">
                {story.competency_tags.map((tag) => (
                  <span key={`comp-${tag}`} className="stage-badge">
                    {tag}
                  </span>
                ))}
                {story.skill_tags.map((tag) => (
                  <span key={`skill-${tag}`} className="tech-tag">
                    {tag}
                  </span>
                ))}
              </div>
            </article>
          ))}
        </div>
      )}

      {(showCreateModal || editingStory) && (
        <StoryFormModal
          initial={editingStory}
          onClose={() => {
            setShowCreateModal(false);
            setEditingStory(null);
          }}
          onSubmit={async (input) => {
            await saveStoryMutation.mutateAsync(input);
          }}
        />
      )}
    </section>
  );
}
