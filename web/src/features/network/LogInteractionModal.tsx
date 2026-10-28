import React, { useState } from 'react';
import {
  INTERACTION_DIRECTIONS,
  INTERACTION_KINDS,
  type Contact,
  type InteractionDirection,
  type InteractionKind,
  type Opportunity,
} from '../../api/types';

interface LogInteractionModalProps {
  contacts: Contact[];
  opportunities: Opportunity[];
  defaultContactId?: string;
  defaultOpportunityId?: string;
  onClose: () => void;
  onSubmit: (input: {
    contactId: string;
    opportunityId?: string;
    kind: InteractionKind;
    direction: InteractionDirection;
    summary: string;
    followUp?: {
      title: string;
      dueAt: string;
    };
  }) => Promise<void>;
}

export function LogInteractionModal({
  contacts,
  opportunities,
  defaultContactId = '',
  defaultOpportunityId = '',
  onClose,
  onSubmit,
}: LogInteractionModalProps) {
  const [contactId, setContactId] = useState(defaultContactId);
  const [oppId, setOppId] = useState(defaultOpportunityId);
  const [kind, setKind] = useState<InteractionKind>('Email');
  const [direction, setDirection] = useState<InteractionDirection>('Outbound');
  const [summary, setSummary] = useState('');
  const [withFollowUp, setWithFollowUp] = useState(false);
  const [followUpTitle, setFollowUpTitle] = useState('');
  const [followUpDueAt, setFollowUpDueAt] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!contactId) {
      setError('Please select a Contact.');
      return;
    }
    if (!summary.trim()) {
      setError('Please provide an interaction summary.');
      return;
    }
    setError(null);
    setSubmitting(true);
    try {
      await onSubmit({
        contactId,
        opportunityId: oppId || undefined,
        kind,
        direction,
        summary: summary.trim(),
        followUp:
          withFollowUp && followUpTitle.trim() && followUpDueAt
            ? {
                title: followUpTitle.trim(),
                dueAt: followUpDueAt,
              }
            : undefined,
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to log Interaction');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-backdrop">
      <div
        className="modal-card"
        role="dialog"
        aria-modal="true"
        aria-label="Log Interaction"
      >
        <div className="modal-header">
          <h2>Log Interaction</h2>
          <button type="button" className="btn-ghost" onClick={onClose}>
            Cancel
          </button>
        </div>

        {error && (
          <div className="alert-error" role="alert">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="form-grid">
          <label className="form-field">
            <span>Contact *</span>
            <select
              aria-label="Network Interaction Contact"
              value={contactId}
              onChange={(e) => setContactId(e.target.value)}
              required
            >
              <option value="">Select Contact...</option>
              {contacts.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name} {c.company_name ? `(${c.company_name})` : ''}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field">
            <span>Opportunity (Optional)</span>
            <select
              aria-label="Network Interaction Opportunity"
              value={oppId}
              onChange={(e) => setOppId(e.target.value)}
            >
              <option value="">None / General Networking</option>
              {opportunities.map((o) => (
                <option key={o.id} value={o.id}>
                  {o.company_name} — {o.role_title}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field">
            <span>Channel / Kind</span>
            <select
              aria-label="Network Interaction Kind"
              value={kind}
              onChange={(e) => setKind(e.target.value as InteractionKind)}
            >
              {INTERACTION_KINDS.map((k) => (
                <option key={k} value={k}>
                  {k}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field">
            <span>Direction</span>
            <select
              aria-label="Network Interaction Direction"
              value={direction}
              onChange={(e) => setDirection(e.target.value as InteractionDirection)}
            >
              {INTERACTION_DIRECTIONS.map((d) => (
                <option key={d} value={d}>
                  {d}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field form-field-full">
            <span>Summary &amp; Key Takeaways *</span>
            <textarea
              aria-label="Network Interaction Summary"
              rows={3}
              required
              placeholder="Discussed team structure, upcoming openings, referral steps..."
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
            />
          </label>

          <label className="form-field form-field-full">
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
              <input
                type="checkbox"
                checked={withFollowUp}
                onChange={(e) => setWithFollowUp(e.target.checked)}
              />
              Schedule follow-up Action Item
            </span>
          </label>

          {withFollowUp && (
            <>
              <label className="form-field">
                <span>Follow-up Title *</span>
                <input
                  type="text"
                  aria-label="Network Follow-up Title"
                  placeholder="e.g. Follow up on referral submission"
                  value={followUpTitle}
                  onChange={(e) => setFollowUpTitle(e.target.value)}
                  required={withFollowUp}
                />
              </label>

              <label className="form-field">
                <span>Follow-up Due Date *</span>
                <input
                  type="date"
                  aria-label="Network Follow-up Due Date"
                  value={followUpDueAt}
                  onChange={(e) => setFollowUpDueAt(e.target.value)}
                  required={withFollowUp}
                />
              </label>
            </>
          )}

          <div className="modal-actions form-field-full">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-primary" disabled={submitting}>
              Save Interaction
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
