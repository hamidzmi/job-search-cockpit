import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  AlertTriangle,
  ArrowRight,
  Calendar,
  CheckCircle2,
  Clock,
  ExternalLink,
  FileEdit,
  Flame,
  MessageSquare,
  Sparkles,
  TrendingUp,
  X,
} from 'lucide-react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import type {
  InteractionDirection,
  InteractionKind,
  Outcome,
  Stage,
  StaleOpportunity,
} from '../../api/types';
import { OUTCOMES, STAGES } from '../../api/types';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/Card';
import { Badge } from '../../components/ui/Badge';
import { formatLocalDateTime } from '../../lib/date';

export function FlightDeckWorkspace() {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const {
    data: flightDeck,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['flight-deck'],
    queryFn: () => api.flightDeck.getFlightDeck(),
  });

  const { data: contacts = [] } = useQuery({
    queryKey: ['contacts'],
    queryFn: () => api.contacts.listContacts(),
  });

  // State for Log Interaction Modal
  const [logInteractionOpp, setLogInteractionOpp] = useState<StaleOpportunity | null>(null);
  const [intContactId, setIntContactId] = useState('');
  const [intKind, setIntKind] = useState<InteractionKind>('Email');
  const [intDirection, setIntDirection] = useState<InteractionDirection>('Outbound');
  const [intSummary, setIntSummary] = useState('');

  // State for Advance Stage Modal
  const [stageModalOpp, setStageModalOpp] = useState<StaleOpportunity | null>(null);
  const [targetStage, setTargetStage] = useState<Stage>('Applied');
  const [stageNote, setStageNote] = useState('');

  // State for Update Outcome Modal
  const [outcomeModalOpp, setOutcomeModalOpp] = useState<StaleOpportunity | null>(null);
  const [targetOutcome, setTargetOutcome] = useState<Outcome>('Ghosted');
  const [outcomeNote, setOutcomeNote] = useState('');

  // Mutations
  const completeActionItemMutation = useMutation({
    mutationFn: async (item: {
      id: string;
      title: string;
      opportunity_id?: string;
      contact_id?: string;
      due_at: string;
      notes?: string;
    }) => {
      return api.actionItems.updateActionItem(item.id, {
        title: item.title,
        opportunity_id: item.opportunity_id,
        contact_id: item.contact_id,
        due_at: item.due_at,
        status: 'completed',
        notes: item.notes,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['flight-deck'] });
    },
  });

  const snoozeActionItemMutation = useMutation({
    mutationFn: async ({
      item,
      days,
    }: {
      item: {
        id: string;
        title: string;
        opportunity_id?: string;
        contact_id?: string;
        due_at: string;
        notes?: string;
      };
      days: number;
    }) => {
      const currentDue = new Date(item.due_at);
      const newDue = new Date(currentDue.getTime() + days * 24 * 60 * 60 * 1000);
      return api.actionItems.updateActionItem(item.id, {
        title: item.title,
        opportunity_id: item.opportunity_id,
        contact_id: item.contact_id,
        due_at: newDue.toISOString(),
        status: 'open',
        notes: item.notes,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['flight-deck'] });
    },
  });

  const logInteractionMutation = useMutation({
    mutationFn: async () => {
      if (!logInteractionOpp || !intContactId || !intSummary.trim()) return;
      return api.interactions.logInteraction({
        opportunity_id: logInteractionOpp.opportunity_id,
        contact_id: intContactId,
        kind: intKind,
        direction: intDirection,
        summary: intSummary.trim(),
      });
    },
    onSuccess: async () => {
      setLogInteractionOpp(null);
      setIntSummary('');
      await queryClient.invalidateQueries({ queryKey: ['flight-deck'] });
    },
  });

  const transitionStageMutation = useMutation({
    mutationFn: async () => {
      if (!stageModalOpp) return;
      return api.opportunities.transitionOpportunity(stageModalOpp.opportunity_id, {
        stage: targetStage,
        note: stageNote.trim() || `Advanced to ${targetStage} from Flight Deck`,
      });
    },
    onSuccess: async () => {
      setStageModalOpp(null);
      setStageNote('');
      await queryClient.invalidateQueries({ queryKey: ['flight-deck'] });
    },
  });

  const updateOutcomeMutation = useMutation({
    mutationFn: async () => {
      if (!outcomeModalOpp) return;
      return api.opportunities.transitionOpportunity(outcomeModalOpp.opportunity_id, {
        outcome: targetOutcome,
        note: outcomeNote.trim() || `Marked as ${targetOutcome} from Flight Deck`,
      });
    },
    onSuccess: async () => {
      setOutcomeModalOpp(null);
      setOutcomeNote('');
      await queryClient.invalidateQueries({ queryKey: ['flight-deck'] });
    },
  });

  if (isLoading) {
    return (
      <div className="workspace-container">
        <p className="text-muted-foreground text-sm">Loading Flight Deck...</p>
      </div>
    );
  }

  if (error || !flightDeck) {
    return (
      <div className="workspace-container">
        <p className="error-banner">Failed to load Flight Deck data.</p>
      </div>
    );
  }

  const { upcoming_rounds, missing_debriefs, action_items, stale_opportunities } = flightDeck;

  return (
    <div className="workspace-container" data-testid="flight-deck-workspace">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Flight Deck</h2>
          <p className="workspace-subtitle">
            Daily command center surfacing imminent Interview Rounds, missing Debriefs, due Action
            Items, and stale Opportunities.
          </p>
        </div>
        <div className="workspace-actions">
          <Link to="/opportunities" className="btn-secondary">
            View Pipeline
          </Link>
        </div>
      </header>

      {/* Cockpit HUD: High-signal daily metrics */}
      <div className="flight-deck-hud">
        <div className="hud-card">
          <div>
            <div className="hud-value">{action_items.length}</div>
            <div className="hud-label">Action Items</div>
          </div>
          <div className="hud-icon-badge hud-badge-action">
            <Clock className="w-5 h-5" />
          </div>
        </div>
        <div className="hud-card">
          <div>
            <div className="hud-value">{upcoming_rounds.length}</div>
            <div className="hud-label">Upcoming Rounds</div>
          </div>
          <div className="hud-icon-badge hud-badge-rounds">
            <Calendar className="w-5 h-5" />
          </div>
        </div>
        <div className="hud-card">
          <div>
            <div className="hud-value">{missing_debriefs.length}</div>
            <div className="hud-label">Debriefs Needed</div>
          </div>
          <div className="hud-icon-badge hud-badge-debriefs">
            <FileEdit className="w-5 h-5" />
          </div>
        </div>
        <div className="hud-card">
          <div>
            <div className="hud-value">{stale_opportunities.length}</div>
            <div className="hud-label">Stale Alerts</div>
          </div>
          <div className="hud-icon-badge hud-badge-stale">
            <AlertTriangle className="w-5 h-5" />
          </div>
        </div>
      </div>

      <div className="flight-deck-grid">
        {/* Left Column: Agenda & Actions */}
        <div className="flight-deck-column">
          {/* Action Items */}
          <section className="flight-deck-section" data-testid="action-items-section">
            <div className="kanban-column-header">
              <h3>
                Action Items <span className="badge-count">{action_items.length}</span>
              </h3>
            </div>
            {action_items.length === 0 ? (
              <div className="panel-card text-muted">No pending action items.</div>
            ) : (
              <div className="timeline-list">
                {action_items.map((item) => (
                  <div
                    key={item.id}
                    className="panel-card"
                    data-testid={`action-item-${item.id}`}
                    style={{
                      borderLeft: item.is_overdue
                        ? '4px solid #ef4444'
                        : item.is_due_today
                        ? '4px solid #f59e0b'
                        : '4px solid #cbd5e1',
                    }}
                  >
                    <div className="card-top-row">
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        {item.company_name && (
                          <span className="company-pill">{item.company_name}</span>
                        )}
                        {item.opportunity_role_title && (
                          <span style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                            {item.opportunity_role_title}
                          </span>
                        )}
                        {item.contact_name && (
                          <span className="tech-tag">Contact: {item.contact_name}</span>
                        )}
                      </div>
                      <div>
                        {item.is_overdue && (
                          <span
                            className="priority-badge badge-overdue"
                            style={{ marginRight: '0.35rem' }}
                          >
                            Overdue
                          </span>
                        )}
                        {item.is_due_today && (
                          <span className="priority-badge badge-due-today">
                            Due Today
                          </span>
                        )}
                      </div>
                    </div>

                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <div>
                        <h4 style={{ margin: '0.25rem 0' }}>{item.title}</h4>
                        <p style={{ margin: 0, fontSize: '0.82rem', color: 'var(--text-muted)' }}>
                          Due: {formatLocalDateTime(item.due_at)}
                        </p>
                      </div>
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button
                          type="button"
                          className="btn-primary"
                          onClick={() => completeActionItemMutation.mutate(item)}
                          aria-label={`Complete ${item.title}`}
                        >
                          Complete
                        </button>
                        <button
                          type="button"
                          className="btn-secondary"
                          onClick={() => snoozeActionItemMutation.mutate({ item, days: 1 })}
                          aria-label={`Snooze ${item.title} 1 day`}
                        >
                          Snooze (+1d)
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* Pipeline Staleness Alerts */}
          <section className="flight-deck-section" data-testid="stale-opportunities-section">
            <div className="kanban-column-header">
              <h3>
                Pipeline Staleness Alerts{' '}
                <span
                  className={`badge-count ${stale_opportunities.length > 0 ? 'badge-stale-alert' : ''}`}
                >
                  {stale_opportunities.length}
                </span>
              </h3>
            </div>
            {stale_opportunities.length === 0 ? (
              <div className="panel-card text-muted">
                No stale opportunities — your pipeline is fully up to date!
              </div>
            ) : (
              <div className="timeline-list">
                {stale_opportunities.map((opp) => (
                  <div
                    key={opp.opportunity_id}
                    className="panel-card"
                    data-testid={`stale-opp-${opp.opportunity_id}`}
                    style={{ borderLeft: '4px solid #ef4444' }}
                  >
                    <div className="card-top-row">
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <span className="company-pill">{opp.company_name}</span>
                        <Link
                          to={`/opportunities/${opp.opportunity_id}`}
                          className="card-title-link"
                        >
                          {opp.role_title}
                        </Link>
                      </div>
                      <div style={{ display: 'flex', gap: '0.35rem' }}>
                        <span className="stage-badge">{opp.stage}</span>
                        <span className="priority-badge">{opp.priority}</span>
                      </div>
                    </div>

                    <div className="stale-reason-text">
                      <span>⚠️ {opp.stale_reason}</span>
                      <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontWeight: 400 }}>
                        ({opp.days_stale} days idle)
                      </span>
                    </div>

                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        marginTop: '0.5rem',
                        borderTop: '1px solid var(--border)',
                        paddingTop: '0.5rem',
                      }}
                    >
                      <span style={{ fontSize: '0.78rem', color: 'var(--text-muted)' }}>
                        Last activity: {formatLocalDateTime(opp.last_activity_at)}
                      </span>
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button
                          type="button"
                          className="btn-primary"
                          onClick={() => {
                            setLogInteractionOpp(opp);
                            setIntContactId(contacts[0]?.id || '');
                            setIntSummary('');
                          }}
                          aria-label={`Log Interaction for ${opp.role_title}`}
                        >
                          Log Interaction
                        </button>
                        <button
                          type="button"
                          className="btn-secondary"
                          onClick={() => {
                            setStageModalOpp(opp);
                            const idx = STAGES.indexOf(opp.stage);
                            const next =
                              idx >= 0 && idx < STAGES.length - 1 ? STAGES[idx + 1] : opp.stage;
                            setTargetStage(next);
                            setStageNote('');
                          }}
                          aria-label={`Advance Stage for ${opp.role_title}`}
                        >
                          Advance Stage
                        </button>
                        <button
                          type="button"
                          className="btn-ghost"
                          onClick={() => {
                            setOutcomeModalOpp(opp);
                            setTargetOutcome('Ghosted');
                            setOutcomeNote('');
                          }}
                          aria-label={`Update Outcome for ${opp.role_title}`}
                        >
                          Update Outcome
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>

        {/* Right Column: Interviews & Debriefs */}
        <div className="flight-deck-column">
          {/* Imminent Interview Rounds (next 48 hours) */}
          <section className="flight-deck-section" data-testid="imminent-rounds-section">
            <div className="kanban-column-header">
              <h3>
                Imminent Interview Rounds (Next 48 Hours){' '}
                <span className="badge-count">{upcoming_rounds.length}</span>
              </h3>
            </div>
            {upcoming_rounds.length === 0 ? (
              <div className="panel-card text-muted">
                No interview rounds scheduled in the next 48 hours.
              </div>
            ) : (
              <div className="timeline-list">
                {upcoming_rounds.map((round) => (
                  <div
                    key={round.id}
                    className="panel-card"
                    data-testid={`upcoming-round-${round.id}`}
                    style={{ borderLeft: '4px solid #0284c7' }}
                  >
                    <div className="card-top-row">
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <span className="company-pill">{round.company_name}</span>
                        <strong>{round.opportunity_role_title}</strong>
                      </div>
                      <span className="stage-badge">{round.round_type}</span>
                    </div>

                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <div>
                        <h4 style={{ margin: '0.25rem 0' }}>{round.title}</h4>
                        <p style={{ margin: '0.15rem 0', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>
                          🕒 {formatLocalDateTime(round.scheduled_at)} ({round.duration_minutes}{' '}
                          min)
                        </p>
                      </div>
                      <Link
                        to={`/opportunities/${round.opportunity_id}?tab=interviews`}
                        className="btn-primary"
                        aria-label={`Prepare for ${round.title}`}
                      >
                        Prepare for Round
                      </Link>
                    </div>

                    <div
                      style={{
                        display: 'flex',
                        gap: '1.5rem',
                        fontSize: '0.82rem',
                        marginTop: '0.35rem',
                      }}
                    >
                      <div>
                        <span style={{ fontWeight: 600, color: 'var(--text-muted)' }}>Interviewers: </span>
                        {round.interviewer_names && round.interviewer_names.length > 0 ? (
                          round.interviewer_names.map((name) => (
                            <span
                              key={name}
                              className="tech-tag"
                              style={{ marginRight: '0.35rem' }}
                            >
                              {name}
                            </span>
                          ))
                        ) : (
                          <span className="text-muted">None linked</span>
                        )}
                      </div>
                      <div>
                        <span style={{ fontWeight: 600, color: 'var(--text-muted)' }}>Stories: </span>
                        {round.story_titles && round.story_titles.length > 0 ? (
                          round.story_titles.map((title) => (
                            <span
                              key={title}
                              className="tech-tag"
                              style={{ marginRight: '0.35rem' }}
                            >
                              {title}
                            </span>
                          ))
                        ) : (
                          <span className="text-muted">None linked</span>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* Missing Debriefs on Past/Completed Rounds */}
          <section className="flight-deck-section" data-testid="missing-debriefs-section">
            <div className="kanban-column-header">
              <h3>
                Completed Rounds Missing Debrief{' '}
                <span
                  className={`badge-count ${missing_debriefs.length > 0 ? 'badge-warning-alert' : ''}`}
                >
                  {missing_debriefs.length}
                </span>
              </h3>
            </div>
            {missing_debriefs.length === 0 ? (
              <div className="panel-card text-muted">
                All completed interview rounds have debriefs recorded.
              </div>
            ) : (
              <div className="timeline-list">
                {missing_debriefs.map((round) => (
                  <div
                    key={round.id}
                    className="panel-card panel-card-debrief"
                    data-testid={`missing-debrief-${round.id}`}
                  >
                    <div className="card-top-row">
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <span className="company-pill">{round.company_name}</span>
                        <strong>{round.opportunity_role_title}</strong>
                      </div>
                      <span className="priority-badge badge-debrief-needed">
                        Debrief Needed
                      </span>
                    </div>
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <div>
                        <h4 style={{ margin: '0.25rem 0' }}>
                          {round.title} ({round.round_type})
                        </h4>
                        <p className="debrief-hint-text">
                          Completed {formatLocalDateTime(round.scheduled_at)} — log questions and
                          reflection while fresh.
                        </p>
                      </div>
                      <Link
                        to={`/opportunities/${round.opportunity_id}?tab=interviews`}
                        className="btn-primary"
                        aria-label={`Log Debrief for ${round.title}`}
                      >
                        Log Debrief
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>
      </div>

      {/* Modal: Log Interaction */}
      {logInteractionOpp && (
        <div className="modal-backdrop" role="dialog" aria-modal="true">
          <div className="modal-card">
            <div className="modal-header">
              <h2>Log Interaction — {logInteractionOpp.company_name}</h2>
              <button
                type="button"
                className="btn-ghost"
                onClick={() => setLogInteractionOpp(null)}
              >
                ✕
              </button>
            </div>
            <form
              onSubmit={async (e) => {
                e.preventDefault();
                await logInteractionMutation.mutateAsync();
              }}
              className="form-grid"
            >
              <label className="form-field-full">
                <span>Contact *</span>
                <select
                  aria-label="Interaction Contact"
                  value={intContactId}
                  onChange={(e) => setIntContactId(e.target.value)}
                  required
                >
                  <option value="">Select a Contact...</option>
                  {contacts.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name} {c.title ? `(${c.title})` : ''}
                    </option>
                  ))}
                </select>
              </label>

              <label className="form-field">
                <span>Kind</span>
                <select
                  aria-label="Interaction Kind"
                  value={intKind}
                  onChange={(e) => setIntKind(e.target.value as InteractionKind)}
                >
                  <option value="Email">Email</option>
                  <option value="DM">DM</option>
                  <option value="Referral Request">Referral Request</option>
                  <option value="Coffee Chat">Coffee Chat</option>
                  <option value="Phone Call">Phone Call</option>
                  <option value="Other">Other</option>
                </select>
              </label>

              <label className="form-field">
                <span>Direction</span>
                <select
                  aria-label="Interaction Direction"
                  value={intDirection}
                  onChange={(e) => setIntDirection(e.target.value as InteractionDirection)}
                >
                  <option value="Outbound">Outbound</option>
                  <option value="Inbound">Inbound</option>
                </select>
              </label>

              <label className="form-field-full">
                <span>Summary *</span>
                <textarea
                  aria-label="Interaction Summary"
                  rows={3}
                  value={intSummary}
                  onChange={(e) => setIntSummary(e.target.value)}
                  placeholder="Summarize the email, note, or conversation..."
                  required
                />
              </label>

              <div className="modal-actions form-field-full">
                <button
                  type="button"
                  className="btn-secondary"
                  onClick={() => setLogInteractionOpp(null)}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-primary"
                  disabled={!intContactId || !intSummary.trim() || logInteractionMutation.isPending}
                >
                  {logInteractionMutation.isPending ? 'Logging...' : 'Log Interaction'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Advance Stage */}
      {stageModalOpp && (
        <div className="modal-backdrop" role="dialog" aria-modal="true">
          <div className="modal-card">
            <div className="modal-header">
              <h2>Advance Stage — {stageModalOpp.role_title}</h2>
              <button
                type="button"
                className="btn-ghost"
                onClick={() => setStageModalOpp(null)}
              >
                ✕
              </button>
            </div>
            <form
              onSubmit={async (e) => {
                e.preventDefault();
                await transitionStageMutation.mutateAsync();
              }}
              className="form-grid"
            >
              <label className="form-field-full">
                <span>Target Stage *</span>
                <select
                  aria-label="Target Stage"
                  value={targetStage}
                  onChange={(e) => setTargetStage(e.target.value as Stage)}
                >
                  {STAGES.map((s) => (
                    <option key={s} value={s}>
                      {s}
                    </option>
                  ))}
                </select>
              </label>

              <label className="form-field-full">
                <span>Transition Note</span>
                <input
                  type="text"
                  aria-label="Stage Transition Note"
                  value={stageNote}
                  onChange={(e) => setStageNote(e.target.value)}
                  placeholder="Optional transition reason or details..."
                />
              </label>

              <div className="modal-actions form-field-full">
                <button
                  type="button"
                  className="btn-secondary"
                  onClick={() => setStageModalOpp(null)}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-primary"
                  disabled={transitionStageMutation.isPending}
                >
                  {transitionStageMutation.isPending ? 'Advancing...' : 'Advance Stage'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: Update Outcome */}
      {outcomeModalOpp && (
        <div className="modal-backdrop" role="dialog" aria-modal="true">
          <div className="modal-card">
            <div className="modal-header">
              <h2>Update Outcome — {outcomeModalOpp.role_title}</h2>
              <button
                type="button"
                className="btn-ghost"
                onClick={() => setOutcomeModalOpp(null)}
              >
                ✕
              </button>
            </div>
            <form
              onSubmit={async (e) => {
                e.preventDefault();
                await updateOutcomeMutation.mutateAsync();
              }}
              className="form-grid"
            >
              <label className="form-field-full">
                <span>Outcome *</span>
                <select
                  aria-label="Target Outcome"
                  value={targetOutcome}
                  onChange={(e) => setTargetOutcome(e.target.value as Outcome)}
                >
                  {OUTCOMES.map((o) => (
                    <option key={o} value={o}>
                      {o}
                    </option>
                  ))}
                </select>
              </label>

              <label className="form-field-full">
                <span>Note</span>
                <input
                  type="text"
                  aria-label="Outcome Note"
                  value={outcomeNote}
                  onChange={(e) => setOutcomeNote(e.target.value)}
                  placeholder="Reason for outcome update (e.g., ghosted after 5 days)..."
                />
              </label>

              <div className="modal-actions form-field-full">
                <button
                  type="button"
                  className="btn-secondary"
                  onClick={() => setOutcomeModalOpp(null)}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-primary"
                  disabled={updateOutcomeMutation.isPending}
                >
                  {updateOutcomeMutation.isPending ? 'Updating...' : 'Update Outcome'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
