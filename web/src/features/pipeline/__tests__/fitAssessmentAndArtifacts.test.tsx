import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CockpitApiProvider } from '../../../api/CockpitApiContext';
import type {
  ApplicationArtifact,
  ApplicationArtifactUpsertInput,
  ArtifactDraftInput,
  FitAssessment,
  FitAssessmentUpsertInput,
  Opportunity,
  Story,
} from '../../../api/types';
import { AppRoutes } from '../../../App';

function setupTailoringWorkspaceTest() {
  const initialLastActivity = '2026-03-10T09:00:00Z';
  const opportunity: Opportunity = {
    id: 'opp-1',
    company_id: 'comp-1',
    company_name: 'Datadog',
    role_title: 'Staff Distributed Systems Engineer',
    seniority: 'Staff',
    location: 'New York / Remote',
    remote_policy: 'Remote',
    priority: 'High',
    stage: 'Applied',
    outcome: 'Active',
    notes: 'Core ingestion team',
    job_posting: {
      source_url: 'https://careers.datadoghq.com/staff-dist-sys',
      raw_text: 'Build high-throughput Go and Kafka ingestion pipelines.',
      requirements: ['8+ years Go', 'Distributed systems at scale'],
      tech_stack_tags: ['Go', 'Kafka', 'Kubernetes'],
      stated_min_comp: 220000,
      stated_max_comp: 260000,
      currency: 'USD',
      has_content: true,
    },
    stage_transitions: [
      {
        id: 'tr-1',
        opportunity_id: 'opp-1',
        from_stage: '',
        to_stage: 'Applied',
        from_outcome: '',
        to_outcome: 'Active',
        note: 'Initial application',
        transitioned_at: initialLastActivity,
      },
    ],
    last_activity_at: initialLastActivity,
    created_at: initialLastActivity,
    updated_at: initialLastActivity,
  };

  const stories: Story[] = [
    {
      id: 'story-1',
      title: 'Zero-Downtime Control Plane Sharding',
      hook: 'Sharded Postgres control plane with zero downtime at 50k TPS',
      situation: 'Single Postgres primary reached write saturation.',
      task: 'Shard horizontally with zero customer impact.',
      action: 'Built dual-write verifier and consistent hash router in Go.',
      result: 'Zero-downtime cutover and 4x write capacity.',
      competency_tags: ['Architecture', 'Technical Leadership'],
      skill_tags: ['Go', 'PostgreSQL'],
      created_at: initialLastActivity,
      updated_at: initialLastActivity,
    },
  ];

  let fitAssessment: FitAssessment = {
    opportunity_id: 'opp-1',
    overall_score: 0,
    overall_summary: '',
    matched_strengths: [],
    gaps_and_risks: [],
    compensation_alignment: 'Meets floor: stated 220000-260000 USD meets or exceeds candidate floor of 200000 USD.',
    recommended_story_ids: [],
    recommended_stories: [],
    model_used: '',
    updated_at: '',
  };

  const artifacts: ApplicationArtifact[] = [];
  let artCounter = 1;

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

    if (url.pathname === '/api/opportunities/opp-1' && method === 'GET') {
      return jsonResponse(opportunity);
    }
    if (url.pathname === '/api/contacts' && method === 'GET') {
      return jsonResponse([]);
    }
    if (url.pathname === '/api/opportunities/opp-1/contacts' && method === 'GET') {
      return jsonResponse([]);
    }
    if (url.pathname === '/api/opportunities/opp-1/rounds' && method === 'GET') {
      return jsonResponse([]);
    }
    if (url.pathname === '/api/interactions' && method === 'GET') {
      return jsonResponse([]);
    }
    if (url.pathname === '/api/action-items' && method === 'GET') {
      return jsonResponse([]);
    }
    if (url.pathname === '/api/stories' && method === 'GET') {
      return jsonResponse(stories);
    }
    if (url.pathname === '/api/settings/ai' && method === 'GET') {
      return jsonResponse({
        provider: 'openrouter',
        base_url: 'https://openrouter.ai/api/v1',
        api_key: 'sk-or-test',
        has_api_key: true,
        is_configured: true,
        fast_extraction_model: 'anthropic/claude-3.5-haiku',
        reasoning_tailoring_model: 'anthropic/claude-3.7-sonnet',
        bookmarked_models: ['anthropic/claude-3.7-sonnet', 'google/gemini-2.5-pro'],
        updated_at: initialLastActivity,
      });
    }

    if (url.pathname === '/api/opportunities/opp-1/fit-assessment' && method === 'GET') {
      return jsonResponse(fitAssessment);
    }

    if (url.pathname === '/api/opportunities/opp-1/fit-assessment/generate' && method === 'POST') {
      const payload = bodyText ? (JSON.parse(bodyText) as { model_override?: string }) : {};
      fitAssessment = {
        opportunity_id: 'opp-1',
        overall_score: 89,
        overall_summary: 'Strong staff-level distributed systems fit for Datadog ingestion.',
        matched_strengths: ['8+ years Go microservices', 'High-throughput sharding leadership'],
        gaps_and_risks: ['Limited production Kafka Streams experience'],
        compensation_alignment: 'Meets floor: $220k-$260k exceeds $200k USD floor.',
        recommended_story_ids: ['story-1'],
        recommended_stories: [
          {
            id: stories[0].id,
            title: stories[0].title,
            hook: stories[0].hook,
            competency_tags: stories[0].competency_tags,
            skill_tags: stories[0].skill_tags,
          },
        ],
        model_used: payload.model_override || 'anthropic/claude-3.7-sonnet',
        updated_at: '2026-03-10T11:00:00Z',
      };
      opportunity.updated_at = '2026-03-10T11:00:00Z';
      return jsonResponse(fitAssessment);
    }

    if (url.pathname === '/api/opportunities/opp-1/fit-assessment' && method === 'PUT') {
      const payload = JSON.parse(bodyText) as FitAssessmentUpsertInput;
      fitAssessment = {
        opportunity_id: 'opp-1',
        overall_score: payload.overall_score,
        overall_summary: payload.overall_summary,
        matched_strengths: payload.matched_strengths,
        gaps_and_risks: payload.gaps_and_risks,
        compensation_alignment: payload.compensation_alignment,
        recommended_story_ids: payload.recommended_story_ids,
        recommended_stories: stories
          .filter((st) => payload.recommended_story_ids.includes(st.id))
          .map((st) => ({
            id: st.id,
            title: st.title,
            hook: st.hook,
            competency_tags: st.competency_tags,
            skill_tags: st.skill_tags,
          })),
        model_used: payload.model_used ?? 'manual',
        updated_at: '2026-03-10T12:00:00Z',
      };
      opportunity.updated_at = '2026-03-10T12:00:00Z';
      return jsonResponse(fitAssessment);
    }

    if (url.pathname === '/api/opportunities/opp-1/artifacts' && method === 'GET') {
      return jsonResponse([...artifacts]);
    }

    if (url.pathname === '/api/opportunities/opp-1/artifacts/draft' && method === 'POST') {
      const payload = JSON.parse(bodyText) as ArtifactDraftInput;
      return jsonResponse({
        kind: payload.kind,
        title: payload.title || 'Datadog Staff Distributed Systems Resume',
        content_markdown:
          '# Alex Rivera\n\n## Summary\nStaff Distributed Systems Engineer tailored for Datadog.\n\n## Highlights\n- Sharded Postgres control plane with zero downtime at 50k TPS',
        model_used: payload.model_override || 'anthropic/claude-3.7-sonnet',
      });
    }

    if (url.pathname === '/api/opportunities/opp-1/artifacts' && method === 'POST') {
      const payload = JSON.parse(bodyText) as ApplicationArtifactUpsertInput;
      const created: ApplicationArtifact = {
        id: `art-${artCounter++}`,
        opportunity_id: 'opp-1',
        kind: payload.kind,
        title: payload.title,
        content_markdown: payload.content_markdown,
        has_attached_file: false,
        file_name: '',
        mime_type: '',
        file_size_bytes: 0,
        download_url: '',
        is_submitted: payload.is_submitted,
        model_used: payload.model_used ?? '',
        created_at: '2026-03-10T13:00:00Z',
        updated_at: '2026-03-10T13:00:00Z',
      };
      artifacts.push(created);
      opportunity.updated_at = '2026-03-10T13:00:00Z';
      return jsonResponse(created, 201);
    }

    const artUpdateMatch = url.pathname.match(
      /^\/api\/opportunities\/opp-1\/artifacts\/([^/]+)$/,
    );
    if (artUpdateMatch && method === 'PUT') {
      const artId = decodeURIComponent(artUpdateMatch[1]);
      const idx = artifacts.findIndex((a) => a.id === artId);
      const payload = JSON.parse(bodyText) as ApplicationArtifactUpsertInput;
      const updated: ApplicationArtifact = {
        ...artifacts[idx],
        kind: payload.kind,
        title: payload.title,
        content_markdown: payload.content_markdown,
        is_submitted: payload.is_submitted,
        updated_at: '2026-03-10T14:00:00Z',
      };
      artifacts[idx] = updated;
      return jsonResponse(updated);
    }

    const artUploadMatch = url.pathname.match(
      /^\/api\/opportunities\/opp-1\/artifacts\/([^/]+)\/file$/,
    );
    if (artUploadMatch && method === 'POST') {
      const artId = decodeURIComponent(artUploadMatch[1]);
      const idx = artifacts.findIndex((a) => a.id === artId);
      const formData = init?.body as FormData;
      const uploaded = formData?.get('file') as File | null;
      const fileName = uploaded?.name || 'compiled-resume.pdf';
      const updated: ApplicationArtifact = {
        ...artifacts[idx],
        has_attached_file: true,
        file_name: fileName,
        mime_type: 'application/pdf',
        file_size_bytes: uploaded?.size || 2048,
        download_url: `/api/opportunities/opp-1/artifacts/${artId}/file`,
        updated_at: '2026-03-10T15:00:00Z',
      };
      artifacts[idx] = updated;
      return jsonResponse(updated);
    }

    return jsonResponse({ error: { code: 'not_found', message: `unhandled ${method} ${url.pathname}` } }, 404);
  });

  vi.stubGlobal('fetch', fetchMock);
  return { fetchMock };
}

