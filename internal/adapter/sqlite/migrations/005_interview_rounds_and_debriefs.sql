CREATE TABLE IF NOT EXISTS interview_rounds (
    id TEXT PRIMARY KEY,
    opportunity_id TEXT NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    round_order INTEGER NOT NULL,
    round_type TEXT NOT NULL,
    title TEXT NOT NULL,
    scheduled_at TEXT NOT NULL DEFAULT '',
    duration_minutes INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'Scheduled',
    prep_notes TEXT NOT NULL DEFAULT '',
    contact_ids_json TEXT NOT NULL DEFAULT '[]',
    story_ids_json TEXT NOT NULL DEFAULT '[]',
    debrief_json TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_interview_rounds_opportunity_order ON interview_rounds(opportunity_id, round_order ASC, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_interview_rounds_status_scheduled ON interview_rounds(status, scheduled_at);
