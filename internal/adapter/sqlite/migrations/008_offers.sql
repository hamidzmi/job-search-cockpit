CREATE TABLE IF NOT EXISTS offers (
    id TEXT PRIMARY KEY,
    opportunity_id TEXT NOT NULL UNIQUE REFERENCES opportunities(id) ON DELETE CASCADE,
    base_salary INTEGER NOT NULL,
    bonus_target_pct INTEGER NOT NULL DEFAULT 0,
    bonus_amount INTEGER NOT NULL DEFAULT 0,
    equity_type TEXT NOT NULL DEFAULT '',
    equity_total_value INTEGER NOT NULL DEFAULT 0,
    equity_vesting_years INTEGER NOT NULL DEFAULT 4,
    sign_on_bonus INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    deadline_at TEXT NOT NULL DEFAULT '',
    benefits_notes TEXT NOT NULL DEFAULT '',
    negotiation_notes TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'Pending',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_offers_opportunity_id ON offers(opportunity_id);
