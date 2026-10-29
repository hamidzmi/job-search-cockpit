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
import { cn } from '../../lib/utils';
import { Button } from '../../components/ui/Button';
import { Badge } from '../../components/ui/Badge';
import {
  ArrowDown,
  ArrowUpRight,
  CalendarDays,
  Check,
  Clock,
  FileText,
  Flame,
  Plus,
} from 'lucide-react';

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
  onAddOpportunity?: (stage?: Stage) => void;
}

const STAGE_CONFIG: Record<
  Stage,
  {
    dot: string;
    borderTop: string;
    badgeBg: string;
    badgeText: string;
    badgeBorder: string;
  }
> = {
  Discovered: {
    dot: 'bg-indigo-500',
    borderTop: 'border-t-indigo-500',
    badgeBg: 'bg-indigo-500/10',
    badgeText: 'text-indigo-600 dark:text-indigo-400',
    badgeBorder: 'border-indigo-500/20',
  },
  Outreach: {
    dot: 'bg-sky-500',
    borderTop: 'border-t-sky-500',
    badgeBg: 'bg-sky-500/10',
    badgeText: 'text-sky-600 dark:text-sky-400',
    badgeBorder: 'border-sky-500/20',
  },
  Applied: {
    dot: 'bg-amber-500',
    borderTop: 'border-t-amber-500',
    badgeBg: 'bg-amber-500/10',
    badgeText: 'text-amber-700 dark:text-amber-400',
    badgeBorder: 'border-amber-500/20',
  },
  Screening: {
    dot: 'bg-teal-500',
    borderTop: 'border-t-teal-500',
    badgeBg: 'bg-teal-500/10',
    badgeText: 'text-teal-600 dark:text-teal-400',
    badgeBorder: 'border-teal-500/20',
  },
  Interviewing: {
    dot: 'bg-purple-500',
    borderTop: 'border-t-purple-500',
    badgeBg: 'bg-purple-500/10',
    badgeText: 'text-purple-600 dark:text-purple-400',
    badgeBorder: 'border-purple-500/20',
  },
  Offer: {
    dot: 'bg-emerald-500',
    borderTop: 'border-t-emerald-500',
    badgeBg: 'bg-emerald-500/10',
    badgeText: 'text-emerald-600 dark:text-emerald-400',
    badgeBorder: 'border-emerald-500/20',
  },
};

