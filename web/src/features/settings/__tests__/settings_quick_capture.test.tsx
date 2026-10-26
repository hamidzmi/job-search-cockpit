import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CockpitApiProvider } from '../../../api/CockpitApiContext';
import type {
  AISettings,
  AISettingsUpsertInput,
  Company,
  ExtractedJobPostingDraft,
  OpenRouterModel,
  Opportunity,
  QuickCaptureExtractInput,
  QuickCaptureSaveInput,
} from '../../../api/types';
import { AppRoutes } from '../../../App';

const CATALOG_MODELS: OpenRouterModel[] = [
  {
    id: 'anthropic/claude-3.5-haiku',
    name: 'Anthropic: Claude 3.5 Haiku',
    description: 'Fast model for structured Job Posting extraction',
    context_length: 200000,
    prompt_pricing: '$0.80/M',
    completion_pricing: '$4.00/M',
  },
  {
    id: 'anthropic/claude-3.7-sonnet',
    name: 'Anthropic: Claude 3.7 Sonnet',
    description: 'Reasoning model for Fit Assessments and tailoring',
    context_length: 200000,
    prompt_pricing: '$3.00/M',
    completion_pricing: '$15.00/M',
  },
  {
    id: 'openai/gpt-4o-mini',
    name: 'OpenAI: GPT-4o Mini',
    description: 'Lightweight fast extraction model',
    context_length: 128000,
    prompt_pricing: '$0.15/M',
    completion_pricing: '$0.60/M',
  },
];

function installStubbedSettingsAndCaptureHandlers(seed?: {
  aiSettings?: Partial<AISettings>;
  companies?: Company[];
  opportunities?: Opportunity[];
  extractHandler?: (input: QuickCaptureExtractInput, settings: AISettings) => ExtractedJobPostingDraft;
}) {
  let aiSettings: AISettings = {
    provider: 'openrouter',
    base_url: '',
    api_key: '',
    has_api_key: false,
    is_configured: false,
    fast_extraction_model: '',
    reasoning_tailoring_model: '',
    bookmarked_models: [],
    updated_at: '2026-10-26T09:00:00Z',
    ...seed?.aiSettings,
  };
  const companies: Company[] = [...(seed?.companies ?? [])];
  const opportunities: Opportunity[] = [...(seed?.opportunities ?? [])];
  const extractCalls: QuickCaptureExtractInput[] = [];
  const saveCalls: QuickCaptureSaveInput[] = [];

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

    if (url.pathname === '/api/settings/ai' && method === 'GET') {
      return jsonResponse(aiSettings);
    }

    if (url.pathname === '/api/settings/ai' && method === 'PUT') {
      const payload = JSON.parse(bodyText) as AISettingsUpsertInput;
      const dedupBookmarks = Array.from(
        new Set(payload.bookmarked_models.map((s) => s.trim()).filter(Boolean)),
      );
      aiSettings = {
        provider: payload.provider,
        base_url: payload.base_url,
        api_key: payload.api_key,
        has_api_key: Boolean(payload.api_key.trim()),
        is_configured:
          Boolean(payload.api_key.trim()) ||
          (payload.provider === 'ollama' && Boolean(payload.base_url.trim())),
        fast_extraction_model: payload.fast_extraction_model,
        reasoning_tailoring_model: payload.reasoning_tailoring_model,
        bookmarked_models: dedupBookmarks,
        updated_at: '2026-10-26T10:00:00Z',
      };
      return jsonResponse(aiSettings);
    }

    if (url.pathname === '/api/settings/ai/openrouter-models' && method === 'GET') {
      const q = (url.searchParams.get('q') ?? '').trim().toLowerCase();
      const filtered = CATALOG_MODELS.filter((m) => {
        if (!q) return true;
        return `${m.id} ${m.name} ${m.description}`.toLowerCase().includes(q);
      });
      return jsonResponse(filtered);
    }

    if (url.pathname === '/api/quick-capture/extract' && method === 'POST') {
      const payload = JSON.parse(bodyText) as QuickCaptureExtractInput;
      extractCalls.push(payload);
      if (seed?.extractHandler) {
        return jsonResponse(seed.extractHandler(payload, aiSettings));
      }
      const modelUsed =
        payload.model_override?.trim() ||
        aiSettings.fast_extraction_model ||
        'anthropic/claude-3.5-haiku';
      return jsonResponse({
        company_name: 'Linear',
        company_domain: 'linear.app',
        role_title: 'Staff Product Infrastructure Engineer',
        seniority: 'Staff',
        location: 'San Francisco, CA',
        remote_policy: 'Remote',
        stated_min_comp: 220000,
        stated_max_comp: 260000,
        currency: 'USD',
        tech_stack_tags: ['TypeScript', 'React', 'GraphQL'],
        requirements: ['7+ years distributed sync engines', 'Deep TypeScript experience'],
        source_url: payload.source_url,
        raw_text: payload.raw_text || 'Cleaned Linear job description',
        extraction_source: aiSettings.is_configured ? 'llm' : 'deterministic',
        model_used: aiSettings.is_configured ? modelUsed : '',
        fetch_warning: '',
        assistant_warning: aiSettings.is_configured
          ? ''
          : 'No AI API key configured; using deterministic parser and manual entry.',
      } satisfies ExtractedJobPostingDraft);
    }

    if (url.pathname === '/api/quick-capture' && method === 'POST') {
      const payload = JSON.parse(bodyText) as QuickCaptureSaveInput;
      saveCalls.push(payload);
      let company = companies.find(
        (c) => c.name.toLowerCase() === payload.company_name.trim().toLowerCase(),
      );
      if (!company) {
        company = {
          id: `comp-${companies.length + 1}`,
          name: payload.company_name.trim(),
          domain: payload.company_domain,
          headquarters: payload.location,
          remote_policy: payload.remote_policy,
          size_stage: '',
          notes: '',
          created_at: '2026-10-26T10:05:00Z',
          updated_at: '2026-10-26T10:05:00Z',
        };
        companies.push(company);
      }
      const created: Opportunity = {
        id: `opp-${opportunities.length + 1}`,
        company_id: company.id,
        company_name: company.name,
        role_title: payload.role_title,
        seniority: payload.seniority,
        location: payload.location,
        remote_policy: payload.remote_policy,
        priority: payload.priority,
        stage: payload.stage ?? 'Discovered',
        outcome: 'Active',
        notes: payload.notes,
        job_posting: {
          ...payload.job_posting,
          has_content: true,
        },
        last_activity_at: '2026-10-26T10:05:00Z',
        created_at: '2026-10-26T10:05:00Z',
        updated_at: '2026-10-26T10:05:00Z',
      };
      opportunities.unshift(created);
      return jsonResponse(created, 201);
    }

    if (url.pathname === '/api/companies' && method === 'GET') {
      return jsonResponse(companies);
    }

    if (url.pathname === '/api/opportunities' && method === 'GET') {
      return jsonResponse(opportunities);
    }

    return jsonResponse({ error: { code: 'not_found', message: 'not found' } }, 404);
  });

  vi.stubGlobal('fetch', fetchMock);
  return { fetchMock, extractCalls, saveCalls, opportunities, companies };
}

