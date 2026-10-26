import React, { useState } from 'react';
import {
  REMOTE_POLICIES,
  type Company,
  type CompanyUpsertInput,
} from '../../api/types';

interface CompanyFormModalProps {
  initial?: Company | null;
  onClose: () => void;
  onSubmit: (input: CompanyUpsertInput) => Promise<void>;
}

export function CompanyFormModal({ initial, onClose, onSubmit }: CompanyFormModalProps) {
  const [name, setName] = useState(initial?.name ?? '');
  const [domain, setDomain] = useState(initial?.domain ?? '');
  const [headquarters, setHeadquarters] = useState(initial?.headquarters ?? '');
  const [remotePolicy, setRemotePolicy] = useState(initial?.remote_policy ?? 'Hybrid');
  const [sizeStage, setSizeStage] = useState(initial?.size_stage ?? '');
  const [notes, setNotes] = useState(initial?.notes ?? '');
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSaving(true);
    try {
      await onSubmit({
        name,
        domain,
        headquarters,
        remote_policy: remotePolicy,
        size_stage: sizeStage,
        notes,
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save Company');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label={initial ? 'Edit Company' : 'New Company'}>
      <div className="modal-card">
        <div className="modal-header">
          <h2>{initial ? `Edit Company: ${initial.name}` : 'New Company'}</h2>
          <button type="button" className="btn-ghost" onClick={onClose}>
            Close
          </button>
        </div>
        {error && <div className="alert-error" role="alert">{error}</div>}
        <form onSubmit={handleSubmit} className="form-grid">
          <label className="form-field">
            <span>Company Name *</span>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Stripe"
            />
          </label>
          <label className="form-field">
            <span>Website / Domain</span>
            <input
              type="text"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              placeholder="e.g. stripe.com"
            />
          </label>
          <label className="form-field">
            <span>Headquarters</span>
            <input
              type="text"
              value={headquarters}
              onChange={(e) => setHeadquarters(e.target.value)}
              placeholder="e.g. San Francisco, CA"
            />
          </label>
          <label className="form-field">
            <span>Remote Policy</span>
            <select value={remotePolicy} onChange={(e) => setRemotePolicy(e.target.value)}>
              {REMOTE_POLICIES.map((policy) => (
                <option key={policy} value={policy}>
                  {policy}
                </option>
              ))}
            </select>
          </label>
          <label className="form-field">
            <span>Size / Stage</span>
            <input
              type="text"
              value={sizeStage}
              onChange={(e) => setSizeStage(e.target.value)}
              placeholder="e.g. Series B (200)"
            />
          </label>
          <label className="form-field form-field-full">
            <span>Notes</span>
            <textarea
              rows={3}
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Culture, engineering org notes, product context..."
            />
          </label>
          <div className="modal-actions form-field-full">
            <button type="button" className="btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn-primary" disabled={saving}>
              {saving ? 'Saving...' : initial ? 'Save Changes' : 'Create Company'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
