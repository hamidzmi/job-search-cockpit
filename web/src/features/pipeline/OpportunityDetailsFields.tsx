import React from 'react';
import {
  PRIORITIES,
  REMOTE_POLICIES,
  type Priority,
  type RemotePolicy,
} from '../../api/types';
import type { OpportunityDetailsDraft } from './opportunityFormState';

interface OpportunityCoreFieldsProps {
  draft: OpportunityDetailsDraft;
  onChange: (next: OpportunityDetailsDraft) => void;
}

export function OpportunityCoreFields({ draft, onChange }: OpportunityCoreFieldsProps) {
  const update = <K extends keyof OpportunityDetailsDraft>(
    key: K,
    value: OpportunityDetailsDraft[K],
  ) => {
    onChange({ ...draft, [key]: value });
  };

  return (
    <>
      <label className="form-field">
        <span>Role Title *</span>
        <input
          type="text"
          required
          value={draft.roleTitle}
          onChange={(e) => update('roleTitle', e.target.value)}
          placeholder="e.g. Staff Backend Engineer"
        />
      </label>
      <label className="form-field">
        <span>Seniority</span>
        <input
          type="text"
          value={draft.seniority}
          onChange={(e) => update('seniority', e.target.value)}
          placeholder="e.g. Staff / L6"
        />
      </label>
      <label className="form-field">
        <span>Location</span>
        <input
          type="text"
          value={draft.location}
          onChange={(e) => update('location', e.target.value)}
          placeholder="e.g. New York, NY"
        />
      </label>
      <label className="form-field">
        <span>Remote Policy</span>
        <select
          value={draft.remotePolicy}
          onChange={(e) => update('remotePolicy', e.target.value as RemotePolicy)}
        >
          {REMOTE_POLICIES.map((policy) => (
            <option key={policy} value={policy}>
              {policy}
            </option>
          ))}
        </select>
      </label>
      <label className="form-field">
        <span>Priority</span>
        <select
          value={draft.priority}
          onChange={(e) => update('priority', e.target.value as Priority)}
        >
          {PRIORITIES.map((priority) => (
            <option key={priority} value={priority}>
              {priority}
            </option>
          ))}
        </select>
      </label>
    </>
  );
}

export function OpportunityJobPostingFields({
  draft,
  onChange,
}: OpportunityCoreFieldsProps) {
  const update = <K extends keyof OpportunityDetailsDraft>(
    key: K,
    value: OpportunityDetailsDraft[K],
  ) => {
    onChange({ ...draft, [key]: value });
  };

  return (
    <>
      <label className="form-field">
        <span>Job Posting URL</span>
        <input
          type="url"
          value={draft.sourceUrl}
          onChange={(e) => update('sourceUrl', e.target.value)}
          placeholder="https://..."
        />
      </label>
      <label className="form-field">
        <span>Stated Min Compensation</span>
        <input
          type="number"
          min={0}
          value={draft.statedMinComp}
          onChange={(e) => update('statedMinComp', e.target.value)}
          placeholder="200000"
        />
      </label>
      <label className="form-field">
        <span>Stated Max Compensation</span>
        <input
          type="number"
          min={0}
          value={draft.statedMaxComp}
          onChange={(e) => update('statedMaxComp', e.target.value)}
          placeholder="250000"
        />
      </label>
      <label className="form-field">
        <span>Currency</span>
        <input
          type="text"
          value={draft.currency}
          onChange={(e) => update('currency', e.target.value)}
          placeholder="USD"
        />
      </label>
      <label className="form-field">
        <span>Tech Stack Tags (comma-separated)</span>
        <input
          type="text"
          value={draft.techStackText}
          onChange={(e) => update('techStackText', e.target.value)}
          placeholder="Go, SQLite, React, TypeScript"
        />
      </label>
      <label className="form-field form-field-full">
        <span>Requirements (one per line)</span>
        <textarea
          rows={3}
          value={draft.requirementsText}
          onChange={(e) => update('requirementsText', e.target.value)}
          placeholder="8+ years backend systems&#10;Distributed consensus experience"
        />
      </label>
      <label className="form-field form-field-full">
        <span>Job Posting Raw Text</span>
        <textarea
          rows={4}
          value={draft.rawText}
          onChange={(e) => update('rawText', e.target.value)}
          placeholder="Paste full Job Posting description..."
        />
      </label>
    </>
  );
}
