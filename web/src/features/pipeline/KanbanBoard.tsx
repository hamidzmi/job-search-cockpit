import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { STAGES, type Opportunity, type Stage } from '../../api/types';
import { formatLocalDate } from '../../lib/date';

interface KanbanBoardProps {
  opportunities: Opportunity[];
  onMoveStage: (opportunityId: string, nextStage: Stage) => Promise<void>;
}

export function KanbanBoard({ opportunities, onMoveStage }: KanbanBoardProps) {
  const activeOpportunities = opportunities.filter((o) => o.outcome === 'Active');
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [dragOverStage, setDragOverStage] = useState<Stage | null>(null);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const handleDragStart = (e: React.DragEvent, id: string) => {
    e.dataTransfer.setData('text/plain', id);
    e.dataTransfer.effectAllowed = 'move';
    setDraggingId(id);
  };

  const handleDragEnd = () => {
    setDraggingId(null);
    setDragOverStage(null);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  };

  const handleDragEnter = (stage: Stage) => {
    setDragOverStage(stage);
  };

  const handleDragLeave = (e: React.DragEvent, stage: Stage) => {
    if (!e.currentTarget.contains(e.relatedTarget as Node)) {
      setDragOverStage((current) => (current === stage ? null : current));
    }
  };

  const handleDrop = async (e: React.DragEvent, targetStage: Stage) => {
    e.preventDefault();
    setDragOverStage(null);
    const oppId = e.dataTransfer.getData('text/plain') || draggingId;
    setDraggingId(null);

    if (!oppId) return;
    const opp = activeOpportunities.find((o) => o.id === oppId);
    if (!opp || opp.stage === targetStage) return;

    await onMoveStage(oppId, targetStage);
    setToastMessage(`Moved ${opp.role_title} to ${targetStage}`);
    setTimeout(() => {
      setToastMessage((msg) => (msg?.includes(opp.role_title) ? null : msg));
    }, 3500);
  };

  return (
    <div className="kanban-board" aria-label="Pipeline Kanban Board">
      {STAGES.map((stage) => {
        const stageItems = activeOpportunities.filter((o) => o.stage === stage);
        const stageLower = stage.toLowerCase();
        const isDropActive = dragOverStage === stage;

        return (
          <section
            key={stage}
            className={`kanban-column kanban-column-${stageLower} ${
              isDropActive ? 'kanban-column-drop-active' : ''
            }`}
            aria-label={`${stage} Stage Column`}
            onDragOver={handleDragOver}
            onDragEnter={() => handleDragEnter(stage)}
            onDragLeave={(e) => handleDragLeave(e, stage)}
            onDrop={(e) => handleDrop(e, stage)}
          >
            <header className="kanban-column-header">
              <h3>{stage}</h3>
              <span className="badge-count">{stageItems.length}</span>
            </header>
            <div className="kanban-column-cards">
              {stageItems.length === 0 ? (
                <p className="empty-column-hint">
                  {isDropActive ? 'Drop card here' : 'No active opportunities'}
                </p>
              ) : (
                stageItems.map((opp) => {
                  const isCurrentDragging = draggingId === opp.id;
                  return (
                    <article
                      key={opp.id}
                      className={`opportunity-card ${
                        isCurrentDragging ? 'is-dragging' : ''
                      }`}
                      draggable
                      onDragStart={(e) => handleDragStart(e, opp.id)}
                      onDragEnd={handleDragEnd}
                      title="Drag to change Stage"
                    >
                      <div className="card-top-row">
                        <span className="company-pill">{opp.company_name}</span>
                        <span
                          className={`priority-badge priority-${opp.priority.toLowerCase()}`}
                        >
                          {opp.priority}
                        </span>
                      </div>
                      <h4 className="card-title">
                        <Link to={`/opportunities/${opp.id}`}>{opp.role_title}</Link>
                      </h4>
                      <div className="card-meta">
                        <span className="tech-tag">
                          {opp.job_posting.has_content ? 'Job Posting attached' : 'Unposted pursuit'}
                        </span>
                        {opp.remote_policy && (
                          <span className="tech-tag">{opp.remote_policy}</span>
                        )}
                        {opp.location && opp.location !== 'Remote' && (
                          <span className="tech-tag">{opp.location}</span>
                        )}
                      </div>
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
                  );
                })
              )}
            </div>
          </section>
        );
      })}

      {toastMessage && (
        <div className="cockpit-toast" role="status" aria-live="polite">
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="#10b981"
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="20 6 9 17 4 12" />
          </svg>
          <span>{toastMessage}</span>
        </div>
      )}
    </div>
  );
}
