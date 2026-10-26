import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CockpitApiProvider } from '../../../api/CockpitApiContext';
import type {
  Contact,
  FlightDeckData,
  Interaction,
  Opportunity,
  StageTransition,
} from '../../../api/types';
import { AppRoutes } from '../../../App';

function installFlightDeckHttpHandlers(seed?: {
  flightDeck?: FlightDeckData;
  contacts?: Contact[];
}) {
  let flightDeck: FlightDeckData = seed?.flightDeck ?? {
    upcoming_rounds: [],
    missing_debriefs: [],
    action_items: [],
    stale_opportunities: [],
  };

  const contacts: Contact[] = seed?.contacts ?? [
    {
      id: 'contact-1',
      company_id: 'comp-1',
      company_name: 'Stripe',
      name: 'Alice Chen',
      title: 'Engineering Director',
      warmth: 'Advocate',
      email: 'alice@stripe.com',
      linkedin_url: '',
      notes: '',
      linked_opportunities: [],
      created_at: '2026-10-20T10:00:00Z',
      updated_at: '2026-10-20T10:00:00Z',
    },
  ];

  const interactions: Interaction[] = [];
  const transitions: StageTransition[] = [];

  const jsonResponse = (body: unknown, status = 200) =>
    Promise.resolve(
      new Response(JSON.stringify(body), {
        status,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const rawUrl = typeof input === 'string' ? input : input.toString();
    const url = new URL(rawUrl, 'http://localhost');
    const method = (init?.method ?? 'GET').toUpperCase();
    const bodyText = typeof init?.body === 'string' ? init.body : '';

    if (url.pathname === '/api/flight-deck' && method === 'GET') {
      return jsonResponse(flightDeck);
    }

    if (url.pathname === '/api/contacts' && method === 'GET') {
      return jsonResponse(contacts);
    }

    // Action item update (complete / snooze)
    const actionMatch = url.pathname.match(/^\/api\/action-items\/([^/]+)$/);
    if (actionMatch && method === 'PUT') {
      const id = decodeURIComponent(actionMatch[1]);
      const payload = JSON.parse(bodyText);
      flightDeck = {
        ...flightDeck,
        action_items: flightDeck.action_items
          .map((item) => (item.id === id ? { ...item, ...payload } : item))
          .filter((item) => item.status === 'open'),
      };
      return jsonResponse({ id, ...payload });
    }

    // Log Interaction
    if (url.pathname === '/api/interactions' && method === 'POST') {
      const payload = JSON.parse(bodyText);
      const created: Interaction = {
        id: `int-${Date.now()}`,
        contact_id: payload.contact_id,
        contact_name: 'Alice Chen',
        opportunity_id: payload.opportunity_id,
        opportunity_role_title: 'Lead Platform Engineer',
        company_name: 'Acme Corp',
        kind: payload.kind,
        direction: payload.direction,
        occurred_at: payload.occurred_at || new Date().toISOString(),
        summary: payload.summary,
        created_at: new Date().toISOString(),
      };
      interactions.push(created);
      // Logging an interaction resolves staleness on the opportunity
      flightDeck = {
        ...flightDeck,
        stale_opportunities: flightDeck.stale_opportunities.filter(
          (s) => s.opportunity_id !== payload.opportunity_id,
        ),
      };
      return jsonResponse({ interaction: created }, 201);
    }

    // Transition Opportunity (Stage or Outcome)
    const transMatch = url.pathname.match(/^\/api\/opportunities\/([^/]+)\/transition$/);
    if (transMatch && method === 'POST') {
      const oppId = decodeURIComponent(transMatch[1]);
      const payload = JSON.parse(bodyText);
      transitions.push({
        id: `tr-${Date.now()}`,
        opportunity_id: oppId,
        from_stage: 'Discovered',
        to_stage: payload.stage || 'Applied',
        from_outcome: 'Active',
        to_outcome: payload.outcome || 'Active',
        note: payload.note || '',
        transitioned_at: new Date().toISOString(),
      });
      // Advancing stage or changing outcome resolves staleness in the flight deck
      flightDeck = {
        ...flightDeck,
        stale_opportunities: flightDeck.stale_opportunities.filter(
          (s) => s.opportunity_id !== oppId,
        ),
      };
      return jsonResponse({ id: oppId, ...payload });
    }

    return jsonResponse({ error: { code: 'not_found', message: 'route not found' } }, 404);
  });

  vi.stubGlobal('fetch', fetchMock);

  return {
    fetchMock,
    getInteractions: () => interactions,
    getTransitions: () => transitions,
  };
}

describe('Flight Deck Daily Command Center Workspace Views', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  function renderFlightDeck(initialUrl = '/') {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false, gcTime: 0 },
      },
    });

    return render(
      <QueryClientProvider client={queryClient}>
        <CockpitApiProvider>
          <MemoryRouter initialEntries={[initialUrl]}>
            <AppRoutes />
          </MemoryRouter>
        </CockpitApiProvider>
      </QueryClientProvider>,
    );
  }

  it('pins imminent interview rounds within 48 hours at the top with details and prep link', async () => {
    installFlightDeckHttpHandlers({
      flightDeck: {
        upcoming_rounds: [
          {
            id: 'round-1',
            opportunity_id: 'opp-stripe',
            company_name: 'Stripe',
            opportunity_role_title: 'Staff Core Payments Engineer',
            round_order: 1,
            round_type: 'Technical',
            title: 'Payment Engine System Design',
            scheduled_at: '2026-10-27T14:00:00Z',
            duration_minutes: 60,
            status: 'Scheduled',
            prep_notes: 'Focus on distributed idempotency',
            contact_ids: ['contact-1'],
            story_ids: ['story-1'],
            interviewer_names: ['Alice Chen'],
            story_titles: ['Scaling Payment Pipeline to 10k RPS'],
          },
        ],
        missing_debriefs: [],
        action_items: [],
        stale_opportunities: [],
      },
    });

    renderFlightDeck('/');

    expect(await screen.findByRole('heading', { level: 2, name: /flight deck/i })).toBeInTheDocument();

    const upcomingSection = screen.getByTestId('imminent-rounds-section');
    expect(within(upcomingSection).getByText(/Payment Engine System Design/i)).toBeInTheDocument();
    expect(within(upcomingSection).getByText('Stripe')).toBeInTheDocument();
    expect(within(upcomingSection).getByText('Staff Core Payments Engineer')).toBeInTheDocument();
    expect(within(upcomingSection).getByText('Alice Chen')).toBeInTheDocument();
    expect(within(upcomingSection).getByText('Scaling Payment Pipeline to 10k RPS')).toBeInTheDocument();

    const prepLink = within(upcomingSection).getByRole('link', { name: /prepare for payment engine system design/i });
    expect(prepLink).toHaveAttribute('href', '/opportunities/opp-stripe?tab=interviews');
  });

  it('surfaces completed interview rounds missing debrief with one-click debrief prompt', async () => {
    installFlightDeckHttpHandlers({
      flightDeck: {
        upcoming_rounds: [],
        missing_debriefs: [
          {
            id: 'round-past-1',
            opportunity_id: 'opp-datadog',
            company_name: 'Datadog',
            opportunity_role_title: 'Principal Ingestion Engineer',
            round_order: 2,
            round_type: 'Technical',
            title: 'High-Throughput Ingestion Coding',
            scheduled_at: '2026-10-26T10:00:00Z',
            duration_minutes: 60,
            status: 'Completed',
          },
        ],
        action_items: [],
        stale_opportunities: [],
      },
    });

    renderFlightDeck('/');

    const debriefSection = await screen.findByTestId('missing-debriefs-section');
    expect(within(debriefSection).getByText(/High-Throughput Ingestion Coding/i)).toBeInTheDocument();
    expect(within(debriefSection).getByText('Datadog')).toBeInTheDocument();
    expect(within(debriefSection).getByText(/Debrief Needed/i)).toBeInTheDocument();

    const debriefLink = within(debriefSection).getByRole('link', { name: /log debrief for high-throughput ingestion coding/i });
    expect(debriefLink).toHaveAttribute('href', '/opportunities/opp-datadog?tab=interviews');
  });

  it('displays open action items highlighting overdue and due-today, allowing complete and snooze inline', async () => {
    const user = userEvent.setup();
    installFlightDeckHttpHandlers({
      flightDeck: {
        upcoming_rounds: [],
        missing_debriefs: [],
        action_items: [
          {
            id: 'action-overdue',
            title: 'Send follow-up email to recruiter',
            opportunity_id: 'opp-stripe',
            opportunity_role_title: 'Staff Payments Engineer',
            company_name: 'Stripe',
            contact_id: 'contact-1',
            contact_name: 'Alice Chen',
            due_at: '2026-10-25T12:00:00Z',
            status: 'open',
            created_at: '2026-10-24T10:00:00Z',
            updated_at: '2026-10-24T10:00:00Z',
            is_overdue: true,
            is_due_today: false,
          },
          {
            id: 'action-due-today',
            title: 'Review system design notes',
            opportunity_id: 'opp-stripe',
            opportunity_role_title: 'Staff Payments Engineer',
            company_name: 'Stripe',
            due_at: '2026-10-26T18:00:00Z',
            status: 'open',
            created_at: '2026-10-25T10:00:00Z',
            updated_at: '2026-10-25T10:00:00Z',
            is_overdue: false,
            is_due_today: true,
          },
        ],
        stale_opportunities: [],
      },
    });

    renderFlightDeck('/');

    const actionSection = await screen.findByTestId('action-items-section');
    expect(within(actionSection).getByText('Send follow-up email to recruiter')).toBeInTheDocument();
    expect(within(actionSection).getByText('Overdue')).toBeInTheDocument();
    expect(within(actionSection).getByText('Review system design notes')).toBeInTheDocument();
    expect(within(actionSection).getByText('Due Today')).toBeInTheDocument();

    // Complete the overdue item
    const completeBtn = within(actionSection).getByRole('button', {
      name: /complete send follow-up email to recruiter/i,
    });
    await user.click(completeBtn);

    await waitFor(() => {
      expect(screen.queryByText('Send follow-up email to recruiter')).not.toBeInTheDocument();
    });

    // Snooze the due-today item
    const snoozeBtn = within(actionSection).getByRole('button', {
      name: /snooze review system design notes 1 day/i,
    });
    await user.click(snoozeBtn);
  });

  it('surfaces stale opportunities and resolves them inline via log interaction, advance stage, and update outcome', async () => {
    const user = userEvent.setup();
    const { getInteractions, getTransitions } = installFlightDeckHttpHandlers({
      flightDeck: {
        upcoming_rounds: [],
        missing_debriefs: [],
        action_items: [],
        stale_opportunities: [
          {
            opportunity_id: 'opp-stale-disc',
            company_id: 'comp-1',
            company_name: 'Acme Corp',
            role_title: 'Lead Platform Engineer',
            stage: 'Discovered',
            outcome: 'Active',
            priority: 'High',
            stale_reason: 'In Discovered for 5 days with no activity',
            days_stale: 5,
            anchor_date: '2026-10-21T10:00:00Z',
            last_activity_at: '2026-10-21T10:00:00Z',
          },
          {
            opportunity_id: 'opp-stale-outreach',
            company_id: 'comp-2',
            company_name: 'Beta Co',
            role_title: 'Senior Backend Engineer',
            stage: 'Outreach',
            outcome: 'Active',
            priority: 'Medium',
            stale_reason: 'In Outreach for 6 days with no interaction',
            days_stale: 6,
            anchor_date: '2026-10-20T10:00:00Z',
            last_activity_at: '2026-10-20T10:00:00Z',
          },
        ],
      },
    });

    renderFlightDeck('/');

    const staleSection = await screen.findByTestId('stale-opportunities-section');
    expect(within(staleSection).getByText('Lead Platform Engineer')).toBeInTheDocument();
    expect(within(staleSection).getByText(/In Discovered for 5 days with no activity/i)).toBeInTheDocument();
    expect(within(staleSection).getByText('Senior Backend Engineer')).toBeInTheDocument();
    expect(within(staleSection).getByText(/In Outreach for 6 days with no interaction/i)).toBeInTheDocument();

    // Inline Action 1: Log Interaction on Discovered opp
    const logIntBtn = within(staleSection).getByRole('button', {
      name: /log interaction for lead platform engineer/i,
    });
    await user.click(logIntBtn);

    // Modal appears
    const summaryInput = screen.getByLabelText(/interaction summary/i);
    await user.type(summaryInput, 'Sent note to recruiter Alice');
    const submitIntBtn = screen.getByRole('button', { name: /^log interaction$/i });
    await user.click(submitIntBtn);

    await waitFor(() => {
      expect(screen.queryByText('Lead Platform Engineer')).not.toBeInTheDocument();
    });
    expect(getInteractions().length).toBe(1);
    expect(getInteractions()[0].summary).toBe('Sent note to recruiter Alice');

    // Inline Action 2: Advance Stage on Outreach opp
    const advanceStageBtn = screen.getByRole('button', {
      name: /advance stage for senior backend engineer/i,
    });
    await user.click(advanceStageBtn);

    const submitAdvanceBtn = screen.getByRole('button', { name: /^advance stage$/i });
    await user.click(submitAdvanceBtn);

    await waitFor(() => {
      expect(screen.queryByText('Senior Backend Engineer')).not.toBeInTheDocument();
    });
    expect(getTransitions().length).toBe(1);
  });
});
