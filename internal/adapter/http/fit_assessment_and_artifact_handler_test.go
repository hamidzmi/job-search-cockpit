package httpadapter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	aiadapter "job-search-cockpit/internal/adapter/ai"
	fsadapter "job-search-cockpit/internal/adapter/fs"
	httpadapter "job-search-cockpit/internal/adapter/http"
	"job-search-cockpit/internal/adapter/sqlite"
	"job-search-cockpit/internal/app"
	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

func newTestServerForTailoring(
	t *testing.T,
	clock *manualClock,
	assistant outbound.AssistantGateway,
	blobStore outbound.ArtifactBlobStore,
) http.Handler {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test-cockpit-tailoring.db")
	store, err := sqlite.OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("failed to open and migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	if blobStore == nil {
		blobStore = fsadapter.NewLocalArtifactBlobStore(filepath.Join(t.TempDir(), "artifacts"))
	}

	return app.NewHTTPHandler(store, app.Dependencies{
		Clock:             clock,
		AssistantGateway:  assistant,
		ArtifactBlobStore: blobStore,
	})
}

func TestFitAssessmentAndApplicationArtifactsLifecycle(t *testing.T) {
	initialAt := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	clock := &manualClock{now: initialAt}

	var capturedAssessCalls []domain.AssistantAssessFitInput
	var capturedDraftCalls []domain.AssistantDraftArtifactInput

	fakeAI := &fakeAssistantGateway{
		assessFitFn: func(_ context.Context, settings domain.AISettings, in domain.AssistantAssessFitInput) (domain.AssistantFitAssessmentDraft, error) {
			capturedAssessCalls = append(capturedAssessCalls, in)
			model := settings.ResolveReasoningModel(in.ModelOverride)
			var recStoryIDs []string
			for _, st := range in.Stories {
				recStoryIDs = append(recStoryIDs, st.ID)
			}
			return domain.AssistantFitAssessmentDraft{
				OverallScore:          88,
				OverallSummary:        "Strong alignment with Staff Backend role in Go and distributed systems.",
				MatchedStrengths:      []string{"8+ years Go distributed systems", "Led multi-region migration"},
				GapsAndRisks:          []string{"Limited direct Kafka Streams experience"},
				CompensationAlignment: "Meets floor: stated 220000-260000 USD exceeds floor of 200000 USD.",
				RecommendedStoryIDs:   recStoryIDs,
				ModelUsed:             model,
			}, nil
		},
		draftArtifactFn: func(_ context.Context, settings domain.AISettings, in domain.AssistantDraftArtifactInput) (domain.AssistantArtifactDraft, error) {
			capturedDraftCalls = append(capturedDraftCalls, in)
			model := settings.ResolveReasoningModel(in.ModelOverride)
			return domain.AssistantArtifactDraft{
				Kind:            in.Kind,
				Title:           firstNonEmptyStr(in.Title, "Tailored Resume — "+in.Opportunity.Opportunity.RoleTitle),
				ContentMarkdown: fmt.Sprintf("# %s\n\nTailored for **%s** at **%s**.\n\n- Highlighted Story count: %d", in.Profile.FullName, in.Opportunity.Opportunity.RoleTitle, in.Opportunity.CompanyName, len(in.Stories)),
				ModelUsed:       model,
			}, nil
		},
	}

	blobDir := filepath.Join(t.TempDir(), "artifacts-blobs")
	blobStore := fsadapter.NewLocalArtifactBlobStore(blobDir)
	srv := newTestServerForTailoring(t, clock, fakeAI, blobStore)

	// 1. Seed Candidate Profile, a Story in the Story Bank, AI Settings, Company, and Opportunity
	profileRec := doJSON(t, srv, http.MethodPut, "/api/profile", map[string]any{
		"full_name":            "Alex Rivera",
		"headline":             "Staff Distributed Systems Engineer",
		"summary":              "10 years building low-latency Go services.",
		"skills":               []string{"Go", "PostgreSQL", "Kubernetes"},
		"target_role_titles":   []string{"Staff Backend Engineer"},
		"location_preferences": []string{"Remote"},
		"remote_preference":    "Remote",
		"compensation_floor":   200000,
		"target_compensation":  240000,
		"currency":             "USD",
		"work_history": []map[string]any{
			{
				"company":        "CloudScale",
				"title":          "Staff Engineer",
				"start_date":     "2022-01",
				"end_date":       "Present",
				"summary":        "Tech lead for control plane",
				"impact_bullets": []string{"Cut p99 latency by 45% across 3 regions"},
			},
		},
	})
	if profileRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK saving profile, got %d: %s", profileRec.Code, profileRec.Body.String())
	}

	storyRec := doJSON(t, srv, http.MethodPost, "/api/stories", map[string]any{
		"title":           "Zero-Downtime Control Plane Sharding",
		"hook":            "Sharded control plane DB with zero customer downtime",
		"situation":       "Single Postgres primary hit write saturation at 50k TPS",
		"task":            "Design and execute live horizontal sharding",
		"action":          "Built dual-write verifier and consistent hash router in Go",
		"result":          "Zero downtime cutover and 4x write headroom",
		"competency_tags": []string{"Architecture", "Technical Leadership"},
		"skill_tags":      []string{"Go", "PostgreSQL"},
	})
	if storyRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created creating story, got %d: %s", storyRec.Code, storyRec.Body.String())
	}
	var createdStory httpadapter.StoryDTO
	_ = json.Unmarshal(storyRec.Body.Bytes(), &createdStory)

	aiCfgRec := doJSON(t, srv, http.MethodPut, "/api/settings/ai", map[string]any{
		"provider":                  "openrouter",
		"api_key":                   "sk-or-test-key",
		"fast_extraction_model":     "anthropic/claude-3.5-haiku",
		"reasoning_tailoring_model": "anthropic/claude-3.7-sonnet",
		"bookmarked_models": []string{
			"anthropic/claude-3.7-sonnet",
			"google/gemini-2.5-pro",
		},
	})
	if aiCfgRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK configuring AI settings, got %d", aiCfgRec.Code)
	}

	compRec := doJSON(t, srv, http.MethodPost, "/api/companies", map[string]any{
		"name":          "Datadog",
		"domain":        "datadoghq.com",
		"remote_policy": "Hybrid",
	})
	if compRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created creating company, got %d", compRec.Code)
	}
	var comp httpadapter.CompanyDTO
	_ = json.Unmarshal(compRec.Body.Bytes(), &comp)

	oppRec := doJSON(t, srv, http.MethodPost, "/api/opportunities", map[string]any{
		"company_id":    comp.ID,
		"role_title":    "Staff Distributed Systems Engineer",
		"seniority":     "Staff",
		"location":      "New York / Remote",
		"remote_policy": "Remote",
		"priority":      "High",
		"stage":         "Discovered",
		"job_posting": map[string]any{
			"source_url":      "https://careers.datadoghq.com/staff-dist-sys",
			"raw_text":        "Build high-throughput ingestion pipelines in Go and Kafka.",
			"requirements":    []string{"8+ years Go", "Distributed storage and streaming"},
			"tech_stack_tags": []string{"Go", "Kafka", "Kubernetes"},
			"stated_min_comp": 220000,
			"stated_max_comp": 260000,
			"currency":        "USD",
		},
	})
	if oppRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created creating opportunity, got %d: %s", oppRec.Code, oppRec.Body.String())
	}
	var opp httpadapter.OpportunityDTO
	_ = json.Unmarshal(oppRec.Body.Bytes(), &opp)

	t.Run("GET fit-assessment before creation returns default structure with evaluated compensation alignment", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+opp.ID+"/fit-assessment", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		var dto httpadapter.FitAssessmentDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if dto.OpportunityID != opp.ID {
			t.Fatalf("expected opportunity_id %q, got %q", opp.ID, dto.OpportunityID)
		}
		if !strings.Contains(dto.CompensationAlignment, "Meets floor") {
			t.Fatalf("expected default compensation alignment to evaluate against floor, got %q", dto.CompensationAlignment)
		}
	})

	t.Run("POST fit-assessment/generate uses AssistantGateway Reasoning model or override, resolves recommended Stories, and preserves Opportunity.LastActivityAt", func(t *testing.T) {
		clock.now = initialAt.Add(2 * time.Hour)
		genRec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+opp.ID+"/fit-assessment/generate", map[string]any{
			"model_override": "google/gemini-2.5-pro",
		})
		if genRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK generating fit assessment, got %d: %s", genRec.Code, genRec.Body.String())
		}
		var generated httpadapter.FitAssessmentDTO
		if err := json.Unmarshal(genRec.Body.Bytes(), &generated); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if generated.OverallScore != 88 || generated.ModelUsed != "google/gemini-2.5-pro" {
			t.Fatalf("unexpected generated FitAssessment: %+v", generated)
		}
		if len(generated.RecommendedStories) != 1 || generated.RecommendedStories[0].ID != createdStory.ID || generated.RecommendedStories[0].Title != createdStory.Title {
			t.Fatalf("expected hydrated RecommendedStories from Story Bank, got %+v", generated.RecommendedStories)
		}
		if len(capturedAssessCalls) != 1 || capturedAssessCalls[0].Profile.FullName != "Alex Rivera" || len(capturedAssessCalls[0].Stories) != 1 {
			t.Fatalf("expected AssessFit to receive CandidateProfile and Story Bank, got %+v", capturedAssessCalls)
		}

		// Verify Opportunity.LastActivityAt did NOT advance, while UpdatedAt did advance
		oppAfterGenRec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+opp.ID, nil)
		var oppAfterGen httpadapter.OpportunityDTO
		_ = json.Unmarshal(oppAfterGenRec.Body.Bytes(), &oppAfterGen)
		if oppAfterGen.LastActivityAt != opp.LastActivityAt {
			t.Fatalf("expected LastActivityAt to remain %q after FitAssessment generation, got %q", opp.LastActivityAt, oppAfterGen.LastActivityAt)
		}
		if oppAfterGen.UpdatedAt == opp.UpdatedAt {
			t.Fatalf("expected UpdatedAt to advance after FitAssessment generation")
		}
	})

	t.Run("PUT fit-assessment manually edits and persists FitAssessment without advancing Opportunity.LastActivityAt", func(t *testing.T) {
		clock.now = initialAt.Add(3 * time.Hour)
		putRec := doJSON(t, srv, http.MethodPut, "/api/opportunities/"+opp.ID+"/fit-assessment", map[string]any{
			"overall_score":          92,
			"overall_summary":        "Manually refined fit summary after speaking with hiring manager.",
			"matched_strengths":      []string{"Go high-throughput ingestion", "Zero-downtime sharding"},
			"gaps_and_risks":         []string{"Need to brush up on Kafka internals"},
			"compensation_alignment": "Confirmed $245k base + equity in recruiter screen.",
			"recommended_story_ids":  []string{createdStory.ID},
			"model_used":             "manual",
		})
		if putRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on PUT fit-assessment, got %d: %s", putRec.Code, putRec.Body.String())
		}
		var edited httpadapter.FitAssessmentDTO
		_ = json.Unmarshal(putRec.Body.Bytes(), &edited)
		if edited.OverallScore != 92 || edited.OverallSummary != "Manually refined fit summary after speaking with hiring manager." {
			t.Fatalf("unexpected edited FitAssessment: %+v", edited)
		}

		getRec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+opp.ID+"/fit-assessment", nil)
		var reloaded httpadapter.FitAssessmentDTO
		_ = json.Unmarshal(getRec.Body.Bytes(), &reloaded)
		if reloaded.OverallScore != 92 || len(reloaded.RecommendedStories) != 1 {
			t.Fatalf("expected persisted FitAssessment on GET, got %+v", reloaded)
		}

		oppCheckRec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+opp.ID, nil)
		var oppCheck httpadapter.OpportunityDTO
		_ = json.Unmarshal(oppCheckRec.Body.Bytes(), &oppCheck)
		if oppCheck.LastActivityAt != opp.LastActivityAt {
			t.Fatalf("expected LastActivityAt unchanged after manual FitAssessment edit, got %q", oppCheck.LastActivityAt)
		}
	})

	var resumeArtifactID string

	t.Run("drafts, creates, edits, and marks ApplicationArtifact as submitted without advancing Opportunity.LastActivityAt", func(t *testing.T) {
		clock.now = initialAt.Add(4 * time.Hour)

		// 1. LLM-assisted draft grounded in Candidate Profile, Story Bank, and Job Posting
		draftRec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+opp.ID+"/artifacts/draft", map[string]any{
			"kind":         "resume",
			"title":        "Datadog Staff Distributed Systems Resume",
			"instructions": "Emphasize zero-downtime Postgres sharding and Go p99 latency wins.",
		})
		if draftRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK drafting artifact, got %d: %s", draftRec.Code, draftRec.Body.String())
		}
		var draftDTO httpadapter.ArtifactDraftDTO
		if err := json.Unmarshal(draftRec.Body.Bytes(), &draftDTO); err != nil {
			t.Fatalf("decode draft failed: %v", err)
		}
		if draftDTO.Kind != "resume" || draftDTO.ModelUsed != "anthropic/claude-3.7-sonnet" {
			t.Fatalf("unexpected draft DTO: %+v", draftDTO)
		}
		if !strings.Contains(draftDTO.ContentMarkdown, "Alex Rivera") || !strings.Contains(draftDTO.ContentMarkdown, "Datadog") {
			t.Fatalf("expected draft content grounded in profile and opportunity, got %q", draftDTO.ContentMarkdown)
		}
		if len(capturedDraftCalls) != 1 || capturedDraftCalls[0].FitAssessment == nil {
			t.Fatalf("expected DraftArtifact to receive FitAssessment context, got %+v", capturedDraftCalls)
		}

		// 2. Create Markdown ApplicationArtifact
		createRec := doJSON(t, srv, http.MethodPost, "/api/opportunities/"+opp.ID+"/artifacts", map[string]any{
			"kind":             draftDTO.Kind,
			"title":            draftDTO.Title,
			"content_markdown": draftDTO.ContentMarkdown,
			"is_submitted":     false,
			"model_used":       draftDTO.ModelUsed,
		})
		if createRec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created creating artifact, got %d: %s", createRec.Code, createRec.Body.String())
		}
		var createdArt httpadapter.ApplicationArtifactDTO
		_ = json.Unmarshal(createRec.Body.Bytes(), &createdArt)
		resumeArtifactID = createdArt.ID
		if createdArt.Title != "Datadog Staff Distributed Systems Resume" || createdArt.IsSubmitted {
			t.Fatalf("unexpected created artifact: %+v", createdArt)
		}

		// 3. Edit and mark as submitted (IsSubmitted = true)
		clock.now = initialAt.Add(5 * time.Hour)
		updateRec := doJSON(t, srv, http.MethodPut, "/api/opportunities/"+opp.ID+"/artifacts/"+resumeArtifactID, map[string]any{
			"kind":             "resume",
			"title":            "Datadog Staff Distributed Systems Resume (Final)",
			"content_markdown": draftDTO.ContentMarkdown + "\n\n## Core Skills\n- Go, Kafka, PostgreSQL",
			"is_submitted":     true,
			"model_used":       draftDTO.ModelUsed,
		})
		if updateRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK updating artifact, got %d: %s", updateRec.Code, updateRec.Body.String())
		}
		var updatedArt httpadapter.ApplicationArtifactDTO
		_ = json.Unmarshal(updateRec.Body.Bytes(), &updatedArt)
		if !updatedArt.IsSubmitted || updatedArt.Title != "Datadog Staff Distributed Systems Resume (Final)" {
			t.Fatalf("expected updated artifact with IsSubmitted=true, got %+v", updatedArt)
		}

		// Verify LastActivityAt was NOT advanced
		oppCheckRec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+opp.ID, nil)
		var oppCheck httpadapter.OpportunityDTO
		_ = json.Unmarshal(oppCheckRec.Body.Bytes(), &oppCheck)
		if oppCheck.LastActivityAt != opp.LastActivityAt {
			t.Fatalf("expected LastActivityAt unchanged after ApplicationArtifact create/update, got %q", oppCheck.LastActivityAt)
		}
	})

	t.Run("uploads external PDF via multipart form to ArtifactBlobStore, downloads it, and rejects non-PDF files", func(t *testing.T) {
		clock.now = initialAt.Add(6 * time.Hour)

		pdfBytes := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "alex-rivera-datadog.pdf")
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(pdfBytes); err != nil {
			t.Fatalf("write pdf bytes: %v", err)
		}
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/opportunities/"+opp.ID+"/artifacts/"+resumeArtifactID+"/file", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK uploading PDF, got %d: %s", rec.Code, rec.Body.String())
		}
		var withPDF httpadapter.ApplicationArtifactDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &withPDF); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if !withPDF.HasAttachedFile || withPDF.FileName != "alex-rivera-datadog.pdf" || withPDF.MimeType != "application/pdf" {
			t.Fatalf("expected attached PDF metadata, got %+v", withPDF)
		}
		if withPDF.FileSizeBytes != int64(len(pdfBytes)) {
			t.Fatalf("expected file_size_bytes %d, got %d", len(pdfBytes), withPDF.FileSizeBytes)
		}
		if withPDF.DownloadURL == "" {
			t.Fatalf("expected non-empty download_url on artifact DTO")
		}

		// Download the attached PDF via GET /api/opportunities/{id}/artifacts/{artifactId}/file
		dlReq := httptest.NewRequest(http.MethodGet, withPDF.DownloadURL, nil)
		dlRec := httptest.NewRecorder()
		srv.ServeHTTP(dlRec, dlReq)
		if dlRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK downloading PDF, got %d: %s", dlRec.Code, dlRec.Body.String())
		}
		if dlRec.Header().Get("Content-Type") != "application/pdf" {
			t.Fatalf("expected Content-Type application/pdf, got %q", dlRec.Header().Get("Content-Type"))
		}
		if !bytes.Equal(dlRec.Body.Bytes(), pdfBytes) {
			t.Fatalf("downloaded PDF bytes do not match uploaded bytes")
		}

		// Reject non-PDF upload with 400 Bad Request
		var badBody bytes.Buffer
		badWriter := multipart.NewWriter(&badBody)
		badPart, _ := badWriter.CreateFormFile("file", "notes.txt")
		_, _ = badPart.Write([]byte("not a pdf file"))
		_ = badWriter.Close()

		badReq := httptest.NewRequest(http.MethodPost, "/api/opportunities/"+opp.ID+"/artifacts/"+resumeArtifactID+"/file", &badBody)
		badReq.Header.Set("Content-Type", badWriter.FormDataContentType())
		badRec := httptest.NewRecorder()
		srv.ServeHTTP(badRec, badReq)
		if badRec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request for non-PDF upload, got %d: %s", badRec.Code, badRec.Body.String())
		}

		// Verify LastActivityAt was still NOT advanced by PDF upload
		oppCheckRec := doJSON(t, srv, http.MethodGet, "/api/opportunities/"+opp.ID, nil)
		var oppCheck httpadapter.OpportunityDTO
		_ = json.Unmarshal(oppCheckRec.Body.Bytes(), &oppCheck)
		if oppCheck.LastActivityAt != opp.LastActivityAt {
			t.Fatalf("expected LastActivityAt unchanged after PDF upload, got %q", oppCheck.LastActivityAt)
		}
	})

	t.Run("exercises real HTTPAssistantGateway AssessFit and DraftArtifact against stubbed OpenRouter server", func(t *testing.T) {
		var callCount int
		openRouterSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			callCount++
			if callCount == 1 {
				_, _ = fmt.Fprintf(w, `{"model":"anthropic/claude-3.7-sonnet","choices":[{"message":{"content":%q}}]}`,
					`{"overall_score":91,"overall_summary":"Strong distributed systems fit.","matched_strengths":["Go","Postgres sharding"],"gaps_and_risks":["None major"],"compensation_alignment":"Exceeds $200k floor","recommended_story_ids":["`+createdStory.ID+`"]}`)
				return
			}
			_, _ = fmt.Fprintf(w, `{"model":"anthropic/claude-3.7-sonnet","choices":[{"message":{"content":%q}}]}`,
				`{"title":"Cover Letter for Datadog","content_markdown":"Dear Hiring Manager,\n\nI led zero-downtime Postgres sharding in Go."}`)
		}))
		defer openRouterSrv.Close()

		realAI := aiadapter.NewAssistantGateway(openRouterSrv.Client())
		realSrv := newTestServerForTailoring(t, clock, realAI, nil)

		doJSON(t, realSrv, http.MethodPut, "/api/settings/ai", map[string]any{
			"provider":                  "openrouter",
			"base_url":                  openRouterSrv.URL,
			"api_key":                   "sk-or-live",
			"reasoning_tailoring_model": "anthropic/claude-3.7-sonnet",
		})
		compRes := doJSON(t, realSrv, http.MethodPost, "/api/companies", map[string]any{"name": "Stripe"})
		var stripeComp httpadapter.CompanyDTO
		_ = json.Unmarshal(compRes.Body.Bytes(), &stripeComp)

		oppRes := doJSON(t, realSrv, http.MethodPost, "/api/opportunities", map[string]any{
			"company_id": stripeComp.ID,
			"role_title": "Staff Infrastructure Engineer",
		})
		var stripeOpp httpadapter.OpportunityDTO
		_ = json.Unmarshal(oppRes.Body.Bytes(), &stripeOpp)

		fitRes := doJSON(t, realSrv, http.MethodPost, "/api/opportunities/"+stripeOpp.ID+"/fit-assessment/generate", map[string]any{})
		if fitRes.Code != http.StatusOK {
			t.Fatalf("expected 200 OK from real HTTPAssistantGateway.AssessFit, got %d: %s", fitRes.Code, fitRes.Body.String())
		}

		draftRes := doJSON(t, realSrv, http.MethodPost, "/api/opportunities/"+stripeOpp.ID+"/artifacts/draft", map[string]any{
			"kind": "cover_letter",
		})
		if draftRes.Code != http.StatusOK {
			t.Fatalf("expected 200 OK from real HTTPAssistantGateway.DraftArtifact, got %d: %s", draftRes.Code, draftRes.Body.String())
		}
		var draftDTO httpadapter.ArtifactDraftDTO
		_ = json.Unmarshal(draftRes.Body.Bytes(), &draftDTO)
		if !strings.Contains(draftDTO.ContentMarkdown, "zero-downtime Postgres sharding") {
			t.Fatalf("unexpected drafted markdown from real HTTPAssistantGateway: %+v", draftDTO)
		}
	})
}
