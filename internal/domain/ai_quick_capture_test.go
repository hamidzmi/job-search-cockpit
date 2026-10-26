package domain_test

import (
	"errors"
	"testing"
	"time"

	"job-search-cockpit/internal/domain"
)

func TestAISettings_ValidationAndModelResolution(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		in                 domain.AISettingsInput
		overrideFast       string
		overrideReasoning  string
		wantErr            error
		wantProvider       domain.AIProvider
		wantConfigured     bool
		wantBookmarks      []string
		wantResolvedFast   string
		wantResolvedReason string
	}{
		{
			name:               "defaults blank provider to openrouter and unconfigured when no api key",
			in:                 domain.AISettingsInput{},
			wantProvider:       domain.AIProviderOpenRouter,
			wantConfigured:     false,
			wantBookmarks:      []string{},
			wantResolvedFast:   "",
			wantResolvedReason: "",
		},
		{
			name: "rejects unsupported AI provider with ErrValidation",
			in: domain.AISettingsInput{
				Provider: domain.AIProvider("unknown-provider"),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "deduplicates bookmarked models and resolves default vs override models",
			in: domain.AISettingsInput{
				Provider:                domain.AIProviderOpenRouter,
				APIKey:                  "  sk-or-v1-secret  ",
				FastExtractionModel:     "anthropic/claude-3.5-haiku",
				ReasoningTailoringModel: "anthropic/claude-3.7-sonnet",
				BookmarkedModels: []string{
					" anthropic/claude-3.5-haiku ",
					"openai/gpt-4o-mini",
					"anthropic/claude-3.5-haiku",
					"",
					"google/gemini-2.5-pro",
				},
			},
			overrideFast:       "openai/gpt-4o-mini",
			overrideReasoning:  "",
			wantProvider:       domain.AIProviderOpenRouter,
			wantConfigured:     true,
			wantBookmarks:      []string{"anthropic/claude-3.5-haiku", "openai/gpt-4o-mini", "google/gemini-2.5-pro"},
			wantResolvedFast:   "openai/gpt-4o-mini",
			wantResolvedReason: "anthropic/claude-3.7-sonnet",
		},
		{
			name: "treats ollama provider with custom base_url as configured without api key",
			in: domain.AISettingsInput{
				Provider:            domain.AIProviderOllama,
				BaseURL:             "http://localhost:11434/v1",
				FastExtractionModel: "llama3.2:latest",
			},
			wantProvider:     domain.AIProviderOllama,
			wantConfigured:   true,
			wantBookmarks:    []string{},
			wantResolvedFast: "llama3.2:latest",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.NewAISettings(tc.in, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Provider != tc.wantProvider {
				t.Errorf("expected provider %q, got %q", tc.wantProvider, got.Provider)
			}
			if got.IsConfigured() != tc.wantConfigured {
				t.Errorf("expected IsConfigured=%v, got %v", tc.wantConfigured, got.IsConfigured())
			}
			if len(got.BookmarkedModels) != len(tc.wantBookmarks) {
				t.Fatalf("expected bookmarks %v, got %v", tc.wantBookmarks, got.BookmarkedModels)
			}
			for i, b := range tc.wantBookmarks {
				if got.BookmarkedModels[i] != b {
					t.Errorf("bookmark[%d]: expected %q, got %q", i, b, got.BookmarkedModels[i])
				}
			}
			if fast := got.ResolveFastExtractionModel(tc.overrideFast); fast != tc.wantResolvedFast {
				t.Errorf("expected resolved fast model %q, got %q", tc.wantResolvedFast, fast)
			}
			if reason := got.ResolveReasoningModel(tc.overrideReasoning); reason != tc.wantResolvedReason {
				t.Errorf("expected resolved reasoning model %q, got %q", tc.wantResolvedReason, reason)
			}
		})
	}
}

func TestFilterOpenRouterModels(t *testing.T) {
	catalog := []domain.OpenRouterModel{
		{
			ID:            "anthropic/claude-3.5-haiku",
			Name:          "Anthropic: Claude 3.5 Haiku",
			Description:   "Fast model for structured extraction",
			ContextLength: 200000,
		},
		{
			ID:            "anthropic/claude-3.7-sonnet",
			Name:          "Anthropic: Claude 3.7 Sonnet",
			Description:   "Hybrid reasoning model for deep tailoring",
			ContextLength: 200000,
		},
		{
			ID:            "openai/gpt-4o-mini",
			Name:          "OpenAI: GPT-4o Mini",
			Description:   "Lightweight multimodal model",
			ContextLength: 128000,
		},
	}

	got := domain.FilterOpenRouterModels(catalog, "haiku")
	if len(got) != 1 || got[0].ID != "anthropic/claude-3.5-haiku" {
		t.Fatalf("expected 1 haiku match, got %+v", got)
	}

	all := domain.FilterOpenRouterModels(catalog, "   ")
	if len(all) != 3 {
		t.Fatalf("expected all 3 models on empty query, got %d", len(all))
	}
}

