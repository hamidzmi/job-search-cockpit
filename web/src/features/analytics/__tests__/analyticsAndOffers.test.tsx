import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CockpitApiProvider } from '../../../api/CockpitApiContext';
import type {
  FunnelAnalytics,
  OfferComparison,
  OfferView,
  Opportunity,
} from '../../../api/types';
import { AppRoutes } from '../../../App';

describe('Analytics, Offers & Backup/Seeding Workspace Views', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders side-by-side offer comparison and funnel analytics on /analytics', async () => {
    const user = userEvent.setup();
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    const mockComparison: OfferComparison = {
      compensation_floor: 240000,
      target_compensation: 290000,
      profile_currency: 'USD',
      offers: [
        {
          offer: {
            id: 'offer-1',
            opportunity_id: 'opp-1',
            base_salary: 200000,
            bonus_target_pct: 15,
            bonus_amount: 0,
            equity_type: 'RSU',
            equity_total_value: 160000,
            equity_vesting_years: 4,
            sign_on_bonus: 25000,
            currency: 'USD',
            deadline_at: '2026-11-15T00:00:00Z',
            benefits_notes: '100% health, 401k match',
            negotiation_notes: 'Countered +$20k equity',
            status: 'Negotiating',
            annual_bonus: 30000,
            annual_equity: 40000,
            recurring_total_comp: 270000,
            year1_total_comp: 295000,
            delta_vs_comp_floor: 30000,
            created_at: '2026-10-26T10:00:00Z',
            updated_at: '2026-10-26T10:00:00Z',
          },
          company_id: 'comp-1',
          company_name: 'Starlight AI',
          opportunity_role_title: 'Staff Infrastructure Engineer',
          opportunity_stage: 'Offer',
          opportunity_outcome: 'Active',
          compensation_floor: 240000,
        },
      ],
    };

    const mockFunnel: FunnelAnalytics = {
      total_opportunities: 4,
      active_opportunities: 1,
      stages: [
        {
          stage: 'Discovered',
          reached_count: 4,
          conversion_from_previous_pct: 100,
          drop_off_count: 0,
          outcome_breakdown: {},
          average_days_in_stage: 3.0,
        },
        {
          stage: 'Outreach',
          reached_count: 4,
          conversion_from_previous_pct: 100,
          drop_off_count: 1,
          outcome_breakdown: { Withdrawn: 1 },
          average_days_in_stage: 3.5,
        },
        {
          stage: 'Applied',
          reached_count: 3,
          conversion_from_previous_pct: 75,
          drop_off_count: 1,
          outcome_breakdown: { Ghosted: 1 },
          average_days_in_stage: 8.0,
        },
        {
          stage: 'Screening',
          reached_count: 2,
          conversion_from_previous_pct: 66.7,
          drop_off_count: 0,
          outcome_breakdown: {},
          average_days_in_stage: 5.0,
        },
        {
          stage: 'Interviewing',
          reached_count: 2,
          conversion_from_previous_pct: 100,
          drop_off_count: 1,
          outcome_breakdown: { Rejected: 1 },
          average_days_in_stage: 6.0,
        },
        {
          stage: 'Offer',
          reached_count: 1,
          conversion_from_previous_pct: 50,
          drop_off_count: 0,
          outcome_breakdown: { Active: 1 },
          average_days_in_stage: 6.0,
        },
      ],
    };

    const mockApi: any = {
      offers: {
        getOfferComparison: vi.fn().mockResolvedValue(mockComparison),
        getOffer: vi.fn(),
        upsertOffer: vi.fn(),
        deleteOffer: vi.fn(),
      },
      analytics: {
        getFunnelAnalytics: vi.fn().mockResolvedValue(mockFunnel),
      },
      backupAndSeed: {
        seedSampleData: vi.fn(),
        getBackupExportUrl: vi.fn((format: string) => `/api/backup/export?format=${format}`),
      },
    };

    render(
      <QueryClientProvider client={queryClient}>
        <CockpitApiProvider api={mockApi}>
          <MemoryRouter initialEntries={['/analytics']}>
            <AppRoutes />
          </MemoryRouter>
        </CockpitApiProvider>
      </QueryClientProvider>,
    );

    // Offers comparison table is visible
    expect(await screen.findByText('Side-by-Side Offer Comparison')).toBeInTheDocument();
    expect(screen.getByText('Starlight AI')).toBeInTheDocument();
    expect(screen.getByText('Staff Infrastructure Engineer')).toBeInTheDocument();
    expect(screen.getByText(/270,000/)).toBeInTheDocument();
    expect(screen.getByText(/295,000/)).toBeInTheDocument();
    expect(screen.getByText(/\+USD 30,000/)).toBeInTheDocument();

    // Switch to Funnel Analytics tab
    const funnelTab = screen.getByRole('tab', { name: /Funnel Analytics/i });
    await user.click(funnelTab);

    // Funnel analytics metrics are visible
    expect(await screen.findByText('Stage-by-Stage Pipeline Funnel')).toBeInTheDocument();
    expect(screen.getByText('Total Opportunities')).toBeInTheDocument();
    expect(screen.getAllByText('4').length).toBeGreaterThan(0);
    expect(screen.getByText('Discovered')).toBeInTheDocument();
    expect(screen.getByText('66.7%')).toBeInTheDocument();
    expect(screen.getByText('Rejected: 1')).toBeInTheDocument();
  });

  it('manages structured offer on OpportunityDetailWorkspace', async () => {
    const user = userEvent.setup();
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    const mockOpp: Opportunity = {
      id: 'opp-1',
      company_id: 'comp-1',
      company_name: 'Acme Cloud',
      role_title: 'Staff Architect',
      seniority: 'Staff',
      location: 'Remote',
      remote_policy: 'Remote',
      priority: 'High',
      stage: 'Offer',
      outcome: 'Active',
      notes: '',
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
      last_activity_at: '2026-10-26T09:00:00Z',
      created_at: '2026-10-26T09:00:00Z',
      updated_at: '2026-10-26T09:00:00Z',
    };

    let savedOfferView: OfferView | null = null;

    const mockApi: any = {
      opportunities: {
        getOpportunity: vi.fn().mockResolvedValue(mockOpp),
        updateOpportunity: vi.fn(),
        transitionOpportunity: vi.fn(),
      },
      offers: {
        getOffer: vi.fn().mockImplementation(() => {
          if (!savedOfferView) {
            const err: any = new Error('Offer not found');
            err.status = 404;
            return Promise.reject(err);
          }
          return Promise.resolve(savedOfferView);
        }),
        upsertOffer: vi.fn().mockImplementation((_oppId, inOffer) => {
          savedOfferView = {
            offer: {
              id: 'offer-123',
              opportunity_id: 'opp-1',
              base_salary: inOffer.base_salary,
              bonus_target_pct: inOffer.bonus_target_pct,
              bonus_amount: inOffer.bonus_amount,
              equity_type: inOffer.equity_type,
              equity_total_value: inOffer.equity_total_value,
              equity_vesting_years: inOffer.equity_vesting_years,
              sign_on_bonus: inOffer.sign_on_bonus,
              currency: inOffer.currency,
              benefits_notes: inOffer.benefits_notes,
              negotiation_notes: inOffer.negotiation_notes,
              status: inOffer.status,
              annual_bonus: (inOffer.base_salary * inOffer.bonus_target_pct) / 100,
              annual_equity: inOffer.equity_total_value / inOffer.equity_vesting_years,
              recurring_total_comp:
                inOffer.base_salary +
                (inOffer.base_salary * inOffer.bonus_target_pct) / 100 +
                inOffer.equity_total_value / inOffer.equity_vesting_years,
              year1_total_comp:
                inOffer.base_salary +
                (inOffer.base_salary * inOffer.bonus_target_pct) / 100 +
                inOffer.equity_total_value / inOffer.equity_vesting_years +
                inOffer.sign_on_bonus,
              delta_vs_comp_floor:
                inOffer.base_salary +
                (inOffer.base_salary * inOffer.bonus_target_pct) / 100 +
                inOffer.equity_total_value / inOffer.equity_vesting_years -
                220000,
              created_at: '2026-10-26T10:00:00Z',
              updated_at: '2026-10-26T10:00:00Z',
            },
            company_id: 'comp-1',
            company_name: 'Acme Cloud',
            opportunity_role_title: 'Staff Architect',
            opportunity_stage: 'Offer',
            opportunity_outcome: 'Active',
            compensation_floor: 220000,
          };
          return Promise.resolve(savedOfferView);
        }),
        deleteOffer: vi.fn().mockImplementation(() => {
          savedOfferView = null;
          return Promise.resolve();
        }),
      },
      tailoring: {
        getFitAssessment: vi.fn().mockRejectedValue({ status: 404 }),
        listArtifacts: vi.fn().mockResolvedValue([]),
      },
      interviewRounds: {
        listInterviewRounds: vi.fn().mockResolvedValue([]),
      },
      contacts: {
        listOpportunityContacts: vi.fn().mockResolvedValue([]),
      },
      interactions: {
        listInteractions: vi.fn().mockResolvedValue([]),
      },
      actionItems: {
        listActionItems: vi.fn().mockResolvedValue([]),
      },
    };

    render(
      <QueryClientProvider client={queryClient}>
        <CockpitApiProvider api={mockApi}>
          <MemoryRouter initialEntries={['/opportunities/opp-1?tab=offer']}>
            <AppRoutes />
          </MemoryRouter>
        </CockpitApiProvider>
      </QueryClientProvider>,
    );

    // Initial empty state
    expect(await screen.findByText('Structured Offer & Compensation')).toBeInTheDocument();
    expect(screen.getByText('No offer recorded for this opportunity yet.')).toBeInTheDocument();

    // Click Record Offer
    const recordBtn = screen.getByRole('button', { name: 'Record Offer' });
    await user.click(recordBtn);

    // Fill form
    const baseInput = screen.getByLabelText('Base Salary');
    await user.type(baseInput, '200000');

    const bonusPctInput = screen.getByLabelText('Bonus Target Percentage');
    await user.type(bonusPctInput, '15');

    const equityValInput = screen.getByLabelText('Equity Total Value');
    await user.type(equityValInput, '160000');

    const saveBtn = screen.getByRole('button', { name: 'Save Offer' });
    await user.click(saveBtn);

    // Expect offer summary values
    expect(await screen.findByText('Offer saved successfully.')).toBeInTheDocument();
    expect(screen.getAllByText(/270,000/).length).toBeGreaterThan(0);
  });

  it('triggers backup download and sample data seeding in SettingsWorkspace', async () => {
    const user = userEvent.setup();
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    const seedFn = vi.fn().mockResolvedValue({ success: true });

    const mockApi: any = {
      aiSettings: {
        getSettings: vi.fn().mockResolvedValue({
          provider: 'openrouter',
          base_url: '',
          api_key: '',
          fast_extraction_model: '',
          reasoning_tailoring_model: '',
          bookmarked_models: [],
          is_configured: false,
        }),
        listOpenRouterModels: vi.fn().mockResolvedValue([]),
      },
      backupAndSeed: {
        seedSampleData: seedFn,
        getBackupExportUrl: vi.fn((format: string) => `/api/backup/export?format=${format}`),
      },
    };

    render(
      <QueryClientProvider client={queryClient}>
        <CockpitApiProvider api={mockApi}>
          <MemoryRouter initialEntries={['/settings']}>
            <AppRoutes />
          </MemoryRouter>
        </CockpitApiProvider>
      </QueryClientProvider>,
    );

    // Backup links
    expect(await screen.findByText('Local Workspace Backup')).toBeInTheDocument();
    const jsonLink = screen.getByRole('link', { name: 'Export JSON Backup' });
    expect(jsonLink).toHaveAttribute('href', '/api/backup/export?format=json');

    const sqliteLink = screen.getByRole('link', { name: 'Export SQLite (.db) Backup' });
    expect(sqliteLink).toHaveAttribute('href', '/api/backup/export?format=sqlite');

    // Sample data seeding
    const seedBtn = screen.getByRole('button', { name: 'Seed Sample Workspace Data' });
    await user.click(seedBtn);

    await waitFor(() => {
      expect(seedFn).toHaveBeenCalled();
    });
    expect(await screen.findByText(/Sample data successfully seeded!/)).toBeInTheDocument();
  });
});
