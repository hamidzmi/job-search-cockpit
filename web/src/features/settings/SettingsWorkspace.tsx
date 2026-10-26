import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useEffect, useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import type { AIProvider, AISettingsUpsertInput } from '../../api/types';

const PROVIDER_LABELS: Record<AIProvider, string> = {
  openrouter: 'OpenRouter (Unified Model Catalog)',
  openai: 'OpenAI (Direct)',
  anthropic: 'Anthropic (Direct)',
  gemini: 'Google Gemini (Direct)',
  ollama: 'Ollama / OpenAI-Compatible Local Endpoint',
};

export function SettingsWorkspace() {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const settingsQuery = useQuery({
    queryKey: ['aiSettings'],
    queryFn: () => api.aiSettings.getSettings(),
  });

  const [provider, setProvider] = useState<AIProvider>('openrouter');
  const [baseUrl, setBaseUrl] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [fastExtractionModel, setFastExtractionModel] = useState('');
  const [reasoningTailoringModel, setReasoningTailoringModel] = useState('');
  const [bookmarkedModels, setBookmarkedModels] = useState<string[]>([]);
  const [customSlug, setCustomSlug] = useState('');
  const [catalogSearch, setCatalogSearch] = useState('');
  const [initialized, setInitialized] = useState(false);
  const [saveMessage, setSaveMessage] = useState<string | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);

  useEffect(() => {
    if (settingsQuery.data && !initialized) {
      setProvider(settingsQuery.data.provider);
      setBaseUrl(settingsQuery.data.base_url);
      setApiKey(settingsQuery.data.api_key);
      setFastExtractionModel(settingsQuery.data.fast_extraction_model);
      setReasoningTailoringModel(settingsQuery.data.reasoning_tailoring_model);
      setBookmarkedModels(settingsQuery.data.bookmarked_models);
      setInitialized(true);
    }
  }, [settingsQuery.data, initialized]);

  const catalogQuery = useQuery({
    queryKey: ['openRouterModels', catalogSearch],
    queryFn: () => api.aiSettings.listOpenRouterModels(catalogSearch),
  });

  const saveMutation = useMutation({
    mutationFn: (input: AISettingsUpsertInput) => api.aiSettings.updateSettings(input),
    onSuccess: (updated) => {
      queryClient.setQueryData(['aiSettings'], updated);
      setProvider(updated.provider);
      setBaseUrl(updated.base_url);
      setApiKey(updated.api_key);
      setFastExtractionModel(updated.fast_extraction_model);
      setReasoningTailoringModel(updated.reasoning_tailoring_model);
      setBookmarkedModels(updated.bookmarked_models);
      setSaveMessage('AI settings saved.');
      setSaveError(null);
    },
    onError: (err) => {
      setSaveError(err instanceof Error ? err.message : 'Failed to save AI settings');
      setSaveMessage(null);
    },
  });

  const toggleBookmark = (modelSlug: string) => {
    const trimmed = modelSlug.trim();
    if (!trimmed) return;
    setBookmarkedModels((prev) =>
      prev.includes(trimmed) ? prev.filter((m) => m !== trimmed) : [...prev, trimmed],
    );
  };

  const handleAddCustomBookmark = () => {
    const trimmed = customSlug.trim();
    if (!trimmed) return;
    if (!bookmarkedModels.includes(trimmed)) {
      setBookmarkedModels((prev) => [...prev, trimmed]);
    }
    setCustomSlug('');
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSaveMessage(null);
    setSaveError(null);
    saveMutation.mutate({
      provider,
      base_url: baseUrl,
      api_key: apiKey,
      fast_extraction_model: fastExtractionModel,
      reasoning_tailoring_model: reasoningTailoringModel,
      bookmarked_models: bookmarkedModels,
    });
  };

  const isConfigured =
    settingsQuery.data?.is_configured ||
    Boolean(apiKey.trim()) ||
    (provider === 'ollama' && Boolean(baseUrl.trim()));

  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Settings</h2>
          <p className="workspace-subtitle">
            OpenRouter &amp; AI provider configuration, model slots, and local workspace preferences.
          </p>
        </div>
        <div className="workspace-actions">
          <span className="stage-badge">
            {isConfigured
              ? 'AI Configured'
              : 'Manual Mode Active — All Cockpit features work without an API key'}
          </span>
        </div>
      </header>

      {saveMessage && (
        <div className="alert-success" role="status">
          {saveMessage}
        </div>
      )}
      {saveError && (
        <div className="alert-error" role="alert">
          {saveError}
        </div>
      )}

      <div className="detail-grid">
        <section className="panel-card" aria-label="AI Provider & Model Slots">
          <h3>AI Provider &amp; Default Model Slots</h3>
          <form onSubmit={handleSubmit} className="form-grid">
            <label className="form-field form-field-full">
              <span>AI Provider</span>
              <select
                value={provider}
                onChange={(e) => setProvider(e.target.value as AIProvider)}
              >
                {(Object.keys(PROVIDER_LABELS) as AIProvider[]).map((p) => (
                  <option key={p} value={p}>
                    {PROVIDER_LABELS[p]}
                  </option>
                ))}
              </select>
            </label>

            <label className="form-field">
              <span>API Key</span>
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="sk-or-v1-... (leave blank for manual mode)"
              />
            </label>

            <label className="form-field">
              <span>Base URL / Endpoint</span>
              <input
                type="text"
                value={baseUrl}
                onChange={(e) => setBaseUrl(e.target.value)}
                placeholder={
                  provider === 'ollama'
                    ? 'http://localhost:11434/v1'
                    : 'https://openrouter.ai/api/v1'
                }
              />
            </label>

            <label className="form-field">
              <span>Fast Extraction Model</span>
              <input
                type="text"
                value={fastExtractionModel}
                onChange={(e) => setFastExtractionModel(e.target.value)}
                placeholder="e.g. anthropic/claude-3.5-haiku"
              />
            </label>

            <label className="form-field">
              <span>Reasoning / Tailoring Model</span>
              <input
                type="text"
                value={reasoningTailoringModel}
                onChange={(e) => setReasoningTailoringModel(e.target.value)}
                placeholder="e.g. anthropic/claude-3.7-sonnet"
              />
            </label>

            <div className="form-field form-field-full">
              <span>Bookmarked Models ({bookmarkedModels.length})</span>
              <div className="card-top-row">
                <label className="form-field" style={{ flex: 1 }}>
                  <span className="sr-only">Custom Model Slug</span>
                  <input
                    type="text"
                    aria-label="Custom Model Slug"
                    value={customSlug}
                    onChange={(e) => setCustomSlug(e.target.value)}
                    placeholder="Type any custom model slug (e.g. deepseek/deepseek-r1)..."
                  />
                </label>
                <button
                  type="button"
                  className="btn-secondary"
                  onClick={handleAddCustomBookmark}
                >
                  Add Bookmark
                </button>
              </div>

              {bookmarkedModels.length > 0 && (
                <div className="tag-list" style={{ marginTop: '0.5rem' }}>
                  {bookmarkedModels.map((slug) => (
                    <div key={slug} className="panel-card" style={{ padding: '0.45rem 0.65rem' }}>
                      <div className="card-top-row">
                        <span className="tech-tag">{slug}</span>
                        <div className="workspace-actions">
                          <button
                            type="button"
                            className="btn-ghost"
                            onClick={() => setFastExtractionModel(slug)}
                          >
                            Use for Fast
                          </button>
                          <button
                            type="button"
                            className="btn-ghost"
                            onClick={() => setReasoningTailoringModel(slug)}
                          >
                            Use for Reasoning
                          </button>
                          <button
                            type="button"
                            className="btn-ghost"
                            onClick={() => toggleBookmark(slug)}
                          >
                            Remove
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="modal-actions form-field-full">
              <button
                type="submit"
                className="btn-primary"
                disabled={saveMutation.isPending}
              >
                {saveMutation.isPending ? 'Saving...' : 'Save AI Settings'}
              </button>
            </div>
          </form>
        </section>

        <section className="panel-card" aria-label="OpenRouter Model Catalog">
          <h3>OpenRouter Model Catalog</h3>
          <label className="form-field">
            <span>Search OpenRouter Models</span>
            <input
              type="search"
              value={catalogSearch}
              onChange={(e) => setCatalogSearch(e.target.value)}
              placeholder="Search by model name, slug, or capability..."
            />
          </label>

          <div className="kanban-column-cards" style={{ marginTop: '0.5rem' }}>
            {(catalogQuery.data ?? []).map((model) => {
              const isBookmarked = bookmarkedModels.includes(model.id);
              return (
                <article key={model.id} className="opportunity-card">
                  <div className="card-top-row">
                    <strong>{model.name}</strong>
                    <span className="company-pill">{model.id}</span>
                  </div>
                  {model.description && (
                    <p className="timeline-note">{model.description}</p>
                  )}
                  <div className="card-meta">
                    {model.context_length > 0 && (
                      <span>Context: {model.context_length.toLocaleString()} tokens</span>
                    )}
                    {model.prompt_pricing && <span>Prompt: {model.prompt_pricing}</span>}
                    {model.completion_pricing && (
                      <span>Completion: {model.completion_pricing}</span>
                    )}
                  </div>
                  <div className="card-footer">
                    <button
                      type="button"
                      className="btn-secondary"
                      aria-label={
                        isBookmarked
                          ? `Remove Bookmark ${model.id}`
                          : `Bookmark ${model.id}`
                      }
                      onClick={() => toggleBookmark(model.id)}
                    >
                      {isBookmarked ? 'Bookmarked' : 'Bookmark'}
                    </button>
                    <div className="workspace-actions">
                      <button
                        type="button"
                        className="btn-ghost"
                        aria-label={`Set ${model.id} as Fast Extraction`}
                        onClick={() => {
                          setFastExtractionModel(model.id);
                          if (!bookmarkedModels.includes(model.id)) {
                            setBookmarkedModels((prev) => [...prev, model.id]);
                          }
                        }}
                      >
                        Fast Extraction
                      </button>
                      <button
                        type="button"
                        className="btn-ghost"
                        aria-label={`Set ${model.id} as Reasoning / Tailoring`}
                        onClick={() => {
                          setReasoningTailoringModel(model.id);
                          if (!bookmarkedModels.includes(model.id)) {
                            setBookmarkedModels((prev) => [...prev, model.id]);
                          }
                        }}
                      >
                        Reasoning / Tailoring
                      </button>
                    </div>
                  </div>
                </article>
              );
            })}
          </div>
        </section>
      </div>
    </div>
  );
}
