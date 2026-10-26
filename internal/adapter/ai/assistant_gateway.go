package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"job-search-cockpit/internal/domain"
)

type HTTPAssistantGateway struct {
	client             *http.Client
	openRouterBaseURL  string
}

func NewAssistantGateway(client *http.Client) *HTTPAssistantGateway {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &HTTPAssistantGateway{
		client:            client,
		openRouterBaseURL: "https://openrouter.ai/api/v1",
	}
}

func (g *HTTPAssistantGateway) ListOpenRouterModels(ctx context.Context, settings domain.AISettings) ([]domain.OpenRouterModel, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/")
	if baseURL == "" || settings.Provider != domain.AIProviderOpenRouter {
		baseURL = g.openRouterBaseURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return defaultOpenRouterCatalog(), nil
	}
	if strings.TrimSpace(settings.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(settings.APIKey))
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return defaultOpenRouterCatalog(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return defaultOpenRouterCatalog(), nil
	}

	var payload struct {
		Data []struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			ContextLength int    `json:"context_length"`
			Pricing       struct {
				Prompt     string `json:"prompt"`
				Completion string `json:"completion"`
			} `json:"pricing"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || len(payload.Data) == 0 {
		return defaultOpenRouterCatalog(), nil
	}

	models := make([]domain.OpenRouterModel, 0, len(payload.Data))
	for _, item := range payload.Data {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = item.ID
		}
		models = append(models, domain.OpenRouterModel{
			ID:                strings.TrimSpace(item.ID),
			Name:              name,
			Description:       strings.TrimSpace(item.Description),
			ContextLength:     item.ContextLength,
			PromptPricing:     strings.TrimSpace(item.Pricing.Prompt),
			CompletionPricing: strings.TrimSpace(item.Pricing.Completion),
		})
	}
	return models, nil
}

func (g *HTTPAssistantGateway) ExtractJobPosting(ctx context.Context, settings domain.AISettings, in domain.AssistantExtractInput) (domain.ExtractedJobPostingDraft, error) {
	if !settings.IsConfigured() {
		return domain.ExtractedJobPostingDraft{}, fmt.Errorf("%w: AI provider is not configured", domain.ErrAssistantUnavailable)
	}

	model := strings.TrimSpace(in.Model)
	if model == "" {
		model = settings.ResolveFastExtractionModel("")
	}
	if model == "" {
		model = defaultFastModelForProvider(settings.Provider)
	}

	systemPrompt := `You are a structured job posting extractor for a software engineering Job Search Cockpit.
Return ONLY valid JSON with fields:
{
  "company_name": "string",
  "company_domain": "string",
  "role_title": "string",
  "seniority": "string",
  "location": "string",
  "remote_policy": "Remote | Hybrid | On-site | \"\"",
  "stated_min_comp": 0,
  "stated_max_comp": 0,
  "currency": "USD",
  "tech_stack_tags": ["string"],
  "requirements": ["string"]
}`

	userPrompt := fmt.Sprintf(
		"Source URL: %s\nDeterministic Hints: Company=%q, Role=%q, Location=%q, RemotePolicy=%q, Comp=%d-%d %s\n\nJob Posting Text:\n%s",
		in.SourceURL,
		in.DeterministicHint.CompanyName,
		in.DeterministicHint.RoleTitle,
		in.DeterministicHint.Location,
		in.DeterministicHint.RemotePolicy,
		in.DeterministicHint.StatedMinComp,
		in.DeterministicHint.StatedMaxComp,
		in.DeterministicHint.Currency,
		in.RawText,
	)

	content, usedModel, err := g.completeChatJSON(ctx, settings, model, systemPrompt, userPrompt, 0.1)
	if err != nil {
		return domain.ExtractedJobPostingDraft{}, err
	}

	var parsed struct {
		CompanyName   string   `json:"company_name"`
		CompanyDomain string   `json:"company_domain"`
		RoleTitle     string   `json:"role_title"`
		Seniority     string   `json:"seniority"`
		Location      string   `json:"location"`
		RemotePolicy  string   `json:"remote_policy"`
		StatedMinComp int      `json:"stated_min_comp"`
		StatedMaxComp int      `json:"stated_max_comp"`
		Currency      string   `json:"currency"`
		TechStackTags []string `json:"tech_stack_tags"`
		Requirements  []string `json:"requirements"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return domain.ExtractedJobPostingDraft{}, fmt.Errorf("%w: failed to parse structured llm json: %v", domain.ErrAssistantUnavailable, err)
	}

	return domain.ExtractedJobPostingDraft{
		CompanyName:      strings.TrimSpace(parsed.CompanyName),
		CompanyDomain:    strings.TrimSpace(parsed.CompanyDomain),
		RoleTitle:        strings.TrimSpace(parsed.RoleTitle),
		Seniority:        strings.TrimSpace(parsed.Seniority),
		Location:         strings.TrimSpace(parsed.Location),
		RemotePolicy:     domain.RemotePolicy(strings.TrimSpace(parsed.RemotePolicy)),
		StatedMinComp:    parsed.StatedMinComp,
		StatedMaxComp:    parsed.StatedMaxComp,
		Currency:         strings.TrimSpace(parsed.Currency),
		TechStackTags:    parsed.TechStackTags,
		Requirements:     parsed.Requirements,
		SourceURL:        in.SourceURL,
		RawText:          in.RawText,
		ExtractionSource: "llm",
		ModelUsed:        usedModel,
	}, nil
}

