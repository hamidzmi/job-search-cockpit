import React from 'react';
import { Link } from 'react-router-dom';
import { STAGES, type Opportunity, type Stage } from '../../api/types';
import { formatLocalDate } from '../../lib/date';

interface KanbanBoardProps {
  opportunities: Opportunity[];
  onMoveStage: (opportunityId: string, nextStage: Stage) => Promise<void>;
}

export function KanbanBoard({ opportunities, onMoveStage }: KanbanBoardProps) {
  const activeOpportunities = opportunities.filter((o) => o.outcome === 'Active');

  return (
    <div className="kanban-board" aria-label="Pipeline Kanban Board">
      {STAGES.map((stage) => {
        const stageItems = activeOpportunities.filter((o) => o.stage === stage);
        return (
          <section
            key={stage}
            className="kanban-column"
            aria-label={`${stage} Stage Column`}
          >
            <header className="kanban-column-header">
              <h3>{stage}</h3>
              <span className="badge-count">{stageItems.length}</span>
            </header>
            <div className="kanban-column-cards">
              {stageItems.length === 0 ? (
                <p className="empty-column-hint">No active opportunities</p>
              ) : (
                stageItems.map((opp) => (
                  <article key={opp.id} className="opportunity-card">
                    <div className="card-top-row">
                      <span className="company-pill">{opp.company_name}</span>
                      <span className={`priority-badge priority-${opp.priority.toLowerCase()}`}>
                        {opp.priority}
                      </span>
                    </div>
                    <h4 className="card-title">
                      <Link to={`/opportunities/${opp.id}`}>{opp.role_title}</Link>
                    </h4>
                    <div className="card-meta">
                      {opp.remote_policy && <span>{opp.remote_policy}</span>}
                      {opp.location && <span>{opp.location}</span>}
                      <span>{opp.job_posting.has_content ? 'Job Posting attached' : 'Unposted pursuit'}</span>
                    </div>
                    {opp.job_posting.tech_stack_tags.length > 0 && (
                      <div className="tag-list">
                        {opp.job_posting.tech_stack_tags.map((tag) => (
                          <span key={tag} className="tech-tag">
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}
                    <div className="card-footer">
                      <span className="timestamp-text">
                        Active {formatLocalDate(opp.last_activity_at)}
                      </span>
                      <label className="stage-select-label">
                        <span className="sr-only">Move {opp.role_title} Stage</span>
                        <select
                          aria-label={`Move ${opp.role_title} Stage`}
                          value={opp.stage}
                          onChange={(e) =>
                            void onMoveStage(opp.id, e.target.value as Stage)
                          }
                        >
                          {STAGES.map((s) => (
                            <option key={s} value={s}>
                              {s}
                            </option>
                          ))}
                        </select>
                      </label>
                    </div>
                  </article>
                ))
              )}
            </div>
          </section>
        );
      })}
    </div>
  );
}
