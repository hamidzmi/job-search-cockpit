import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useEffect, useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import {
  ARTIFACT_KINDS,
  type ApplicationArtifact,
  type ArtifactKind,
} from '../../api/types';

interface OpportunityFitAndArtifactsProps {
  opportunityId: string;
}

const ARTIFACT_KIND_LABELS: Record<ArtifactKind, string> = {
  resume: 'Resume',
  cover_letter: 'Cover Letter',
  screening_qa: 'Screening Q&A',
  other: 'Other',
};

function renderMarkdownBlocks(markdown: string): React.ReactNode[] {
  const lines = markdown.split('\n');
  const blocks: React.ReactNode[] = [];
  let listItems: string[] = [];
  let keyCounter = 0;

  const flushList = () => {
    if (listItems.length > 0) {
      const items = [...listItems];
      listItems = [];
      blocks.push(
        <ul key={`list-${keyCounter++}`} className="ats-list">
          {items.map((item, idx) => (
            <li key={idx}>{item}</li>
          ))}
        </ul>,
      );
    }
  };

  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (!line) {
      flushList();
      continue;
    }
    if (line.startsWith('### ')) {
      flushList();
      blocks.push(<h3 key={`h3-${keyCounter++}`}>{line.slice(4)}</h3>);
    } else if (line.startsWith('## ')) {
      flushList();
      blocks.push(<h2 key={`h2-${keyCounter++}`}>{line.slice(3)}</h2>);
    } else if (line.startsWith('# ')) {
      flushList();
      blocks.push(<h1 key={`h1-${keyCounter++}`}>{line.slice(2)}</h1>);
    } else if (line.startsWith('- ') || line.startsWith('* ')) {
      listItems.push(line.slice(2));
    } else {
      flushList();
      blocks.push(<p key={`p-${keyCounter++}`}>{line}</p>);
    }
  }
  flushList();
  return blocks;
}

