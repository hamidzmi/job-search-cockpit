import React, { useState } from 'react';
import {
  type ActionItem,
  type Contact,
  type Opportunity,
} from '../../api/types';

interface ActionItemModalProps {
  initial?: ActionItem | null;
  contacts: Contact[];
  opportunities: Opportunity[];
  defaultContactId?: string;
  defaultOpportunityId?: string;
  onClose: () => void;
  onSubmit: (input: {
    title: string;
    contactId?: string;
    opportunityId?: string;
    dueAt: string;
    notes?: string;
  }) => Promise<void>;
}

export function ActionItemModal({
  initial,
  contacts,
  opportunities,
  defaultContactId = '',
  defaultOpportunityId = '',
  onClose,
  onSubmit,
}: ActionItemModalProps) {
  const [title, setTitle] = useState(initial?.title ?? '');
  const [contactId, setContactId] = useState(
    initial?.contact_id ?? defaultContactId,
  );
  const [opportunityId, setOpportunityId] = useState(
    initial?.opportunity_id ?? defaultOpportunityId,
  );
  const [dueAt, setDueAt] = useState(
    initial?.due_at ? initial.due_at.slice(0, 10) : '',
  );
  const [notes, setNotes] = useState(initial?.notes ?? '');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const isEdit = Boolean(initial);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setError('Title is required.');
      return;
    }
    if (!dueAt) {
      setError('Due date is required.');
      return;
    }
    setError(null);
    setSubmitting(true);
    try {
      await onSubmit({
        title: title.trim(),
        contactId: contactId || undefined,
        opportunityId: opportunityId || undefined,
        dueAt,
        notes: notes.trim() || undefined,
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save Action Item');
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
        aria-label={isEdit ? 'Edit Action Item' : 'New Action Item'}
      >
        <div className="modal-header">
          <h2>{isEdit ? 'Edit Action Item' : 'New Action Item'}</h2>
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
          <label className="form-field form-field-full">
            <span>Action Item Title *</span>
            <input
              type="text"
              aria-label="Network Action Item Title"
              required
              placeholder="e.g. Follow up with hiring manager after screen"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
          </label>

          <label className="form-field">
            <span>Due Date *</span>
            <input
              type="date"
              aria-label="Network Action Item Due Date"
              required
              value={dueAt}
              onChange={(e) => setDueAt(e.target.value)}
            />
          </label>

          <label className="form-field">
            <span>Contact (Optional)</span>
            <select
              aria-label="Network Action Item Contact"
              value={contactId}
              onChange={(e) => setContactId(e.target.value)}
            >
              <option value="">None</option>
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
              aria-label="Network Action Item Opportunity"
              value={opportunityId}
              onChange={(e) => setOpportunityId(e.target.value)}
            >
              <option value="">None</option>
              {opportunities.map((o) => (
                <option key={o.id} value={o.id}>
                  {o.company_name} — {o.role_title}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field form-field-full">
            <span>Task Notes (Optional)</span>
            <textarea
              rows={2}
              placeholder="Extra context or links..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
            />
          </label>

          <div className="modal-actions form-field-full">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-primary" disabled={submitting}>
              {isEdit ? 'Update Action Item' : 'Add Action Item'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
