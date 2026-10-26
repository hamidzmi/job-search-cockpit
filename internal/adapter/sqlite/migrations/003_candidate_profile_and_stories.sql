CREATE TABLE IF NOT EXISTS candidate_profile (
    id TEXT PRIMARY KEY,
    full_name TEXT NOT NULL DEFAULT '',
    headline TEXT NOT NULL DEFAULT '',
    summary TEXT NOT NULL DEFAULT '',
    work_history_json TEXT NOT NULL DEFAULT '[]',
    skills_json TEXT NOT NULL DEFAULT '[]',
    target_role_titles_json TEXT NOT NULL DEFAULT '[]',
    location_preferences_json TEXT NOT NULL DEFAULT '[]',
    remote_preference TEXT NOT NULL DEFAULT '',
    compensation_floor INTEGER NOT NULL DEFAULT 0,
    target_compensation INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS stories (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    hook TEXT NOT NULL DEFAULT '',
    situation TEXT NOT NULL DEFAULT '',
    task TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL DEFAULT '',
    result TEXT NOT NULL DEFAULT '',
    competency_tags_json TEXT NOT NULL DEFAULT '[]',
    skill_tags_json TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_stories_updated_at ON stories(updated_at DESC);
