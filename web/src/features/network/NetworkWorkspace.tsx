import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useCockpitApi } from '../../api/CockpitApiContext';
import {
  CONTACT_WARMTHS,
  INTERACTION_DIRECTIONS,
  INTERACTION_KINDS,
  STAKEHOLDER_ROLES,
  type ActionItem,
  type Company,
  type CompanyUpsertInput,
  type Contact,
  type ContactUpsertInput,
  type ContactWarmth,
  type InteractionDirection,
  type InteractionKind,
  type StakeholderRole,
} from '../../api/types';
import { formatLocalDate, formatLocalDateTime } from '../../lib/date';
import { CompanyFormModal } from '../pipeline/CompanyFormModal';
import { ContactFormModal } from './ContactFormModal';

export function NetworkWorkspace() {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const [editingCompany, setEditingCompany] = useState<Company | null>(null);
  const [showCreateCompanyModal, setShowCreateCompanyModal] = useState(false);

  const [editingContact, setEditingContact] = useState<Contact | null>(null);
  const [showCreateContactModal, setShowCreateContactModal] = useState(false);

  const [contactSearch, setContactSearch] = useState('');
  const [warmthFilter, setWarmthFilter] = useState<ContactWarmth | ''>('');
  const [companyFilter, setCompanyFilter] = useState('');

  const [linkOppByContact, setLinkOppByContact] = useState<Record<string, string>>({});
  const [linkRoleByContact, setLinkRoleByContact] = useState<Record<string, StakeholderRole>>({});

  // Quick Interaction logging form in Network workspace
  const [intContactId, setIntContactId] = useState('');
  const [intOppId, setIntOppId] = useState('');
  const [intKind, setIntKind] = useState<InteractionKind>('Email');
  const [intDirection, setIntDirection] = useState<InteractionDirection>('Outbound');
  const [intSummary, setIntSummary] = useState('');
  const [intWithFollowUp, setIntWithFollowUp] = useState(false);
  const [intFollowUpTitle, setIntFollowUpTitle] = useState('');
  const [intFollowUpDueAt, setIntFollowUpDueAt] = useState('');

  // Standalone Action Item creation/editing in Network workspace
  const [actionTitle, setActionTitle] = useState('');
  const [actionContactId, setActionContactId] = useState('');
  const [actionOppId, setActionOppId] = useState('');
  const [actionDueAt, setActionDueAt] = useState('');
  const [editingActionItem, setEditingActionItem] = useState<ActionItem | null>(null);

  const companiesQuery = useQuery({
    queryKey: ['companies'],
    queryFn: () => api.companies.listCompanies(),
  });

  const opportunitiesQuery = useQuery({
    queryKey: ['opportunities', {}],
    queryFn: () => api.opportunities.listOpportunities(),
  });

  const contactsQuery = useQuery({
    queryKey: [
      'contacts',
      { q: contactSearch, warmth: warmthFilter, company_id: companyFilter },
    ],
    queryFn: () =>
      api.contacts.listContacts({
        q: contactSearch || undefined,
        warmth: warmthFilter || undefined,
        company_id: companyFilter || undefined,
      }),
  });

  const interactionsQuery = useQuery({
    queryKey: ['interactions', {}],
    queryFn: () => api.interactions.listInteractions(),
  });

  const actionItemsQuery = useQuery({
    queryKey: ['action-items', {}],
    queryFn: () => api.actionItems.listActionItems(),
  });

  const saveCompanyMutation = useMutation({
    mutationFn: async (input: CompanyUpsertInput) => {
      if (editingCompany) {
        return api.companies.updateCompany(editingCompany.id, input);
      }
      return api.companies.createCompany(input);
    },
    onSuccess: async () => {
      setEditingCompany(null);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['companies'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
        queryClient.invalidateQueries({ queryKey: ['contacts'] }),
      ]);
    },
  });

  const saveContactMutation = useMutation({
    mutationFn: async (input: ContactUpsertInput) => {
      if (editingContact) {
        return api.contacts.updateContact(editingContact.id, input);
      }
      return api.contacts.createContact(input);
    },
    onSuccess: async () => {
      setEditingContact(null);
      await queryClient.invalidateQueries({ queryKey: ['contacts'] });
    },
  });

  const linkOpportunityMutation = useMutation({
    mutationFn: async ({
      contactId,
      opportunityId,
      role,
    }: {
      contactId: string;
      opportunityId: string;
      role: StakeholderRole;
    }) =>
      api.contacts.linkContactToOpportunity(opportunityId, {
        contact_id: contactId,
        role,
      }),
    onSuccess: async (_, vars) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['contacts'] }),
        queryClient.invalidateQueries({
          queryKey: ['opportunity-contacts', vars.opportunityId],
        }),
      ]);
    },
  });

  const logInteractionMutation = useMutation({
    mutationFn: async () =>
      api.interactions.logInteraction({
        contact_id: intContactId,
        opportunity_id: intOppId || undefined,
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
        queryClient.invalidateQueries({ queryKey: ['interactions'] }),
        queryClient.invalidateQueries({ queryKey: ['action-items'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunity'] }),
      ]);
    },
  });

  const saveActionItemMutation = useMutation({
    mutationFn: async () => {
      if (editingActionItem) {
        return api.actionItems.updateActionItem(editingActionItem.id, {
          title: actionTitle,
          contact_id: actionContactId || undefined,
          opportunity_id: actionOppId || undefined,
          due_at: actionDueAt,
          status: editingActionItem.status,
          notes: editingActionItem.notes,
        });
      }
      return api.actionItems.createActionItem({
        title: actionTitle,
        contact_id: actionContactId || undefined,
        opportunity_id: actionOppId || undefined,
        due_at: actionDueAt,
      });
    },
    onSuccess: async () => {
      setEditingActionItem(null);
      setActionTitle('');
      setActionContactId('');
      setActionOppId('');
      setActionDueAt('');
      await queryClient.invalidateQueries({ queryKey: ['action-items'] });
    },
  });

  const updateActionStatusMutation = useMutation({
    mutationFn: async ({
      item,
      status,
    }: {
      item: ActionItem;
      status: ActionItem['status'];
    }) =>
      api.actionItems.updateActionItem(item.id, {
        title: item.title,
        contact_id: item.contact_id || undefined,
        opportunity_id: item.opportunity_id || undefined,
        due_at: item.due_at,
        status,
        notes: item.notes,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['action-items'] });
    },
  });

  const companies = companiesQuery.data ?? [];
  const opportunities = opportunitiesQuery.data ?? [];
  const contacts = contactsQuery.data ?? [];
  const interactions = interactionsQuery.data ?? [];
  const actionItems = actionItemsQuery.data ?? [];

  const handleLinkContactToOpp = async (contact: Contact) => {
    const selectedOppId = linkOppByContact[contact.id] ?? opportunities[0]?.id ?? '';
    if (!selectedOppId) return;
    const selectedRole = linkRoleByContact[contact.id] ?? 'Referrer';
    await linkOpportunityMutation.mutateAsync({
      contactId: contact.id,
      opportunityId: selectedOppId,
      role: selectedRole,
    });
  };

  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Network &amp; Companies</h2>
          <p className="workspace-subtitle">
            Manage target Companies, global Contacts, Opportunity stakeholder roles, Interactions, and Action Items.
          </p>
        </div>
        <div className="workspace-actions">
          <button
            type="button"
            className="btn-secondary"
            onClick={() => setShowCreateContactModal(true)}
          >
            + New Contact
          </button>
          <button
            type="button"
            className="btn-primary"
            onClick={() => setShowCreateCompanyModal(true)}
          >
            + New Company
          </button>
        </div>
      </header>

      <section aria-label="Global Contacts Directory">
        <div className="card-top-row">
          <h3>Global Contacts ({contacts.length})</h3>
        </div>
        <div className="filter-bar">
          <label className="filter-control">
            <span>Search Contacts</span>
            <input
              type="text"
              aria-label="Search Contacts"
              placeholder="Name, title, email, or notes..."
              value={contactSearch}
              onChange={(e) => setContactSearch(e.target.value)}
            />
          </label>
          <label className="filter-control">
            <span>Filter by Warmth</span>
            <select
              aria-label="Filter by Warmth"
              value={warmthFilter}
              onChange={(e) => setWarmthFilter(e.target.value as ContactWarmth | '')}
            >
              <option value="">All Warmth Levels</option>
              {CONTACT_WARMTHS.map((w) => (
                <option key={w} value={w}>
                  {w}
                </option>
              ))}
            </select>
          </label>
          <label className="filter-control">
            <span>Filter by Company</span>
            <select
              aria-label="Filter Contacts by Company"
              value={companyFilter}
              onChange={(e) => setCompanyFilter(e.target.value)}
            >
              <option value="">All Companies</option>
              {companies.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
        </div>

        {contacts.length === 0 ? (
          <p className="empty-state">No Contacts match the current filters.</p>
        ) : (
          <div className="company-grid">
            {contacts.map((contact) => {
              const contactInteractions = interactions.filter(
                (i) => i.contact_id === contact.id,
              );
              const selectedOppId =
                linkOppByContact[contact.id] ?? opportunities[0]?.id ?? '';
              const selectedRole = linkRoleByContact[contact.id] ?? 'Referrer';

              return (
                <article
                  key={contact.id}
                  className="panel-card"
                  aria-label={`Contact: ${contact.name}`}
                >
                  <div className="card-top-row">
                    <div>
                      <h3>{contact.name}</h3>
                      <div className="card-meta">
                        {contact.title && <span>{contact.title}</span>}
                        {contact.company_name && (
                          <span className="company-pill">{contact.company_name}</span>
                        )}
                        <span className="stage-badge">{contact.warmth}</span>
                      </div>
                    </div>
                    <button
                      type="button"
                      className="btn-secondary"
                      onClick={() => setEditingContact(contact)}
                    >
                      Edit Contact
                    </button>
                  </div>

                  <div className="card-meta">
                    {contact.email && <span>{contact.email}</span>}
                    {contact.linkedin_url && (
                      <a href={contact.linkedin_url} target="_blank" rel="noreferrer">
                        LinkedIn
                      </a>
                    )}
                  </div>
                  {contact.notes && <p className="company-notes">{contact.notes}</p>}

                  <div className="company-opportunities">
                    <h4>Linked Opportunities ({contact.linked_opportunities.length})</h4>
                    {contact.linked_opportunities.length === 0 ? (
                      <p className="timestamp-text">Not linked to any Opportunity yet.</p>
                    ) : (
                      <ul>
                        {contact.linked_opportunities.map((link) => (
                          <li key={link.opportunity_id}>
                            <Link to={`/opportunities/${link.opportunity_id}`}>
                              {link.role_title}
                            </Link>{' '}
                            <span className="company-pill">{link.company_name}</span>{' '}
                            <span className="stage-badge">{link.role}</span>{' '}
                            <span className="outcome-pill">{link.stage}</span>
                          </li>
                        ))}
                      </ul>
                    )}

                    {opportunities.length > 0 && (
                      <div className="form-grid" style={{ marginTop: '0.5rem' }}>
                        <label className="form-field">
                          <span>Opportunity</span>
                          <select
                            aria-label={`Link Opportunity for ${contact.name}`}
                            value={selectedOppId}
                            onChange={(e) =>
                              setLinkOppByContact((prev) => ({
                                ...prev,
                                [contact.id]: e.target.value,
                              }))
                            }
                          >
                            {opportunities.map((opp) => (
                              <option key={opp.id} value={opp.id}>
                                {opp.company_name} — {opp.role_title}
                              </option>
                            ))}
                          </select>
                        </label>
                        <label className="form-field">
                          <span>Role</span>
                          <select
                            aria-label={`Stakeholder Role for ${contact.name}`}
                            value={selectedRole}
                            onChange={(e) =>
                              setLinkRoleByContact((prev) => ({
                                ...prev,
                                [contact.id]: e.target.value as StakeholderRole,
                              }))
                            }
                          >
                            {STAKEHOLDER_ROLES.map((r) => (
                              <option key={r} value={r}>
                                {r}
                              </option>
                            ))}
                          </select>
                        </label>
                        <div className="form-field-full">
                          <button
                            type="button"
                            className="btn-secondary"
                            onClick={() => handleLinkContactToOpp(contact)}
                          >
                            Link to Opportunity
                          </button>
                        </div>
                      </div>
                    )}
                  </div>

                  <div>
                    <h4>Interactions ({contactInteractions.length})</h4>
                    {contactInteractions.length === 0 ? (
                      <p className="timestamp-text">No Interactions logged yet.</p>
                    ) : (
                      <ul className="timeline-list">
                        {contactInteractions.map((intr) => (
                          <li key={intr.id} className="timeline-item">
                            <div className="timeline-header">
                              <strong>
                                {intr.kind} ({intr.direction})
                              </strong>
                              <span className="timestamp-text">
                                {formatLocalDateTime(intr.occurred_at)}
                              </span>
                            </div>
                            <p className="timeline-note">{intr.summary}</p>
                            {intr.opportunity_role_title && (
                              <span className="timestamp-text">
                                Opportunity: {intr.opportunity_role_title}
                              </span>
                            )}
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                </article>
              );
            })}
          </div>
        )}
      </section>

      <div className="detail-grid">
        <section className="panel-card" aria-label="Log Network Interaction">
          <h3>Log Interaction</h3>
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
                aria-label="Network Interaction Contact"
                value={intContactId}
                onChange={(e) => setIntContactId(e.target.value)}
                required
              >
                <option value="">Select Contact...</option>
                {contacts.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="form-field">
              <span>Opportunity (Optional)</span>
              <select
                aria-label="Network Interaction Opportunity"
                value={intOppId}
                onChange={(e) => setIntOppId(e.target.value)}
              >
                <option value="">None</option>
                {opportunities.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.company_name} — {o.role_title}
                  </option>
                ))}
              </select>
            </label>
            <label className="form-field">
              <span>Kind</span>
              <select
                aria-label="Network Interaction Kind"
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
                aria-label="Network Interaction Direction"
                value={intDirection}
                onChange={(e) => setIntDirection(e.target.value as InteractionDirection)}
              >
                {INTERACTION_DIRECTIONS.map((d) => (
                  <option key={d} value={d}>
                    {d}
                  </option>
                ))}
              </select>
            </label>
            <label className="form-field form-field-full">
              <span>Summary *</span>
              <textarea
                aria-label="Network Interaction Summary"
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
                Schedule follow-up Action Item
              </span>
            </label>
            {intWithFollowUp && (
              <>
                <label className="form-field">
                  <span>Follow-up Title</span>
                  <input
                    type="text"
                    aria-label="Network Follow-up Title"
                    value={intFollowUpTitle}
                    onChange={(e) => setIntFollowUpTitle(e.target.value)}
                  />
                </label>
                <label className="form-field">
                  <span>Follow-up Due Date</span>
                  <input
                    type="date"
                    aria-label="Network Follow-up Due Date"
                    value={intFollowUpDueAt}
                    onChange={(e) => setIntFollowUpDueAt(e.target.value)}
                  />
                </label>
              </>
            )}
            <div className="form-field-full">
              <button type="submit" className="btn-primary">
                Save Interaction
              </button>
            </div>
          </form>
        </section>

        <section className="panel-card" aria-label="Network Action Items">
          <h3>Action Items ({actionItems.length})</h3>
          <form
            onSubmit={async (e) => {
              e.preventDefault();
              if (!actionTitle.trim() || !actionDueAt) return;
              await saveActionItemMutation.mutateAsync();
            }}
            className="form-grid"
          >
            <label className="form-field">
              <span>Action Item Title *</span>
              <input
                type="text"
                aria-label="Network Action Item Title"
                required
                value={actionTitle}
                onChange={(e) => setActionTitle(e.target.value)}
              />
            </label>
            <label className="form-field">
              <span>Due Date *</span>
              <input
                type="date"
                aria-label="Network Action Item Due Date"
                required
                value={actionDueAt}
                onChange={(e) => setActionDueAt(e.target.value)}
              />
            </label>
            <label className="form-field">
              <span>Contact</span>
              <select
                aria-label="Network Action Item Contact"
                value={actionContactId}
                onChange={(e) => setActionContactId(e.target.value)}
              >
                <option value="">None</option>
                {contacts.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="form-field">
              <span>Opportunity</span>
              <select
                aria-label="Network Action Item Opportunity"
                value={actionOppId}
                onChange={(e) => setActionOppId(e.target.value)}
              >
                <option value="">None</option>
                {opportunities.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.company_name} — {o.role_title}
                  </option>
                ))}
              </select>
            </label>
            <div className="form-field-full">
              <button type="submit" className="btn-primary">
                {editingActionItem ? 'Update Action Item' : 'Add Action Item'}
              </button>
            </div>
          </form>

          {actionItems.length > 0 && (
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
                    {item.opportunity_role_title && (
                      <span>Opportunity: {item.opportunity_role_title}</span>
                    )}
                  </div>
                  <div className="workspace-actions">
                    <button
                      type="button"
                      className="btn-ghost"
                      onClick={() => {
                        setEditingActionItem(item);
                        setActionTitle(item.title);
                        setActionContactId(item.contact_id);
                        setActionOppId(item.opportunity_id);
                        setActionDueAt(item.due_at.slice(0, 10));
                      }}
                    >
                      Edit
                    </button>
                    {item.status === 'open' && (
                      <>
                        <button
                          type="button"
                          className="btn-secondary"
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
      </div>

      <section aria-label="Target Companies Directory">
        <h3>Companies ({companies.length})</h3>
        {companiesQuery.isLoading ? (
          <p>Loading Companies...</p>
        ) : companies.length === 0 ? (
          <p className="empty-state">No Companies added yet.</p>
        ) : (
          <div className="company-grid">
            {companies.map((c) => {
              const companyOpps = opportunities.filter((o) => o.company_id === c.id);
              return (
                <article key={c.id} className="panel-card">
                  <div className="card-top-row">
                    <h3>{c.name}</h3>
                    <button
                      type="button"
                      className="btn-secondary"
                      onClick={() => setEditingCompany(c)}
                    >
                      Edit Company
                    </button>
                  </div>
                  <div className="card-meta">
                    {c.domain && <span>{c.domain}</span>}
                    {c.headquarters && <span>{c.headquarters}</span>}
                    {c.remote_policy && <span>{c.remote_policy}</span>}
                    {c.size_stage && <span>{c.size_stage}</span>}
                  </div>
                  {c.notes && <p className="company-notes">{c.notes}</p>}
                  <div className="company-opportunities">
                    <h4>Opportunities ({companyOpps.length})</h4>
                    {companyOpps.length === 0 ? (
                      <p className="timestamp-text">No opportunities tracked yet.</p>
                    ) : (
                      <ul>
                        {companyOpps.map((opp) => (
                          <li key={opp.id}>
                            <Link to={`/opportunities/${opp.id}`}>{opp.role_title}</Link>{' '}
                            <span className="stage-badge">{opp.stage}</span>{' '}
                            <span className="outcome-pill">{opp.outcome}</span>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                </article>
              );
            })}
          </div>
        )}
      </section>

      {(showCreateCompanyModal || editingCompany) && (
        <CompanyFormModal
          initial={editingCompany}
          onClose={() => {
            setShowCreateCompanyModal(false);
            setEditingCompany(null);
          }}
          onSubmit={async (input) => {
            await saveCompanyMutation.mutateAsync(input);
          }}
        />
      )}

      {(showCreateContactModal || editingContact) && (
        <ContactFormModal
          initial={editingContact}
          companies={companies}
          onClose={() => {
            setShowCreateContactModal(false);
            setEditingContact(null);
          }}
          onSubmit={async (input) => {
            await saveContactMutation.mutateAsync(input);
          }}
        />
      )}
    </div>
  );
}
