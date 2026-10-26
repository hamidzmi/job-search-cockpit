package httpadapter_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	httpadapter "job-search-cockpit/internal/adapter/http"
)

func TestCandidateProfileAPI(t *testing.T) {
	t0 := time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)
	clock := &manualClock{now: t0}
	srv := newTestServer(t, clock)

	t.Run("GET /api/profile returns default empty Candidate Profile before first save", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodGet, "/api/profile", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		var got httpadapter.CandidateProfileDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if got.Currency != "USD" || got.CompensationFloor != 0 || len(got.WorkHistory) != 0 || len(got.Skills) != 0 {
			t.Errorf("unexpected default profile: %+v", got)
		}
	})

	putCases := []struct {
		name       string
		payload    map[string]any
		wantStatus int
	}{
		{
			name: "rejects target compensation below compensation floor with 400",
			payload: map[string]any{
				"compensation_floor":  260000,
				"target_compensation": 210000,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "rejects work history entry without company or title with 400",
			payload: map[string]any{
				"work_history": []map[string]any{
					{
						"company": "Stripe",
						"title":   "   ",
					},
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "saves canonical Candidate Profile and returns 200",
			payload: map[string]any{
				"full_name": "Alex Rivera",
				"headline":  "Staff Distributed Systems Engineer",
				"summary":   "12 years architecting high-throughput payment ledgers and cloud storage engines.",
				"work_history": []map[string]any{
					{
						"company":    "Stripe",
						"title":      "Staff Software Engineer",
						"start_date": "2022-01",
						"end_date":   "Present",
						"summary":    "Tech lead for core settlement ledger.",
						"impact_bullets": []string{
							"Cut global settlement p99 latency by 42% across 3 regions",
							"Led online migration of $40B/yr ledger across 16 shards with zero downtime",
						},
					},
					{
						"company":    "Datadog",
						"title":      "Senior Software Engineer",
						"start_date": "2019-03",
						"end_date":   "2021-12",
						"summary":    "Metrics ingestion pipeline.",
						"impact_bullets": []string{
							"Scaled Kafka ingestion throughput 3.5x while reducing compute cost by 28%",
						},
					},
				},
				"skills":               []string{"Go", "Rust", "Kafka", "PostgreSQL", "Distributed Systems"},
				"target_role_titles":   []string{"Staff Backend Engineer", "Principal Systems Engineer"},
				"location_preferences": []string{"San Francisco, CA", "Seattle, WA", "Remote (US)"},
				"remote_preference":    "Remote or Hybrid",
				"compensation_floor":   240000,
				"target_compensation":  290000,
				"currency":             "USD",
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range putCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, srv, http.MethodPut, "/api/profile", tc.payload)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}

	t.Run("GET /api/profile returns saved Candidate Profile with all fields persisted", func(t *testing.T) {
		rec := doJSON(t, srv, http.MethodGet, "/api/profile", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		var got httpadapter.CandidateProfileDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if got.FullName != "Alex Rivera" || got.CompensationFloor != 240000 || got.TargetCompensation != 290000 {
			t.Errorf("unexpected profile fields: %+v", got)
		}
		if got.RemotePreference != "Remote or Hybrid" {
			t.Errorf("expected RemotePreference 'Remote or Hybrid', got %q", got.RemotePreference)
		}
		if len(got.WorkHistory) != 2 || len(got.WorkHistory[0].ImpactBullets) != 2 {
			t.Errorf("unexpected work history: %+v", got.WorkHistory)
		}
		wantTitles := []string{"Staff Backend Engineer", "Principal Systems Engineer"}
		if !reflect.DeepEqual(got.TargetRoleTitles, wantTitles) {
			t.Errorf("expected TargetRoleTitles %v, got %v", wantTitles, got.TargetRoleTitles)
		}
		if !got.UpdatedAt.Equal(t0) {
			t.Errorf("expected UpdatedAt %v, got %v", t0, got.UpdatedAt)
		}
	})
}

func TestStoriesAPI_CRUDAndFiltering(t *testing.T) {
	t0 := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	clock := &manualClock{now: t0}
	srv := newTestServer(t, clock)

	createCases := []struct {
		name       string
		payload    map[string]any
		wantStatus int
		wantTitle  string
	}{
		{
			name:       "rejects blank story title with 400",
			payload:    map[string]any{"title": "   "},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "creates System Design & Leadership STAR story",
			payload: map[string]any{
				"title":           "Zero-Downtime Ledger Sharding",
				"hook":            "Sharded a $40B/yr payment ledger across 16 PostgreSQL clusters with zero downtime.",
				"situation":       "Single primary ledger database hit 85% write saturation ahead of peak holiday traffic.",
				"task":            "Design and execute horizontal sharding in 10 weeks without halting live transactions.",
				"action":          "Built a dual-write shadow verifier in Go with deterministic replay and phased traffic migration.",
				"result":          "Reduced p99 write latency by 58% and survived 3.2x Black Friday peak with zero incidents.",
				"competency_tags": []string{"System Design", "Leadership", "Ownership"},
				"skill_tags":      []string{"Go", "PostgreSQL", "Distributed Systems"},
			},
			wantStatus: http.StatusCreated,
			wantTitle:  "Zero-Downtime Ledger Sharding",
		},
		{
			name: "creates Conflict & Ownership STAR story",
			payload: map[string]any{
				"title":           "Unblocking Cross-Team Schema Deadlock",
				"hook":            "Aligned 3 platform teams on a unified event schema in 48 hours after a 3-week impasse.",
				"situation":       "Ingestion and billing teams disagreed on Kafka protobuf partitioning guarantees.",
				"task":            "Resolve the architectural disagreement before the Q3 launch freeze.",
				"action":          "Prototyped benchmark harness in Rust comparing both partitioning strategies under skew.",
				"result":          "Achieved unanimous sign-off and shipped the pipeline 1 week ahead of schedule.",
				"competency_tags": []string{"Conflict", "Ownership"},
				"skill_tags":      []string{"Kafka", "Rust", "Protobuf"},
			},
			wantStatus: http.StatusCreated,
			wantTitle:  "Unblocking Cross-Team Schema Deadlock",
		},
	}

	var storyIDs []string
	for _, tc := range createCases {
		t.Run(tc.name, func(t *testing.T) {
			clock.now = clock.now.Add(time.Minute)
			rec := doJSON(t, srv, http.MethodPost, "/api/stories", tc.payload)
			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.wantStatus, rec.Code, rec.Body.String())
			}
			if tc.wantStatus == http.StatusCreated {
				var dto httpadapter.StoryDTO
				if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
					t.Fatalf("decode failed: %v", err)
				}
				if dto.Title != tc.wantTitle {
					t.Errorf("expected title %q, got %q", tc.wantTitle, dto.Title)
				}
				storyIDs = append(storyIDs, dto.ID)
			}
		})
	}

	firstID := storyIDs[0]
	secondID := storyIDs[1]

	t.Run("updates story via PUT and retrieves via GET /api/stories/{id}", func(t *testing.T) {
		clock.now = clock.now.Add(10 * time.Minute)
		updateRec := doJSON(t, srv, http.MethodPut, "/api/stories/"+firstID, map[string]any{
			"title":           "Zero-Downtime Multi-Region Ledger Sharding",
			"hook":            "Sharded a $40B/yr payment ledger across 16 PostgreSQL clusters with zero downtime.",
			"situation":       "Single primary ledger database hit 85% write saturation ahead of peak holiday traffic.",
			"task":            "Design and execute horizontal sharding in 10 weeks without halting live transactions.",
			"action":          "Built a dual-write shadow verifier in Go with deterministic replay and phased traffic migration.",
			"result":          "Reduced p99 write latency by 62% and survived 4x Black Friday peak with zero incidents.",
			"competency_tags": []string{"System Design", "Leadership"},
			"skill_tags":      []string{"Go", "PostgreSQL", "Distributed Systems"},
		})
		if updateRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on PUT story, got %d: %s", updateRec.Code, updateRec.Body.String())
		}

		getRec := doJSON(t, srv, http.MethodGet, "/api/stories/"+firstID, nil)
		if getRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on GET story, got %d", getRec.Code)
		}
		var got httpadapter.StoryDTO
		_ = json.Unmarshal(getRec.Body.Bytes(), &got)
		if got.Title != "Zero-Downtime Multi-Region Ledger Sharding" {
			t.Errorf("expected updated title, got %q", got.Title)
		}
	})

	filterCases := []struct {
		name       string
		query      string
		wantTitles []string
	}{
		{
			name:       "lists all stories when no filter is provided",
			query:      "",
			wantTitles: []string{"Zero-Downtime Multi-Region Ledger Sharding", "Unblocking Cross-Team Schema Deadlock"},
		},
		{
			name:       "filters by keyword query in Result or Action",
			query:      "?q=shadow+verifier",
			wantTitles: []string{"Zero-Downtime Multi-Region Ledger Sharding"},
		},
		{
			name:       "filters by competency_tag",
			query:      "?competency_tag=Conflict",
			wantTitles: []string{"Unblocking Cross-Team Schema Deadlock"},
		},
		{
			name:       "filters by skill_tag",
			query:      "?skill_tag=PostgreSQL",
			wantTitles: []string{"Zero-Downtime Multi-Region Ledger Sharding"},
		},
		{
			name:       "filters by combined keyword, competency_tag, and skill_tag",
			query:      "?q=protobuf&competency_tag=Ownership&skill_tag=Kafka",
			wantTitles: []string{"Unblocking Cross-Team Schema Deadlock"},
		},
	}

	for _, tc := range filterCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doJSON(t, srv, http.MethodGet, "/api/stories"+tc.query, nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
			}
			var got []httpadapter.StoryDTO
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if len(got) != len(tc.wantTitles) {
				t.Fatalf("expected %d stories, got %d: %+v", len(tc.wantTitles), len(got), got)
			}
			for i, wantTitle := range tc.wantTitles {
				if got[i].Title != wantTitle {
					t.Errorf("story[%d]: expected title %q, got %q", i, wantTitle, got[i].Title)
				}
			}
		})
	}

	t.Run("deletes story via DELETE /api/stories/{id} and returns 404 thereafter", func(t *testing.T) {
		delRec := doJSON(t, srv, http.MethodDelete, "/api/stories/"+secondID, nil)
		if delRec.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content on DELETE story, got %d: %s", delRec.Code, delRec.Body.String())
		}

		delAgain := doJSON(t, srv, http.MethodDelete, "/api/stories/"+secondID, nil)
		if delAgain.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found on second DELETE, got %d", delAgain.Code)
		}

		listRec := doJSON(t, srv, http.MethodGet, "/api/stories", nil)
		var remaining []httpadapter.StoryDTO
		_ = json.Unmarshal(listRec.Body.Bytes(), &remaining)
		if len(remaining) != 1 || remaining[0].ID != firstID {
			t.Fatalf("expected only first story to remain, got %+v", remaining)
		}
	})
}
