import React, { useState } from 'react';
import {
  STAGES,
  type Company,
  type OpportunityCreateInput,
  type Stage,
} from '../../api/types';
import {
  OpportunityCoreFields,
  OpportunityJobPostingFields,
} from './OpportunityDetailsFields';
import {
  buildOpportunityDetailsInput,
  createEmptyDetailsDraft,
} from './opportunityFormState';

interface OpportunityFormModalProps {
  companies: Company[];
  selectedCompanyId?: string;
  onRequestNewCompany: () => void;
  onClose: () => void;
  onSubmit: (input: OpportunityCreateInput) => Promise<void>;
}

export function OpportunityFormModal({
  companies,
  selectedCompanyId,
  onRequestNewCompany,
  onClose,
  onSubmit,
}: OpportunityFormModalProps) {
  const [companyId, setCompanyId] = useState(selectedCompanyId ?? companies[0]?.id ?? '');
  const [draft, setDraft] = useState(() => createEmptyDetailsDraft('High', 'Remote'));
  const [stage, setStage] = useState<Stage>('Discovered');
  const [transitionNote, setTransitionNote] = useState('');
  const [includeJobPosting, setIncludeJobPosting] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const activeCompanyId = selectedCompanyId || companyId || companies[0]?.id || '';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSaving(true);
    try {
      const detailsInput = buildOpportunityDetailsInput(
        activeCompanyId,
        draft,
        includeJobPosting,
      );
      await onSubmit({
        ...detailsInput,
        stage,
        transition_note: transitionNote,
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create Opportunity');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="New Opportunity">
      <div className="modal-card modal-card-wide">
        <div className="modal-header">
          <h2>New Opportunity</h2>
          <button type="button" className="btn-ghost" onClick={onClose}>
            Close
          </button>
        </div>
        {error && <div className="alert-error" role="alert">{error}</div>}
        <form onSubmit={handleSubmit} className="form-grid">
          <div className="form-field">
            <div className="card-top-row">
              <label htmlFor="opportunity-company-select">Company *</label>
              <button
                type="button"
                className="btn-ghost"
                onClick={onRequestNewCompany}
              >
                + New Company
              </button>
            </div>
            <select
              id="opportunity-company-select"
              required
              value={activeCompanyId}
              onChange={(e) => setCompanyId(e.target.value)}
            >
              <option value="" disabled>
                {companies.length === 0 ? 'Click "+ New Company" first...' : 'Select Company...'}
              </option>
              {companies.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>

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
          <label className="form-field form-field-full">
            <span>Initial Stage Transition Note</span>
            <input
              type="text"
              value={transitionNote}
              onChange={(e) => setTransitionNote(e.target.value)}
              placeholder="e.g. Warm intro from engineering manager"
            />
          </label>
          <label className="form-field form-field-full">
            <span>Pursuit Notes</span>
            <textarea
              rows={2}
              value={draft.notes}
              onChange={(e) => setDraft({ ...draft, notes: e.target.value })}
              placeholder="Team focus, hiring manager context..."
            />
          </label>

          <div className="form-field-full checkbox-row">
            <label>
              <input
                type="checkbox"
                checked={includeJobPosting}
                onChange={(e) => setIncludeJobPosting(e.target.checked)}
              />{' '}
              Attach formal Job Posting details now
            </label>
          </div>

          {includeJobPosting && (
            <OpportunityJobPostingFields draft={draft} onChange={setDraft} />
          )}

          <div className="modal-actions form-field-full">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-primary" disabled={saving || !activeCompanyId}>
              {saving ? 'Creating...' : 'Create Opportunity'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
