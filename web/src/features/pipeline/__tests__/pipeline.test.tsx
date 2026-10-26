import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CockpitApiProvider } from '../../../api/CockpitApiContext';
import type {
  Company,
  CompanyUpsertInput,
  JobPosting,
  JobPostingInput,
  Opportunity,
  OpportunityCreateInput,
  OpportunityTransitionInput,
  OpportunityUpdateInput,
  StageTransition,
} from '../../../api/types';
import { AppRoutes } from '../../../App';

function hasJobPostingContent(jp: JobPostingInput): boolean {
  return Boolean(
    jp.source_url.trim() ||
      jp.raw_text.trim() ||
      jp.requirements.length > 0 ||
      jp.tech_stack_tags.length > 0 ||
      jp.stated_min_comp > 0 ||
      jp.stated_max_comp > 0,
  );
}

function toJobPosting(jp: JobPostingInput): JobPosting {
  return {
    ...jp,
    has_content: hasJobPostingContent(jp),
  };
}

function installStubbedHttpHandlers(seed?: {
  companies?: Company[];
  opportunities?: Opportunity[];
  transitions?: Record<string, StageTransition[]>;
}) {
  const companies: Company[] = [...(seed?.companies ?? [])];
  const opportunities: Opportunity[] = [...(seed?.opportunities ?? [])];
  const transitionsByOpp: Record<string, StageTransition[]> = {
    ...(seed?.transitions ?? {}),
  };
  let idCounter = 100;
  let tick = 1;
  const nextIso = () => {
    const day = String(20 + Math.floor(tick / 10)).padStart(2, '0');
    const hour = String(10 + (tick % 10)).padStart(2, '0');
    tick += 1;
    return `2026-10-${day}T${hour}:00:00Z`;
  };

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

    if (url.pathname === '/api/companies' && method === 'GET') {
      return jsonResponse([...companies]);
    }

    if (url.pathname === '/api/companies' && method === 'POST') {
      const payload = JSON.parse(bodyText) as CompanyUpsertInput;
      const now = nextIso();
      const created: Company = {
        id: `comp-${++idCounter}`,
        ...payload,
        created_at: now,
        updated_at: now,
      };
      companies.push(created);
      return jsonResponse(created, 201);
    }

    const companyMatch = url.pathname.match(/^\/api\/companies\/([^/]+)$/);
    if (companyMatch && method === 'PUT') {
      const id = decodeURIComponent(companyMatch[1]);
      const idx = companies.findIndex((c) => c.id === id);
      if (idx === -1) {
        return jsonResponse({ error: { code: 'not_found', message: 'not found' } }, 404);
      }
      const payload = JSON.parse(bodyText) as CompanyUpsertInput;
      const updated: Company = {
        ...companies[idx],
        ...payload,
        updated_at: nextIso(),
      };
      companies[idx] = updated;
      for (const opp of opportunities) {
        if (opp.company_id === id) {
          opp.company_name = updated.name;
        }
      }
      return jsonResponse(updated);
    }

    if (url.pathname === '/api/opportunities' && method === 'GET') {
      const stage = url.searchParams.get('stage') ?? '';
      const outcome = url.searchParams.get('outcome') ?? '';
      const priority = url.searchParams.get('priority') ?? '';
      const remotePolicy = url.searchParams.get('remote_policy') ?? '';
      const companyId = url.searchParams.get('company_id') ?? '';
      const q = (url.searchParams.get('q') ?? '').trim().toLowerCase();

      const filtered = opportunities.filter((o) => {
        if (stage && o.stage !== stage) return false;
        if (outcome && o.outcome !== outcome) return false;
        if (priority && o.priority !== priority) return false;
        if (remotePolicy && o.remote_policy !== remotePolicy) return false;
        if (companyId && o.company_id !== companyId) return false;
        if (q) {
          const hay = [
            o.role_title,
            o.company_name,
            o.notes,
            ...o.job_posting.tech_stack_tags,
          ]
            .join(' ')
            .toLowerCase();
          if (!hay.includes(q)) return false;
        }
        return true;
      });
      return jsonResponse(filtered);
    }

    if (url.pathname === '/api/opportunities' && method === 'POST') {
      const payload = JSON.parse(bodyText) as OpportunityCreateInput;
      const company = companies.find((c) => c.id === payload.company_id);
      if (!company) {
        return jsonResponse({ error: { code: 'not_found', message: 'company not found' } }, 404);
      }
      const id = `opp-${++idCounter}`;
      const now = nextIso();
      const stage = payload.stage ?? 'Discovered';
      const initialTr: StageTransition = {
        id: `tr-${++idCounter}`,
        opportunity_id: id,
        from_stage: '',
        to_stage: stage,
        from_outcome: '',
        to_outcome: 'Active',
        note: payload.transition_note,
        transitioned_at: now,
      };
      transitionsByOpp[id] = [initialTr];
      const created: Opportunity = {
        id,
        company_id: payload.company_id,
        company_name: company.name,
        role_title: payload.role_title,
        seniority: payload.seniority,
        location: payload.location,
        remote_policy: payload.remote_policy,
        priority: payload.priority,
        stage,
        outcome: 'Active',
        notes: payload.notes,
        job_posting: toJobPosting(payload.job_posting),
        last_activity_at: now,
        created_at: now,
        updated_at: now,
      };
      opportunities.unshift(created);
      return jsonResponse(created, 201);
    }

    const oppDetailMatch = url.pathname.match(/^\/api\/opportunities\/([^/]+)$/);
    if (oppDetailMatch && method === 'GET') {
      const id = decodeURIComponent(oppDetailMatch[1]);
      const found = opportunities.find((o) => o.id === id);
      if (!found) {
        return jsonResponse({ error: { code: 'not_found', message: 'not found' } }, 404);
      }
      return jsonResponse({
        ...found,
        stage_transitions: transitionsByOpp[id] ?? [],
      });
    }

    if (oppDetailMatch && method === 'PUT') {
      const id = decodeURIComponent(oppDetailMatch[1]);
      const idx = opportunities.findIndex((o) => o.id === id);
      if (idx === -1) {
        return jsonResponse({ error: { code: 'not_found', message: 'not found' } }, 404);
      }
      const payload = JSON.parse(bodyText) as OpportunityUpdateInput;
      const company = companies.find((c) => c.id === payload.company_id);
      const updated: Opportunity = {
        ...opportunities[idx],
        company_id: payload.company_id,
        company_name: company?.name ?? opportunities[idx].company_name,
        role_title: payload.role_title,
        seniority: payload.seniority,
        location: payload.location,
        remote_policy: payload.remote_policy,
        priority: payload.priority,
        notes: payload.notes,
        job_posting: toJobPosting(payload.job_posting),
        // Preserve last_activity_at on detail edits, advance updated_at
        updated_at: nextIso(),
      };
      opportunities[idx] = updated;
      return jsonResponse(updated);
    }

    const oppTransitionMatch = url.pathname.match(/^\/api\/opportunities\/([^/]+)\/transition$/);
    if (oppTransitionMatch && method === 'POST') {
      const id = decodeURIComponent(oppTransitionMatch[1]);
      const idx = opportunities.findIndex((o) => o.id === id);
      if (idx === -1) {
        return jsonResponse({ error: { code: 'not_found', message: 'not found' } }, 404);
      }
      const payload = JSON.parse(bodyText) as OpportunityTransitionInput;
      const current = opportunities[idx];
      const nextStage = payload.stage ?? current.stage;
      const nextOutcome = payload.outcome ?? current.outcome;
      const now = nextIso();
      const tr: StageTransition = {
        id: `tr-${++idCounter}`,
        opportunity_id: id,
        from_stage: current.stage,
        to_stage: nextStage,
        from_outcome: current.outcome,
        to_outcome: nextOutcome,
        note: payload.note,
        transitioned_at: now,
      };
      transitionsByOpp[id] = [...(transitionsByOpp[id] ?? []), tr];
      const updated: Opportunity = {
        ...current,
        stage: nextStage,
        outcome: nextOutcome,
        last_activity_at: now,
        updated_at: now,
      };
      opportunities[idx] = updated;
      return jsonResponse(updated);
    }

    return jsonResponse({ error: { code: 'not_found', message: `unhandled ${method} ${url.pathname}` } }, 404);
  });

  vi.stubGlobal('fetch', fetchMock);
  return { fetchMock, companies, opportunities };
}

