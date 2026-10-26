CREATE TABLE IF NOT EXISTS fit_assessments (
    opportunity_id TEXT PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    overall_score INTEGER NOT NULL DEFAULT 0,
    overall_summary TEXT NOT NULL DEFAULT '',
    matched_strengths_json TEXT NOT NULL DEFAULT '[]',
    gaps_and_risks_json TEXT NOT NULL DEFAULT '[]',
    compensation_alignment TEXT NOT NULL DEFAULT '',
    recommended_story_ids_json TEXT NOT NULL DEFAULT '[]',
    model_used TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS application_artifacts (
    id TEXT PRIMARY KEY,
    opportunity_id TEXT NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    title TEXT NOT NULL,
    content_markdown TEXT NOT NULL DEFAULT '',
    file_path TEXT NOT NULL DEFAULT '',
    file_name TEXT NOT NULL DEFAULT '',
    mime_type TEXT NOT NULL DEFAULT '',
    file_size_bytes INTEGER NOT NULL DEFAULT 0,
    is_submitted INTEGER NOT NULL DEFAULT 0,
    model_used TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_application_artifacts_opportunity_id
    ON application_artifacts(opportunity_id, updated_at DESC);