func (g *HTTPAssistantGateway) AssessFit(ctx context.Context, settings domain.AISettings, in domain.AssistantAssessFitInput) (domain.AssistantFitAssessmentDraft, error) {
	if !settings.IsConfigured() {
		return domain.AssistantFitAssessmentDraft{}, fmt.Errorf("%w: AI provider is not configured", domain.ErrAssistantUnavailable)
	}

	model := settings.ResolveReasoningModel(in.ModelOverride)
	if model == "" {
		model = defaultReasoningModelForProvider(settings.Provider)
	}

	systemPrompt := `You are a career coach and staff engineering hiring bar evaluator for a Job Search Cockpit.
Compare the Opportunity and Job Posting against the Candidate Profile and Story Bank.
Return ONLY valid JSON with fields:
{
  "overall_score": 85,
  "overall_summary": "string",
  "matched_strengths": ["string"],
  "gaps_and_risks": ["string"],
  "compensation_alignment": "string",
  "recommended_story_ids": ["story-id"]
}`

	userPrompt := buildFitAssessmentPrompt(in)
	content, usedModel, err := g.completeChatJSON(ctx, settings, model, systemPrompt, userPrompt, 0.2)
	if err != nil {
		return domain.AssistantFitAssessmentDraft{}, err
	}

	var parsed struct {
		OverallScore          int      `json:"overall_score"`
		OverallSummary        string   `json:"overall_summary"`
		MatchedStrengths      []string `json:"matched_strengths"`
		GapsAndRisks          []string `json:"gaps_and_risks"`
		CompensationAlignment string   `json:"compensation_alignment"`
		RecommendedStoryIDs   []string `json:"recommended_story_ids"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return domain.AssistantFitAssessmentDraft{}, fmt.Errorf("%w: failed to parse fit assessment json: %v", domain.ErrAssistantUnavailable, err)
	}

	compAlign := strings.TrimSpace(parsed.CompensationAlignment)
	if compAlign == "" {
		compAlign = domain.EvaluateCompensationAlignment(in.Opportunity.Opportunity.JobPosting, in.Profile)
	}

	score := parsed.OverallScore
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return domain.AssistantFitAssessmentDraft{
		OverallScore:          score,
		OverallSummary:        strings.TrimSpace(parsed.OverallSummary),
		MatchedStrengths:      parsed.MatchedStrengths,
		GapsAndRisks:          parsed.GapsAndRisks,
		CompensationAlignment: compAlign,
		RecommendedStoryIDs:   parsed.RecommendedStoryIDs,
		ModelUsed:             usedModel,
	}, nil
}

func (g *HTTPAssistantGateway) DraftArtifact(ctx context.Context, settings domain.AISettings, in domain.AssistantDraftArtifactInput) (domain.AssistantArtifactDraft, error) {
	if !settings.IsConfigured() {
		return domain.AssistantArtifactDraft{}, fmt.Errorf("%w: AI provider is not configured", domain.ErrAssistantUnavailable)
	}

	model := settings.ResolveReasoningModel(in.ModelOverride)
	if model == "" {
		model = defaultReasoningModelForProvider(settings.Provider)
	}

	systemPrompt := `You are an expert technical resume and application writer for a Job Search Cockpit.
Draft a tailored, ATS-friendly Markdown Application Artifact grounded strictly in the Candidate Profile, Story Bank, and Job Posting.
Return ONLY valid JSON with fields:
{
  "title": "string",
  "content_markdown": "string"
}`

	userPrompt := buildArtifactDraftPrompt(in)
	content, usedModel, err := g.completeChatJSON(ctx, settings, model, systemPrompt, userPrompt, 0.3)
	if err != nil {
		return domain.AssistantArtifactDraft{}, err
	}

	var parsed struct {
		Title           string `json:"title"`
		ContentMarkdown string `json:"content_markdown"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		// Graceful fallback if the LLM returned raw Markdown directly
		parsed.ContentMarkdown = content
	}

	title := strings.TrimSpace(parsed.Title)
	if title == "" {
		title = strings.TrimSpace(in.Title)
	}
	if title == "" {
		title = fmt.Sprintf("Tailored %s — %s", in.Kind, in.Opportunity.Opportunity.RoleTitle)
	}

	return domain.AssistantArtifactDraft{
		Kind:            in.Kind,
		Title:           title,
		ContentMarkdown: strings.TrimSpace(parsed.ContentMarkdown),
		ModelUsed:       usedModel,
	}, nil
}