export function OpportunityFitAndArtifacts({
  opportunityId,
}: OpportunityFitAndArtifactsProps) {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const [fitModelOverride, setFitModelOverride] = useState('');
  const [fitScore, setFitScore] = useState(0);
  const [fitSummary, setFitSummary] = useState('');
  const [matchedStrengthsText, setMatchedStrengthsText] = useState('');
  const [gapsAndRisksText, setGapsAndRisksText] = useState('');
  const [compensationAlignment, setCompensationAlignment] = useState('');
  const [recommendedStoryIds, setRecommendedStoryIds] = useState<string[]>([]);
  const [fitModelUsed, setFitModelUsed] = useState('');
  const [fitStatus, setFitStatus] = useState<string | null>(null);
  const [fitError, setFitError] = useState<string | null>(null);

  const [editingArtifactId, setEditingArtifactId] = useState<string | null>(null);
  const [artKind, setArtKind] = useState<ArtifactKind>('resume');
  const [artTitle, setArtTitle] = useState('');
  const [artInstructions, setArtInstructions] = useState('');
  const [artModelOverride, setArtModelOverride] = useState('');
  const [artContentMarkdown, setArtContentMarkdown] = useState('');
  const [artIsSubmitted, setArtIsSubmitted] = useState(false);
  const [artModelUsed, setArtModelUsed] = useState('');
  const [artStatus, setArtStatus] = useState<string | null>(null);
  const [artError, setArtError] = useState<string | null>(null);

  const [previewArtifact, setPreviewArtifact] =
    useState<ApplicationArtifact | null>(null);

  const aiSettingsQuery = useQuery({
    queryKey: ['ai-settings'],
    queryFn: () => api.aiSettings.getSettings(),
  });

  const storiesQuery = useQuery({
    queryKey: ['stories', {}],
    queryFn: () => api.stories.listStories(),
  });

  const fitQuery = useQuery({
    queryKey: ['fit-assessment', opportunityId],
    queryFn: () => api.tailoring.getFitAssessment(opportunityId),
    enabled: Boolean(opportunityId),
  });

  const artifactsQuery = useQuery({
    queryKey: ['application-artifacts', opportunityId],
    queryFn: () => api.tailoring.listArtifacts(opportunityId),
    enabled: Boolean(opportunityId),
  });

  useEffect(() => {
    const fa = fitQuery.data;
    if (!fa) return;
    setFitScore(fa.overall_score);
    setFitSummary(fa.overall_summary);
    setMatchedStrengthsText((fa.matched_strengths ?? []).join('\n'));
    setGapsAndRisksText((fa.gaps_and_risks ?? []).join('\n'));
    setCompensationAlignment(fa.compensation_alignment);
    setRecommendedStoryIds(fa.recommended_story_ids ?? []);
    setFitModelUsed(fa.model_used ?? '');
  }, [fitQuery.data]);

  const generateFitMutation = useMutation({
    mutationFn: () =>
      api.tailoring.generateFitAssessment(
        opportunityId,
        fitModelOverride || undefined,
      ),
    onSuccess: async (generated) => {
      setFitScore(generated.overall_score);
      setFitSummary(generated.overall_summary);
      setMatchedStrengthsText((generated.matched_strengths ?? []).join('\n'));
      setGapsAndRisksText((generated.gaps_and_risks ?? []).join('\n'));
      setCompensationAlignment(generated.compensation_alignment);
      setRecommendedStoryIds(generated.recommended_story_ids ?? []);
      setFitModelUsed(generated.model_used ?? '');
      setFitStatus('Fit Assessment generated.');
      setFitError(null);
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['fit-assessment', opportunityId],
        }),
        queryClient.invalidateQueries({
          queryKey: ['opportunity', opportunityId],
        }),
      ]);
    },
    onError: (err) => {
      setFitError(
        err instanceof Error ? err.message : 'Failed to generate Fit Assessment',
      );
    },
  });

  const saveFitMutation = useMutation({
    mutationFn: () =>
      api.tailoring.saveFitAssessment(opportunityId, {
        overall_score: Number(fitScore) || 0,
        overall_summary: fitSummary,
        matched_strengths: matchedStrengthsText
          .split('\n')
          .map((s) => s.trim())
          .filter(Boolean),
        gaps_and_risks: gapsAndRisksText
          .split('\n')
          .map((s) => s.trim())
          .filter(Boolean),
        compensation_alignment: compensationAlignment,
        recommended_story_ids: recommendedStoryIds,
        model_used: fitModelUsed || 'manual',
      }),
    onSuccess: async () => {
      setFitStatus('Fit Assessment saved.');
      setFitError(null);
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['fit-assessment', opportunityId],
        }),
        queryClient.invalidateQueries({
          queryKey: ['opportunity', opportunityId],
        }),
      ]);
    },
    onError: (err) => {
      setFitError(
        err instanceof Error ? err.message : 'Failed to save Fit Assessment',
      );
    },
  });

  const draftArtifactMutation = useMutation({
    mutationFn: () =>
      api.tailoring.draftArtifact(opportunityId, {
        kind: artKind,
        title: artTitle || undefined,
        existing_content: artContentMarkdown || undefined,
        instructions: artInstructions || undefined,
        model_override: artModelOverride || undefined,
      }),
    onSuccess: (draft) => {
      if (draft.title) {
        setArtTitle(draft.title);
      }
      setArtContentMarkdown(draft.content_markdown);
      setArtModelUsed(draft.model_used);
      setArtStatus('AI draft ready in editor.');
      setArtError(null);
    },
    onError: (err) => {
      setArtError(
        err instanceof Error ? err.message : 'Failed to draft Application Artifact',
      );
    },
  });

  const saveArtifactMutation = useMutation({
    mutationFn: () => {
      const payload = {
        kind: artKind,
        title: artTitle.trim() || `${ARTIFACT_KIND_LABELS[artKind]} Draft`,
        content_markdown: artContentMarkdown,
        is_submitted: artIsSubmitted,
        model_used: artModelUsed || undefined,
      };
      if (editingArtifactId) {
        return api.tailoring.updateArtifact(
          opportunityId,
          editingArtifactId,
          payload,
        );
      }
      return api.tailoring.createArtifact(opportunityId, payload);
    },
    onSuccess: async () => {
      setEditingArtifactId(null);
      setArtTitle('');
      setArtInstructions('');
      setArtContentMarkdown('');
      setArtIsSubmitted(false);
      setArtModelUsed('');
      setArtStatus('Application Artifact saved.');
      setArtError(null);
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['application-artifacts', opportunityId],
        }),
        queryClient.invalidateQueries({
          queryKey: ['opportunity', opportunityId],
        }),
      ]);
    },
    onError: (err) => {
      setArtError(
        err instanceof Error ? err.message : 'Failed to save Application Artifact',
      );
    },
  });

  const toggleSubmittedMutation = useMutation({
    mutationFn: (art: ApplicationArtifact) =>
      api.tailoring.updateArtifact(opportunityId, art.id, {
        kind: art.kind,
        title: art.title,
        content_markdown: art.content_markdown,
        is_submitted: !art.is_submitted,
        model_used: art.model_used,
      }),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['application-artifacts', opportunityId],
        }),
        queryClient.invalidateQueries({
          queryKey: ['opportunity', opportunityId],
        }),
      ]);
    },
  });

  const uploadPdfMutation = useMutation({
    mutationFn: ({ artifactId, file }: { artifactId: string; file: File }) =>
      api.tailoring.uploadArtifactPDF(opportunityId, artifactId, file),
    onSuccess: async () => {
      setArtStatus('External PDF uploaded and attached.');
      setArtError(null);
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['application-artifacts', opportunityId],
        }),
        queryClient.invalidateQueries({
          queryKey: ['opportunity', opportunityId],
        }),
      ]);
    },
    onError: (err) => {
      setArtError(err instanceof Error ? err.message : 'Failed to upload PDF');
    },
  });

  const aiSettings = aiSettingsQuery.data;
  const availableModels = Array.from(
    new Set(
      [
        aiSettings?.reasoning_tailoring_model,
        ...(aiSettings?.bookmarked_models ?? []),
      ].filter((m): m is string => Boolean(m && m.trim())),
    ),
  );

  const stories = storiesQuery.data ?? [];
  const recommendedStories = fitQuery.data?.recommended_stories ?? [];
  const artifacts = artifactsQuery.data ?? [];

  const toggleStoryRecommendation = (storyId: string) => {
    setRecommendedStoryIds((prev) =>
      prev.includes(storyId)
        ? prev.filter((id) => id !== storyId)
        : [...prev, storyId],
    );
  };

  return (
    <>
      <section
        className="panel-card panel-card-full"
        aria-label="Opportunity Fit Assessment"
      >
        <div className="timeline-header">
          <div>
            <h3>Fit Assessment</h3>
            <p className="panel-hint">
              Evaluate this Job Posting against your Candidate Profile and Story Bank without advancing Last Activity.
            </p>
          </div>
          <div className="workspace-actions">
            <label className="form-field" style={{ minWidth: '220px' }}>
              <span>Reasoning Model</span>
              <select
                aria-label="Fit Assessment Model"
                value={fitModelOverride}
                onChange={(e) => setFitModelOverride(e.target.value)}
              >
                <option value="">
                  Default ({aiSettings?.reasoning_tailoring_model || 'Reasoning Model'})
                </option>
                {availableModels.map((m) => (
                  <option key={m} value={m}>
                    {m}
                  </option>
                ))}
              </select>
            </label>
            <button
              type="button"
              className="btn-secondary"
              disabled={generateFitMutation.isPending}
              onClick={() => generateFitMutation.mutate()}
            >
              {generateFitMutation.isPending
                ? 'Analyzing Fit...'
                : 'Generate Fit Assessment with AI'}
            </button>
          </div>
        </div>

        {fitError && <div className="alert-error">{fitError}</div>}
        {fitStatus && <div className="alert-success">{fitStatus}</div>}

        <form
          onSubmit={async (e) => {
            e.preventDefault();
            await saveFitMutation.mutateAsync();
          }}
          className="form-grid"
        >
          <label className="form-field">
            <span>Overall Fit Score (0–100)</span>
            <input
              type="number"
              aria-label="Overall Fit Score"
              min={0}
              max={100}
              value={fitScore}
              onChange={(e) => setFitScore(Number(e.target.value))}
            />
          </label>
          <label className="form-field">
            <span>Compensation Alignment vs. Floor</span>
            <input
              type="text"
              aria-label="Compensation Alignment"
              value={compensationAlignment}
              onChange={(e) => setCompensationAlignment(e.target.value)}
              placeholder="Evaluated against Candidate Profile floor..."
            />
          </label>
          <label className="form-field form-field-full">
            <span>Overall Fit Summary</span>
            <textarea
              aria-label="Overall Fit Summary"
              rows={2}
              value={fitSummary}
              onChange={(e) => setFitSummary(e.target.value)}
              placeholder="Summary of role alignment, seniority fit, and strategic positioning..."
            />
          </label>
          <label className="form-field">
            <span>Matched Strengths (one per line)</span>
            <textarea
              aria-label="Matched Strengths"
              rows={3}
              value={matchedStrengthsText}
              onChange={(e) => setMatchedStrengthsText(e.target.value)}
            />
          </label>
          <label className="form-field">
            <span>Requirement Gaps &amp; Risks (one per line)</span>
            <textarea
              aria-label="Requirement Gaps and Risks"
              rows={3}
              value={gapsAndRisksText}
              onChange={(e) => setGapsAndRisksText(e.target.value)}
            />
          </label>

          <div className="form-field-full">
            <span>Recommended Stories to Emphasize</span>
            {stories.length === 0 ? (
              <p className="timestamp-text">
                No Stories in Story Bank yet. Add STAR Stories in Candidate Profile &amp; Story Bank.
              </p>
            ) : (
              <div className="tag-list" style={{ marginTop: '0.35rem' }}>
                {stories.map((st) => (
                  <label key={st.id} className="tech-tag" style={{ cursor: 'pointer' }}>
                    <input
                      type="checkbox"
                      checked={recommendedStoryIds.includes(st.id)}
                      onChange={() => toggleStoryRecommendation(st.id)}
                    />{' '}
                    {st.title}
                  </label>
                ))}
              </div>
            )}

            {recommendedStories.length > 0 && (
              <ul className="timeline-list" style={{ marginTop: '0.65rem' }}>
                {recommendedStories.map((st) => (
                  <li key={st.id} className="timeline-item">
                    <strong>{st.title}</strong>
                    {st.hook && <p className="timeline-note">{st.hook}</p>}
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div className="form-field-full">
            <button type="submit" className="btn-primary">
              Save Fit Assessment
            </button>
          </div>
        </form>
      </section>

      <section
        className="panel-card panel-card-full"
        aria-label="Opportunity Application Artifacts"
      >
        <h3>Application Artifacts (Tailoring, Print-to-PDF &amp; PDF Upload)</h3>
        {artError && <div className="alert-error">{artError}</div>}
        {artStatus && <div className="alert-success">{artStatus}</div>}

        <form
          onSubmit={async (e) => {
            e.preventDefault();
            await saveArtifactMutation.mutateAsync();
          }}
          className="form-grid"
        >
          <label className="form-field">
            <span>Artifact Kind</span>
            <select
              aria-label="Artifact Kind"
              value={artKind}
              onChange={(e) => setArtKind(e.target.value as ArtifactKind)}
            >
              {ARTIFACT_KINDS.map((k) => (
                <option key={k} value={k}>
                  {ARTIFACT_KIND_LABELS[k]}
                </option>
              ))}
            </select>
          </label>
          <label className="form-field">
            <span>Artifact Title *</span>
            <input
              type="text"
              aria-label="Artifact Title"
              value={artTitle}
              onChange={(e) => setArtTitle(e.target.value)}
              placeholder="e.g., Tailored Staff Backend Resume"
            />
          </label>
          <label className="form-field">
            <span>Tailoring Instructions (for AI Draft)</span>
            <input
              type="text"
              aria-label="Tailoring Instructions"
              value={artInstructions}
              onChange={(e) => setArtInstructions(e.target.value)}
              placeholder="e.g., Emphasize distributed systems and Go latency wins..."
            />
          </label>
          <div className="form-field">
            <span>AI Drafting Model &amp; Action</span>
            <div className="workspace-actions">
              <select
                aria-label="Artifact Drafting Model"
                value={artModelOverride}
                onChange={(e) => setArtModelOverride(e.target.value)}
              >
                <option value="">
                  Default ({aiSettings?.reasoning_tailoring_model || 'Reasoning Model'})
                </option>
                {availableModels.map((m) => (
                  <option key={m} value={m}>
                    {m}
                  </option>
                ))}
              </select>
              <button
                type="button"
                className="btn-secondary"
                disabled={draftArtifactMutation.isPending}
                onClick={() => draftArtifactMutation.mutate()}
              >
                {draftArtifactMutation.isPending ? 'Drafting...' : 'Draft with AI'}
              </button>
            </div>
          </div>

          <label className="form-field form-field-full">
            <span>Markdown Content</span>
            <textarea
              aria-label="Artifact Markdown Content"
              rows={8}
              value={artContentMarkdown}
              onChange={(e) => setArtContentMarkdown(e.target.value)}
              placeholder="# Candidate Name&#10;&#10;Write or AI-draft tailored Markdown content here..."
            />
          </label>

          <label className="form-field form-field-full">
            <span>
              <input
                type="checkbox"
                aria-label="Mark Artifact as Submitted"
                checked={artIsSubmitted}
                onChange={(e) => setArtIsSubmitted(e.target.checked)}
              />{' '}
              Submitted with application
            </span>
          </label>

          <div className="form-field-full workspace-actions">
            <button type="submit" className="btn-primary">
              {editingArtifactId
                ? 'Update Application Artifact'
                : 'Save Application Artifact'}
            </button>
            {editingArtifactId && (
              <button
                type="button"
                className="btn-ghost"
                onClick={() => {
                  setEditingArtifactId(null);
                  setArtTitle('');
                  setArtContentMarkdown('');
                  setArtIsSubmitted(false);
                }}
              >
                Cancel Edit
              </button>
            )}
          </div>
        </form>

        <h4 style={{ marginTop: '1rem', marginBottom: '0.5rem' }}>
          Saved Application Artifacts ({artifacts.length})
        </h4>
        {artifacts.length === 0 ? (
          <p className="timestamp-text">
            No Application Artifacts created for this Opportunity yet.
          </p>
        ) : (
          <ul className="timeline-list">
            {artifacts.map((art) => (
              <li key={art.id} className="timeline-item">
                <div className="timeline-header">
                  <div>
                    <strong>{art.title}</strong>{' '}
                    <span className="stage-badge">
                      {ARTIFACT_KIND_LABELS[art.kind] ?? art.kind}
                    </span>
                    {art.is_submitted && (
                      <span
                        className="outcome-pill"
                        data-testid={`artifact-submitted-${art.id}`}
                        style={{ marginLeft: '0.4rem' }}
                      >
                        Submitted
                      </span>
                    )}
                  </div>
                  <div className="workspace-actions">
                    <button
                      type="button"
                      className="btn-secondary"
                      aria-label={`Print Preview ${art.title}`}
                      onClick={() => setPreviewArtifact(art)}
                    >
                      ATS Print Preview / Export PDF
                    </button>
                    <button
                      type="button"
                      className="btn-secondary"
                      aria-label={`Mark ${art.title} as Submitted`}
                      onClick={() => toggleSubmittedMutation.mutate(art)}
                    >
                      {art.is_submitted ? 'Unmark Submitted' : 'Mark as Submitted'}
                    </button>
                    <button
                      type="button"
                      className="btn-ghost"
                      aria-label={`Edit ${art.title}`}
                      onClick={() => {
                        setEditingArtifactId(art.id);
                        setArtKind(art.kind);
                        setArtTitle(art.title);
                        setArtContentMarkdown(art.content_markdown);
                        setArtIsSubmitted(art.is_submitted);
                        setArtModelUsed(art.model_used);
                      }}
                    >
                      Edit
                    </button>
                  </div>
                </div>

                <div className="card-meta" style={{ alignItems: 'center' }}>
                  <label className="btn-ghost" style={{ cursor: 'pointer', padding: '0.2rem 0.5rem' }}>
                    <span>Attach External PDF</span>
                    <input
                      type="file"
                      accept="application/pdf,.pdf"
                      aria-label={`Upload PDF for ${art.title}`}
                      style={{ display: 'none' }}
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file) {
                          uploadPdfMutation.mutate({ artifactId: art.id, file });
                        }
                      }}
                    />
                  </label>
                  {art.has_attached_file && (
                    <a
                      href={art.download_url}
                      target="_blank"
                      rel="noreferrer"
                      className="table-role-link"
                    >
                      Attached PDF: {art.file_name}
                    </a>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      {previewArtifact && (
        <div
          className="modal-backdrop ats-print-backdrop"
          role="dialog"
          aria-modal="true"
          aria-label="ATS Print-to-PDF Preview"
        >
          <div className="modal-card modal-card-wide ats-print-modal">
            <div className="modal-header ats-no-print">
              <div>
                <h2>ATS Print-to-PDF Preview: {previewArtifact.title}</h2>
                <p className="timestamp-text">
                  Clean single-column layout formatted for browser Print to PDF export.
                </p>
              </div>
              <div className="workspace-actions">
                <button
                  type="button"
                  className="btn-primary"
                  onClick={() => window.print()}
                >
                  Print / Save as PDF
                </button>
                <button
                  type="button"
                  className="btn-secondary"
                  onClick={() => setPreviewArtifact(null)}
                >
                  Close Preview
                </button>
              </div>
            </div>
            <article className="ats-print-sheet">
              {renderMarkdownBlocks(previewArtifact.content_markdown)}
            </article>
          </div>
        </div>
      )}
    </>
  );
}
