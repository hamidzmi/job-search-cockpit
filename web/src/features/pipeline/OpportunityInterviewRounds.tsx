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
}

export function OpportunityInterviewRounds({
  opportunityId,
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
  // (ignoring Cancelled rounds and the round currently being edited)
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
    mutationFn: async () => {
      const payload = {
        round_order: numericRoundOrder,
        round_type: roundType,
        title: title.trim() || roundType,
        scheduled_at: scheduledAt,
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
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
      ]);
      const nextOrder =
        rounds.reduce((max, r) => Math.max(max, r.round_order), numericRoundOrder) + 1;
      resetRoundForm(nextOrder);
    },
  });

  const deleteRoundMutation = useMutation({
    mutationFn: (roundId: string) => api.interviewRounds.deleteInterviewRound(roundId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ['interview-rounds', opportunityId],
      });
    },
  });

  const saveDebriefMutation = useMutation({
    mutationFn: (input: { roundId: string; payload: DebriefUpsertInput }) =>
      api.interviewRounds.recordDebrief(input.roundId, input.payload),
    onSuccess: async () => {
      setDebriefRound(null);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['interview-rounds', opportunityId] }),
        queryClient.invalidateQueries({ queryKey: ['question-bank'] }),
      ]);
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
      prev.includes(contactId)
        ? prev.filter((id) => id !== contactId)
        : [...prev, contactId],
    );
  };

  const toggleStory = (storyId: string) => {
    setSelectedStoryIds((prev) =>
      prev.includes(storyId)
        ? prev.filter((id) => id !== storyId)
        : [...prev, storyId],
    );
  };

  return (
    <section
      className="panel-card panel-card-full"
      aria-label="Opportunity Interview Rounds and Debriefs"
    >
      <h3>Interview Rounds &amp; Debriefs ({rounds.length})</h3>
      <p className="panel-hint">
        Order company-specific Interview Rounds, attach interviewer Contacts and STAR Stories (with automatic same-loop repeat warnings), and capture post-interview Debriefs.
      </p>

      {rounds.length === 0 ? (
        <p className="timestamp-text">No Interview Rounds scheduled for this Opportunity yet.</p>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem' }}>
          {rounds.map((r) => (
            <article key={r.id} className="panel-card">
              <div className="card-top-row">
                <div>
                  <h4>
                    Round {r.round_order}: {r.title}
                  </h4>
                  <div className="card-meta">
                    <span className="stage-badge">{r.round_type}</span>
                    <span className="outcome-pill">{r.status}</span>
                    {r.scheduled_at && (
                      <span>Scheduled: {formatLocalDateTime(r.scheduled_at)}</span>
                    )}
                    {r.duration_minutes > 0 && <span>{r.duration_minutes} mins</span>}
                  </div>
                </div>
                <div className="workspace-actions">
                  <button
                    type="button"
                    className="btn-secondary"
                    onClick={() => openDebriefModal(r)}
                  >
                    Record / Edit Debrief
                  </button>
                  <button
                    type="button"
                    className="btn-ghost"
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

              {r.prep_notes && <p className="timeline-note">Prep: {r.prep_notes}</p>}

              {r.contacts.length > 0 && (
                <div>
                  <strong style={{ fontSize: '0.82rem' }}>Interviewer Contacts:</strong>{' '}
                  {r.contacts.map((c) => `${c.name}${c.title ? ` (${c.title})` : ''}`).join(', ')}
                </div>
              )}

              {r.stories.length > 0 && (
                <div>
                  <strong style={{ fontSize: '0.82rem' }}>Planned / Used Stories:</strong>
                  <ul className="timeline-list" style={{ marginTop: '0.35rem' }}>
                    {r.stories.map((st) => (
                      <li key={st.story_id} className="timeline-item">
                        <div className="timeline-header">
                          <span>{st.title}</span>
                          {st.used_in_earlier_rounds.length > 0 && (
                            <span
                              className="outcome-badge outcome-on-hold"
                              role="status"
                            >
                              Repeated Story: Used in{' '}
                              {st.used_in_earlier_rounds
                                .map((p) => `Round ${p.round_order} (${p.title})`)
                                .join(', ')}
                            </span>
                          )}
                        </div>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {r.debrief && (
                <div
                  style={{
                    borderTop: '1px solid #e2e8f0',
                    paddingTop: '0.65rem',
                    marginTop: '0.35rem',
                  }}
                >
                  <div className="timeline-header">
                    <strong>Post-Interview Debrief</strong>
                    <span className="stage-badge">
                      Confidence: {r.debrief.confidence_score} / 5
                    </span>
                  </div>
                  {r.debrief.went_well && (
                    <p className="timeline-note">
                      <strong>Went Well:</strong> {r.debrief.went_well}
                    </p>
                  )}
                  {r.debrief.to_improve && (
                    <p className="timeline-note">
                      <strong>To Improve:</strong> {r.debrief.to_improve}
                    </p>
                  )}
                  {r.debrief.signals && (
                    <p className="timeline-note">
                      <strong>Signals:</strong> {r.debrief.signals}
                    </p>
                  )}
                  {r.debrief.follow_up_notes && (
                    <p className="timeline-note">
                      <strong>Follow-up:</strong> {r.debrief.follow_up_notes}
                    </p>
                  )}
                  {r.debrief.questions_asked.length > 0 && (
                    <div style={{ marginTop: '0.4rem' }}>
                      <strong style={{ fontSize: '0.8rem' }}>
                        Questions Asked ({r.debrief.questions_asked.length}):
                      </strong>
                      <ul className="timeline-list">
                        {r.debrief.questions_asked.map((q, idx) => (
                          <li key={idx} className="timeline-item">
                            <div>
                              <strong>{q.question}</strong>{' '}
                              {q.category && (
                                <span className="tech-tag">{q.category}</span>
                              )}
                              {q.answered_with_story_title && (
                                <span className="stage-badge">
                                  Answered with: {q.answered_with_story_title}
                                </span>
                              )}
                            </div>
                            {q.notes && <p className="timeline-note">{q.notes}</p>}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}
            </article>
          ))}
        </div>
      )}

      <hr style={{ border: 'none', borderTop: '1px solid #e2e8f0', width: '100%' }} />

      <form
        onSubmit={async (e) => {
          e.preventDefault();
          await saveRoundMutation.mutateAsync();
        }}
        className="form-grid"
      >
        <h4 className="form-field-full" style={{ margin: 0 }}>
          {editingRound
            ? `Edit Round ${editingRound.round_order}: ${editingRound.title}`
            : 'Schedule / Add Interview Round'}
        </h4>

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
          />
        </label>

        <div className="form-field">
          <span>Participating Interviewer Contacts</span>
          {allContacts.length === 0 ? (
            <p className="timestamp-text">No Contacts in directory yet.</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
              {allContacts.map((c) => (
                <label key={c.id} style={{ fontSize: '0.84rem' }}>
                  <input
                    type="checkbox"
                    aria-label={`Interviewer Contact: ${c.name}`}
                    checked={selectedContactIds.includes(c.id)}
                    onChange={() => toggleContact(c.id)}
                  />{' '}
                  {c.name} {c.title ? `(${c.title})` : ''}
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
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
              {allStories.map((st) => {
                const priors = priorUsageByStoryId[st.id] ?? [];
                return (
                  <div key={st.id} style={{ fontSize: '0.84rem' }}>
                    <label>
                      <input
                        type="checkbox"
                        aria-label={`Planned Story: ${st.title}`}
                        checked={selectedStoryIds.includes(st.id)}
                        onChange={() => toggleStory(st.id)}
                      />{' '}
                      {st.title}
                    </label>
                    {priors.length > 0 && (
                      <span
                        className="outcome-badge outcome-on-hold"
                        style={{ marginLeft: '0.5rem' }}
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

        <div className="form-field-full workspace-actions">
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
                Close
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
    </section>
  );
}
