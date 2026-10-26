import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CockpitApiProvider } from '../../../api/CockpitApiContext';
import type {
  CandidateProfile,
  CandidateProfileUpsertInput,
  Story,
  StoryUpsertInput,
} from '../../../api/types';
import { AppRoutes } from '../../../App';

function installStubbedProfileAndStoryHandlers(seed?: {
  profile?: CandidateProfile;
  stories?: Story[];
}) {
  let profile: CandidateProfile = seed?.profile ?? {
    full_name: '',
    headline: '',
    summary: '',
    work_history: [],
    skills: [],
    target_role_titles: [],
    location_preferences: [],
    remote_preference: '',
    compensation_floor: 0,
    target_compensation: 0,
    currency: 'USD',
    updated_at: '2026-10-25T09:00:00Z',
  };
  let stories: Story[] = [...(seed?.stories ?? [])];
  let idCounter = 200;

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

    if (url.pathname === '/api/profile' && method === 'GET') {
      return jsonResponse(profile);
    }

    if (url.pathname === '/api/profile' && method === 'PUT') {
      const payload = JSON.parse(bodyText) as CandidateProfileUpsertInput;
      profile = {
        ...payload,
        updated_at: '2026-10-26T12:00:00Z',
      };
      return jsonResponse(profile);
    }

    if (url.pathname === '/api/stories' && method === 'GET') {
      const q = (url.searchParams.get('q') ?? '').trim().toLowerCase();
      const compTag = (url.searchParams.get('competency_tag') ?? '').trim().toLowerCase();
      const skillTag = (url.searchParams.get('skill_tag') ?? '').trim().toLowerCase();

      const filtered = stories.filter((s) => {
        if (
          compTag &&
          !s.competency_tags.some((tag) => tag.trim().toLowerCase() === compTag)
        ) {
          return false;
        }
        if (skillTag && !s.skill_tags.some((tag) => tag.trim().toLowerCase() === skillTag)) {
          return false;
        }
        if (q) {
          const hay = [
            s.title,
            s.hook,
            s.situation,
            s.task,
            s.action,
            s.result,
            ...s.competency_tags,
            ...s.skill_tags,
          ]
            .join(' ')
            .toLowerCase();
          if (!hay.includes(q)) return false;
        }
        return true;
      });
      return jsonResponse(filtered);
    }

    if (url.pathname === '/api/stories' && method === 'POST') {
      const payload = JSON.parse(bodyText) as StoryUpsertInput;
      const created: Story = {
        id: `story-${++idCounter}`,
        ...payload,
        created_at: '2026-10-26T10:00:00Z',
        updated_at: '2026-10-26T10:00:00Z',
      };
      stories.unshift(created);
      return jsonResponse(created, 201);
    }

    const storyMatch = url.pathname.match(/^\/api\/stories\/([^/]+)$/);
    if (storyMatch && method === 'PUT') {
      const id = decodeURIComponent(storyMatch[1]);
      const idx = stories.findIndex((s) => s.id === id);
      if (idx === -1) {
        return jsonResponse({ error: { code: 'not_found', message: 'not found' } }, 404);
      }
      const payload = JSON.parse(bodyText) as StoryUpsertInput;
      const updated: Story = {
        ...stories[idx],
        ...payload,
        updated_at: '2026-10-26T11:00:00Z',
      };
      stories[idx] = updated;
      return jsonResponse(updated);
    }

    if (storyMatch && method === 'DELETE') {
      const id = decodeURIComponent(storyMatch[1]);
      const idx = stories.findIndex((s) => s.id === id);
      if (idx === -1) {
        return jsonResponse({ error: { code: 'not_found', message: 'not found' } }, 404);
      }
      stories = stories.filter((s) => s.id !== id);
      return jsonResponse(null, 204);
    }

    return jsonResponse(
      { error: { code: 'not_found', message: `unhandled ${method} ${url.pathname}` } },
      404,
    );
  });

  vi.stubGlobal('fetch', fetchMock);
  return { fetchMock, getProfile: () => profile, getStories: () => stories };
}