func TestMergeCaptureDraft_FallbackAndPrecedence(t *testing.T) {
	page := domain.ParsedJobPage{
		SourceURL:     "https://boards.greenhouse.io/stripe/jobs/42",
		CompanyName:   "Stripe",
		CompanyDomain: "stripe.com",
		RoleTitle:     "Staff Infrastructure Engineer",
		Location:      "Seattle, WA",
		RemotePolicy:  domain.RemotePolicyHybrid,
		StatedMinComp: 220000,
		StatedMaxComp: 260000,
		Currency:      "USD",
		TechStackTags: []string{"Go", "Ruby"},
		Requirements:  []string{"8+ years distributed systems"},
		CleanedText:   "Stripe is hiring a Staff Infrastructure Engineer in Seattle.",
	}

	t.Run("falls back to deterministic page metadata when LLM draft is nil", func(t *testing.T) {
		draft := domain.MergeCaptureDraft(page.SourceURL, "", page, nil)
		if draft.CompanyName != "Stripe" || draft.RoleTitle != "Staff Infrastructure Engineer" {
			t.Fatalf("unexpected deterministic draft: %+v", draft)
		}
		if draft.ExtractionSource != "deterministic" {
			t.Errorf("expected ExtractionSource='deterministic', got %q", draft.ExtractionSource)
		}
		if draft.RawText != page.CleanedText {
			t.Errorf("expected RawText to use page CleanedText, got %q", draft.RawText)
		}
	})

	t.Run("merges LLM fields over deterministic base while keeping deterministic fallbacks for blank LLM fields", func(t *testing.T) {
		llm := &domain.ExtractedJobPostingDraft{
			CompanyName:   "Stripe",
			RoleTitle:     "Staff Core Infrastructure Engineer",
			Seniority:     "Staff",
			TechStackTags: []string{"Go", "Ruby", "Kafka"},
			Requirements:  []string{"8+ years distributed systems", "High-reliability storage"},
			ModelUsed:     "anthropic/claude-3.5-haiku",
		}
		draft := domain.MergeCaptureDraft(page.SourceURL, "", page, llm)
		if draft.RoleTitle != "Staff Core Infrastructure Engineer" {
			t.Errorf("expected LLM role title, got %q", draft.RoleTitle)
		}
		if draft.Seniority != "Staff" {
			t.Errorf("expected LLM seniority 'Staff', got %q", draft.Seniority)
		}
		if draft.Location != "Seattle, WA" || draft.StatedMinComp != 220000 {
			t.Errorf("expected deterministic location and comp fallback, got location=%q minComp=%d", draft.Location, draft.StatedMinComp)
		}
		if draft.ExtractionSource != "llm" || draft.ModelUsed != "anthropic/claude-3.5-haiku" {
			t.Errorf("expected llm source and model, got source=%q model=%q", draft.ExtractionSource, draft.ModelUsed)
		}
	})

	t.Run("heuristically extracts fields from pasted raw text when URL fetch failed and no LLM is configured", func(t *testing.T) {
		rawText := `Company: Linear
Role: Senior Product Engineer
Location: Remote (US/EU)
Compensation: $190,000 - $230,000 USD
Requirements:
- 5+ years building TypeScript and React applications
- Experience with GraphQL and sync engines`

		draft := domain.MergeCaptureDraft("https://linkedin.com/jobs/view/999", rawText, domain.ParsedJobPage{}, nil)
		if draft.CompanyName != "Linear" {
			t.Errorf("expected CompanyName 'Linear', got %q", draft.CompanyName)
		}
		if draft.RoleTitle != "Senior Product Engineer" {
			t.Errorf("expected RoleTitle 'Senior Product Engineer', got %q", draft.RoleTitle)
		}
		if draft.Seniority != "Senior" {
			t.Errorf("expected Seniority 'Senior', got %q", draft.Seniority)
		}
		if draft.RemotePolicy != domain.RemotePolicyRemote {
			t.Errorf("expected RemotePolicy 'Remote', got %q", draft.RemotePolicy)
		}
		if draft.StatedMinComp != 190000 || draft.StatedMaxComp != 230000 {
			t.Errorf("expected comp 190000-230000, got %d-%d", draft.StatedMinComp, draft.StatedMaxComp)
		}
		if len(draft.Requirements) != 2 {
			t.Errorf("expected 2 bullet requirements, got %+v", draft.Requirements)
		}
		if len(draft.TechStackTags) == 0 {
			t.Errorf("expected extracted tech stack tags (TypeScript, React, GraphQL), got empty")
		}
	})
}
