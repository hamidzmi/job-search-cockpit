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
        <div className="stories-grid">
          {stories.map((story) => (
            <article key={story.id} className="panel-card story-card">
              <div className="card-top-row">
                <h4 style={{ margin: 0, fontSize: '1rem' }}>{story.title}</h4>
                <div className="workspace-actions">
                  <button
                    type="button"
                    className="btn-secondary"
                    style={{ padding: '0.25rem 0.55rem', fontSize: '0.78rem' }}
                    onClick={() => setEditingStory(story)}
                  >
                    Edit Story
                  </button>
                  <button
                    type="button"
                    className="btn-ghost"
                    style={{ padding: '0.25rem 0.55rem', fontSize: '0.78rem' }}
                    onClick={() => deleteStoryMutation.mutate(story.id)}
                  >
                    Delete Story
                  </button>
                </div>
              </div>

              {story.hook && (
                <div className="story-hook-banner">
                  ⚡ &ldquo;{story.hook}&rdquo;
                </div>
              )}

              <div className="star-grid">
                {story.situation && (
                  <div className="star-line">
                    <span className="star-label">Situation</span>
                    <div>{story.situation}</div>
                  </div>
                )}
                {story.task && (
                  <div className="star-line">
                    <span className="star-label">Task</span>
                    <div>{story.task}</div>
                  </div>
                )}
                {story.action && (
                  <div className="star-line">
                    <span className="star-label">Action</span>
                    <div>{story.action}</div>
                  </div>
                )}
                {story.result && (
                  <div className="star-line">
                    <span className="star-label">Result</span>
                    <div className="story-result-text">{story.result}</div>
                  </div>
                )}
              </div>

              <div className="tag-list" style={{ marginTop: 'auto', paddingTop: '0.35rem' }}>
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
