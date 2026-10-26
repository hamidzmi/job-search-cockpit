import React, { useState } from 'react';
import type { Story, StoryUpsertInput } from '../../api/types';

interface StoryFormModalProps {
  initial?: Story | null;
  onClose: () => void;
  onSubmit: (input: StoryUpsertInput) => Promise<void>;
}

function splitTags(text: string): string[] {
  return text
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

export function StoryFormModal({ initial, onClose, onSubmit }: StoryFormModalProps) {
  const isEditing = Boolean(initial);
  const [title, setTitle] = useState(initial?.title ?? '');
  const [hook, setHook] = useState(initial?.hook ?? '');
  const [situation, setSituation] = useState(initial?.situation ?? '');
  const [task, setTask] = useState(initial?.task ?? '');
  const [action, setAction] = useState(initial?.action ?? '');
  const [result, setResult] = useState(initial?.result ?? '');
  const [competencyTagsText, setCompetencyTagsText] = useState(
    initial?.competency_tags.join(', ') ?? '',
  );
  const [skillTagsText, setSkillTagsText] = useState(
    initial?.skill_tags.join(', ') ?? '',
  );
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const dialogTitle = isEditing ? 'Edit Story' : 'New Story';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setError('Story title is required.');
      return;
    }
    setError(null);
    setSubmitting(true);
    try {
      await onSubmit({
        title: title.trim(),
        hook: hook.trim(),
        situation: situation.trim(),
        task: task.trim(),
        action: action.trim(),
        result: result.trim(),
        competency_tags: splitTags(competencyTagsText),
        skill_tags: splitTags(skillTagsText),
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save Story');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-backdrop">
      <div
        className="modal-card modal-card-wide"
        role="dialog"
        aria-modal="true"
        aria-label={dialogTitle}
      >
        <div className="modal-header">
          <h2>{dialogTitle}</h2>
          <button type="button" className="btn-ghost" onClick={onClose}>
            Close
          </button>
        </div>

        {error && <div className="alert-error">{error}</div>}

        <form onSubmit={handleSubmit} className="form-grid">
          <label className="form-field form-field-full">
            Story Title *
            <input
              type="text"
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Zero-Downtime Ledger Sharding"
            />
          </label>

          <label className="form-field form-field-full">
            Hook
            <input
              type="text"
              value={hook}
              onChange={(e) => setHook(e.target.value)}
              placeholder="One-line elevator pitch for this STAR narrative"
            />
          </label>

          <label className="form-field form-field-full">
            Situation
            <textarea
              rows={2}
              value={situation}
              onChange={(e) => setSituation(e.target.value)}
              placeholder="Context, constraints, and business stakes..."
            />
          </label>

          <label className="form-field form-field-full">
            Task
            <textarea
              rows={2}
              value={task}
              onChange={(e) => setTask(e.target.value)}
              placeholder="Specific responsibility and goal..."
            />
          </label>

          <label className="form-field form-field-full">
            Action
            <textarea
              rows={3}
              value={action}
              onChange={(e) => setAction(e.target.value)}
              placeholder="Technical decisions, trade-offs, and leadership steps taken..."
            />
          </label>

          <label className="form-field form-field-full">
            Result (Quantified Impact)
            <textarea
              rows={2}
              value={result}
              onChange={(e) => setResult(e.target.value)}
              placeholder="Measurable outcomes, latency/cost improvements, business impact..."
            />
          </label>

          <label className="form-field">
            Competency Tags (comma-separated)
            <input
              type="text"
              value={competencyTagsText}
              onChange={(e) => setCompetencyTagsText(e.target.value)}
              placeholder="Leadership, System Design, Conflict, Ownership"
            />
          </label>

          <label className="form-field">
            Skill Tags (comma-separated)
            <input
              type="text"
              value={skillTagsText}
              onChange={(e) => setSkillTagsText(e.target.value)}
              placeholder="Go, PostgreSQL, Kafka, Distributed Systems"
            />
          </label>

          <div className="modal-actions form-field-full">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-primary" disabled={submitting}>
              {isEditing ? 'Save Story' : 'Create Story'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
