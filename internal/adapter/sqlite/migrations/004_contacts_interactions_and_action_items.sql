CREATE TABLE IF NOT EXISTS contacts (
    id TEXT PRIMARY KEY,
    company_id TEXT REFERENCES companies(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    linkedin_url TEXT NOT NULL DEFAULT '',
    warmth TEXT NOT NULL DEFAULT 'Warm',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_contacts_company_id ON contacts(company_id);
CREATE INDEX IF NOT EXISTS idx_contacts_warmth ON contacts(warmth);

CREATE TABLE IF NOT EXISTS opportunity_contacts (
    opportunity_id TEXT NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    contact_id TEXT NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    linked_at TEXT NOT NULL,
    PRIMARY KEY (opportunity_id, contact_id)
);

CREATE INDEX IF NOT EXISTS idx_opportunity_contacts_contact_id ON opportunity_contacts(contact_id);

CREATE TABLE IF NOT EXISTS interactions (
    id TEXT PRIMARY KEY,
    contact_id TEXT NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    opportunity_id TEXT REFERENCES opportunities(id) ON DELETE SET NULL,
    kind TEXT NOT NULL,
    direction TEXT NOT NULL DEFAULT 'Outbound',
    occurred_at TEXT NOT NULL,
    summary TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_interactions_contact_id ON interactions(contact_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_interactions_opportunity_id ON interactions(opportunity_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS action_items (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    opportunity_id TEXT REFERENCES opportunities(id) ON DELETE CASCADE,
    contact_id TEXT REFERENCES contacts(id) ON DELETE SET NULL,
    due_at TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_action_items_status_due_at ON action_items(status, due_at);
CREATE INDEX IF NOT EXISTS idx_action_items_opportunity_id ON action_items(opportunity_id);
CREATE INDEX IF NOT EXISTS idx_action_items_contact_id ON action_items(contact_id);
