import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CockpitApiProvider } from '../../../api/CockpitApiContext';
import type {
  Company,
  Contact,
  DebriefUpsertInput,
  InterviewRound,
  InterviewRoundUpsertInput,
  Opportunity,
  QuestionBankEntry,
  Story,
  StoryPriorRoundRef,
} from '../../../api/types';
import { AppRoutes } from '../../../App';

function computeRepeatedStories(rounds: InterviewRound[]): Record<string, Record<string, StoryPriorRoundRef[]>> {
  const result: Record<string, Record<string, StoryPriorRoundRef[]>> = {};
  const byOpp: Record<string, InterviewRound[]> = {};
  for (const r of rounds) {
    if (r.status === 'Cancelled') continue;
    byOpp[r.opportunity_id] = [...(byOpp[r.opportunity_id] ?? []), r];
  }
  for (const oppRounds of Object.values(byOpp)) {
    oppRounds.sort((a, b) => a.round_order - b.round_order);
    const seen: Record<string, StoryPriorRoundRef[]> = {};
    for (const r of oppRounds) {
      for (const sid of r.story_ids) {
        if (seen[sid] && seen[sid].length > 0) {
          if (!result[r.id]) result[r.id] = {};
          result[r.id][sid] = [...seen[sid]];
        }
      }
      for (const sid of r.story_ids) {
        seen[sid] = [
          ...(seen[sid] ?? []),
          {
            round_id: r.id,
            round_order: r.round_order,
            round_type: r.round_type,
            title: r.title,
          },
        ];
      }
    }
  }
  return result;
}

