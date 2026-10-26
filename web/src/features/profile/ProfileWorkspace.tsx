import React from 'react';
import { CandidateProfilePanel } from './CandidateProfilePanel';
import { QuestionBankPanel } from './QuestionBankPanel';
import { StoryBankPanel } from './StoryBankPanel';

export function ProfileWorkspace() {
  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Candidate Profile &amp; Story Bank</h2>
          <p className="workspace-subtitle">
            Canonical work history, compensation floor, and STAR Story Bank.
          </p>
        </div>
      </header>

      <CandidateProfilePanel />
      <StoryBankPanel />
      <QuestionBankPanel />
    </div>
  );
}
