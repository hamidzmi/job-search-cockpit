import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useCockpitApi } from '../../api/CockpitApiContext';
import {
  CONTACT_WARMTHS,
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
import { ActionItemModal } from './ActionItemModal';
import { ContactFormModal } from './ContactFormModal';
import { LogInteractionModal } from './LogInteractionModal';

type NetworkTab = 'contacts' | 'companies' | 'activity';

function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length === 0 || !parts[0]) return '?';
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function NetworkWorkspace() {
  const api = useCockpitApi();
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();

  const tabParam = searchParams.get('tab') as NetworkTab | null;
  const activeTab: NetworkTab =
    tabParam && ['contacts', 'companies', 'activity'].includes(tabParam)
      ? tabParam
      : 'contacts';

  const setTab = (tab: NetworkTab) => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.set('tab', tab);
        return next;
      },
      { replace: true },
    );
  };

  const [editingCompany, setEditingCompany] = useState<Company | null>(null);
  const [showCreateCompanyModal, setShowCreateCompanyModal] = useState(false);

  const [editingContact, setEditingContact] = useState<Contact | null>(null);
  const [showCreateContactModal, setShowCreateContactModal] = useState(false);

  const [showLogInteractionModal, setShowLogInteractionModal] = useState(false);
  const [logInteractionContactId, setLogInteractionContactId] = useState('');
  const [logInteractionOppId, setLogInteractionOppId] = useState('');

  const [showActionItemModal, setShowActionItemModal] = useState(false);
  const [editingActionItem, setEditingActionItem] = useState<ActionItem | null>(null);

  const [contactSearch, setContactSearch] = useState('');
  const [warmthFilter, setWarmthFilter] = useState<ContactWarmth | ''>('');
  const [companyFilter, setCompanyFilter] = useState('');

  const [linkOppByContact, setLinkOppByContact] = useState<Record<string, string>>({});
  const [linkRoleByContact, setLinkRoleByContact] = useState<Record<string, StakeholderRole>>({});

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
      setShowCreateCompanyModal(false);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['companies'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
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
      setShowCreateContactModal(false);
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
    mutationFn: async (input: {
      contactId: string;
      opportunityId?: string;
      kind: InteractionKind;
      direction: InteractionDirection;
      summary: string;
      followUp?: {
        title: string;
        dueAt: string;
      };
    }) =>
      api.interactions.logInteraction({
        contact_id: input.contactId,
        opportunity_id: input.opportunityId,
        kind: input.kind,
        direction: input.direction,
        summary: input.summary,
        follow_up: input.followUp
          ? {
              title: input.followUp.title,
              due_at: input.followUp.dueAt,
            }
          : undefined,
      }),
    onSuccess: async () => {
      setShowLogInteractionModal(false);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['interactions'] }),
        queryClient.invalidateQueries({ queryKey: ['action-items'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunity'] }),
      ]);
    },
  });

  const saveActionItemMutation = useMutation({
    mutationFn: async (input: {
      title: string;
      contactId?: string;
      opportunityId?: string;
      dueAt: string;
      notes?: string;
    }) => {
      if (editingActionItem) {
        return api.actionItems.updateActionItem(editingActionItem.id, {
          title: input.title,
          contact_id: input.contactId,
          opportunity_id: input.opportunityId,
          due_at: input.dueAt,
          status: editingActionItem.status,
          notes: input.notes ?? editingActionItem.notes,
        });
      }
      return api.actionItems.createActionItem({
        title: input.title,
        contact_id: input.contactId,
        opportunity_id: input.opportunityId,
        due_at: input.dueAt,
        notes: input.notes,
      });
    },
    onSuccess: async () => {
      setEditingActionItem(null);
      setShowActionItemModal(false);
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

  const openActionItems = actionItems.filter((a) => a.status === 'open');

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

  const openLogInteractionForContact = (contactId: string) => {
    setLogInteractionContactId(contactId);
    setLogInteractionOppId('');
    setShowLogInteractionModal(true);
  };

  return (
    <div className="workspace-container">
      {/* Top Header */}
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
            onClick={() => {
              setLogInteractionContactId('');
              setLogInteractionOppId('');
              setShowLogInteractionModal(true);
            }}
          >
            + Log Interaction
          </button>
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

      {/* Top HUD Banner */}
      <div className="flight-deck-hud" role="region" aria-label="Network Health HUD">
        <div className="hud-card" onClick={() => setTab('contacts')} style={{ cursor: 'pointer' }}>
          <div>
            <div className="hud-value">{contacts.length}</div>
            <div className="hud-label">Global Contacts</div>
          </div>
          <div className="hud-icon-badge" style={{ background: '#e0f2fe', color: '#0369a1' }}>
            👥
          </div>
        </div>

        <div className="hud-card" onClick={() => setTab('companies')} style={{ cursor: 'pointer' }}>
          <div>
            <div className="hud-value">{companies.length}</div>
            <div className="hud-label">Target Companies</div>
          </div>
          <div className="hud-icon-badge" style={{ background: '#f1f5f9', color: '#334155' }}>
            🏢
          </div>
        </div>

        <div className="hud-card" onClick={() => setTab('activity')} style={{ cursor: 'pointer' }}>
          <div>
            <div className="hud-value">{openActionItems.length}</div>
            <div className="hud-label">Open Action Items</div>
          </div>
          <div className="hud-icon-badge" style={{ background: '#fef3c7', color: '#b45309' }}>
            ⚡
          </div>
        </div>

        <div className="hud-card" onClick={() => setTab('activity')} style={{ cursor: 'pointer' }}>
          <div>
            <div className="hud-value">{interactions.length}</div>
            <div className="hud-label">Logged Interactions</div>
          </div>
          <div className="hud-icon-badge" style={{ background: '#ecfdf5', color: '#047857' }}>
            💬
          </div>
        </div>
      </div>

      {/* Workspace Sub-Tabs */}
      <div
        className="opportunity-tab-bar"
        role="tablist"
        aria-label="Network Workspace Tabs"
      >
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'contacts'}
          className={`btn-tab ${activeTab === 'contacts' ? 'active' : ''}`}
          onClick={() => setTab('contacts')}
        >
          Contacts ({contacts.length})
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'companies'}
          className={`btn-tab ${activeTab === 'companies' ? 'active' : ''}`}
          onClick={() => setTab('companies')}
        >
          Companies ({companies.length})
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'activity'}
          className={`btn-tab ${activeTab === 'activity' ? 'active' : ''}`}
          onClick={() => setTab('activity')}
        >
          Activity &amp; Tasks ({openActionItems.length} open)
        </button>
      </div>

      {/* TAB 1: Contacts Directory */}
      {activeTab === 'contacts' && (
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
            <div className="contacts-grid">
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
                    className="panel-card contact-card"
                    aria-label={`Contact: ${contact.name}`}
                  >
                    {/* Zone 1: Header with Avatar & Warmth */}
                    <div className="contact-card-header">
                      <div className="contact-avatar-title">
                        <div className="contact-avatar">
                          {getInitials(contact.name)}
                        </div>
                        <div>
                          <h3 style={{ margin: 0, fontSize: '0.98rem' }}>{contact.name}</h3>
                          <div className="card-meta" style={{ marginTop: '0.2rem' }}>
                            {contact.title && <span>{contact.title}</span>}
                            {contact.company_name && (
                              <span className="company-pill">{contact.company_name}</span>
                            )}
                            <span
                              className={`stage-badge warmth-badge warmth-${contact.warmth.toLowerCase()}`}
                            >
                              {contact.warmth}
                            </span>
                          </div>
                        </div>
                      </div>
                      <button
                        type="button"
                        className="btn-secondary"
                        style={{ padding: '0.3rem 0.6rem', fontSize: '0.78rem' }}
                        onClick={() => setEditingContact(contact)}
                      >
                        Edit Contact
                      </button>
                    </div>

                    {/* Zone 2: Meta Row (Email, LinkedIn, Notes) */}
                    {(contact.email || contact.linkedin_url) && (
                      <div className="contact-meta-row">
                        {contact.email && (
                          <a href={`mailto:${contact.email}`} className="contact-meta-link">
                            ✉️ <span>{contact.email}</span>
                          </a>
                        )}
                        {contact.linkedin_url && (
                          <a
                            href={contact.linkedin_url}
                            target="_blank"
                            rel="noreferrer"
                            className="contact-meta-link"
                          >
                            🔗 LinkedIn
                          </a>
                        )}
                      </div>
                    )}
                    {contact.notes && (
                      <p
                        className="company-notes"
                        style={{
                          margin: 0,
                          fontSize: '0.8rem',
                          display: '-webkit-box',
                          WebkitLineClamp: 2,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden',
                        }}
                      >
                        {contact.notes}
                      </p>
                    )}

                    {/* Zone 3: Linked Opportunities */}
                    <div className="contact-opps-section">
                      <div className="contact-opps-header">
                        <h4>Linked Opportunities ({contact.linked_opportunities.length})</h4>
                      </div>
                      {contact.linked_opportunities.length === 0 ? (
                        <p className="timestamp-text" style={{ margin: 0 }}>
                          Not linked to any Opportunity yet.
                        </p>
                      ) : (
                        <ul className="contact-opp-list">
                          {contact.linked_opportunities.map((link) => (
                            <li key={link.opportunity_id} className="contact-opp-item">
                              <Link
                                to={`/opportunities/${link.opportunity_id}`}
                                style={{ fontWeight: 600 }}
                              >
                                {link.role_title}
                              </Link>{' '}
                              <span className="company-pill">{link.company_name}</span>{' '}
                              <span className="stage-badge">{link.role}</span>{' '}
                              <span className="outcome-pill">{link.stage}</span>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>

                    {/* Zone 4: Compact Link Form */}
                    {opportunities.length > 0 && (
                      <div className="link-opp-compact-panel">
                        <span
                          style={{
                            fontSize: '0.75rem',
                            fontWeight: 600,
                            color: 'var(--text-secondary)',
                          }}
                        >
                          Link to Opportunity
                        </span>
                        <div className="form-grid" style={{ marginTop: 0, gap: '0.35rem' }}>
                          <label className="form-field" style={{ minWidth: '110px' }}>
                            <span style={{ fontSize: '0.72rem' }}>Opportunity</span>
                            <select
                              aria-label={`Link Opportunity for ${contact.name}`}
                              value={selectedOppId}
                              style={{ fontSize: '0.8rem', padding: '0.25rem 0.4rem' }}
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
                          <label className="form-field" style={{ minWidth: '90px' }}>
                            <span style={{ fontSize: '0.72rem' }}>Role</span>
                            <select
                              aria-label={`Stakeholder Role for ${contact.name}`}
                              value={selectedRole}
                              style={{ fontSize: '0.8rem', padding: '0.25rem 0.4rem' }}
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
                              style={{ width: '100%', fontSize: '0.78rem', padding: '0.3rem' }}
                              onClick={() => handleLinkContactToOpp(contact)}
                            >
                              Link to Opportunity
                            </button>
                          </div>
                        </div>
                      </div>
                    )}

                    {/* Zone 5: Footer Actions & Interaction Count */}
                    <div className="contact-card-footer">
                      <span className="interaction-count-badge">
                        💬 {contactInteractions.length}{' '}
                        {contactInteractions.length === 1 ? 'interaction' : 'interactions'}
                      </span>
                      <button
                        type="button"
                        className="btn-ghost"
                        style={{ fontSize: '0.78rem', padding: '0.2rem 0.5rem' }}
                        onClick={() => openLogInteractionForContact(contact.id)}
                      >
                        + Log Interaction
                      </button>
                    </div>
                  </article>
                );
              })}
            </div>
          )}
        </section>
      )}

      {/* TAB 2: Companies Directory */}
      {activeTab === 'companies' && (
        <section aria-label="Target Companies Directory">
          <div className="card-top-row">
            <h3>Companies ({companies.length})</h3>
          </div>

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
                      <div>
                        <h3>{c.name}</h3>
                        <div className="card-meta">
                          {c.domain && (
                            <span className="company-pill">
                              🌐 {c.domain}
                            </span>
                          )}
                          {c.remote_policy && (
                            <span className="stage-badge">{c.remote_policy}</span>
                          )}
                          {c.size_stage && (
                            <span className="outcome-pill">{c.size_stage}</span>
                          )}
                        </div>
                      </div>
                      <button
                        type="button"
                        className="btn-secondary"
                        onClick={() => setEditingCompany(c)}
                      >
                        Edit Company
                      </button>
                    </div>

                    {c.headquarters && (
                      <div className="card-meta">
                        <span>📍 {c.headquarters}</span>
                      </div>
                    )}
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
      )}

      {/* TAB 3: Activity & Tasks */}
      {activeTab === 'activity' && (
        <div className="detail-grid">
          {/* Recent Interactions Timeline */}
          <section className="panel-card" aria-label="Network Interaction History">
            <div className="card-top-row">
              <h3>Interaction History ({interactions.length})</h3>
              <button
                type="button"
                className="btn-primary"
                onClick={() => {
                  setLogInteractionContactId('');
                  setLogInteractionOppId('');
                  setShowLogInteractionModal(true);
                }}
              >
                + Log Interaction
              </button>
            </div>

            {interactions.length === 0 ? (
              <p className="empty-state">No Interactions recorded yet.</p>
            ) : (
              <ul className="timeline-list">
                {interactions.map((intr) => (
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
                    <div className="card-meta">
                      {intr.contact_name && (
                        <span className="company-pill">👤 {intr.contact_name}</span>
                      )}
                      {intr.opportunity_role_title && (
                        <span>💼 {intr.opportunity_role_title}</span>
                      )}
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>

          {/* Action Items List */}
          <section className="panel-card" aria-label="Network Action Items">
            <div className="card-top-row">
              <h3>Action Items ({actionItems.length})</h3>
              <button
                type="button"
                className="btn-primary"
                onClick={() => {
                  setEditingActionItem(null);
                  setShowActionItemModal(true);
                }}
              >
                + New Action Item
              </button>
            </div>

            {actionItems.length === 0 ? (
              <p className="empty-state">No Action Items added yet.</p>
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
                          setShowActionItemModal(true);
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
      )}

      {/* Company Form Modal */}
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

      {/* Contact Form Modal */}
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

      {/* Log Interaction Modal */}
      {showLogInteractionModal && (
        <LogInteractionModal
          contacts={contacts}
          opportunities={opportunities}
          defaultContactId={logInteractionContactId}
          defaultOpportunityId={logInteractionOppId}
          onClose={() => {
            setShowLogInteractionModal(false);
            setLogInteractionContactId('');
            setLogInteractionOppId('');
          }}
          onSubmit={async (input) => {
            await logInteractionMutation.mutateAsync(input);
          }}
        />
      )}

      {/* Action Item Modal */}
      {showActionItemModal && (
        <ActionItemModal
          initial={editingActionItem}
          contacts={contacts}
          opportunities={opportunities}
          onClose={() => {
            setShowActionItemModal(false);
            setEditingActionItem(null);
          }}
          onSubmit={async (input) => {
            await saveActionItemMutation.mutateAsync(input);
          }}
        />
      )}
    </div>
  );
}
