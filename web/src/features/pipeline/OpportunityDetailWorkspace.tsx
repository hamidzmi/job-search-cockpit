import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useEffect, useState } from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import { useCockpitApi } from '../../api/CockpitApiContext';
import {
  OUTCOMES,
  STAGES,
  type Outcome,
  type Stage,
} from '../../api/types';
import { formatLocalDateTime } from '../../lib/date';
import {
  OpportunityCoreFields,
  OpportunityJobPostingFields,
} from './OpportunityDetailsFields';
import { OpportunityContactsAndActions } from './OpportunityContactsAndActions';
import { OpportunityFitAndArtifacts } from './OpportunityFitAndArtifacts';
import { OpportunityInterviewRounds } from './OpportunityInterviewRounds';
import { OpportunityOfferSection } from './OpportunityOfferSection';
import {
  buildOpportunityDetailsInput,
  createEmptyDetailsDraft,
  draftFromOpportunity,
} from './opportunityFormState';

type TabKey = 'overview' | 'interviews' | 'fit' | 'offer' | 'contacts';

const TABS: { id: TabKey; label: string }[] = [
  { id: 'overview', label: 'Overview & Details' },
  { id: 'interviews', label: 'Interviews & Debriefs' },
  { id: 'fit', label: 'Fit & Artifacts' },
  { id: 'offer', label: 'Offer & Compensation' },
  { id: 'contacts', label: 'Contacts & Actions' },
];

