import React from 'react';
import { Link } from 'react-router-dom';

export function FlightDeckWorkspace() {
  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Flight Deck</h2>
          <p className="workspace-subtitle">
            Daily command center surfacing imminent Interview Rounds, missing Debriefs, due Action Items, and stale Opportunities.
          </p>
        </div>
        <div className="workspace-actions">
          <Link to="/opportunities" className="btn-primary">
            Open Pipeline
          </Link>
        </div>
      </header>
    </div>
  );
}

export function AnalyticsWorkspace() {
  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Analytics &amp; Offers</h2>
          <p className="workspace-subtitle">
            Funnel conversion analytics from Stage Transitions and side-by-side Offer comparisons.
          </p>
        </div>
      </header>
    </div>
  );
}

export function SettingsWorkspace() {
  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Settings</h2>
          <p className="workspace-subtitle">
            OpenRouter &amp; AI provider configuration, local data backup, and sample workspace seeding.
          </p>
        </div>
      </header>
    </div>
  );
}
