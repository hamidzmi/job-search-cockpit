import React, { useState } from 'react';
import {
  CONTACT_WARMTHS,
  type Company,
  type Contact,
  type ContactUpsertInput,
  type ContactWarmth,
} from '../../api/types';

interface ContactFormModalProps {
  initial?: Contact | null;
  companies: Company[];
  onClose: () => void;
  onSubmit: (input: ContactUpsertInput) => Promise<void>;
}

export function ContactFormModal({
  initial,
  companies,
  onClose,
  onSubmit,
}: ContactFormModalProps) {
  const [name, setName] = useState(initial?.name ?? '');
  const [companyId, setCompanyId] = useState(initial?.company_id ?? '');
  const [title, setTitle] = useState(initial?.title ?? '');
  const [email, setEmail] = useState(initial?.email ?? '');
  const [linkedinUrl, setLinkedinUrl] = useState(initial?.linkedin_url ?? '');
  const [warmth, setWarmth] = useState<ContactWarmth>(initial?.warmth ?? 'Warm');
  const [notes, setNotes] = useState(initial?.notes ?? '');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const isEdit = Boolean(initial);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await onSubmit({
        company_id: companyId,
        name,
        title,
        email,
        linkedin_url: linkedinUrl,
        warmth,
        notes,
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save Contact');
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
        aria-label={isEdit ? 'Edit Contact' : 'New Contact'}
      >
        <div className="modal-header">
          <h2>{isEdit ? 'Edit Contact' : 'New Contact'}</h2>
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
            <span>Contact Name *</span>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </label>
          <label className="form-field">
            <span>Current Company</span>
            <select
              value={companyId}
              onChange={(e) => setCompanyId(e.target.value)}
            >
              <option value="">Independent / Unaffiliated</option>
              {companies.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
          <label className="form-field">
            <span>Title / Role</span>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
          </label>
          <label className="form-field">
            <span>Warmth Level</span>
            <select
              value={warmth}
              onChange={(e) => setWarmth(e.target.value as ContactWarmth)}
            >
              {CONTACT_WARMTHS.map((w) => (
                <option key={w} value={w}>
                  {w}
                </option>
              ))}
            </select>
          </label>
          <label className="form-field">
            <span>Email</span>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </label>
          <label className="form-field">
            <span>LinkedIn URL</span>
            <input
              type="url"
              value={linkedinUrl}
              onChange={(e) => setLinkedinUrl(e.target.value)}
            />
          </label>
          <label className="form-field form-field-full">
            <span>Contact Notes</span>
            <textarea
              rows={2}
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
            />
          </label>
          <div className="modal-actions form-field-full">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-primary" disabled={submitting}>
              {isEdit ? 'Save Contact' : 'Create Contact'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