function renderOpportunityWorkspace() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <CockpitApiProvider>
        <MemoryRouter initialEntries={['/opportunities/opp-1']}>
          <AppRoutes />
        </MemoryRouter>
      </CockpitApiProvider>
    </QueryClientProvider>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('Opportunity Fit Assessment & Application Artifacts Workspace', () => {
  it('generates and manually edits a Fit Assessment with Story Bank recommendations and compensation alignment', async () => {
    const user = userEvent.setup();
    setupTailoringWorkspaceTest();
    renderOpportunityWorkspace();

    const fitPanel = await screen.findByRole('region', {
      name: 'Opportunity Fit Assessment',
    });

    // Shows initial compensation alignment against floor
    expect(
      await within(fitPanel).findByDisplayValue(/Meets floor/i),
    ).toBeInTheDocument();

    // Select bookmarked model and generate Fit Assessment with AI
    const modelSelect = within(fitPanel).getByLabelText('Fit Assessment Model');
    await user.selectOptions(modelSelect, 'google/gemini-2.5-pro');
    await user.click(
      within(fitPanel).getByRole('button', { name: /Generate Fit Assessment/i }),
    );

    expect(
      await within(fitPanel).findByDisplayValue(
        'Strong staff-level distributed systems fit for Datadog ingestion.',
      ),
    ).toBeInTheDocument();
    expect(within(fitPanel).getByLabelText('Overall Fit Score')).toHaveValue(89);
    expect(
      within(fitPanel).getByRole('checkbox', {
        name: 'Zero-Downtime Control Plane Sharding',
      }),
    ).toBeChecked();
    expect(
      within(fitPanel).getByText(
        'Sharded Postgres control plane with zero downtime at 50k TPS',
      ),
    ).toBeInTheDocument();

    // Manually refine the Fit Assessment and save
    const summaryInput = within(fitPanel).getByLabelText('Overall Fit Summary');
    await user.clear(summaryInput);
    await user.type(
      summaryInput,
      'Updated after recruiter call: strong fit for core ingestion team.',
    );
    await user.click(
      within(fitPanel).getByRole('button', { name: /Save Fit Assessment/i }),
    );

    expect(
      await within(fitPanel).findByText(/Fit Assessment saved/i),
    ).toBeInTheDocument();
  });

  it('drafts a Markdown Application Artifact with AI, opens ATS print-to-PDF preview, uploads external PDF, and marks as submitted', async () => {
    const user = userEvent.setup();
    setupTailoringWorkspaceTest();
    renderOpportunityWorkspace();

    const artifactsPanel = await screen.findByRole('region', {
      name: 'Opportunity Application Artifacts',
    });

    // Draft a tailored resume with AI
    await user.selectOptions(
      within(artifactsPanel).getByLabelText('Artifact Kind'),
      'resume',
    );
    await user.type(
      within(artifactsPanel).getByLabelText('Tailoring Instructions'),
      'Highlight zero-downtime Postgres sharding.',
    );
    await user.click(
      within(artifactsPanel).getByRole('button', { name: /Draft with AI/i }),
    );

    const markdownEditor = within(artifactsPanel).getByLabelText(
      'Artifact Markdown Content',
    ) as HTMLTextAreaElement;
    expect(markdownEditor.value).toContain(
      'Staff Distributed Systems Engineer tailored for Datadog.',
    );

    // Save the drafted Markdown Application Artifact
    await user.click(
      within(artifactsPanel).getByRole('button', { name: /Save Application Artifact/i }),
    );

    expect(
      await within(artifactsPanel).findByText(
        'Datadog Staff Distributed Systems Resume',
      ),
    ).toBeInTheDocument();

    // Open clean ATS Print-to-PDF preview mode
    const printSpy = vi.spyOn(window, 'print').mockImplementation(() => {});
    await user.click(
      within(artifactsPanel).getByRole('button', {
        name: 'Print Preview Datadog Staff Distributed Systems Resume',
      }),
    );

    const previewDialog = await screen.findByRole('dialog', {
      name: 'ATS Print-to-PDF Preview',
    });
    expect(
      within(previewDialog).getByText(
        /Staff Distributed Systems Engineer tailored for Datadog/i,
      ),
    ).toBeInTheDocument();

    await user.click(
      within(previewDialog).getByRole('button', { name: /Print \/ Save as PDF/i }),
    );
    expect(printSpy).toHaveBeenCalledTimes(1);
    printSpy.mockRestore();

    await user.click(
      within(previewDialog).getByRole('button', { name: /Close Preview/i }),
    );

    // Upload external PDF file to the Application Artifact
    const fileInput = within(artifactsPanel).getByLabelText(
      'Upload PDF for Datadog Staff Distributed Systems Resume',
    );
    const pdfFile = new File(['%PDF-1.7 fake pdf content'], 'alex-rivera-datadog.pdf', {
      type: 'application/pdf',
    });
    await user.upload(fileInput, pdfFile);

    const downloadLink = await within(artifactsPanel).findByRole('link', {
      name: /alex-rivera-datadog\.pdf/i,
    });
    expect(downloadLink).toHaveAttribute(
      'href',
      '/api/opportunities/opp-1/artifacts/art-1/file',
    );

    // Mark the artifact as submitted
    await user.click(
      within(artifactsPanel).getByRole('button', {
        name: 'Mark Datadog Staff Distributed Systems Resume as Submitted',
      }),
    );

    expect(
      await within(artifactsPanel).findByTestId('artifact-submitted-art-1'),
    ).toHaveTextContent('Submitted');
  });
});
