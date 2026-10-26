CREATE TABLE IF NOT EXISTS companies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    domain TEXT NOT NULL DEFAULT '',
    headquarters TEXT NOT NULL DEFAULT '',
    remote_policy TEXT NOT NULL DEFAULT '',
    size_stage TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS opportunities (
    id TEXT PRIMARY KEY,
    company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    role_title TEXT NOT NULL,
    seniority TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    remote_policy TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'Medium',
    stage TEXT NOT NULL,
    outcome TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    raw_text TEXT NOT NULL DEFAULT '',
    requirements_json TEXT NOT NULL DEFAULT '[]',
    tech_stack_tags_json TEXT NOT NULL DEFAULT '[]',
    stated_min_comp INTEGER NOT NULL DEFAULT 0,
    stated_max_comp INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT '',
    last_activity_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_opportunities_company_id ON opportunities(company_id);
CREATE INDEX IF NOT EXISTS idx_opportunities_stage_outcome ON opportunities(stage, outcome);

CREATE TABLE IF NOT EXISTS stage_transitions (
    id TEXT PRIMARY KEY,
    opportunity_id TEXT NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    from_stage TEXT NOT NULL DEFAULT '',
    to_stage TEXT NOT NULL,
    from_outcome TEXT NOT NULL DEFAULT '',
    to_outcome TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    transitioned_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_stage_transitions_opportunity_id ON stage_transitions(opportunity_id, transitioned_at);
