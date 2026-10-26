import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import { STAGES, type Stage } from '../../api/types';
import { OpportunityCoreFields } from '../pipeline/OpportunityDetailsFields';
import {
  buildJobPostingInput,
  createEmptyDetailsDraft,
  type OpportunityDetailsDraft,
} from '../pipeline/opportunityFormState';

interface QuickCaptureModalProps {
  onClose: () => void;
}

export function QuickCaptureModal({ onClose }: QuickCaptureModalProps) {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const settingsQuery = useQuery({
    queryKey: ['aiSettings'],
    queryFn: () => api.aiSettings.getSettings(),
  });

  const [companyName, setCompanyName] = useState('');
  const [companyDomain, setCompanyDomain] = useState('');
  const [draft, setDraft] = useState<OpportunityDetailsDraft>(() =>
    createEmptyDetailsDraft('High', 'Remote'),
  );
  const [stage, setStage] = useState<Stage>('Discovered');
  const [modelOverride, setModelOverride] = useState('');

  const [fetchWarning, setFetchWarning] = useState('');
  const [assistantWarning, setAssistantWarning] = useState('');
  const [error, setError] = useState<string | null>(null);

  const extractMutation = useMutation({
    mutationFn: () =>
      api.quickCapture.extractJobPosting({
        source_url: draft.sourceUrl,
        raw_text: draft.rawText,
        model_override: modelOverride || undefined,
      }),
    onSuccess: (extracted) => {
      setError(null);
      setFetchWarning(extracted.fetch_warning || '');
      setAssistantWarning(extracted.assistant_warning || '');

      if (extracted.company_name) {
        setCompanyName(extracted.company_name);
      }
      if (extracted.company_domain) {
        setCompanyDomain(extracted.company_domain);
      }
      setDraft((prev) => ({
        ...prev,
        roleTitle: extracted.role_title || prev.roleTitle,
        seniority: extracted.seniority || prev.seniority,
        location: extracted.location || prev.location,
        remotePolicy: extracted.remote_policy || prev.remotePolicy,
        sourceUrl: extracted.source_url || prev.sourceUrl,
        rawText: extracted.raw_text || prev.rawText,
        statedMinComp: extracted.stated_min_comp
          ? String(extracted.stated_min_comp)
          : prev.statedMinComp,
        statedMaxComp: extracted.stated_max_comp
          ? String(extracted.stated_max_comp)
          : prev.statedMaxComp,
        currency: extracted.currency || prev.currency || 'USD',
        techStackText:
          extracted.tech_stack_tags.length > 0
            ? extracted.tech_stack_tags.join(', ')
            : prev.techStackText,
        requirementsText:
          extracted.requirements.length > 0
            ? extracted.requirements.join('\n')
            : prev.requirementsText,
      }));
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Extraction failed');
    },
  });

  const saveMutation = useMutation({
    mutationFn: () =>
      api.quickCapture.captureOpportunity({
        company_name: companyName,
        company_domain: companyDomain,
        role_title: draft.roleTitle,
        seniority: draft.seniority,
        location: draft.location,
        remote_policy: draft.remotePolicy,
        priority: draft.priority,
        stage,
        transition_note: 'Captured via Quick Capture',
        notes: draft.notes,
        job_posting: buildJobPostingInput(draft, true),
      }),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
        queryClient.invalidateQueries({ queryKey: ['companies'] }),
      ]);
      onClose();
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to save Opportunity');
    },
  });

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    saveMutation.mutate();
  };

  const bookmarkedModels = settingsQuery.data?.bookmarked_models ?? [];
  const defaultFastModel = settingsQuery.data?.fast_extraction_model || '';

  return (
    <div
      className="modal-backdrop"
      role="dialog"
      aria-modal="true"
      aria-label="Quick Capture"
    >
      <div className="modal-card modal-card-wide">
        <div className="modal-header">
          <h2>Quick Capture</h2>
          <button type="button" className="btn-ghost" onClick={onClose}>
            Close
          </button>
        </div>

        {error && (
          <div className="alert-error" role="alert">
            {error}
          </div>
        )}
        {fetchWarning && (
          <div className="alert-error" role="status">
            {fetchWarning}
          </div>
        )}
        {assistantWarning && (
          <div className="alert-success" role="status">
            {assistantWarning}
          </div>
        )}

        <div className="panel-card" style={{ marginBottom: '1rem' }}>
          <div className="form-grid">
            <label className="form-field">
              <span>Job Posting URL</span>
              <input
                type="url"
                value={draft.sourceUrl}
                onChange={(e) => setDraft({ ...draft, sourceUrl: e.target.value })}
                placeholder="https://boards.greenhouse.io/..."
              />
            </label>

            <label className="form-field">
              <span>Extraction Model</span>
              <select
                value={modelOverride}
                onChange={(e) => setModelOverride(e.target.value)}
              >
                <option value="">
                  {defaultFastModel
                    ? `Default Fast Extraction (${defaultFastModel})`
                    : 'Default Fast Extraction Model'}
                </option>
                {bookmarkedModels.map((slug) => (
                  <option key={slug} value={slug}>
                    {slug}
                  </option>
                ))}
              </select>
            </label>

            <label className="form-field form-field-full">
              <span>Paste Job Posting Text</span>
              <textarea
                rows={3}
                value={draft.rawText}
                onChange={(e) => setDraft({ ...draft, rawText: e.target.value })}
                placeholder="Paste raw Job Posting text (used alongside URL or as fallback when a job board blocks automated fetching)..."
              />
            </label>

            <div className="modal-actions form-field-full">
              <button
                type="button"
                className="btn-secondary"
                disabled={
                  extractMutation.isPending ||
                  (!draft.sourceUrl.trim() && !draft.rawText.trim())
                }
                onClick={() => extractMutation.mutate()}
              >
                {extractMutation.isPending
                  ? 'Extracting...'
                  : 'Extract Job Posting'}
              </button>
            </div>
          </div>
        </div>

        <form onSubmit={handleSave} className="form-grid">
          <label className="form-field">
            <span>Company Name *</span>
            <input
              type="text"
              required
              value={companyName}
              onChange={(e) => setCompanyName(e.target.value)}
              placeholder="e.g. Stripe"
            />
          </label>

          <label className="form-field">
            <span>Company Domain</span>
            <input
              type="text"
              value={companyDomain}
              onChange={(e) => setCompanyDomain(e.target.value)}
              placeholder="e.g. stripe.com"
            />
          </label>

          <OpportunityCoreFields draft={draft} onChange={setDraft} />

          <label className="form-field">
            <span>Initial Stage</span>
            <select value={stage} onChange={(e) => setStage(e.target.value as Stage)}>
              {STAGES.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field">
            <span>Stated Min Compensation</span>
            <input
              type="number"
              min={0}
              value={draft.statedMinComp}
              onChange={(e) => setDraft({ ...draft, statedMinComp: e.target.value })}
              placeholder="200000"
            />
          </label>

          <label className="form-field">
            <span>Stated Max Compensation</span>
            <input
              type="number"
              min={0}
              value={draft.statedMaxComp}
              onChange={(e) => setDraft({ ...draft, statedMaxComp: e.target.value })}
              placeholder="250000"
            />
          </label>

          <label className="form-field">
            <span>Currency</span>
            <input
              type="text"
              value={draft.currency}
              onChange={(e) => setDraft({ ...draft, currency: e.target.value })}
              placeholder="USD"
            />
          </label>

          <label className="form-field">
            <span>Tech Stack Tags (comma-separated)</span>
            <input
              type="text"
              value={draft.techStackText}
              onChange={(e) => setDraft({ ...draft, techStackText: e.target.value })}
              placeholder="Go, TypeScript, React"
            />
          </label>

          <label className="form-field form-field-full">
            <span>Requirements (one per line)</span>
            <textarea
              rows={3}
              value={draft.requirementsText}
              onChange={(e) => setDraft({ ...draft, requirementsText: e.target.value })}
              placeholder="Key role requirements..."
            />
          </label>

          <label className="form-field form-field-full">
            <span>Pursuit Notes</span>
            <textarea
              rows={2}
              value={draft.notes}
              onChange={(e) => setDraft({ ...draft, notes: e.target.value })}
              placeholder="Referral or application context..."
            />
          </label>

          <div className="modal-actions form-field-full">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button
              type="submit"
              className="btn-primary"
              disabled={
                saveMutation.isPending ||
                !companyName.trim() ||
                !draft.roleTitle.trim()
              }
            >
              {saveMutation.isPending ? 'Saving...' : 'Save Opportunity'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
