import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useMemo, useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import type {
  CompanyUpsertInput,
  OpportunityCreateInput,
  OpportunityFilter,
  Outcome,
  Stage,
} from '../../api/types';
import { CompanyFormModal } from './CompanyFormModal';
import { KanbanBoard } from './KanbanBoard';
import { OpportunityFormModal } from './OpportunityFormModal';
import { OpportunityTable } from './OpportunityTable';
import { Button } from '../../components/ui/Button';
import { Kanban, Table2, Plus, Search, X } from 'lucide-react';
import { cn } from '../../lib/utils';

type ViewMode = 'board' | 'table';

export function PipelineWorkspace() {
  const api = useCockpitApi();
  const queryClient = useQueryClient();
  const [viewMode, setViewMode] = useState<ViewMode>('board');
  const [filter, setFilter] = useState<OpportunityFilter>({});
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState<string | null>(null);

  const [showNewCompany, setShowNewCompany] = useState(false);
  const [showNewOpportunity, setShowNewOpportunity] = useState(false);
  const [newOppStage, setNewOppStage] = useState<Stage | undefined>(undefined);
  const [lastCreatedCompanyId, setLastCreatedCompanyId] = useState<string | undefined>(
    undefined,
  );

  const activeFilter = viewMode === 'table' ? filter : {};

  const companiesQuery = useQuery({
    queryKey: ['companies'],
    queryFn: () => api.companies.listCompanies(),
  });

  const opportunitiesQuery = useQuery({
    queryKey: ['opportunities', activeFilter],
    queryFn: () => api.opportunities.listOpportunities(activeFilter),
  });

  const createCompanyMutation = useMutation({
    mutationFn: (input: CompanyUpsertInput) => api.companies.createCompany(input),
    onSuccess: async (created) => {
      setLastCreatedCompanyId(created.id);
      await queryClient.invalidateQueries({ queryKey: ['companies'] });
    },
  });

  const createOpportunityMutation = useMutation({
    mutationFn: (input: OpportunityCreateInput) => api.opportunities.createOpportunity(input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['opportunities'] });
    },
  });

  const moveStageMutation = useMutation({
    mutationFn: ({ opportunityId, nextStage }: { opportunityId: string; nextStage: Stage }) =>
      api.opportunities.transitionOpportunity(opportunityId, {
        stage: nextStage,
        note: `Moved to ${nextStage} on Kanban board`,
      }),
    onSuccess: async (_, variables) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunity', variables.opportunityId] }),
      ]);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to transition Opportunity');
    },
  });

  const changeOutcomeMutation = useMutation({
    mutationFn: ({
      opportunityId,
      nextOutcome,
      note,
    }: {
      opportunityId: string;
      nextOutcome: Outcome;
      note?: string;
    }) =>
      api.opportunities.transitionOpportunity(opportunityId, {
        outcome: nextOutcome,
        note: note ?? `Marked as ${nextOutcome} on Kanban board`,
      }),
    onSuccess: async (_, variables) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunity', variables.opportunityId] }),
      ]);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to transition Opportunity outcome');
    },
  });

  const handleCreateCompany = async (input: CompanyUpsertInput) => {
    await createCompanyMutation.mutateAsync(input);
  };

  const handleCreateOpportunity = async (input: OpportunityCreateInput) => {
    await createOpportunityMutation.mutateAsync(input);
  };

  const handleMoveStage = async (opportunityId: string, nextStage: Stage) => {
    setError(null);
    await moveStageMutation.mutateAsync({ opportunityId, nextStage });
  };

  const handleChangeOutcome = async (
    opportunityId: string,
    nextOutcome: Outcome,
    note?: string,
  ) => {
    setError(null);
    await changeOutcomeMutation.mutateAsync({ opportunityId, nextOutcome, note });
  };

  const companies = companiesQuery.data ?? [];
  const opportunities = opportunitiesQuery.data ?? [];
  const loading = companiesQuery.isLoading || opportunitiesQuery.isLoading;

  const filteredOpportunities = useMemo(() => {
    if (!searchQuery.trim() || viewMode !== 'board') return opportunities;
    const q = searchQuery.toLowerCase().trim();
    return opportunities.filter(
      (opp) =>
        opp.role_title.toLowerCase().includes(q) ||
        opp.company_name.toLowerCase().includes(q) ||
        opp.notes.toLowerCase().includes(q) ||
        opp.job_posting.tech_stack_tags.some((tag) => tag.toLowerCase().includes(q)),
    );
  }, [opportunities, searchQuery, viewMode]);

  const displayError =
    error ??
    (companiesQuery.error instanceof Error
      ? companiesQuery.error.message
      : opportunitiesQuery.error instanceof Error
        ? opportunitiesQuery.error.message
        : null);

  return (
    <div className="workspace-container">
      {/* Studio Admin Pipeline Toolbar */}
      <header className="workspace-header flex flex-col md:flex-row md:items-center justify-between gap-4 pb-2 border-b border-border/40">
        <div>
          <h2 className="workspace-title text-xl font-bold tracking-tight">Pipeline</h2>
          <p className="workspace-subtitle text-xs text-muted-foreground mt-0.5">
            Track active Opportunities across 6 company-agnostic macro Stages and inspect all Outcomes.
          </p>
        </div>

        <div className="workspace-actions flex flex-wrap items-center gap-2.5">
          {/* Studio Admin View Switcher */}
          <div
            className="view-toggle inline-flex items-center gap-1 bg-muted/60 dark:bg-muted/40 p-1 rounded-lg border border-border"
            role="group"
            aria-label="Pipeline View Mode"
          >
            <button
              type="button"
              className={cn(
                'btn-toggle inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md transition-all cursor-pointer',
                viewMode === 'board'
                  ? 'btn-toggle-active bg-background text-foreground shadow-xs font-semibold'
                  : 'text-muted-foreground hover:text-foreground',
              )}
              onClick={() => setViewMode('board')}
            >
              <Kanban className="size-3.5" />
              <span>Kanban Board</span>
            </button>
            <button
              type="button"
              className={cn(
                'btn-toggle inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md transition-all cursor-pointer',
                viewMode === 'table'
                  ? 'btn-toggle-active bg-background text-foreground shadow-xs font-semibold'
                  : 'text-muted-foreground hover:text-foreground',
              )}
              onClick={() => setViewMode('table')}
            >
              <Table2 className="size-3.5" />
              <span>Table View</span>
            </button>
          </div>

          {/* Quick Search on Board */}
          {viewMode === 'board' && (
            <div className="relative flex items-center">
              <Search className="size-3.5 absolute left-2.5 text-muted-foreground pointer-events-none" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search board..."
                aria-label="Search board opportunities"
                className="h-8 pl-8 pr-7 text-xs rounded-md bg-background border border-input text-foreground w-40 sm:w-48 focus:w-56 transition-all focus:outline-none focus:ring-1 focus:ring-ring"
              />
              {searchQuery && (
                <button
                  type="button"
                  onClick={() => setSearchQuery('')}
                  title="Clear search"
                  aria-label="Clear search"
                  className="absolute right-2 text-muted-foreground hover:text-foreground p-0.5"
                >
                  <X className="size-3" />
                </button>
              )}
            </div>
          )}

          {/* New Opportunity Primary Button */}
          <Button
            type="button"
            className="btn-primary gap-1.5 h-8 text-xs font-semibold"
            onClick={() => {
              setNewOppStage(undefined);
              setShowNewOpportunity(true);
            }}
          >
            <Plus className="size-3.5" />
            <span>+ New Opportunity</span>
          </Button>
        </div>
      </header>

      {displayError && (
        <div className="alert-error" role="alert">
          {displayError}
        </div>
      )}

      {loading ? (
        <p className="loading-state">Loading pipeline...</p>
      ) : viewMode === 'board' ? (
        <KanbanBoard
          opportunities={filteredOpportunities}
          onMoveStage={handleMoveStage}
          onChangeOutcome={handleChangeOutcome}
          onAddOpportunity={(stage) => {
            setNewOppStage(stage);
            setShowNewOpportunity(true);
          }}
        />
      ) : (
        <OpportunityTable
          opportunities={opportunities}
          companies={companies}
          filter={filter}
          onFilterChange={setFilter}
        />
      )}

      {showNewOpportunity && (
        <OpportunityFormModal
          companies={companies}
          selectedCompanyId={lastCreatedCompanyId}
          initialStage={newOppStage}
          onRequestNewCompany={() => setShowNewCompany(true)}
          onClose={() => {
            setShowNewOpportunity(false);
            setNewOppStage(undefined);
          }}
          onSubmit={handleCreateOpportunity}
        />
      )}

      {showNewCompany && (
        <CompanyFormModal
          onClose={() => setShowNewCompany(false)}
          onSubmit={handleCreateCompany}
        />
      )}
    </div>
  );
}
