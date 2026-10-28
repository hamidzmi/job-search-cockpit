import React, { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  OUTCOMES,
  STAGES,
  type Opportunity,
  type Outcome,
  type Stage,
} from '../../api/types';
import { formatLocalDate } from '../../lib/date';

interface ToastState {
  message: string;
  onUndo?: () => Promise<void>;
}

interface KanbanBoardProps {
  opportunities: Opportunity[];
  onMoveStage: (opportunityId: string, nextStage: Stage) => Promise<void>;
  onChangeOutcome?: (
    opportunityId: string,
    nextOutcome: Outcome,
    note?: string,
  ) => Promise<void>;
}

export function KanbanBoard({
  opportunities,
  onMoveStage,
  onChangeOutcome,
}: KanbanBoardProps) {
  const activeOpportunities = opportunities.filter((o) => o.outcome === 'Active');
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [dragOverStage, setDragOverStage] = useState<Stage | null>(null);
  const [toast, setToast] = useState<ToastState | null>(null);
  const toastTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (toastTimerRef.current) {
        clearTimeout(toastTimerRef.current);
      }
    };
  }, []);

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
    if (toastTimerRef.current) {
      clearTimeout(toastTimerRef.current);
    }
    setToast({
      message: `Moved ${opp.role_title} to ${targetStage}`,
    });
    toastTimerRef.current = setTimeout(() => {
      setToast(null);
    }, 3500);
  };

  const handleOutcomeChange = async (opp: Opportunity, nextOutcome: Outcome) => {
    if (nextOutcome === opp.outcome || !onChangeOutcome) return;
    const previousOutcome = opp.outcome;
    await onChangeOutcome(opp.id, nextOutcome);

    if (toastTimerRef.current) {
      clearTimeout(toastTimerRef.current);
    }
    setToast({
      message: `Marked ${opp.role_title} as ${nextOutcome}`,
      onUndo: async () => {
        await onChangeOutcome(opp.id, previousOutcome, 'Reverted to Active (Undo)');
      },
    });
    toastTimerRef.current = setTimeout(() => {
      setToast(null);
    }, 6000);
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
                        <label className="outcome-select-label">
                          <span className="sr-only">Set {opp.role_title} Outcome</span>
                          <select
                            aria-label={`Set ${opp.role_title} Outcome`}
                            value={opp.outcome}
                            onChange={(e) =>
                              void handleOutcomeChange(opp, e.target.value as Outcome)
                            }
                          >
                            {OUTCOMES.map((o) => (
                              <option key={o} value={o}>
                                {o}
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

      {toast && (
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
          <span>{toast.message}</span>
          {toast.onUndo && (
            <button
              type="button"
              className="toast-undo-btn"
              onClick={async () => {
                const undoFn = toast.onUndo;
                setToast(null);
                if (undoFn) {
                  await undoFn();
                }
              }}
            >
              Undo
            </button>
          )}
        </div>
      )}
    </div>
  );
}