func (g *HTTPAssistantGateway) completeChatJSON(
	ctx context.Context,
	settings domain.AISettings,
	model string,
	systemPrompt string,
	userPrompt string,
	temperature float64,
) (string, string, error) {
	endpoint := resolveChatCompletionsEndpoint(settings)
	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": temperature,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("encode llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", "", fmt.Errorf("%w: build request: %v", domain.ErrAssistantUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(settings.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(settings.APIKey))
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("%w: llm request failed: %v", domain.ErrAssistantUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", "", fmt.Errorf("%w: status %d: %s", domain.ErrAssistantUnavailable, resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var chatResp struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil || len(chatResp.Choices) == 0 {
		return "", "", fmt.Errorf("%w: invalid llm response payload", domain.ErrAssistantUnavailable)
	}

	usedModel := strings.TrimSpace(chatResp.Model)
	if usedModel == "" {
		usedModel = model
	}

	return cleanJSONFence(chatResp.Choices[0].Message.Content), usedModel, nil
}

func buildFitAssessmentPrompt(in domain.AssistantAssessFitInput) string {
	opp := in.Opportunity.Opportunity
	jp := opp.JobPosting
	var sb strings.Builder
	fmt.Fprintf(&sb, "Company: %s\nRole: %s (%s)\nLocation: %s (%s)\nStated Comp: %d-%d %s\n",
		in.Opportunity.CompanyName, opp.RoleTitle, opp.Seniority, opp.Location, opp.RemotePolicy,
		jp.StatedMinComp, jp.StatedMaxComp, jp.Currency,
	)
	fmt.Fprintf(&sb, "Requirements: %s\nTech Stack: %s\nJob Posting Raw Text:\n%s\n\n",
		strings.Join(jp.Requirements, "; "), strings.Join(jp.TechStackTags, ", "), jp.RawText,
	)
	fmt.Fprintf(&sb, "Candidate Profile: %s (%s)\nSummary: %s\nSkills: %s\nComp Floor: %d %s (Target: %d %s)\n",
		in.Profile.FullName, in.Profile.Headline, in.Profile.Summary,
		strings.Join(in.Profile.Skills, ", "),
		in.Profile.CompensationFloor, in.Profile.Currency, in.Profile.TargetCompensation, in.Profile.Currency,
	)
	for _, wh := range in.Profile.WorkHistory {
		fmt.Fprintf(&sb, "- %s at %s (%s - %s): %s | %s\n",
			wh.Title, wh.Company, wh.StartDate, wh.EndDate, wh.Summary, strings.Join(wh.ImpactBullets, "; "),
		)
	}
	sb.WriteString("\nStory Bank:\n")
	for _, st := range in.Stories {
		fmt.Fprintf(&sb, "- ID=%s | Title=%s | Hook=%s | Result=%s | Skills=%s\n",
			st.ID, st.Title, st.Hook, st.Result, strings.Join(st.SkillTags, ", "),
		)
	}
	return sb.String()
}

