import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useEffect, useMemo, useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import {
  INTERVIEW_ROUND_STATUSES,
  INTERVIEW_ROUND_TYPES,
  type DebriefQuestionInput,
  type DebriefUpsertInput,
  type InterviewRound,
  type InterviewRoundStatus,
  type StoryPriorRoundRef,
} from '../../api/types';
import { formatLocalDateTime } from '../../lib/date';

interface OpportunityInterviewRoundsProps {
  opportunityId: string;
  companyId?: string;
  companyName?: string;
}

export function OpportunityInterviewRounds({
  opportunityId,
  companyId,
  companyName,
}: OpportunityInterviewRoundsProps) {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const [editingRound, setEditingRound] = useState<InterviewRound | null>(null);
  const [roundOrder, setRoundOrder] = useState('1');
  const [roundType, setRoundType] = useState<string>('Recruiter Screen');
  const [title, setTitle] = useState('Recruiter Screen');
  const [scheduledAt, setScheduledAt] = useState('');
  const [durationMinutes, setDurationMinutes] = useState(45);
  const [status, setStatus] = useState<InterviewRoundStatus>('Scheduled');
  const [prepNotes, setPrepNotes] = useState('');
  const [selectedContactIds, setSelectedContactIds] = useState<string[]>([]);
  const [selectedStoryIds, setSelectedStoryIds] = useState<string[]>([]);

  // Quick Add Contact Modal State
  const [showQuickContactModal, setShowQuickContactModal] = useState(false);
  const [quickName, setQuickName] = useState('');
  const [quickTitle, setQuickTitle] = useState('');
  const [quickEmail, setQuickEmail] = useState('');
  const [quickLinkedin, setQuickLinkedin] = useState('');
  const [quickContactSubmitting, setQuickContactSubmitting] = useState(false);
  const [quickContactError, setQuickContactError] = useState<string | null>(null);

  const [debriefRound, setDebriefRound] = useState<InterviewRound | null>(null);
  const [questionsAsked, setQuestionsAsked] = useState<DebriefQuestionInput[]>([]);
  const [wentWell, setWentWell] = useState('');
  const [toImprove, setToImprove] = useState('');
  const [signals, setSignals] = useState('');
  const [confidenceScore, setConfidenceScore] = useState(4);
  const [followUpNotes, setFollowUpNotes] = useState('');

  const roundsQuery = useQuery({
    queryKey: ['interview-rounds', opportunityId],
    queryFn: () => api.interviewRounds.listInterviewRounds(opportunityId),
    enabled: Boolean(opportunityId),
  });

  const contactsQuery = useQuery({
    queryKey: ['contacts', {}],
    queryFn: () => api.contacts.listContacts(),
  });

  const storiesQuery = useQuery({
    queryKey: ['stories', {}],
    queryFn: () => api.stories.listStories({}),
  });

  const rounds = roundsQuery.data ?? [];
  const allContacts = contactsQuery.data ?? [];
  const allStories = storiesQuery.data ?? [];

  useEffect(() => {
    if (!editingRound) {
      const maxOrder = rounds.reduce((max, r) => Math.max(max, r.round_order), 0);
      setRoundOrder(String(maxOrder + 1));
    }
  }, [rounds, editingRound]);

  const numericRoundOrder = Number(roundOrder) || 1;

  // Compute live prior-round usage for the currently selected roundOrder in the form
  const priorUsageByStoryId = useMemo(() => {
    const map: Record<string, StoryPriorRoundRef[]> = {};
    const sorted = [...rounds]
      .filter(
        (r) =>
          r.status !== 'Cancelled' &&
          r.id !== editingRound?.id &&
          r.round_order < numericRoundOrder,
      )
      .sort((a, b) => a.round_order - b.round_order);

    for (const r of sorted) {
      for (const sid of r.story_ids) {
        map[sid] = [
          ...(map[sid] ?? []),
          {
            round_id: r.id,
            round_order: r.round_order,
            round_type: r.round_type,
            title: r.title,
          },
        ];
      }
    }
    return map;
  }, [rounds, numericRoundOrder, editingRound]);

  const resetRoundForm = (nextOrder: number) => {
    setEditingRound(null);
    setRoundOrder(String(nextOrder));
    setRoundType('Recruiter Screen');
    setTitle('Recruiter Screen');
    setScheduledAt('');
    setDurationMinutes(45);
    setStatus('Scheduled');
    setPrepNotes('');
    setSelectedContactIds([]);
    setSelectedStoryIds([]);
  };

  const saveRoundMutation = useMutation({
    mutationFn: () => {
      const payload = {
        opportunity_id: opportunityId,
        round_order: Number(roundOrder) || 1,
        round_type: roundType,
        title: title.trim() || roundType,
        scheduled_at: scheduledAt ? new Date(scheduledAt).toISOString() : '',
        duration_minutes: durationMinutes,
        status,
        prep_notes: prepNotes,
        contact_ids: selectedContactIds,
        story_ids: selectedStoryIds,
      };
      if (editingRound) {
        return api.interviewRounds.updateInterviewRound(editingRound.id, payload);
      }
      return api.interviewRounds.createInterviewRound(opportunityId, payload);
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['interview-rounds', opportunityId] }),
        queryClient.invalidateQueries({ queryKey: ['opportunity', opportunityId] }),
        queryClient.invalidateQueries({ queryKey: ['flight-deck'] }),
      ]);
      resetRoundForm(rounds.length + (editingRound ? 1 : 2));
    },
  });

  const deleteRoundMutation = useMutation({
    mutationFn: (roundId: string) => api.interviewRounds.deleteInterviewRound(roundId),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['interview-rounds', opportunityId] }),
        queryClient.invalidateQueries({ queryKey: ['flight-deck'] }),
      ]);
    },
  });

  const saveDebriefMutation = useMutation({
    mutationFn: ({
      roundId,
      payload,
    }: {
      roundId: string;
      payload: DebriefUpsertInput;
    }) => api.interviewRounds.recordDebrief(roundId, payload),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['interview-rounds', opportunityId] }),
        queryClient.invalidateQueries({ queryKey: ['flight-deck'] }),
        queryClient.invalidateQueries({ queryKey: ['question-bank'] }),
      ]);
      setDebriefRound(null);
    },
  });

  const openDebriefModal = (round: InterviewRound) => {
    setDebriefRound(round);
    if (round.debrief) {
      setQuestionsAsked(
        round.debrief.questions_asked.map((q) => ({
          question: q.question,
          category: q.category,
          answered_with_story_id: q.answered_with_story_id,
          notes: q.notes,
        })),
      );
      setWentWell(round.debrief.went_well);
      setToImprove(round.debrief.to_improve);
      setSignals(round.debrief.signals);
      setConfidenceScore(round.debrief.confidence_score);
      setFollowUpNotes(round.debrief.follow_up_notes);
    } else {
      setQuestionsAsked([]);
      setWentWell('');
      setToImprove('');
      setSignals('');
      setConfidenceScore(4);
      setFollowUpNotes('');
    }
  };

  const toggleContact = (contactId: string) => {
    setSelectedContactIds((prev) =>
      prev.includes(contactId) ? prev.filter((id) => id !== contactId) : [...prev, contactId],
    );
  };

  const toggleStory = (storyId: string) => {
    setSelectedStoryIds((prev) =>
      prev.includes(storyId) ? prev.filter((id) => id !== storyId) : [...prev, storyId],
    );
  };

  const handleQuickContactSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!quickName.trim()) return;
    setQuickContactError(null);
    setQuickContactSubmitting(true);
    try {
      const createdContact = await api.contacts.createContact({
        name: quickName.trim(),
        company_id: companyId || '',
        title: quickTitle.trim(),
        email: quickEmail.trim(),
        linkedin_url: quickLinkedin.trim(),
        warmth: 'Warm',
        notes: '',
      });

      try {
        await api.contacts.linkContactToOpportunity(opportunityId, {
          contact_id: createdContact.id,
          role: 'Interviewer',
        });
      } catch {
        // Non-blocking if linking is already present or optional
      }

      setSelectedContactIds((prev) =>
        prev.includes(createdContact.id) ? prev : [...prev, createdContact.id]
      );

      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['contacts'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunity-contacts', opportunityId] }),
        queryClient.invalidateQueries({ queryKey: ['opportunity', opportunityId] }),
      ]);

      setQuickName('');
      setQuickTitle('');
      setQuickEmail('');
      setQuickLinkedin('');
      setShowQuickContactModal(false);
    } catch (err) {
      setQuickContactError(err instanceof Error ? err.message : 'Failed to create Contact');
    } finally {
      setQuickContactSubmitting(false);
    }
  };

  return (
    <section
      className="panel-card panel-card-full"
      aria-label="Opportunity Interview Rounds and Debriefs"
      style={{ background: 'transparent', border: 'none', padding: 0, boxShadow: 'none' }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '0.75rem', marginBottom: '1rem' }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '1.2rem', fontWeight: 700 }}>
            Interview Loop ({rounds.length} {rounds.length === 1 ? 'Round' : 'Rounds'})
          </h3>
          <p className="panel-hint" style={{ margin: '0.25rem 0 0' }}>
            Structured loop stages, linked STAR stories, same-loop repeat prevention, and post-round debriefs.
          </p>
        </div>
      </div>

      {rounds.length === 0 ? (
        <div className="panel-card" style={{ textAlign: 'center', padding: '2rem 1rem' }}>
          <p style={{ margin: 0, color: 'var(--text-muted)', fontSize: '0.9rem' }}>
            No Interview Rounds scheduled for this Opportunity yet. Use the form below to add your first round.
          </p>
        </div>
      ) : (
        <div className="interview-loop-timeline">
          {rounds.map((r) => {
            const isCompleted = r.status === 'Completed';
            const isCancelled = r.status === 'Cancelled';
            const cardClass = isCompleted
              ? 'round-card-completed'
              : isCancelled
              ? 'round-card-cancelled'
              : 'round-card-scheduled';

            return (
              <article key={r.id} className={`round-card-modern ${cardClass}`}>
                <div className="card-top-row">
                  <div>
                    <h4 style={{ margin: 0, fontSize: '1.05rem', fontWeight: 700 }}>
                      Round {r.round_order}: {r.title}
                    </h4>
                    <div className="card-meta" style={{ marginTop: '0.35rem' }}>
                      <span className="stage-badge">{r.round_type}</span>
                      <span className="outcome-pill">{r.status}</span>
                      {r.scheduled_at && (
                        <span>🕒 {formatLocalDateTime(r.scheduled_at)}</span>
                      )}
                      {r.duration_minutes > 0 && <span>({r.duration_minutes} mins)</span>}
                    </div>
                  </div>
                  <div className="workspace-actions">
                    <button
                      type="button"
                      className="btn-primary"
                      onClick={() => openDebriefModal(r)}
                    >
                      {r.debrief ? 'Edit Debrief' : 'Record / Edit Debrief'}
                    </button>
                    <button
                      type="button"
                      className="btn-secondary"
                      onClick={() => {
                        setEditingRound(r);
                        setRoundOrder(String(r.round_order));
                        setRoundType(r.round_type);
                        setTitle(r.title);
                        setScheduledAt(r.scheduled_at ? r.scheduled_at.slice(0, 16) : '');
                        setDurationMinutes(r.duration_minutes);
                        setStatus(r.status);
                        setPrepNotes(r.prep_notes);
                        setSelectedContactIds(r.contact_ids);
                        setSelectedStoryIds(r.story_ids);
                      }}
                    >
                      Edit Round
                    </button>
                    <button
                      type="button"
                      className="btn-ghost"
                      onClick={() => deleteRoundMutation.mutate(r.id)}
                    >
                      Delete
                    </button>
                  </div>
                </div>

                {r.prep_notes && (
                  <div className="round-prep-box">
                    <strong>Preparation Notes:</strong> {r.prep_notes}
                  </div>
                )}

                {r.contacts.length > 0 && (
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap', fontSize: '0.84rem' }}>
                    <span style={{ fontWeight: 600, color: 'var(--text-secondary)' }}>Interviewers:</span>
                    {r.contacts.map((c) => (
                      <span key={c.contact_id} className="company-pill">
                        {c.name} {c.title ? `(${c.title})` : ''}
                      </span>
                    ))}
                  </div>
                )}

                {r.stories.length > 0 && (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                    <span style={{ fontWeight: 600, fontSize: '0.84rem', color: 'var(--text-secondary)' }}>
                      Planned Stories:
                    </span>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.4rem' }}>
                      {r.stories.map((st) => (
                        <div
                          key={st.story_id}
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '0.4rem',
                            background: '#f8fafc',
                            border: '1px solid #e2e8f0',
                            borderRadius: '0.45rem',
                            padding: '0.25rem 0.55rem',
                            fontSize: '0.82rem',
                          }}
                        >
                          <span>{st.title}</span>
                          {st.used_in_earlier_rounds.length > 0 && (
                            <span
                              className="priority-badge priority-medium"
                              role="status"
                              style={{ fontSize: '0.72rem' }}
                            >
                              Repeated Story: Used in{' '}
                              {st.used_in_earlier_rounds
                                .map((p) => `Round ${p.round_order} (${p.title})`)
                                .join(', ')}
                            </span>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {r.debrief && (
                  <div className="debrief-summary-box">
                    <div className="timeline-header">
                      <strong style={{ color: '#166534', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                        <span>✓</span> Debrief Completed
                      </strong>
                      <span className="priority-badge" style={{ background: '#dcfce7', color: '#166534', border: '1px solid #86efac' }}>
                        Confidence: {r.debrief.confidence_score} / 5
                      </span>
                    </div>
                    {r.debrief.went_well && (
                      <p style={{ margin: '0.2rem 0', fontSize: '0.84rem', color: '#14532d' }}>
                        <strong>Went Well:</strong> {r.debrief.went_well}
                      </p>
                    )}
                    {r.debrief.to_improve && (
                      <p style={{ margin: '0.2rem 0', fontSize: '0.84rem', color: '#14532d' }}>
                        <strong>To Improve:</strong> {r.debrief.to_improve}
                      </p>
                    )}
                    {r.debrief.signals && (
                      <p style={{ margin: '0.2rem 0', fontSize: '0.84rem', color: '#14532d' }}>
                        <strong>Signals:</strong> {r.debrief.signals}
                      </p>
                    )}
                    {r.debrief.follow_up_notes && (
                      <p style={{ margin: '0.2rem 0', fontSize: '0.84rem', color: '#14532d' }}>
                        <strong>Follow-up:</strong> {r.debrief.follow_up_notes}
                      </p>
                    )}
                    {r.debrief.questions_asked.length > 0 && (
                      <div style={{ marginTop: '0.35rem', borderTop: '1px solid #bbf7d0', paddingTop: '0.35rem' }}>
                        <span style={{ fontSize: '0.78rem', fontWeight: 700, color: '#166534' }}>
                          Questions Recorded ({r.debrief.questions_asked.length}):
                        </span>
                        <ul style={{ margin: '0.25rem 0 0', paddingLeft: '1.25rem', fontSize: '0.82rem', color: '#14532d' }}>
                          {r.debrief.questions_asked.map((q, idx) => (
                            <li key={idx}>
                              <strong>{q.question}</strong> {q.category && `[${q.category}]`}
                            </li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                )}
              </article>
            );
          })}
        </div>
      )}

      {/* Schedule / Add / Edit Round Section */}
      <div
        className="panel-card"
        style={{
          marginTop: '1.75rem',
          background: '#ffffff',
          borderRadius: '0.75rem',
          border: '1px solid var(--border-subtle)',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
          <h4 style={{ margin: 0, fontSize: '1rem', fontWeight: 700 }}>
            {editingRound
              ? `Edit Round ${editingRound.round_order}: ${editingRound.title}`
              : 'Schedule / Add Interview Round'}
          </h4>
          {editingRound && (
            <button
              type="button"
              className="btn-ghost"
              onClick={() => resetRoundForm(rounds.length + 1)}
            >
              Cancel Edit
            </button>
          )}
        </div>

        <form
          onSubmit={async (e) => {
            e.preventDefault();
            await saveRoundMutation.mutateAsync();
          }}
          className="form-grid"
        >
          <label className="form-field">
            <span>Round Order *</span>
            <input
              type="number"
              aria-label="Interview Round Order"
              min={1}
              required
              value={roundOrder}
              onChange={(e) => setRoundOrder(e.target.value)}
            />
          </label>

          <label className="form-field">
            <span>Round Type *</span>
            <select
              aria-label="Interview Round Type"
              value={roundType}
              onChange={(e) => {
                const nextType = e.target.value;
                setRoundType(nextType);
                if (!title.trim() || INTERVIEW_ROUND_TYPES.includes(title as never)) {
                  setTitle(nextType);
                }
              }}
            >
              {INTERVIEW_ROUND_TYPES.map((rt) => (
                <option key={rt} value={rt}>
                  {rt}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field">
            <span>Custom Title / Label *</span>
            <input
              type="text"
              aria-label="Interview Round Title"
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
          </label>

          <label className="form-field">
            <span>Status</span>
            <select
              aria-label="Interview Round Status"
              value={status}
              onChange={(e) => setStatus(e.target.value as InterviewRoundStatus)}
            >
              {INTERVIEW_ROUND_STATUSES.map((st) => (
                <option key={st} value={st}>
                  {st}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field">
            <span>Scheduled Date/Time</span>
            <input
              type="datetime-local"
              aria-label="Interview Scheduled At"
              value={scheduledAt}
              onChange={(e) => setScheduledAt(e.target.value)}
            />
          </label>

          <label className="form-field">
            <span>Duration (Minutes)</span>
            <input
              type="number"
              aria-label="Interview Duration Minutes"
              min={0}
              value={durationMinutes}
              onChange={(e) => setDurationMinutes(Number(e.target.value) || 0)}
            />
          </label>

          <label className="form-field form-field-full">
            <span>Preparation Notes</span>
            <textarea
              aria-label="Preparation Notes"
              rows={2}
              value={prepNotes}
              onChange={(e) => setPrepNotes(e.target.value)}
              placeholder="Key technical topics to cover, questions for interviewer, reminders..."
            />
          </label>

          <div className="form-field">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.25rem' }}>
              <span>Participating Interviewer Contacts</span>
              <button
                type="button"
                className="btn-secondary"
                style={{ padding: '0.2rem 0.55rem', fontSize: '0.75rem' }}
                onClick={() => {
                  setQuickContactError(null);
                  setShowQuickContactModal(true);
                }}
              >
                + Quick Add Contact
              </button>
            </div>
            {allContacts.length === 0 ? (
              <p className="timestamp-text">No Contacts in directory yet.</p>
            ) : (
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '0.35rem',
                  maxHeight: '140px',
                  overflowY: 'auto',
                  border: '1px solid #e2e8f0',
                  borderRadius: '0.45rem',
                  padding: '0.5rem',
                  background: '#f8fafc',
                }}
              >
                {allContacts.map((c) => (
                  <label key={c.id} style={{ fontSize: '0.84rem', display: 'flex', alignItems: 'center', gap: '0.4rem', cursor: 'pointer' }}>
                    <input
                      type="checkbox"
                      aria-label={`Interviewer Contact: ${c.name}`}
                      checked={selectedContactIds.includes(c.id)}
                      onChange={() => toggleContact(c.id)}
                    />{' '}
                    <span>{c.name} {c.title ? `(${c.title})` : ''}</span>
                  </label>
                ))}
              </div>
            )}
          </div>

          <div className="form-field">
            <span>Planned / Used Stories (Story Bank)</span>
            {allStories.length === 0 ? (
              <p className="timestamp-text">No Stories in Story Bank yet.</p>
            ) : (
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '0.4rem',
                  maxHeight: '140px',
                  overflowY: 'auto',
                  border: '1px solid #e2e8f0',
                  borderRadius: '0.45rem',
                  padding: '0.5rem',
                  background: '#f8fafc',
                }}
              >
                {allStories.map((st) => {
                  const priors = priorUsageByStoryId[st.id] ?? [];
                  return (
                    <div key={st.id} style={{ fontSize: '0.84rem' }}>
                      <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', cursor: 'pointer' }}>
                        <input
                          type="checkbox"
                          aria-label={`Planned Story: ${st.title}`}
                          checked={selectedStoryIds.includes(st.id)}
                          onChange={() => toggleStory(st.id)}
                        />{' '}
                        <span>{st.title}</span>
                      </label>
                      {priors.length > 0 && (
                        <span
                          className="priority-badge priority-medium"
                          style={{ marginLeft: '1.4rem', marginTop: '0.15rem', display: 'inline-block' }}
                        >
                          Already used in{' '}
                          {priors
                            .map((p) => `Round ${p.round_order} (${p.title})`)
                            .join(', ')}
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          <div className="form-field-full workspace-actions" style={{ marginTop: '0.5rem' }}>
            <button type="submit" className="btn-primary">
              {editingRound ? 'Save Interview Round' : 'Add Interview Round'}
            </button>
            {editingRound && (
              <button
                type="button"
                className="btn-ghost"
                onClick={() => resetRoundForm(rounds.length + 1)}
              >
                Cancel Edit
              </button>
            )}
          </div>
        </form>
      </div>

      {debriefRound && (
        <div
          className="modal-backdrop"
          role="dialog"
          aria-modal="true"
          aria-label={`Debrief: ${debriefRound.title}`}
        >
          <div className="modal-card">
            <div className="modal-header">
              <h3>Debrief: {debriefRound.title}</h3>
              <button
                type="button"
                className="btn-ghost"
                onClick={() => setDebriefRound(null)}
              >
                ✕
              </button>
            </div>

            <form
              onSubmit={async (e) => {
                e.preventDefault();
                await saveDebriefMutation.mutateAsync({
                  roundId: debriefRound.id,
                  payload: {
                    questions_asked: questionsAsked,
                    went_well: wentWell,
                    to_improve: toImprove,
                    signals,
                    confidence_score: confidenceScore,
                    follow_up_notes: followUpNotes,
                  },
                });
              }}
              className="form-grid"
            >
              <div className="form-field-full">
                <div className="card-top-row">
                  <h4>Structured Questions Asked ({questionsAsked.length})</h4>
                  <button
                    type="button"
                    className="btn-secondary"
                    onClick={() =>
                      setQuestionsAsked((prev) => [
                        ...prev,
                        {
                          question: '',
                          category: '',
                          answered_with_story_id: '',
                          notes: '',
                        },
                      ])
                    }
                  >
                    + Add Interview Question
                  </button>
                </div>

                {questionsAsked.map((q, idx) => (
                  <div
                    key={idx}
                    className="panel-card"
                    style={{ marginTop: '0.5rem' }}
                  >
                    <div className="form-grid">
                      <label className="form-field form-field-full">
                        <span>Question Asked {idx + 1} *</span>
                        <input
                          type="text"
                          aria-label={`Question Asked ${idx + 1}`}
                          required
                          value={q.question}
                          onChange={(e) => {
                            const next = [...questionsAsked];
                            next[idx] = { ...next[idx], question: e.target.value };
                            setQuestionsAsked(next);
                          }}
                        />
                      </label>
                      <label className="form-field">
                        <span>Category</span>
                        <input
                          type="text"
                          aria-label={`Question Category ${idx + 1}`}
                          value={q.category}
                          onChange={(e) => {
                            const next = [...questionsAsked];
                            next[idx] = { ...next[idx], category: e.target.value };
                            setQuestionsAsked(next);
                          }}
                          placeholder="e.g., System Design, Behavioral"
                        />
                      </label>
                      <label className="form-field">
                        <span>Answered With Story</span>
                        <select
                          aria-label={`Answered With Story ${idx + 1}`}
                          value={q.answered_with_story_id}
                          onChange={(e) => {
                            const next = [...questionsAsked];
                            next[idx] = {
                              ...next[idx],
                              answered_with_story_id: e.target.value,
                            };
                            setQuestionsAsked(next);
                          }}
                        >
                          <option value="">None</option>
                          {allStories.map((st) => (
                            <option key={st.id} value={st.id}>
                              {st.title}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label className="form-field form-field-full">
                        <span>Question Notes</span>
                        <input
                          type="text"
                          aria-label={`Question Notes ${idx + 1}`}
                          value={q.notes}
                          onChange={(e) => {
                            const next = [...questionsAsked];
                            next[idx] = { ...next[idx], notes: e.target.value };
                            setQuestionsAsked(next);
                          }}
                        />
                      </label>
                    </div>
                  </div>
                ))}
              </div>

              <label className="form-field">
                <span>What Went Well</span>
                <textarea
                  aria-label="What Went Well"
                  rows={2}
                  value={wentWell}
                  onChange={(e) => setWentWell(e.target.value)}
                />
              </label>

              <label className="form-field">
                <span>Areas to Improve</span>
                <textarea
                  aria-label="Areas to Improve"
                  rows={2}
                  value={toImprove}
                  onChange={(e) => setToImprove(e.target.value)}
                />
              </label>

              <label className="form-field">
                <span>Company &amp; Team Signals</span>
                <textarea
                  aria-label="Company & Team Signals"
                  rows={2}
                  value={signals}
                  onChange={(e) => setSignals(e.target.value)}
                />
              </label>

              <label className="form-field">
                <span>Confidence Score (1-5)</span>
                <select
                  aria-label="Confidence Score (1-5)"
                  value={String(confidenceScore)}
                  onChange={(e) => setConfidenceScore(Number(e.target.value))}
                >
                  {[1, 2, 3, 4, 5].map((n) => (
                    <option key={n} value={String(n)}>
                      {n}
                    </option>
                  ))}
                </select>
              </label>

              <label className="form-field form-field-full">
                <span>Follow-up Notes</span>
                <textarea
                  aria-label="Follow-up Notes"
                  rows={2}
                  value={followUpNotes}
                  onChange={(e) => setFollowUpNotes(e.target.value)}
                />
              </label>

              <div className="modal-actions form-field-full">
                <button
                  type="button"
                  className="btn-secondary"
                  onClick={() => setDebriefRound(null)}
                >
                  Cancel
                </button>
                <button type="submit" className="btn-primary">
                  Save Debrief
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
      {showQuickContactModal && (
        <div
          className="modal-backdrop"
          role="dialog"
          aria-modal="true"
          aria-label="Quick Add Interviewer"
        >
          <div className="modal-card" style={{ maxWidth: '480px' }}>
            <div className="modal-header">
              <h3>Quick Add Interviewer</h3>
              <button
                type="button"
                className="btn-ghost"
                onClick={() => setShowQuickContactModal(false)}
              >
                ✕
              </button>
            </div>

            <p className="panel-hint" style={{ marginTop: 0 }}>
              Add a new interviewer{companyName ? ` to ${companyName}` : ''}. They will be automatically linked as an Interviewer and selected for this round.
            </p>

            {quickContactError && (
              <div className="alert-error" role="alert">
                {quickContactError}
              </div>
            )}

            <form onSubmit={handleQuickContactSubmit} className="form-grid">
              <label className="form-field form-field-full">
                <span>Full Name *</span>
                <input
                  type="text"
                  required
                  aria-label="Contact Name"
                  value={quickName}
                  onChange={(e) => setQuickName(e.target.value)}
                  placeholder="e.g., Sarah Connor"
                  autoFocus
                />
              </label>

              <label className="form-field form-field-full">
                <span>Role / Title</span>
                <input
                  type="text"
                  aria-label="Contact Title"
                  value={quickTitle}
                  onChange={(e) => setQuickTitle(e.target.value)}
                  placeholder="e.g., Staff Distributed Systems Engineer"
                />
              </label>

              <label className="form-field">
                <span>Email</span>
                <input
                  type="email"
                  aria-label="Contact Email"
                  value={quickEmail}
                  onChange={(e) => setQuickEmail(e.target.value)}
                  placeholder="sarah@company.com"
                />
              </label>

              <label className="form-field">
                <span>LinkedIn URL</span>
                <input
                  type="url"
                  aria-label="Contact LinkedIn"
                  value={quickLinkedin}
                  onChange={(e) => setQuickLinkedin(e.target.value)}
                  placeholder="https://linkedin.com/in/sarah"
                />
              </label>

              <div className="modal-actions form-field-full" style={{ marginTop: '0.5rem' }}>
                <button
                  type="button"
                  className="btn-secondary"
                  onClick={() => setShowQuickContactModal(false)}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn-primary"
                  disabled={quickContactSubmitting}
                >
                  {quickContactSubmitting ? 'Saving...' : 'Add & Select Interviewer'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </section>
  );
}