function renderProfileWorkspace() {
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
        <MemoryRouter initialEntries={['/profile']}>
          <AppRoutes />
        </MemoryRouter>
      </CockpitApiProvider>
    </QueryClientProvider>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('Candidate Profile & Story Bank Workspace (/profile over /api/*)', () => {
  it('views and saves the canonical Candidate Profile including work history with impact bullets, skills, target preferences, and compensation floor', async () => {
    const user = userEvent.setup();
    const { getProfile } = installStubbedProfileAndStoryHandlers();

    renderProfileWorkspace();

    expect(
      await screen.findByRole('heading', { name: /Candidate Profile & Story Bank/i }),
    ).toBeInTheDocument();

    const profileSection = await screen.findByRole('region', { name: 'Candidate Profile' });

    await user.type(within(profileSection).getByLabelText(/Full Name/i), 'Alex Rivera');
    await user.type(
      within(profileSection).getByLabelText(/Headline/i),
      'Staff Distributed Systems Engineer',
    );
    await user.type(
      within(profileSection).getByLabelText(/Profile Summary/i),
      '12 years building low-latency payment ledgers and cloud storage engines.',
    );
    await user.type(
      within(profileSection).getByLabelText(/Skills \(comma-separated\)/i),
      'Go, Rust, Kafka, PostgreSQL',
    );
    await user.type(
      within(profileSection).getByLabelText(/Target Role Titles \(comma-separated\)/i),
      'Staff Backend Engineer, Principal Systems Engineer',
    );
    await user.type(
      within(profileSection).getByLabelText(/Location Preferences \(comma-separated\)/i),
      'San Francisco, CA, Remote (US)',
    );
    await user.type(
      within(profileSection).getByLabelText(/Remote Preference/i),
      'Remote or Hybrid',
    );

    const floorInput = within(profileSection).getByLabelText(/Minimum Compensation Floor/i);
    await user.clear(floorInput);
    await user.type(floorInput, '240000');

    const targetCompInput = within(profileSection).getByLabelText(/Target Compensation/i);
    await user.clear(targetCompInput);
    await user.type(targetCompInput, '285000');

    await user.click(
      within(profileSection).getByRole('button', { name: '+ Add Work History Entry' }),
    );

    const entryRegion = within(profileSection).getByRole('group', {
      name: 'Work History Entry 1',
    });
    await user.type(within(entryRegion).getByLabelText(/Company/i), 'Stripe');
    await user.type(within(entryRegion).getByLabelText(/Role Title/i), 'Staff Software Engineer');
    await user.type(within(entryRegion).getByLabelText(/Start Date/i), '2022-01');
    await user.type(within(entryRegion).getByLabelText(/End Date/i), 'Present');
    await user.type(
      within(entryRegion).getByLabelText(/Impact Bullets \(one per line\)/i),
      'Cut p99 settlement latency by 42% across 3 regions\nLed online sharding of $40B/yr ledger',
    );

    await user.click(
      within(profileSection).getByRole('button', { name: 'Save Candidate Profile' }),
    );

    expect(await screen.findByText('Candidate Profile saved.')).toBeInTheDocument();

    const saved = getProfile();
    expect(saved.full_name).toBe('Alex Rivera');
    expect(saved.compensation_floor).toBe(240000);
    expect(saved.target_compensation).toBe(285000);
    expect(saved.skills).toEqual(['Go', 'Rust', 'Kafka', 'PostgreSQL']);
    expect(saved.target_role_titles).toEqual([
      'Staff Backend Engineer',
      'Principal Systems Engineer',
    ]);
    expect(saved.work_history).toHaveLength(1);
    expect(saved.work_history[0].company).toBe('Stripe');
    expect(saved.work_history[0].impact_bullets).toEqual([
      'Cut p99 settlement latency by 42% across 3 regions',
      'Led online sharding of $40B/yr ledger',
    ]);
  });

  it('creates, edits, searches, filters by Competency Tags and Skill Tags, and deletes Stories in the Story Bank', async () => {
    const user = userEvent.setup();
    installStubbedProfileAndStoryHandlers({
      stories: [
        {
          id: 'story-existing',
          title: 'Unblocking Cross-Team Schema Deadlock',
          hook: 'Aligned 3 platform teams on a unified event schema in 48 hours.',
          situation: 'Ingestion and billing teams disagreed on Kafka protobuf partitioning.',
          task: 'Resolve the architectural impasse before Q3 freeze.',
          action: 'Built a Rust benchmark suite comparing both partitioning strategies.',
          result: 'Achieved unanimous sign-off and launched 1 week early.',
          competency_tags: ['Conflict', 'Ownership'],
          skill_tags: ['Kafka', 'Rust'],
          created_at: '2026-10-20T09:00:00Z',
          updated_at: '2026-10-20T09:00:00Z',
        },
      ],
    });

    renderProfileWorkspace();

    const storyBankSection = await screen.findByRole('region', { name: 'Story Bank' });
    expect(
      await within(storyBankSection).findByRole('heading', {
        name: 'Unblocking Cross-Team Schema Deadlock',
      }),
    ).toBeInTheDocument();

    // Create a new Story
    await user.click(within(storyBankSection).getByRole('button', { name: '+ New Story' }));
    const createModal = screen.getByRole('dialog', { name: 'New Story' });

    await user.type(
      within(createModal).getByLabelText(/Story Title/i),
      'Zero-Downtime Ledger Sharding',
    );
    await user.type(
      within(createModal).getByLabelText(/Hook/i),
      'Sharded $40B/yr ledger across 16 clusters with zero customer downtime.',
    );
    await user.type(
      within(createModal).getByLabelText(/Situation/i),
      'Primary PostgreSQL cluster reached 85% write saturation.',
    );
    await user.type(
      within(createModal).getByLabelText(/Task/i),
      'Execute horizontal sharding in 10 weeks before Black Friday.',
    );
    await user.type(
      within(createModal).getByLabelText(/Action/i),
      'Built dual-write shadow verifier in Go and phased traffic cutover.',
    );
    await user.type(
      within(createModal).getByLabelText(/Result \(Quantified Impact\)/i),
      'Reduced p99 write latency by 58% and handled 3.2x peak load.',
    );
    await user.type(
      within(createModal).getByLabelText(/Competency Tags \(comma-separated\)/i),
      'System Design, Leadership',
    );
    await user.type(
      within(createModal).getByLabelText(/Skill Tags \(comma-separated\)/i),
      'Go, PostgreSQL',
    );
    await user.click(within(createModal).getByRole('button', { name: 'Create Story' }));

    expect(
      await within(storyBankSection).findByRole('heading', {
        name: 'Zero-Downtime Ledger Sharding',
      }),
    ).toBeInTheDocument();

    // Edit the newly created Story
    const ledgerCard = within(storyBankSection)
      .getByRole('heading', { name: 'Zero-Downtime Ledger Sharding' })
      .closest('article')!;
    await user.click(within(ledgerCard).getByRole('button', { name: 'Edit Story' }));

    const editModal = screen.getByRole('dialog', { name: 'Edit Story' });
    const resultInput = within(editModal).getByLabelText(/Result \(Quantified Impact\)/i);
    await user.clear(resultInput);
    await user.type(
      resultInput,
      'Reduced p99 write latency by 62% and handled 4x peak load.',
    );
    await user.click(within(editModal).getByRole('button', { name: 'Save Story' }));

    expect(
      await within(storyBankSection).findByText(
        /Reduced p99 write latency by 62% and handled 4x peak load\./i,
      ),
    ).toBeInTheDocument();

    // Filter by Competency Tag
    await user.selectOptions(
      within(storyBankSection).getByLabelText('Filter by Competency Tag'),
      'Conflict',
    );
    expect(
      await within(storyBankSection).findByRole('heading', {
        name: 'Unblocking Cross-Team Schema Deadlock',
      }),
    ).toBeInTheDocument();
    expect(
      within(storyBankSection).queryByRole('heading', {
        name: 'Zero-Downtime Ledger Sharding',
      }),
    ).not.toBeInTheDocument();

    // Reset Competency Tag filter and filter by Skill Tag
    await user.selectOptions(
      within(storyBankSection).getByLabelText('Filter by Competency Tag'),
      '',
    );
    await user.selectOptions(
      within(storyBankSection).getByLabelText('Filter by Skill Tag'),
      'PostgreSQL',
    );
    expect(
      await within(storyBankSection).findByRole('heading', {
        name: 'Zero-Downtime Ledger Sharding',
      }),
    ).toBeInTheDocument();
    expect(
      within(storyBankSection).queryByRole('heading', {
        name: 'Unblocking Cross-Team Schema Deadlock',
      }),
    ).not.toBeInTheDocument();

    // Reset Skill Tag filter and search by keyword
    await user.selectOptions(within(storyBankSection).getByLabelText('Filter by Skill Tag'), '');
    const searchInput = within(storyBankSection).getByLabelText('Search Stories');
    await user.type(searchInput, 'shadow verifier');
    expect(
      await within(storyBankSection).findByRole('heading', {
        name: 'Zero-Downtime Ledger Sharding',
      }),
    ).toBeInTheDocument();
    expect(
      within(storyBankSection).queryByRole('heading', {
        name: 'Unblocking Cross-Team Schema Deadlock',
      }),
    ).not.toBeInTheDocument();

    // Clear search and delete a Story
    await user.clear(searchInput);
    const conflictCard = (
      await within(storyBankSection).findByRole('heading', {
        name: 'Unblocking Cross-Team Schema Deadlock',
      })
    ).closest('article')!;
    await user.click(within(conflictCard).getByRole('button', { name: 'Delete Story' }));

    expect(
      within(storyBankSection).queryByRole('heading', {
        name: 'Unblocking Cross-Team Schema Deadlock',
      }),
    ).not.toBeInTheDocument();
    expect(
      within(storyBankSection).getByRole('heading', {
        name: 'Zero-Downtime Ledger Sharding',
      }),
    ).toBeInTheDocument();
  });
});