export function OpportunityDetailWorkspace() {
  const { id = '' } = useParams<{ id: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const tabParam = searchParams.get('tab') as TabKey | null;
  const activeTab: TabKey =
    tabParam && ['overview', 'interviews', 'fit', 'offer', 'contacts'].includes(tabParam)
      ? tabParam
      : 'overview';

  const [error, setError] = useState<string | null>(null);
  const [nextStage, setNextStage] = useState<Stage>('Discovered');
  const [nextOutcome, setNextOutcome] = useState<Outcome>('Active');
  const [transitionNote, setTransitionNote] = useState('');
  const [transitionStatus, setTransitionStatus] = useState<string | null>(null);

  const [draft, setDraft] = useState(() => createEmptyDetailsDraft('Medium', 'Remote'));
  const [detailsStatus, setDetailsStatus] = useState<string | null>(null);

  const {
    data: opportunity,
    isLoading,
    error: queryError,
  } = useQuery({
    queryKey: ['opportunity', id],
    queryFn: () => api.opportunities.getOpportunity(id),
    enabled: Boolean(id),
  });

  useEffect(() => {
    if (!opportunity) return;
    setNextStage(opportunity.stage);
    setNextOutcome(opportunity.outcome);
    setDraft(draftFromOpportunity(opportunity));
  }, [opportunity]);

  const transitionMutation = useMutation({
    mutationFn: () =>
      api.opportunities.transitionOpportunity(id, {
        stage: nextStage,
        outcome: nextOutcome,
        note: transitionNote,
      }),
    onSuccess: async () => {
      setTransitionNote('');
      setTransitionStatus('Stage / Outcome transition recorded.');
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['opportunity', id] }),
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
      ]);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to transition Opportunity');
    },
  });

  const detailsMutation = useMutation({
    mutationFn: () =>
      api.opportunities.updateOpportunity(
        id,
        buildOpportunityDetailsInput(opportunity?.company_id ?? '', draft, true),
      ),
    onSuccess: async () => {
      setDetailsStatus('Opportunity & Job Posting details saved.');
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['opportunity', id] }),
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
      ]);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to update Opportunity');
    },
  });

  const handleTransitionSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!opportunity) return;
    setTransitionStatus(null);
    setError(null);
    await transitionMutation.mutateAsync();
  };

  const handleDetailsSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!opportunity) return;
    setDetailsStatus(null);
    setError(null);
    await detailsMutation.mutateAsync();
  };

  const setTab = (key: TabKey) => {
    setSearchParams({ tab: key });
  };

  if (isLoading) {
    return (
      <div className="workspace-container">
        <p>Loading Opportunity...</p>
      </div>
    );
  }

  if (!opportunity) {
    const loadMessage =
      queryError instanceof Error ? queryError.message : 'Opportunity not found';
    return (
      <div className="workspace-container">
        <p className="alert-error">{error ?? loadMessage}</p>
        <Link to="/opportunities" className="btn-secondary">
          Back to Pipeline
        </Link>
      </div>
    );
  }

  const transitions = opportunity.stage_transitions ?? [];

  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <div className="breadcrumb-row">
            <Link to="/opportunities" className="breadcrumb-link">
              Pipeline
            </Link>
            <span>/</span>
            <span className="company-pill">{opportunity.company_name}</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginTop: '0.25rem' }}>
            <h2 className="workspace-title">{opportunity.role_title}</h2>
            <span
              className={`priority-badge priority-${opportunity.priority.toLowerCase()}`}
            >
              {opportunity.priority}
            </span>
          </div>
          <div className="detail-badges" style={{ marginTop: '0.4rem' }}>
            <span className="stage-badge">Stage: {opportunity.stage}</span>
            <span
              className={`outcome-badge outcome-${opportunity.outcome
                .toLowerCase()
                .replace(/\s+/g, '-')}`}
            >
              Outcome: {opportunity.outcome}
            </span>
            <span className="timestamp-text" data-testid="detail-last-activity">
              Last Activity: {formatLocalDateTime(opportunity.last_activity_at)}
            </span>
            <span className="timestamp-text" data-testid="detail-updated-at">
              Updated: {formatLocalDateTime(opportunity.updated_at)}
            </span>
          </div>
        </div>
      </header>

      {error && (
        <div className="alert-error" role="alert">
          {error}
        </div>
      )}

      {/* Tab Pill Navigation */}
      <div
        className="opportunity-tab-bar"
        role="tablist"
        aria-label="Opportunity Workspace Tabs"
      >
        {TABS.map((tab) => (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={activeTab === tab.id}
            className={`btn-tab ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setTab(tab.id)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* 1. Overview Tab */}
      {activeTab === 'overview' && (
        <div className="detail-grid">
          <section className="panel-card" aria-label="Stage and Outcome Transition">
            <h3>Update Stage &amp; Outcome (Orthogonal)</h3>
            <p className="panel-hint">
              Closing an Opportunity (e.g. Rejected, Ghosted, Withdrawn) preserves the Stage it reached and appends an immutable Stage Transition.
            </p>
            {transitionStatus && <div className="alert-success">{transitionStatus}</div>}
            <form onSubmit={handleTransitionSubmit} className="form-grid">
              <label className="form-field">
                <span>Stage</span>
                <select
                  aria-label="Target Stage"
                  value={nextStage}
                  onChange={(e) => setNextStage(e.target.value as Stage)}
                >
                  {STAGES.map((s) => (
                    <option key={s} value={s}>
                      {s}
                    </option>
                  ))}
                </select>
              </label>
              <label className="form-field">
                <span>Outcome</span>
                <select
                  aria-label="Target Outcome"
                  value={nextOutcome}
                  onChange={(e) => setNextOutcome(e.target.value as Outcome)}
                >
                  {OUTCOMES.map((o) => (
                    <option key={o} value={o}>
                      {o}
                    </option>
                  ))}
                </select>
              </label>
              <label className="form-field form-field-full">
                <span>Transition Context Note</span>
                <input
                  type="text"
                  aria-label="Transition Note"
                  value={transitionNote}
                  onChange={(e) => setTransitionNote(e.target.value)}
                  placeholder="Context for this Stage or Outcome transition..."
                />
              </label>
              <div className="form-field-full">
                <button type="submit" className="btn-primary">
                  Record Transition
                </button>
              </div>
            </form>
          </section>

          <section className="panel-card" aria-label="Stage Transition Timeline">
            <h3>Stage Transition History</h3>
            <ol className="timeline-list">
              {transitions.map((tr) => (
                <li key={tr.id} className="timeline-item">
                  <div className="timeline-header">
                    <strong>
                      {tr.from_stage
                        ? `${tr.from_stage} → ${tr.to_stage}`
                        : `Initial Stage: ${tr.to_stage}`}
                    </strong>
                    <span className="outcome-pill">
                      {tr.from_outcome && tr.from_outcome !== tr.to_outcome
                        ? `${tr.from_outcome} → ${tr.to_outcome}`
                        : tr.to_outcome}
                    </span>
                  </div>
                  {tr.note && <p className="timeline-note">{tr.note}</p>}
                  <time className="timeline-time" dateTime={tr.transitioned_at}>
                    {formatLocalDateTime(tr.transitioned_at)}
                  </time>
                </li>
              ))}
            </ol>
          </section>

          <section
            className="panel-card panel-card-full"
            aria-label="Opportunity and Job Posting Details"
          >
            <h3>Opportunity &amp; Job Posting Details</h3>
            {detailsStatus && <div className="alert-success">{detailsStatus}</div>}
            <form onSubmit={handleDetailsSubmit} className="form-grid">
              <OpportunityCoreFields draft={draft} onChange={setDraft} />
              <OpportunityJobPostingFields draft={draft} onChange={setDraft} />
              <label className="form-field form-field-full">
                <span>Pursuit Notes</span>
                <textarea
                  rows={2}
                  value={draft.notes}
                  onChange={(e) => setDraft({ ...draft, notes: e.target.value })}
                />
              </label>
              <div className="form-field-full">
                <button type="submit" className="btn-primary">
                  Save Details &amp; Job Posting
                </button>
              </div>
            </form>
          </section>
        </div>
      )}

      {/* 2. Interviews Tab */}
      {activeTab === 'interviews' && (
        <OpportunityInterviewRounds
          opportunityId={opportunity.id}
          companyId={opportunity.company_id}
          companyName={opportunity.company_name}
        />
      )}

      {/* 3. Fit & Artifacts Tab */}
      {activeTab === 'fit' && (
        <OpportunityFitAndArtifacts opportunityId={opportunity.id} />
      )}

      {/* 4. Offer Tab */}
      {activeTab === 'offer' && (
        <OpportunityOfferSection opportunityId={opportunity.id} />
      )}

      {/* 5. Contacts & Actions Tab */}
      {activeTab === 'contacts' && (
        <OpportunityContactsAndActions opportunityId={opportunity.id} />
      )}
    </div>
  );
}