function renderCockpit(initialEntries = ['/opportunities']) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <CockpitApiProvider>
        <MemoryRouter initialEntries={initialEntries}>
          <AppRoutes />
        </MemoryRouter>
      </CockpitApiProvider>
    </QueryClientProvider>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('Pipeline Spine Workspace Views (over /api/* HTTP handlers)', () => {
  it('renders persistent navigation shell, Active-only Kanban columns, moves stage, and filters Outcomes in table view', async () => {
    const user = userEvent.setup();
    installStubbedHttpHandlers({
      companies: [
        {
          id: 'comp-1',
          name: 'Stripe',
          domain: 'stripe.com',
          headquarters: 'San Francisco, CA',
          remote_policy: 'Hybrid',
          size_stage: 'Late Stage',
          notes: '',
          created_at: '2026-10-20T09:00:00Z',
          updated_at: '2026-10-20T09:00:00Z',
        },
      ],
      opportunities: [
        {
          id: 'opp-active',
          company_id: 'comp-1',
          company_name: 'Stripe',
          role_title: 'Staff Core Payments Engineer',
          seniority: 'Staff',
          location: 'Remote',
          remote_policy: 'Remote',
          priority: 'High',
          stage: 'Discovered',
          outcome: 'Active',
          notes: 'Exploratory referral',
          job_posting: {
            source_url: '',
            raw_text: '',
            requirements: [],
            tech_stack_tags: ['Go', 'Kafka'],
            stated_min_comp: 0,
            stated_max_comp: 0,
            currency: '',
            has_content: true,
          },
          last_activity_at: '2026-10-20T09:00:00Z',
          created_at: '2026-10-20T09:00:00Z',
          updated_at: '2026-10-20T09:00:00Z',
        },
        {
          id: 'opp-rejected',
          company_id: 'comp-1',
          company_name: 'Stripe',
          role_title: 'Staff Billing Engineer',
          seniority: 'Staff',
          location: 'Seattle, WA',
          remote_policy: 'Hybrid',
          priority: 'Medium',
          stage: 'Interviewing',
          outcome: 'Rejected',
          notes: 'Closed after loop',
          job_posting: {
            source_url: '',
            raw_text: '',
            requirements: [],
            tech_stack_tags: [],
            stated_min_comp: 0,
            stated_max_comp: 0,
            currency: '',
            has_content: false,
          },
          last_activity_at: '2026-10-21T09:00:00Z',
          created_at: '2026-10-18T09:00:00Z',
          updated_at: '2026-10-21T09:00:00Z',
        },
      ],
    });

    renderCockpit(['/opportunities']);

    expect(await screen.findByText('Staff Core Payments Engineer')).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: /Workspace Navigation/i })).toBeInTheDocument();
    expect(screen.queryByText('Staff Billing Engineer')).not.toBeInTheDocument();

    const moveSelect = screen.getByLabelText('Move Staff Core Payments Engineer Stage');
    await user.selectOptions(moveSelect, 'Outreach');

    const outreachCol = screen.getByRole('region', { name: 'Outreach Stage Column' });
    expect(await within(outreachCol).findByText('Staff Core Payments Engineer')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Table View' }));
    expect(await screen.findByText('Staff Billing Engineer')).toBeInTheDocument();

    await user.selectOptions(screen.getByLabelText('Filter by Outcome'), 'Rejected');
    expect(await screen.findByText('Staff Billing Engineer')).toBeInTheDocument();
    expect(screen.queryByText('Staff Core Payments Engineer')).not.toBeInTheDocument();
  });

  it('creates and edits a Company in /network and creates multiple Opportunities (unposted and posted) under that Company', async () => {
    const user = userEvent.setup();
    installStubbedHttpHandlers();

    renderCockpit(['/network']);

    expect(await screen.findByText('No Companies added yet.')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: '+ New Company' }));

    const createDialog = screen.getByRole('dialog', { name: 'New Company' });
    await user.type(within(createDialog).getByLabelText(/Company Name/i), 'Datadog');
    await user.type(within(createDialog).getByLabelText(/Website \/ Domain/i), 'datadoghq.com');
    await user.click(within(createDialog).getByRole('button', { name: 'Create Company' }));

    expect(await screen.findByRole('heading', { name: 'Datadog' })).toBeInTheDocument();

    // Edit the Company
    await user.click(screen.getByRole('button', { name: 'Edit Company' }));
    const editDialog = screen.getByRole('dialog', { name: 'Edit Company' });
    const nameInput = within(editDialog).getByLabelText(/Company Name/i);
    await user.clear(nameInput);
    await user.type(nameInput, 'Datadog Inc');
    await user.click(within(editDialog).getByRole('button', { name: 'Save Changes' }));

    expect(await screen.findByRole('heading', { name: 'Datadog Inc' })).toBeInTheDocument();

    // Navigate to Pipeline and create two distinct Opportunities under Datadog Inc
    await user.click(screen.getByRole('link', { name: 'Pipeline' }));
    await user.click(await screen.findByRole('button', { name: '+ New Opportunity' }));

    const oppModal1 = screen.getByRole('dialog', { name: 'New Opportunity' });
    await user.type(
      within(oppModal1).getByLabelText(/Role Title/i),
      'Staff Metrics Ingestion Engineer',
    );
    await user.type(
      within(oppModal1).getByLabelText(/Initial Stage Transition Note/i),
      'Referred by principal engineer',
    );
    await user.click(within(oppModal1).getByRole('button', { name: 'Create Opportunity' }));

    const discoveredCol = await screen.findByRole('region', { name: 'Discovered Stage Column' });
    expect(await within(discoveredCol).findByText('Staff Metrics Ingestion Engineer')).toBeInTheDocument();
    expect(within(discoveredCol).getByText('Unposted pursuit')).toBeInTheDocument();

    // Create second distinct Opportunity with a formal Job Posting attached
    await user.click(screen.getByRole('button', { name: '+ New Opportunity' }));
    const oppModal2 = screen.getByRole('dialog', { name: 'New Opportunity' });
    await user.type(
      within(oppModal2).getByLabelText(/Role Title/i),
      'Principal Distributed Tracing Engineer',
    );
    await user.selectOptions(within(oppModal2).getByLabelText('Initial Stage'), 'Applied');
    await user.click(
      within(oppModal2).getByLabelText(/Attach formal Job Posting details now/i),
    );
    await user.type(
      within(oppModal2).getByLabelText(/Job Posting URL/i),
      'https://careers.datadoghq.com/detail/123',
    );
    await user.type(
      within(oppModal2).getByLabelText(/Tech Stack Tags/i),
      'Go, Rust, eBPF',
    );
    await user.click(within(oppModal2).getByRole('button', { name: 'Create Opportunity' }));

    const appliedCol = screen.getByRole('region', { name: 'Applied Stage Column' });
    expect(await within(appliedCol).findByText('Principal Distributed Tracing Engineer')).toBeInTheDocument();
    expect(within(appliedCol).getByText('Job Posting attached')).toBeInTheDocument();
  });

  it('preserves LastActivityAt on detail/JobPosting edits while advancing UpdatedAt, and records orthogonal transitions with history', async () => {
    const user = userEvent.setup();
    installStubbedHttpHandlers({
      companies: [
        {
          id: 'comp-1',
          name: 'Anthropic',
          domain: 'anthropic.com',
          headquarters: 'San Francisco, CA',
          remote_policy: 'Hybrid',
          size_stage: 'Growth',
          notes: '',
          created_at: '2026-10-20T09:00:00Z',
          updated_at: '2026-10-20T09:00:00Z',
        },
      ],
      opportunities: [
        {
          id: 'opp-1',
          company_id: 'comp-1',
          company_name: 'Anthropic',
          role_title: 'Staff Inference Systems Engineer',
          seniority: 'Staff',
          location: 'San Francisco, CA',
          remote_policy: 'Hybrid',
          priority: 'High',
          stage: 'Interviewing',
          outcome: 'Active',
          notes: 'Initial exploratory pursuit',
          job_posting: {
            source_url: '',
            raw_text: '',
            requirements: [],
            tech_stack_tags: [],
            stated_min_comp: 0,
            stated_max_comp: 0,
            currency: 'USD',
            has_content: false,
          },
          last_activity_at: '2026-10-20T09:00:00Z',
          created_at: '2026-10-20T09:00:00Z',
          updated_at: '2026-10-20T09:00:00Z',
        },
      ],
      transitions: {
        'opp-1': [
          {
            id: 'tr-1',
            opportunity_id: 'opp-1',
            from_stage: '',
            to_stage: 'Discovered',
            from_outcome: '',
            to_outcome: 'Active',
            note: 'Initial discovery',
            transitioned_at: '2026-10-20T09:00:00Z',
          },
          {
            id: 'tr-2',
            opportunity_id: 'opp-1',
            from_stage: 'Discovered',
            to_stage: 'Interviewing',
            from_outcome: 'Active',
            to_outcome: 'Active',
            note: 'Scheduled technical loop',
            transitioned_at: '2026-10-20T09:00:00Z',
          },
        ],
      },
    });

    renderCockpit(['/opportunities/opp-1']);

    expect(
      await screen.findByRole('heading', { name: 'Staff Inference Systems Engineer' }),
    ).toBeInTheDocument();

    const initialLastActivityText = screen.getByTestId('detail-last-activity').textContent;
    const initialUpdatedText = screen.getByTestId('detail-updated-at').textContent;

    // Attach JobPosting and edit notes without changing Stage/Outcome
    const detailsSection = screen.getByRole('region', {
      name: 'Opportunity and Job Posting Details',
    });
    await user.type(
      within(detailsSection).getByLabelText(/Job Posting URL/i),
      'https://anthropic.com/jobs/inference-1',
    );
    await user.click(
      within(detailsSection).getByRole('button', { name: /Save Details & Job Posting/i }),
    );

    expect(
      await screen.findByText('Opportunity & Job Posting details saved.'),
    ).toBeInTheDocument();
    expect(screen.getByTestId('detail-last-activity').textContent).toBe(initialLastActivityText);
    expect(screen.getByTestId('detail-updated-at').textContent).not.toBe(initialUpdatedText);

    // Transition Outcome to Rejected while preserving Interviewing Stage
    await user.selectOptions(screen.getByLabelText('Target Outcome'), 'Rejected');
    await user.type(
      screen.getByLabelText('Transition Note'),
      'Position filled after final systems round',
    );
    await user.click(screen.getByRole('button', { name: 'Record Transition' }));

    expect(
      await screen.findByText('Stage / Outcome transition recorded.'),
    ).toBeInTheDocument();
    expect(screen.getByText('Stage: Interviewing')).toBeInTheDocument();
    expect(screen.getByText('Outcome: Rejected')).toBeInTheDocument();
    expect(
      screen.getByText('Position filled after final systems round'),
    ).toBeInTheDocument();
    expect(screen.getByTestId('detail-last-activity').textContent).not.toBe(
      initialLastActivityText,
    );
  });
});
