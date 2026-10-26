CREATE TABLE IF NOT EXISTS ai_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    provider TEXT NOT NULL DEFAULT 'openrouter',
    base_url TEXT NOT NULL DEFAULT '',
    api_key TEXT NOT NULL DEFAULT '',
    fast_extraction_model TEXT NOT NULL DEFAULT '',
    reasoning_tailoring_model TEXT NOT NULL DEFAULT '',
    bookmarked_models_json TEXT NOT NULL DEFAULT '[]',
    updated_at TEXT NOT NULL
);
