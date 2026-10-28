import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useEffect, useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import type { CandidateProfile, CandidateProfileUpsertInput } from '../../api/types';
import { formatLocalDateTime } from '../../lib/date';

interface WorkHistoryFormItem {
  company: string;
  title: string;
  start_date: string;
  end_date: string;
  summary: string;
  impact_bullets_text: string;
}

function splitCommaList(value: string): string[] {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

function splitLines(value: string): string[] {
  return value
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean);
}

export function CandidateProfilePanel() {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const profileQuery = useQuery({
    queryKey: ['profile'],
    queryFn: () => api.profile.getProfile(),
  });

  const [fullName, setFullName] = useState('');
  const [headline, setHeadline] = useState('');
  const [summary, setSummary] = useState('');
  const [skillsText, setSkillsText] = useState('');
  const [targetRolesText, setTargetRolesText] = useState('');
  const [locationsText, setLocationsText] = useState('');
  const [remotePreference, setRemotePreference] = useState('');
  const [compensationFloor, setCompensationFloor] = useState('0');
  const [targetCompensation, setTargetCompensation] = useState('0');
  const [currency, setCurrency] = useState('USD');
  const [workHistory, setWorkHistory] = useState<WorkHistoryFormItem[]>([]);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const syncFromProfile = (p: CandidateProfile) => {
    setFullName(p.full_name);
    setHeadline(p.headline);
    setSummary(p.summary);
    setSkillsText(p.skills.join(', '));
    setTargetRolesText(p.target_role_titles.join(', '));
    setLocationsText(p.location_preferences.join(', '));
    setRemotePreference(p.remote_preference);
    setCompensationFloor(String(p.compensation_floor));
    setTargetCompensation(String(p.target_compensation));
    setCurrency(p.currency || 'USD');
    setWorkHistory(
      p.work_history.map((entry) => ({
        company: entry.company,
        title: entry.title,
        start_date: entry.start_date,
        end_date: entry.end_date,
        summary: entry.summary,
        impact_bullets_text: entry.impact_bullets.join('\n'),
      })),
    );
  };

  useEffect(() => {
    if (profileQuery.data) {
      syncFromProfile(profileQuery.data);
    }
  }, [profileQuery.data]);

  const saveMutation = useMutation({
    mutationFn: (input: CandidateProfileUpsertInput) => api.profile.saveProfile(input),
    onSuccess: async (saved) => {
      setErrorMessage(null);
      setStatusMessage('Candidate Profile saved.');
      syncFromProfile(saved);
      await queryClient.invalidateQueries({ queryKey: ['profile'] });
    },
    onError: (err) => {
      setStatusMessage(null);
      setErrorMessage(err instanceof Error ? err.message : 'Failed to save Candidate Profile');
    },
  });

  const handleAddEntry = () => {
    setWorkHistory((prev) => [
      ...prev,
      {
        company: '',
        title: '',
        start_date: '',
        end_date: '',
        summary: '',
        impact_bullets_text: '',
      },
    ]);
  };

  const handleUpdateEntry = (
    idx: number,
    field: keyof WorkHistoryFormItem,
    value: string,
  ) => {
    setWorkHistory((prev) =>
      prev.map((item, i) => (i === idx ? { ...item, [field]: value } : item)),
    );
  };

  const handleRemoveEntry = (idx: number) => {
    setWorkHistory((prev) => prev.filter((_, i) => i !== idx));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setStatusMessage(null);
    setErrorMessage(null);

    const payload: CandidateProfileUpsertInput = {
      full_name: fullName,
      headline,
      summary,
      skills: splitCommaList(skillsText),
      target_role_titles: splitCommaList(targetRolesText),
      location_preferences: splitCommaList(locationsText),
      remote_preference: remotePreference,
      compensation_floor: Number(compensationFloor) || 0,
      target_compensation: Number(targetCompensation) || 0,
      currency: currency || 'USD',
      work_history: workHistory.map((entry) => ({
        company: entry.company,
        title: entry.title,
        start_date: entry.start_date,
        end_date: entry.end_date,
        summary: entry.summary,
        impact_bullets: splitLines(entry.impact_bullets_text),
      })),
    };

    try {
      await saveMutation.mutateAsync(payload);
    } catch {
      // handled in onError
    }
  };

  return (
    <section className="panel-card" aria-label="Candidate Profile">
      <div className="card-top-row">
        <div>
          <h3>Candidate Profile</h3>
          <p className="timestamp-text">
            Canonical work history, skills, target role preferences, and compensation floor.
            {profileQuery.data?.updated_at
              ? ` Last updated: ${formatLocalDateTime(profileQuery.data.updated_at)}`
              : ''}
          </p>
        </div>
      </div>

      {statusMessage && <div className="alert-success">{statusMessage}</div>}
      {errorMessage && <div className="alert-error">{errorMessage}</div>}

      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
        {/* Group 1: Professional Positioning */}
        <div className="profile-card-group" style={{ margin: 0 }}>
          <h4 style={{ margin: '0 0 0.85rem', color: 'var(--text-main)' }}>
            Professional Identity &amp; Positioning
          </h4>
          <div className="form-grid">
            <label className="form-field">
              Full Name
              <input
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                placeholder="Alex Rivera"
              />
            </label>

            <label className="form-field">
              Headline
              <input
                type="text"
                value={headline}
                onChange={(e) => setHeadline(e.target.value)}
                placeholder="Staff Distributed Systems Engineer"
              />
            </label>

            <label className="form-field form-field-full">
              Profile Summary
              <textarea
                rows={3}
                value={summary}
                onChange={(e) => setSummary(e.target.value)}
                placeholder="Concise summary of core strengths, domain expertise, and career arc..."
              />
            </label>

            <label className="form-field">
              Skills (comma-separated)
              <input
                type="text"
                value={skillsText}
                onChange={(e) => setSkillsText(e.target.value)}
                placeholder="Go, Rust, Kafka, PostgreSQL, Distributed Systems"
              />
            </label>

            <label className="form-field">
              Target Role Titles (comma-separated)
              <input
                type="text"
                value={targetRolesText}
                onChange={(e) => setTargetRolesText(e.target.value)}
                placeholder="Staff Backend Engineer, Principal Systems Engineer"
              />
            </label>

            <label className="form-field">
              Location Preferences (comma-separated)
              <input
                type="text"
                value={locationsText}
                onChange={(e) => setLocationsText(e.target.value)}
                placeholder="San Francisco, CA, Seattle, WA, Remote (US)"
              />
            </label>

            <label className="form-field">
              Remote Preference
              <input
                type="text"
                value={remotePreference}
                onChange={(e) => setRemotePreference(e.target.value)}
                placeholder="Remote or Hybrid"
              />
            </label>
          </div>
        </div>

        {/* Group 2: Compensation Architecture */}
        <div className="profile-card-group" style={{ margin: 0 }}>
          <div className="card-top-row" style={{ marginBottom: '0.85rem' }}>
            <h4 style={{ margin: 0, color: 'var(--text-main)' }}>
              Compensation Architecture
            </h4>
            {Number(compensationFloor) > 0 && (
              <span className="outcome-pill" style={{ fontWeight: 600 }}>
                Floor: {currency} {Number(compensationFloor).toLocaleString()}
              </span>
            )}
          </div>
          <div className="form-grid">
            <label className="form-field">
              Minimum Compensation Floor
              <input
                type="number"
                min={0}
                value={compensationFloor}
                onChange={(e) => setCompensationFloor(e.target.value)}
              />
            </label>

            <label className="form-field">
              Target Compensation
              <input
                type="number"
                min={0}
                value={targetCompensation}
                onChange={(e) => setTargetCompensation(e.target.value)}
              />
            </label>

            <label className="form-field">
              Currency
              <input
                type="text"
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
                placeholder="USD"
              />
            </label>
          </div>
        </div>

        {/* Group 3: Work History & Impact Bullets */}
        <div className="profile-card-group" style={{ margin: 0 }}>
          <div className="card-top-row" style={{ marginBottom: '0.85rem' }}>
            <h4 style={{ margin: 0, color: 'var(--text-main)' }}>
              Work History &amp; Quantified Impact Bullets ({workHistory.length})
            </h4>
            <button
              type="button"
              className="btn-secondary"
              onClick={handleAddEntry}
            >
              + Add Work History Entry
            </button>
          </div>

          {workHistory.length === 0 ? (
            <p className="timestamp-text">
              No work history entries added yet. Add roles and impact bullets above.
            </p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem' }}>
              {workHistory.map((entry, idx) => (
                <fieldset
                  key={idx}
                  aria-label={`Work History Entry ${idx + 1}`}
                  className="work-history-card"
                  style={{ margin: 0 }}
                >
                  <div className="card-top-row">
                    <strong>
                      Role #{idx + 1}: {entry.title || 'Untitled Role'}{' '}
                      {entry.company ? `@ ${entry.company}` : ''}
                    </strong>
                    <button
                      type="button"
                      className="btn-ghost"
                      onClick={() => handleRemoveEntry(idx)}
                    >
                      Remove Entry
                    </button>
                  </div>
                  <div className="form-grid">
                    <label className="form-field">
                      Company *
                      <input
                        type="text"
                        required
                        value={entry.company}
                        onChange={(e) => handleUpdateEntry(idx, 'company', e.target.value)}
                      />
                    </label>
                    <label className="form-field">
                      Role Title *
                      <input
                        type="text"
                        required
                        value={entry.title}
                        onChange={(e) => handleUpdateEntry(idx, 'title', e.target.value)}
                      />
                    </label>
                    <label className="form-field">
                      Start Date
                      <input
                        type="text"
                        value={entry.start_date}
                        onChange={(e) => handleUpdateEntry(idx, 'start_date', e.target.value)}
                        placeholder="2022-01"
                      />
                    </label>
                    <label className="form-field">
                      End Date
                      <input
                        type="text"
                        value={entry.end_date}
                        onChange={(e) => handleUpdateEntry(idx, 'end_date', e.target.value)}
                        placeholder="Present"
                      />
                    </label>
                    <label className="form-field form-field-full">
                      Role Summary
                      <input
                        type="text"
                        value={entry.summary}
                        onChange={(e) => handleUpdateEntry(idx, 'summary', e.target.value)}
                      />
                    </label>
                    <label className="form-field form-field-full">
                      Impact Bullets (one per line)
                      <textarea
                        rows={3}
                        value={entry.impact_bullets_text}
                        onChange={(e) =>
                          handleUpdateEntry(idx, 'impact_bullets_text', e.target.value)
                        }
                        placeholder="Cut p99 latency by 42%&#10;Led migration across 16 shards"
                      />
                    </label>
                  </div>
                </fieldset>
              ))}
            </div>
          )}
        </div>

        <div className="modal-actions form-field-full">
          <button
            type="submit"
            className="btn-primary"
            disabled={saveMutation.isPending}
          >
            Save Candidate Profile
          </button>
        </div>
      </form>
    </section>
  );
}