function installInterviewAndQuestionBankHandlers() {
  const companies: Company[] = [
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
  ];
  const opportunities: Opportunity[] = [
    {
      id: 'opp-1',
      company_id: 'comp-1',
      company_name: 'Stripe',
      role_title: 'Staff Core Payments Engineer',
      seniority: 'Staff',
      location: 'San Francisco, CA',
      remote_policy: 'Hybrid',
      priority: 'High',
      stage: 'Interviewing',
      outcome: 'Active',
      notes: '',
      job_posting: {
        source_url: '',
        raw_text: '',
        requirements: [],
        tech_stack_tags: ['Go', 'PostgreSQL'],
        stated_min_comp: 0,
        stated_max_comp: 0,
        currency: 'USD',
        has_content: true,
      },
      last_activity_at: '2026-10-20T09:00:00Z',
      created_at: '2026-10-20T09:00:00Z',
      updated_at: '2026-10-20T09:00:00Z',
    },
  ];
  const contacts: Contact[] = [
    {
      id: 'cnt-1',
      company_id: 'comp-1',
      company_name: 'Stripe',
      name: 'Alice Chen',
      title: 'Staff Payments Engineer',
      email: 'alice@stripe.com',
      linkedin_url: '',
      warmth: 'Advocate',
      notes: '',
      linked_opportunities: [],
      created_at: '2026-10-20T09:00:00Z',
      updated_at: '2026-10-20T09:00:00Z',
    },
  ];
  const stories: Story[] = [
    {
      id: 'story-ledger',
      title: 'Zero-Downtime Ledger Sharding',
      hook: 'Sharded $40B/yr ledger across 16 clusters with zero downtime.',
      situation: 'Single primary hit 85% saturation.',
      task: 'Shard in 10 weeks.',
      action: 'Built shadow verifier in Go.',
      result: 'Cut latency 58%.',
      competency_tags: ['System Design', 'Leadership'],
      skill_tags: ['Go', 'PostgreSQL'],
      created_at: '2026-10-20T09:00:00Z',
      updated_at: '2026-10-20T09:00:00Z',
    },
    {
      id: 'story-conflict',
      title: 'Unblocking Cross-Team Schema Deadlock',
      hook: 'Aligned 3 platform teams on unified event schema.',
      situation: 'Teams disagreed on partitioning.',
      task: 'Resolve before launch freeze.',
      action: 'Prototyped Rust benchmark harness.',
      result: 'Shipped 1 week early.',
      competency_tags: ['Conflict', 'Ownership'],
      skill_tags: ['Kafka', 'Rust'],
      created_at: '2026-10-20T09:00:00Z',
      updated_at: '2026-10-20T09:00:00Z',
    },
  ];

  const rounds: InterviewRound[] = [];
  let idCounter = 500;

  const hydrateRounds = (oppId: string): InterviewRound[] => {
    const oppRounds = rounds
      .filter((r) => r.opportunity_id === oppId)
      .sort((a, b) => a.round_order - b.round_order);
    const warnings = computeRepeatedStories(oppRounds);
    return oppRounds.map((r) => ({
      ...r,
      contacts: r.contact_ids
        .map((cid) => contacts.find((c) => c.id === cid))
        .filter((c): c is Contact => Boolean(c))
        .map((c) => ({
          contact_id: c.id,
          name: c.name,
          title: c.title,
          email: c.email,
          warmth: c.warmth,
          company_name: c.company_name,
        })),
      stories: r.story_ids
        .map((sid) => stories.find((s) => s.id === sid))
        .filter((s): s is Story => Boolean(s))
        .map((s) => ({
          story_id: s.id,
          title: s.title,
          hook: s.hook,
          competency_tags: s.competency_tags,
          skill_tags: s.skill_tags,
          used_in_earlier_rounds: warnings[r.id]?.[s.id] ?? [],
        })),
    }));
  };

  const jsonResponse = (body: unknown, status = 200) =>
    Promise.resolve(
      new Response(status === 204 ? null : JSON.stringify(body), {
        status,
        headers: status === 204 ? {} : { 'Content-Type': 'application/json' },
      }),
    );

  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const rawUrl = typeof input === 'string' ? input : input.toString();
    const url = new URL(rawUrl, 'http://localhost');
    const method = (init?.method ?? 'GET').toUpperCase();
    const bodyText = typeof init?.body === 'string' ? init.body : '';

    if (url.pathname === '/api/opportunities/opp-1' && method === 'GET') {
      return jsonResponse({ ...opportunities[0], stage_transitions: [] });
    }
    if (url.pathname === '/api/contacts' && method === 'GET') {
      return jsonResponse(contacts);
    }
    if (url.pathname === '/api/opportunities/opp-1/contacts' && method === 'GET') {
      return jsonResponse([]);
    }
    if (url.pathname === '/api/interactions' && method === 'GET') {
      return jsonResponse([]);
    }
    if (url.pathname === '/api/action-items' && method === 'GET') {
      return jsonResponse([]);
    }
    if (url.pathname === '/api/profile' && method === 'GET') {
      return jsonResponse({
        full_name: 'Alex Rivera',
        headline: 'Staff Engineer',
        summary: '',
        work_history: [],
        skills: ['Go'],
        target_role_titles: ['Staff Engineer'],
        location_preferences: ['Remote'],
        remote_preference: 'Remote',
        compensation_floor: 240000,
        target_compensation: 280000,
        currency: 'USD',
        updated_at: '2026-10-25T09:00:00Z',
      });
    }
    if (url.pathname === '/api/stories' && method === 'GET') {
      return jsonResponse(stories);
    }
    if (url.pathname === '/api/settings/ai' && method === 'GET') {
      return jsonResponse({
        provider: 'openrouter',
        base_url: '',
        api_key: '',
        has_api_key: false,
        is_configured: false,
        fast_extraction_model: '',
        reasoning_tailoring_model: '',
        bookmarked_models: [],
        updated_at: '2026-10-25T09:00:00Z',
      });
    }
    if (url.pathname === '/api/opportunities/opp-1/fit-assessment' && method === 'GET') {
      return jsonResponse({
        opportunity_id: 'opp-1',
        overall_score: 0,
        overall_summary: '',
        matched_strengths: [],
        gaps_and_risks: [],
        compensation_alignment: '',
        recommended_story_ids: [],
        recommended_stories: [],
        model_used: '',
        updated_at: '',
      });
    }
    if (url.pathname === '/api/opportunities/opp-1/artifacts' && method === 'GET') {
      return jsonResponse([]);
    }

    const oppRoundsMatch = url.pathname.match(/^\/api\/opportunities\/([^/]+)\/rounds$/);
    if (oppRoundsMatch && method === 'GET') {
      return jsonResponse(hydrateRounds(decodeURIComponent(oppRoundsMatch[1])));
    }
    if (oppRoundsMatch && method === 'POST') {
      const oppId = decodeURIComponent(oppRoundsMatch[1]);
      const payload = JSON.parse(bodyText) as InterviewRoundUpsertInput;
      const created: InterviewRound = {
        id: `rnd-${++idCounter}`,
        opportunity_id: oppId,
        round_order: payload.round_order,
        round_type: payload.round_type,
        title: payload.title || payload.round_type,
        scheduled_at: payload.scheduled_at,
        duration_minutes: payload.duration_minutes,
        status: payload.status,
        prep_notes: payload.prep_notes,
        contact_ids: payload.contact_ids,
        story_ids: payload.story_ids,
        contacts: [],
        stories: [],
        created_at: '2026-10-26T10:00:00Z',
        updated_at: '2026-10-26T10:00:00Z',
      };
      rounds.push(created);
      const hydrated = hydrateRounds(oppId).find((r) => r.id === created.id)!;
      return jsonResponse(hydrated, 201);
    }

    const roundMatch = url.pathname.match(/^\/api\/rounds\/([^/]+)$/);
    if (roundMatch && method === 'PUT') {
      const roundId = decodeURIComponent(roundMatch[1]);
      const idx = rounds.findIndex((r) => r.id === roundId);
      const payload = JSON.parse(bodyText) as InterviewRoundUpsertInput;
      rounds[idx] = {
        ...rounds[idx],
        round_order: payload.round_order,
        round_type: payload.round_type,
        title: payload.title || payload.round_type,
        scheduled_at: payload.scheduled_at,
        duration_minutes: payload.duration_minutes,
        status: payload.status,
        prep_notes: payload.prep_notes,
        contact_ids: payload.contact_ids,
        story_ids: payload.story_ids,
        updated_at: '2026-10-26T11:00:00Z',
      };
      const hydrated = hydrateRounds(rounds[idx].opportunity_id).find(
        (r) => r.id === roundId,
      )!;
      return jsonResponse(hydrated);
    }

    const debriefMatch = url.pathname.match(/^\/api\/rounds\/([^/]+)\/debrief$/);
    if (debriefMatch && method === 'PUT') {
      const roundId = decodeURIComponent(debriefMatch[1]);
      const idx = rounds.findIndex((r) => r.id === roundId);
      const payload = JSON.parse(bodyText) as DebriefUpsertInput;
      rounds[idx] = {
        ...rounds[idx],
        debrief: {
          questions_asked: payload.questions_asked.map((q) => ({
            ...q,
            answered_with_story_title:
              stories.find((s) => s.id === q.answered_with_story_id)?.title ?? '',
          })),
          went_well: payload.went_well,
          to_improve: payload.to_improve,
          signals: payload.signals,
          confidence_score: payload.confidence_score,
          follow_up_notes: payload.follow_up_notes,
          updated_at: '2026-10-26T12:00:00Z',
        },
        updated_at: '2026-10-26T12:00:00Z',
      };
      const hydrated = hydrateRounds(rounds[idx].opportunity_id).find(
        (r) => r.id === roundId,
      )!;
      return jsonResponse(hydrated);
    }

    if (url.pathname === '/api/questions' && method === 'GET') {
      const qFilter = (url.searchParams.get('q') ?? '').trim().toLowerCase();
      const catFilter = (url.searchParams.get('category') ?? '').trim().toLowerCase();
      const entries: QuestionBankEntry[] = [];
      for (const r of rounds) {
        if (!r.debrief) continue;
        const opp = opportunities.find((o) => o.id === r.opportunity_id)!;
        const comp = companies.find((c) => c.id === opp.company_id)!;
        for (const q of r.debrief.questions_asked) {
          const entry: QuestionBankEntry = {
            round_id: r.id,
            opportunity_id: opp.id,
            opportunity_role_title: opp.role_title,
            company_id: comp.id,
            company_name: comp.name,
            round_order: r.round_order,
            round_type: r.round_type,
            round_title: r.title,
            question: q.question,
            category: q.category,
            answered_with_story_id: q.answered_with_story_id,
            answered_with_story_title: q.answered_with_story_title,
            notes: q.notes,
            recorded_at: r.debrief.updated_at,
          };
          if (catFilter && entry.category.toLowerCase() !== catFilter) continue;
          if (qFilter) {
            const hay = [
              entry.question,
              entry.category,
              entry.company_name,
              entry.opportunity_role_title,
              entry.round_type,
              entry.answered_with_story_title,
              entry.notes,
            ]
              .join(' ')
              .toLowerCase();
            if (!hay.includes(qFilter)) continue;
          }
          entries.push(entry);
        }
      }
      return jsonResponse(entries);
    }

    return jsonResponse(
      { error: { code: 'not_found', message: `unhandled ${method} ${url.pathname}` } },
      404,
    );
  });

  vi.stubGlobal('fetch', fetchMock);
  return { fetchMock, rounds };
}