export function KanbanBoard({
  opportunities,
  onMoveStage,
  onChangeOutcome,
  onAddOpportunity,
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

  const renderPriorityBadge = (priority: Opportunity['priority']) => {
    switch (priority) {
      case 'High':
        return (
          <span className="priority-badge priority-high" title="Priority: High">
            <Flame className="size-3 text-red-500 shrink-0" />
            <span>High</span>
          </span>
        );
      case 'Medium':
        return (
          <span className="priority-badge priority-medium" title="Priority: Medium">
            <Clock className="size-3 text-amber-500 shrink-0" />
            <span>Medium</span>
          </span>
        );
      case 'Low':
      default:
        return (
          <span className="priority-badge priority-low" title="Priority: Low">
            <ArrowDown className="size-3 text-slate-400 shrink-0" />
            <span>Low</span>
          </span>
        );
    }
  };

  return (
    <div
      className="kanban-board w-full overflow-x-auto pb-4 scrollbar-thin"
      aria-label="Pipeline Kanban Board"
    >
      <div className="inline-grid grid-cols-[repeat(6,minmax(285px,1fr))] min-w-max gap-3.5 items-start">
        {STAGES.map((stage) => {
          const stageItems = activeOpportunities.filter((o) => o.stage === stage);
          const stageLower = stage.toLowerCase();
          const isDropActive = dragOverStage === stage;
          const config = STAGE_CONFIG[stage];

          return (
            <section
              key={stage}
              className={cn(
                'kanban-column w-[285px] sm:w-[300px] flex flex-col rounded-xl border bg-muted/40 dark:bg-muted/20 min-h-[580px] transition-all',
                `kanban-column-${stageLower}`,
                config.borderTop,
                'border-t-[3px]',
                isDropActive &&
                  'kanban-column-drop-active ring-2 ring-primary border-primary bg-primary/5 shadow-md',
              )}
              aria-label={`${stage} Stage Column`}
              onDragOver={handleDragOver}
              onDragEnter={() => handleDragEnter(stage)}
              onDragLeave={(e) => handleDragLeave(e, stage)}
              onDrop={(e) => handleDrop(e, stage)}
            >
              {/* Column Header */}
              <header className="kanban-column-header flex items-center justify-between p-3 border-b border-border/50">
                <div className="flex items-center gap-2 min-w-0">
                  <span
                    className={cn('size-2 rounded-full shrink-0', config.dot)}
                    aria-hidden="true"
                  />
                  <h3 className="font-semibold text-sm tracking-tight text-foreground truncate">
                    {stage}
                  </h3>
                  <span
                    className={cn(
                      'badge-count text-xs font-semibold px-2 py-0.5 rounded-full border tabular-nums',
                      config.badgeBg,
                      config.badgeText,
                      config.badgeBorder,
                    )}
                  >
                    {stageItems.length}
                  </span>
                </div>
                {onAddOpportunity && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    className="text-muted-foreground hover:text-foreground"
                    title={`Add Opportunity in ${stage}`}
                    aria-label={`Add Opportunity in ${stage}`}
                    onClick={() => onAddOpportunity(stage)}
                  >
                    <Plus className="size-3.5" />
                  </Button>
                )}
              </header>

              {/* Column Cards Container */}
              <div className="kanban-column-cards flex flex-col gap-2.5 p-2.5 flex-1 overflow-y-auto">
                {stageItems.length === 0 ? (
                  <div className="empty-column-hint border border-dashed rounded-lg p-5 text-center text-xs text-muted-foreground bg-background/50 flex flex-col items-center justify-center gap-1">
                    <span>
                      {isDropActive ? 'Drop card here' : 'No active opportunities'}
                    </span>
                  </div>
                ) : (
                  stageItems.map((opp) => {
                    const isCurrentDragging = draggingId === opp.id;
                    return (
                      <article
                        key={opp.id}
                        className={cn(
                          'opportunity-card bg-card text-card-foreground flex flex-col gap-2.5 rounded-xl border border-border p-3.5 shadow-xs hover:shadow-md transition-all cursor-grab active:cursor-grabbing select-none',
                          isCurrentDragging &&
                            'is-dragging opacity-40 ring-2 ring-primary border-primary',
                        )}
                        draggable
                        onDragStart={(e) => handleDragStart(e, opp.id)}
                        onDragEnd={handleDragEnd}
                        title="Drag to change Stage"
                      >
                        {/* Top row: Company & Priority */}
                        <div className="card-top-row flex items-center justify-between gap-2">
                          <span className="company-pill text-xs font-semibold truncate max-w-[170px]">
                            {opp.company_name}
                          </span>
                          {renderPriorityBadge(opp.priority)}
                        </div>

                        {/* Title */}
                        <h4 className="card-title text-sm font-semibold leading-snug line-clamp-2">
                          <Link
                            to={`/opportunities/${opp.id}`}
                            className="hover:text-primary transition-colors"
                          >
                            {opp.role_title}
                          </Link>
                        </h4>

                        {/* Metadata Tags */}
                        <div className="card-meta flex flex-wrap gap-1.5 text-[11px]">
                          <span className="tech-tag inline-flex items-center gap-1 text-[11px]">
                            <FileText className="size-3 shrink-0 text-muted-foreground" />
                            <span>
                              {opp.job_posting.has_content
                                ? 'Job Posting attached'
                                : 'Unposted pursuit'}
                            </span>
                          </span>
                          {opp.remote_policy && (
                            <span className="tech-tag">{opp.remote_policy}</span>
                          )}
                          {opp.location && opp.location !== 'Remote' && (
                            <span className="tech-tag truncate max-w-[130px]">
                              {opp.location}
                            </span>
                          )}
                        </div>

                        <div className="border-t border-border/60 my-0.5" />

                        {/* Card Footer: Activity date & Outcome Select */}
                        <div className="card-footer flex items-center justify-between gap-1.5 pt-1 border-t-0">
                          <span className="timestamp-text text-muted-foreground flex items-center gap-1 text-[11px] shrink min-w-0 truncate">
                            <CalendarDays className="size-3 shrink-0" />
                            <span className="truncate">Active {formatLocalDate(opp.last_activity_at)}</span>
                          </span>

                          <div className="flex items-center gap-1 shrink-0">
                            <label className="outcome-select-label flex items-center">
                              <span className="sr-only">Set {opp.role_title} Outcome</span>
                              <select
                                aria-label={`Set ${opp.role_title} Outcome`}
                                value={opp.outcome}
                                onChange={(e) =>
                                  void handleOutcomeChange(opp, e.target.value as Outcome)
                                }
                                className="h-7 text-xs rounded-md bg-background border border-input px-1.5 py-0.5 text-foreground cursor-pointer focus:outline-none focus:ring-1 focus:ring-ring font-medium max-w-[95px]"
                              >
                                {OUTCOMES.map((o) => (
                                  <option key={o} value={o}>
                                    {o}
                                  </option>
                                ))}
                              </select>
                            </label>

                            <Link
                              to={`/opportunities/${opp.id}`}
                              className="p-1 rounded-md text-muted-foreground hover:text-primary hover:bg-accent transition-colors shrink-0"
                              title="View Opportunity details"
                              aria-label="View Details"
                            >
                              <ArrowUpRight className="size-3.5" />
                            </Link>
                          </div>
                        </div>
                      </article>
                    );
                  })
                )}
              </div>
            </section>
          );
        })}
      </div>

      {/* Modern floating toast feedback */}
      {toast && (
        <div
          className="cockpit-toast fixed bottom-5 right-5 z-50 flex items-center gap-3 bg-popover text-popover-foreground border border-border shadow-xl px-4 py-3 rounded-xl animate-in fade-in slide-in-from-bottom-2"
          role="status"
          aria-live="polite"
        >
          <div className="size-6 rounded-full bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0">
            <Check className="size-3.5 stroke-[2.5]" />
          </div>
          <span className="text-sm font-medium">{toast.message}</span>
          {toast.onUndo && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="toast-undo-btn h-7 px-2.5 text-xs font-semibold text-primary hover:bg-primary/10 ml-1"
              onClick={async () => {
                const undoFn = toast.onUndo;
                setToast(null);
                if (undoFn) {
                  await undoFn();
                }
              }}
            >
              Undo
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