func buildArtifactDraftPrompt(in domain.AssistantDraftArtifactInput) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Artifact Kind: %s\nRequested Title: %s\nInstructions: %s\n",
		in.Kind, in.Title, in.Instructions,
	)
	if strings.TrimSpace(in.ExistingContent) != "" {
		fmt.Fprintf(&sb, "Existing Draft to Refine:\n%s\n\n", in.ExistingContent)
	}
	if in.FitAssessment != nil {
		fmt.Fprintf(&sb, "Fit Assessment Summary: %s (Matched Strengths: %s)\n\n",
			in.FitAssessment.OverallSummary, strings.Join(in.FitAssessment.MatchedStrengths, "; "),
		)
	}
	sb.WriteString(buildFitAssessmentPrompt(domain.AssistantAssessFitInput{
		Opportunity: in.Opportunity,
		Profile:     in.Profile,
		Stories:     in.Stories,
	}))
	return sb.String()
}

var defaultReasoningModels = map[domain.AIProvider]string{
	domain.AIProviderOpenAI:    "gpt-4o",
	domain.AIProviderAnthropic: "claude-3-7-sonnet-latest",
	domain.AIProviderGemini:    "gemini-2.5-pro",
	domain.AIProviderOllama:    "llama3.2:latest",
}

var defaultFastModels = map[domain.AIProvider]string{
	domain.AIProviderOpenAI:    "gpt-4o-mini",
	domain.AIProviderAnthropic: "claude-3-5-haiku-latest",
	domain.AIProviderGemini:    "gemini-2.5-flash",
	domain.AIProviderOllama:    "llama3.2:latest",
}

var defaultChatEndpoints = map[domain.AIProvider]string{
	domain.AIProviderOpenAI:    "https://api.openai.com/v1/chat/completions",
	domain.AIProviderAnthropic: "https://api.anthropic.com/v1/chat/completions",
	domain.AIProviderGemini:    "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
	domain.AIProviderOllama:    "http://localhost:11434/v1/chat/completions",
}

func defaultReasoningModelForProvider(p domain.AIProvider) string {
	if m, ok := defaultReasoningModels[p]; ok {
		return m
	}
	return "anthropic/claude-3.7-sonnet"
}

func resolveChatCompletionsEndpoint(settings domain.AISettings) string {
	base := strings.TrimRight(strings.TrimSpace(settings.BaseURL), "/")
	if base != "" {
		if strings.HasSuffix(base, "/chat/completions") {
			return base
		}
		return base + "/chat/completions"
	}
	if ep, ok := defaultChatEndpoints[settings.Provider]; ok {
		return ep
	}
	return "https://openrouter.ai/api/v1/chat/completions"
}

func defaultFastModelForProvider(p domain.AIProvider) string {
	if m, ok := defaultFastModels[p]; ok {
		return m
	}
	return "anthropic/claude-3.5-haiku"
}

func cleanJSONFence(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func defaultOpenRouterCatalog() []domain.OpenRouterModel {
	return []domain.OpenRouterModel{
		{
			ID:                "anthropic/claude-3.5-haiku",
			Name:              "Anthropic: Claude 3.5 Haiku",
			Description:       "Fast, low-latency model ideal for Quick Capture Job Posting extraction",
			ContextLength:     200000,
			PromptPricing:     "$0.80/M",
			CompletionPricing: "$4.00/M",
		},
		{
			ID:                "anthropic/claude-3.7-sonnet",
			Name:              "Anthropic: Claude 3.7 Sonnet",
			Description:       "High-reasoning model ideal for Fit Assessments and Application Artifact tailoring",
			ContextLength:     200000,
			PromptPricing:     "$3.00/M",
			CompletionPricing: "$15.00/M",
		},
		{
			ID:                "openai/gpt-4o-mini",
			Name:              "OpenAI: GPT-4o Mini",
			Description:       "Fast and affordable structured extraction model",
			ContextLength:     128000,
			PromptPricing:     "$0.15/M",
			CompletionPricing: "$0.60/M",
		},
		{
			ID:                "google/gemini-2.5-pro",
			Name:              "Google: Gemini 2.5 Pro",
			Description:       "Long-context reasoning model for deep resume and cover letter tailoring",
			ContextLength:     1000000,
			PromptPricing:     "$1.25/M",
			CompletionPricing: "$10.00/M",
		},
		{
			ID:                "deepseek/deepseek-r1",
			Name:              "DeepSeek: R1",
			Description:       "Open-weights reasoning model for gap analysis and interview prep",
			ContextLength:     128000,
			PromptPricing:     "$0.55/M",
			CompletionPricing: "$2.19/M",
		},
	}
}