function renderCockpit(initialEntries = ['/settings']) {
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

describe('AI Settings & Quick Capture Workspace Views', () => {
  it('configures OpenRouter credentials, searches the live model catalog, bookmarks models, adds custom slugs, and assigns Fast Extraction & Reasoning slots', async () => {
    const user = userEvent.setup();
    installStubbedSettingsAndCaptureHandlers();

    renderCockpit(['/settings']);

    expect(await screen.findByText(/Manual Mode Active/i)).toBeInTheDocument();

    // Configure OpenRouter API Key
    await user.type(screen.getByLabelText(/API Key/i), 'sk-or-v1-my-secret');

    // Search the OpenRouter model catalog
    const searchInput = screen.getByLabelText(/Search OpenRouter Models/i);
    await user.type(searchInput, 'haiku');

    expect(await screen.findByText('Anthropic: Claude 3.5 Haiku')).toBeInTheDocument();
    expect(screen.queryByText('Anthropic: Claude 3.7 Sonnet')).not.toBeInTheDocument();

    // Bookmark Haiku and set as Fast Extraction model from catalog
    await user.click(
      screen.getByRole('button', { name: 'Bookmark anthropic/claude-3.5-haiku' }),
    );
    await user.click(
      screen.getByRole('button', {
        name: 'Set anthropic/claude-3.5-haiku as Fast Extraction',
      }),
    );

    // Clear catalog search and set Claude 3.7 Sonnet as Reasoning / Tailoring model
    await user.clear(searchInput);
    expect(await screen.findByText('Anthropic: Claude 3.7 Sonnet')).toBeInTheDocument();
    await user.click(
      screen.getByRole('button', {
        name: 'Set anthropic/claude-3.7-sonnet as Reasoning / Tailoring',
      }),
    );

    // Type a custom model slug and bookmark it
    await user.type(
      screen.getByLabelText(/Custom Model Slug/i),
      'deepseek/deepseek-r1-distill-llama-70b',
    );
    await user.click(screen.getByRole('button', { name: 'Add Bookmark' }));

    // Save AI Settings
    await user.click(screen.getByRole('button', { name: 'Save AI Settings' }));

    expect(await screen.findByText('AI settings saved.')).toBeInTheDocument();
    expect(screen.getByText(/AI Configured/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Fast Extraction Model/i)).toHaveValue(
      'anthropic/claude-3.5-haiku',
    );
    expect(screen.getByLabelText(/Reasoning \/ Tailoring Model/i)).toHaveValue(
      'anthropic/claude-3.7-sonnet',
    );
  });

  it('opens global Quick Capture modal from Pipeline, overrides extraction model from bookmarked models, extracts URL + text, and saves Opportunity', async () => {
    const user = userEvent.setup();
    const { extractCalls, saveCalls } = installStubbedSettingsAndCaptureHandlers({
      aiSettings: {
        provider: 'openrouter',
        api_key: 'sk-or-v1-configured',
        has_api_key: true,
        is_configured: true,
        fast_extraction_model: 'anthropic/claude-3.5-haiku',
        reasoning_tailoring_model: 'anthropic/claude-3.7-sonnet',
        bookmarked_models: ['anthropic/claude-3.5-haiku', 'openai/gpt-4o-mini'],
      },
    });

    renderCockpit(['/opportunities']);

    // Open Quick Capture from global sidebar button
    await user.click(await screen.findByRole('button', { name: /Quick Capture/i }));

    const dialog = await screen.findByRole('dialog', { name: 'Quick Capture' });

    await user.type(
      within(dialog).getByLabelText(/Job Posting URL/i),
      'https://linear.app/careers/staff-infra',
    );
    await user.type(
      within(dialog).getByLabelText(/Paste Job Posting Text/i),
      'Linear is looking for a Staff Product Infrastructure Engineer.',
    );

    // Select bookmarked model override on the fly
    const modelSelect = within(dialog).getByLabelText(/Extraction Model/i);
    await user.selectOptions(modelSelect, 'openai/gpt-4o-mini');

    await user.click(within(dialog).getByRole('button', { name: /Extract Job Posting/i }));

    // Verify extracted fields populate the draft form
    const companyInput = within(dialog).getByLabelText(/Company Name/i);
    expect(companyInput).toHaveValue('Linear');
    expect(within(dialog).getByLabelText(/Role Title/i)).toHaveValue(
      'Staff Product Infrastructure Engineer',
    );
    expect(extractCalls).toHaveLength(1);
    expect(extractCalls[0]).toEqual({
      source_url: 'https://linear.app/careers/staff-infra',
      raw_text: 'Linear is looking for a Staff Product Infrastructure Engineer.',
      model_override: 'openai/gpt-4o-mini',
    });

    // Candidate reviews/edits a field before saving
    const roleInput = within(dialog).getByLabelText(/Role Title/i);
    await user.clear(roleInput);
    await user.type(roleInput, 'Staff Sync Engine Engineer');

    await user.click(within(dialog).getByRole('button', { name: /Save Opportunity/i }));

    expect(saveCalls).toHaveLength(1);
    expect(saveCalls[0].company_name).toBe('Linear');
    expect(saveCalls[0].role_title).toBe('Staff Sync Engine Engineer');

    // Verify new Opportunity appears on the Pipeline Kanban board
    const discoveredCol = await screen.findByRole('region', {
      name: 'Discovered Stage Column',
    });
    expect(
      await within(discoveredCol).findByText('Staff Sync Engine Engineer'),
    ).toBeInTheDocument();
  });

  it('seamlessly falls back to deterministic/manual fields when URL is blocked and no LLM key is configured', async () => {
    const user = userEvent.setup();
    installStubbedSettingsAndCaptureHandlers({
      aiSettings: {
        provider: 'openrouter',
        api_key: '',
        has_api_key: false,
        is_configured: false,
      },
      extractHandler: (input) => ({
        company_name: 'Tailscale',
        company_domain: 'tailscale.com',
        role_title: 'Senior Network Systems Engineer',
        seniority: 'Senior',
        location: 'Remote',
        remote_policy: 'Remote',
        stated_min_comp: 195000,
        stated_max_comp: 235000,
        currency: 'USD',
        tech_stack_tags: ['Go', 'WireGuard'],
        requirements: ['5+ years Go and networking protocols'],
        source_url: input.source_url,
        raw_text: input.raw_text,
        extraction_source: 'manual_fallback',
        model_used: '',
        fetch_warning:
          'Could not fetch https://linkedin.com/jobs/view/123 (HTTP 403); falling back to pasted text / manual fields.',
        assistant_warning:
          'No AI API key configured; using deterministic parser and manual entry.',
      }),
    });

    renderCockpit(['/opportunities']);

    await user.click(await screen.findByRole('button', { name: /Quick Capture/i }));
    const dialog = await screen.findByRole('dialog', { name: 'Quick Capture' });

    await user.type(
      within(dialog).getByLabelText(/Job Posting URL/i),
      'https://linkedin.com/jobs/view/123',
    );
    await user.type(
      within(dialog).getByLabelText(/Paste Job Posting Text/i),
      'Company: Tailscale\nRole: Senior Network Systems Engineer\nCompensation: $195,000 - $235,000',
    );

    await user.click(within(dialog).getByRole('button', { name: /Extract Job Posting/i }));

    expect(
      await within(dialog).findByText(/falling back to pasted text/i),
    ).toBeInTheDocument();
    expect(
      within(dialog).getByText(/No AI API key configured/i),
    ).toBeInTheDocument();
    expect(within(dialog).getByLabelText(/Company Name/i)).toHaveValue('Tailscale');
    expect(within(dialog).getByLabelText(/Role Title/i)).toHaveValue(
      'Senior Network Systems Engineer',
    );

    await user.click(within(dialog).getByRole('button', { name: /Save Opportunity/i }));

    const discoveredCol = await screen.findByRole('region', {
      name: 'Discovered Stage Column',
    });
    expect(
      await within(discoveredCol).findByText('Senior Network Systems Engineer'),
    ).toBeInTheDocument();
  });
});