function renderCockpit(initialEntries = ['/opportunities/opp-1']) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
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

describe('Interview Rounds, Loop Story De-duplication & Debrief Question Bank', () => {
  it('adds ordered Interview Rounds with linked Contacts and Stories, shows non-blocking repeat warnings across non-cancelled rounds of the same Opportunity, records structured Debriefs, and browses aggregated questions alongside the Story Bank', async () => {
    const user = userEvent.setup();
    installInterviewAndQuestionBankHandlers();

    renderCockpit(['/opportunities/opp-1']);

    expect(
      await screen.findByRole('heading', { name: 'Staff Core Payments Engineer' }),
    ).toBeInTheDocument();

    const roundsSection = await screen.findByRole('region', {
      name: 'Opportunity Interview Rounds and Debriefs',
    });

    // 1. Add Round 1: Hiring Manager (Completed) with Alice Chen and Zero-Downtime Ledger Sharding
    await user.selectOptions(
      within(roundsSection).getByLabelText('Interview Round Type'),
      'Hiring Manager',
    );
    await user.clear(within(roundsSection).getByLabelText('Interview Round Title'));
    await user.type(
      within(roundsSection).getByLabelText('Interview Round Title'),
      'HM Deep Dive',
    );
    await user.selectOptions(
      within(roundsSection).getByLabelText('Interview Round Status'),
      'Completed',
    );
    await user.type(
      within(roundsSection).getByLabelText('Preparation Notes'),
      'Cover ledger availability',
    );
    await user.click(
      within(roundsSection).getByLabelText('Interviewer Contact: Alice Chen'),
    );
    await user.click(
      within(roundsSection).getByLabelText('Planned Story: Zero-Downtime Ledger Sharding'),
    );
    await user.click(
      within(roundsSection).getByRole('button', { name: 'Add Interview Round' }),
    );

    expect(
      await within(roundsSection).findByRole('heading', {
        name: /Round 1: HM Deep Dive/i,
      }),
    ).toBeInTheDocument();

    // 2. While preparing Round 2, check the live repeat warning in the Story selector and after adding Round 2
    expect(
      within(roundsSection).getByText(
        /Already used in Round 1 \(HM Deep Dive\)/i,
      ),
    ).toBeInTheDocument();

    await user.selectOptions(
      within(roundsSection).getByLabelText('Interview Round Type'),
      'System Design',
    );
    await user.clear(within(roundsSection).getByLabelText('Interview Round Title'));
    await user.type(
      within(roundsSection).getByLabelText('Interview Round Title'),
      'Distributed Settlement Design',
    );
    await user.selectOptions(
      within(roundsSection).getByLabelText('Interview Round Status'),
      'Scheduled',
    );
    await user.click(
      within(roundsSection).getByLabelText('Planned Story: Zero-Downtime Ledger Sharding'),
    );
    await user.click(
      within(roundsSection).getByLabelText(
        'Planned Story: Unblocking Cross-Team Schema Deadlock',
      ),
    );
    await user.click(
      within(roundsSection).getByRole('button', { name: 'Add Interview Round' }),
    );

    const round2Card = (
      await within(roundsSection).findByRole('heading', {
        name: /Round 2: Distributed Settlement Design/i,
      })
    ).closest('article')!;

    expect(
      within(round2Card).getByText(/Repeated Story: Used in Round 1 \(HM Deep Dive\)/i),
    ).toBeInTheDocument();

    // 2b. Edit & reorder Round 2 to verify editing RoundOrder / Status works in UI
    await user.click(within(round2Card).getByRole('button', { name: 'Edit Round' }));
    const orderInput = within(roundsSection).getByLabelText('Interview Round Order');
    await user.clear(orderInput);
    await user.type(orderInput, '3');
    await user.selectOptions(
      within(roundsSection).getByLabelText('Interview Round Status'),
      'Completed',
    );
    await user.click(
      within(roundsSection).getByRole('button', { name: 'Save Interview Round' }),
    );
    expect(
      await within(roundsSection).findByRole('heading', {
        name: /Round 3: Distributed Settlement Design/i,
      }),
    ).toBeInTheDocument();

    // 3. Record a structured 1:1 Debrief on Round 1
    const round1Card = within(roundsSection)
      .getByRole('heading', { name: /Round 1: HM Deep Dive/i })
      .closest('article')!;

    await user.click(
      within(round1Card).getByRole('button', { name: /Record \/ Edit Debrief/i }),
    );

    const debriefDialog = screen.getByRole('dialog', {
      name: /Debrief: HM Deep Dive/i,
    });
    await user.click(
      within(debriefDialog).getByRole('button', { name: '+ Add Interview Question' }),
    );
    await user.type(
      within(debriefDialog).getByLabelText('Question Asked 1'),
      'How do you guarantee zero-downtime schema and shard migrations?',
    );
    await user.type(
      within(debriefDialog).getByLabelText('Question Category 1'),
      'System Design',
    );
    await user.selectOptions(
      within(debriefDialog).getByLabelText('Answered With Story 1'),
      'story-ledger',
    );
    await user.type(
      within(debriefDialog).getByLabelText('Question Notes 1'),
      'Discussed shadow dual-write verification',
    );
    await user.type(
      within(debriefDialog).getByLabelText('What Went Well'),
      'Clear quantification of latency and zero-incident cutover',
    );
    await user.type(
      within(debriefDialog).getByLabelText('Areas to Improve'),
      'Keep initial context under 90 seconds',
    );
    await user.type(
      within(debriefDialog).getByLabelText('Company & Team Signals'),
      'High urgency around multi-region active-active ledger',
    );
    await user.selectOptions(
      within(debriefDialog).getByLabelText('Confidence Score (1-5)'),
      '5',
    );
    await user.type(
      within(debriefDialog).getByLabelText('Follow-up Notes'),
      'Send follow-up note to Alice',
    );
    await user.click(within(debriefDialog).getByRole('button', { name: 'Save Debrief' }));

    expect(
      await within(round1Card).findByText(
        /How do you guarantee zero-downtime schema and shard migrations\?/i,
      ),
    ).toBeInTheDocument();
    expect(within(round1Card).getByText(/Confidence: 5 \/ 5/i)).toBeInTheDocument();

    // 4. Navigate to Candidate Profile & Story Bank (/profile) and browse aggregated Debrief Question Bank
    await user.click(
      screen.getByRole('link', { name: /Candidate Profile & Story Bank/i }),
    );

    const questionBankSection = await screen.findByRole('region', {
      name: 'Interview Question Bank',
    });
    expect(
      await within(questionBankSection).findByText(
        'How do you guarantee zero-downtime schema and shard migrations?',
      ),
    ).toBeInTheDocument();
    expect(within(questionBankSection).getByText('Stripe')).toBeInTheDocument();
    expect(
      within(questionBankSection).getByText('Staff Core Payments Engineer'),
    ).toBeInTheDocument();
    expect(within(questionBankSection).getByText('Hiring Manager')).toBeInTheDocument();
    expect(
      within(questionBankSection).getByText(/Story: Zero-Downtime Ledger Sharding/i),
    ).toBeInTheDocument();
  });
});
