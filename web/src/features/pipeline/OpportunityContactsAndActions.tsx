import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import {
  INTERACTION_DIRECTIONS,
  INTERACTION_KINDS,
  STAKEHOLDER_ROLES,
  type ActionItem,
  type InteractionDirection,
  type InteractionKind,
  type StakeholderRole,
} from '../../api/types';
import { formatLocalDate, formatLocalDateTime } from '../../lib/date';

interface OpportunityContactsAndActionsProps {
  opportunityId: string;
}

export function OpportunityContactsAndActions({
  opportunityId,
}: OpportunityContactsAndActionsProps) {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const [selectedContactId, setSelectedContactId] = useState('');
  const [selectedRole, setSelectedRole] = useState<StakeholderRole>('Referrer');
  const [linkNotes, setLinkNotes] = useState('');

  const [intContactId, setIntContactId] = useState('');
  const [intKind, setIntKind] = useState<InteractionKind>('Email');
  const [intDirection, setIntDirection] = useState<InteractionDirection>('Outbound');
  const [intSummary, setIntSummary] = useState('');
  const [intWithFollowUp, setIntWithFollowUp] = useState(false);
  const [intFollowUpTitle, setIntFollowUpTitle] = useState('');
  const [intFollowUpDueAt, setIntFollowUpDueAt] = useState('');

  const [actionTitle, setActionTitle] = useState('');
  const [actionContactId, setActionContactId] = useState('');
  const [actionDueAt, setActionDueAt] = useState('');
  const [actionNotes, setActionNotes] = useState('');
  const [editingActionItem, setEditingActionItem] = useState<ActionItem | null>(null);

  const allContactsQuery = useQuery({
    queryKey: ['contacts', {}],
    queryFn: () => api.contacts.listContacts(),
  });

  const oppContactsQuery = useQuery({
    queryKey: ['opportunity-contacts', opportunityId],
    queryFn: () => api.contacts.listOpportunityContacts(opportunityId),
    enabled: Boolean(opportunityId),
  });

  const interactionsQuery = useQuery({
    queryKey: ['interactions', { opportunity_id: opportunityId }],
    queryFn: () => api.interactions.listInteractions({ opportunity_id: opportunityId }),
    enabled: Boolean(opportunityId),
  });

  const actionItemsQuery = useQuery({
    queryKey: ['action-items', { opportunity_id: opportunityId }],
    queryFn: () => api.actionItems.listActionItems({ opportunity_id: opportunityId }),
    enabled: Boolean(opportunityId),
  });

  const linkMutation = useMutation({
    mutationFn: () =>
      api.contacts.linkContactToOpportunity(opportunityId, {
        contact_id: selectedContactId,
        role: selectedRole,
        notes: linkNotes,
      }),
    onSuccess: async () => {
      setLinkNotes('');
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['opportunity-contacts', opportunityId],
        }),
        queryClient.invalidateQueries({ queryKey: ['contacts'] }),
      ]);
    },
  });

  const unlinkMutation = useMutation({
    mutationFn: (contactId: string) =>
      api.contacts.unlinkContactFromOpportunity(opportunityId, contactId),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['opportunity-contacts', opportunityId],
        }),
        queryClient.invalidateQueries({ queryKey: ['contacts'] }),
      ]);
    },
  });

  const logInteractionMutation = useMutation({
    mutationFn: () =>
      api.interactions.logInteraction({
        contact_id: intContactId,
        opportunity_id: opportunityId,
        kind: intKind,
        direction: intDirection,
        summary: intSummary,
        follow_up:
          intWithFollowUp && intFollowUpTitle.trim() && intFollowUpDueAt
            ? {
                title: intFollowUpTitle,
                due_at: intFollowUpDueAt,
              }
            : undefined,
      }),
    onSuccess: async () => {
      setIntSummary('');
      setIntWithFollowUp(false);
      setIntFollowUpTitle('');
      setIntFollowUpDueAt('');
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['interactions', { opportunity_id: opportunityId }],
        }),
        queryClient.invalidateQueries({ queryKey: ['interactions'] }),
        queryClient.invalidateQueries({
          queryKey: ['action-items', { opportunity_id: opportunityId }],
        }),
        queryClient.invalidateQueries({ queryKey: ['action-items'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunity', opportunityId] }),
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
      ]);
    },
  });

  const saveActionMutation = useMutation({
    mutationFn: () => {
      if (editingActionItem) {
        return api.actionItems.updateActionItem(editingActionItem.id, {
          title: actionTitle,
          opportunity_id: opportunityId,
          contact_id: actionContactId || undefined,
          due_at: actionDueAt,
          status: editingActionItem.status,
          notes: actionNotes,
        });
      }
      return api.actionItems.createActionItem({
        title: actionTitle,
        opportunity_id: opportunityId,
        contact_id: actionContactId || undefined,
        due_at: actionDueAt,
        notes: actionNotes,
      });
    },
    onSuccess: async () => {
      setEditingActionItem(null);
      setActionTitle('');
      setActionContactId('');
      setActionDueAt('');
      setActionNotes('');
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['action-items', { opportunity_id: opportunityId }],
        }),
        queryClient.invalidateQueries({ queryKey: ['action-items'] }),
      ]);
    },
  });

  const updateActionStatusMutation = useMutation({
    mutationFn: ({
      item,
      status,
    }: {
      item: ActionItem;
      status: ActionItem['status'];
    }) =>
      api.actionItems.updateActionItem(item.id, {
        title: item.title,
        opportunity_id: item.opportunity_id || undefined,
        contact_id: item.contact_id || undefined,
        due_at: item.due_at,
        status,
        notes: item.notes,
      }),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['action-items', { opportunity_id: opportunityId }],
        }),
        queryClient.invalidateQueries({ queryKey: ['action-items'] }),
      ]);
    },
  });

  const allContacts = allContactsQuery.data ?? [];
  const oppContacts = oppContactsQuery.data ?? [];
  const interactions = interactionsQuery.data ?? [];
  const actionItems = actionItemsQuery.data ?? [];

  return (
    <>
      <section
        className="panel-card"
        aria-label="Opportunity Contacts and Interactions"
      >
        <h3>Contacts &amp; Interactions</h3>

        <div>
          <h4>Linked Stakeholders ({oppContacts.length})</h4>
          {oppContacts.length === 0 ? (
            <p className="timestamp-text">No Contacts linked to this Opportunity yet.</p>
          ) : (
            <ul className="timeline-list">
              {oppContacts.map((oc) => (
                <li key={oc.contact_id} className="timeline-item">
                  <div className="timeline-header">
                    <strong>{oc.name}</strong>
                    <span className="stage-badge">{oc.role}</span>
                  </div>
                  <div className="card-meta">
                    {oc.title && <span>{oc.title}</span>}
                    {oc.email && <span>{oc.email}</span>}
                    <span className="outcome-pill">{oc.warmth}</span>
                    <button
                      type="button"
                      className="btn-ghost"
                      onClick={() => unlinkMutation.mutate(oc.contact_id)}
                    >
                      Unlink
                    </button>
                  </div>
                  {oc.notes && <p className="timeline-note">{oc.notes}</p>}
                </li>
              ))}
            </ul>
          )}

          <form
            onSubmit={async (e) => {
              e.preventDefault();
              if (!selectedContactId) return;
              await linkMutation.mutateAsync();
            }}
            className="form-grid"
            style={{ marginTop: '0.75rem' }}
          >
            <label className="form-field">
              <span>Contact</span>
              <select
                aria-label="Select Contact to Link"
                value={selectedContactId}
                onChange={(e) => setSelectedContactId(e.target.value)}
              >
                <option value="">Select Contact...</option>
                {allContacts.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name} {c.company_name ? `(${c.company_name})` : ''}
                  </option>
                ))}
              </select>
            </label>
            <label className="form-field">
              <span>Role</span>
              <select
                aria-label="Stakeholder Role"
                value={selectedRole}
                onChange={(e) => setSelectedRole(e.target.value as StakeholderRole)}
              >
                {STAKEHOLDER_ROLES.map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </select>
            </label>
            <label className="form-field form-field-full">
              <span>Role Context Note</span>
              <input
                type="text"
                aria-label="Stakeholder Role Note"
                value={linkNotes}
                onChange={(e) => setLinkNotes(e.target.value)}
                placeholder="e.g., Submitted internal referral..."
              />
            </label>
            <div className="form-field-full">
              <button type="submit" className="btn-secondary">
                Link Contact
              </button>
            </div>
          </form>
        </div>

        <hr style={{ border: 'none', borderTop: '1px solid #e2e8f0', width: '100%' }} />

        <div>
          <h4>Log Interaction (Advances Last Activity)</h4>
          <form
            onSubmit={async (e) => {
              e.preventDefault();
              if (!intContactId || !intSummary.trim()) return;
              await logInteractionMutation.mutateAsync();
            }}
            className="form-grid"
          >
            <label className="form-field">
              <span>Contact *</span>
              <select
                aria-label="Interaction Contact"
                value={intContactId}
                onChange={(e) => setIntContactId(e.target.value)}
                required
              >
                <option value="">Select Contact...</option>
                {allContacts.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
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
                aria-label="Interaction Direction"
                value={intDirection}
                onChange={(e) =>
                  setIntDirection(e.target.value as InteractionDirection)
                }
              >
                {INTERACTION_DIRECTIONS.map((d) => (
                  <option key={d} value={d}>
                    {d}
                  </option>
                ))}
              </select>
            </label>
            <label className="form-field form-field-full">
              <span>Interaction Summary *</span>
              <textarea
                aria-label="Interaction Summary"
                rows={2}
                required
                value={intSummary}
                onChange={(e) => setIntSummary(e.target.value)}
              />
            </label>
            <label className="form-field form-field-full">
              <span>
                <input
                  type="checkbox"
                  checked={intWithFollowUp}
                  onChange={(e) => setIntWithFollowUp(e.target.checked)}
                />{' '}
                Schedule a follow-up Action Item
              </span>
            </label>
            {intWithFollowUp && (
              <>
                <label className="form-field">
                  <span>Follow-up Action Item Title</span>
                  <input
                    type="text"
                    aria-label="Follow-up Action Item Title"
                    value={intFollowUpTitle}
                    onChange={(e) => setIntFollowUpTitle(e.target.value)}
                  />
                </label>
                <label className="form-field">
                  <span>Follow-up Due Date</span>
                  <input
                    type="date"
                    aria-label="Follow-up Due Date"
                    value={intFollowUpDueAt}
                    onChange={(e) => setIntFollowUpDueAt(e.target.value)}
                  />
                </label>
              </>
            )}
            <div className="form-field-full">
              <button type="submit" className="btn-primary">
                Log Interaction
              </button>
            </div>
          </form>

          <h4>Logged Interactions ({interactions.length})</h4>
          {interactions.length === 0 ? (
            <p className="timestamp-text">No Interactions logged for this Opportunity yet.</p>
          ) : (
            <ul className="timeline-list">
              {interactions.map((intr) => (
                <li key={intr.id} className="timeline-item">
                  <div className="timeline-header">
                    <strong>
                      {intr.contact_name} — {intr.kind} ({intr.direction})
                    </strong>
                    <span className="timestamp-text">
                      {formatLocalDateTime(intr.occurred_at)}
                    </span>
                  </div>
                  <p className="timeline-note">{intr.summary}</p>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>

      <section className="panel-card" aria-label="Opportunity Action Items">
        <h3>Action Items ({actionItems.length})</h3>
        <form
          onSubmit={async (e) => {
            e.preventDefault();
            if (!actionTitle.trim() || !actionDueAt) return;
            await saveActionMutation.mutateAsync();
          }}
          className="form-grid"
        >
          <label className="form-field">
            <span>Action Item Title *</span>
            <input
              type="text"
              aria-label="Action Item Title"
              required
              value={actionTitle}
              onChange={(e) => setActionTitle(e.target.value)}
            />
          </label>
          <label className="form-field">
            <span>Due Date *</span>
            <input
              type="date"
              aria-label="Action Item Due Date"
              required
              value={actionDueAt}
              onChange={(e) => setActionDueAt(e.target.value)}
            />
          </label>
          <label className="form-field">
            <span>Contact (Optional)</span>
            <select
              aria-label="Action Item Contact"
              value={actionContactId}
              onChange={(e) => setActionContactId(e.target.value)}
            >
              <option value="">None</option>
              {allContacts.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
          <label className="form-field">
            <span>Notes</span>
            <input
              type="text"
              aria-label="Action Item Notes"
              value={actionNotes}
              onChange={(e) => setActionNotes(e.target.value)}
            />
          </label>
          <div className="form-field-full">
            <button type="submit" className="btn-primary">
              {editingActionItem ? 'Update Action Item' : 'Create Action Item'}
            </button>
          </div>
        </form>

        {actionItems.length === 0 ? (
          <p className="timestamp-text">No Action Items for this Opportunity yet.</p>
        ) : (
          <ul className="timeline-list" style={{ marginTop: '0.75rem' }}>
            {actionItems.map((item) => (
              <li key={item.id} className="timeline-item">
                <div className="timeline-header">
                  <strong>{item.title}</strong>
                  <span className="outcome-pill">{item.status}</span>
                </div>
                <div className="card-meta">
                  <span>Due: {formatLocalDate(item.due_at)}</span>
                  {item.contact_name && <span>Contact: {item.contact_name}</span>}
                </div>
                {item.notes && <p className="timeline-note">{item.notes}</p>}
                <div className="workspace-actions">
                  <button
                    type="button"
                    className="btn-ghost"
                    aria-label={`Edit ${item.title}`}
                    onClick={() => {
                      setEditingActionItem(item);
                      setActionTitle(item.title);
                      setActionContactId(item.contact_id);
                      setActionDueAt(item.due_at.slice(0, 10));
                      setActionNotes(item.notes);
                    }}
                  >
                    Edit
                  </button>
                  {item.status === 'open' && (
                    <>
                      <button
                        type="button"
                        className="btn-secondary"
                        aria-label={`Complete ${item.title}`}
                        onClick={() =>
                          updateActionStatusMutation.mutate({
                            item,
                            status: 'completed',
                          })
                        }
                      >
                        Complete
                      </button>
                      <button
                        type="button"
                        className="btn-ghost"
                        aria-label={`Dismiss ${item.title}`}
                        onClick={() =>
                          updateActionStatusMutation.mutate({
                            item,
                            status: 'dismissed',
                          })
                        }
                      >
                        Dismiss
                      </button>
                    </>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </>
  );
}
