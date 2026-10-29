import { useQuery } from '@tanstack/react-query';
import React from 'react';
import { useSearchParams } from 'react-router-dom';
import { useCockpitApi } from '../../api/CockpitApiContext';
import { CandidateProfilePanel } from './CandidateProfilePanel';
import { QuestionBankPanel } from './QuestionBankPanel';
import { StoryBankPanel } from './StoryBankPanel';

type ProfileTab = 'profile' | 'stories' | 'questions';

export function ProfileWorkspace() {
  const api = useCockpitApi();
  const [searchParams, setSearchParams] = useSearchParams();

  const tabParam = searchParams.get('tab') as ProfileTab | null;
  const activeTab: ProfileTab =
    tabParam && ['profile', 'stories', 'questions'].includes(tabParam)
      ? tabParam
      : 'profile';

  const setTab = (tab: ProfileTab) => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.set('tab', tab);
        return next;
      },
      { replace: true },
    );
  };

  const profileQuery = useQuery({
    queryKey: ['profile'],
    queryFn: () => api.profile.getProfile(),
  });

  const storiesQuery = useQuery({
    queryKey: ['stories', {}],
    queryFn: () => api.stories.listStories({}),
  });

  const questionsQuery = useQuery({
    queryKey: ['question-bank', {}],
    queryFn: () => api.interviewRounds.listQuestionBank({}),
  });

  const profile = profileQuery.data;
  const stories = storiesQuery.data ?? [];
  const questions = questionsQuery.data ?? [];

  const rolesCount = profile?.work_history?.length ?? 0;
  const skillsCount = profile?.skills?.length ?? 0;

  return (
    <div className="workspace-container">
      {/* Workspace Header */}
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Candidate Profile &amp; Story Bank</h2>
          <p className="workspace-subtitle">
            Canonical work history, compensation floor, and STAR Story Bank for tailored pursuits and interview loops.
          </p>
        </div>
      </header>

      {/* Top HUD Banner */}
      <div className="flight-deck-hud" role="region" aria-label="Profile Health HUD">
        <div
          className="hud-card"
          onClick={() => setTab('profile')}
          style={{ cursor: 'pointer' }}
        >
          <div>
            <div className="hud-value">{rolesCount}</div>
            <div className="hud-label">Work History Roles</div>
          </div>
          <div className="hud-icon-badge hud-badge-rounds">
            💼
          </div>
        </div>

        <div
          className="hud-card"
          onClick={() => setTab('profile')}
          style={{ cursor: 'pointer' }}
        >
          <div>
            <div className="hud-value">{skillsCount}</div>
            <div className="hud-label">Technical Skills</div>
          </div>
          <div className="hud-icon-badge hud-badge-stale">
            🎯
          </div>
        </div>

        <div
          className="hud-card"
          onClick={() => setTab('stories')}
          style={{ cursor: 'pointer' }}
        >
          <div>
            <div className="hud-value">{stories.length}</div>
            <div className="hud-label">STAR Stories</div>
          </div>
          <div className="hud-icon-badge hud-badge-action">
            📖
          </div>
        </div>

        <div
          className="hud-card"
          onClick={() => setTab('questions')}
          style={{ cursor: 'pointer' }}
        >
          <div>
            <div className="hud-value">{questions.length}</div>
            <div className="hud-label">Question Bank</div>
          </div>
          <div className="hud-icon-badge hud-badge-success">
            ❓
          </div>
        </div>
      </div>

      {/* Workspace Sub-Tabs */}
      <div
        className="opportunity-tab-bar"
        role="tablist"
        aria-label="Profile Workspace Tabs"
      >
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'profile'}
          className={`btn-tab ${activeTab === 'profile' ? 'active' : ''}`}
          onClick={() => setTab('profile')}
        >
          Candidate Profile
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'stories'}
          className={`btn-tab ${activeTab === 'stories' ? 'active' : ''}`}
          onClick={() => setTab('stories')}
        >
          STAR Story Bank ({stories.length})
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'questions'}
          className={`btn-tab ${activeTab === 'questions' ? 'active' : ''}`}
          onClick={() => setTab('questions')}
        >
          Interview Question Bank ({questions.length})
        </button>
      </div>

      {/* Tab Panels */}
      {activeTab === 'profile' && <CandidateProfilePanel />}
      {activeTab === 'stories' && <StoryBankPanel />}
      {activeTab === 'questions' && <QuestionBankPanel />}
    </div>
  );
}
