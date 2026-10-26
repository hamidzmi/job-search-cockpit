import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import type {
  CompanyUpsertInput,
  OpportunityCreateInput,
  OpportunityFilter,
  Stage,
} from '../../api/types';
import { CompanyFormModal } from './CompanyFormModal';
import { KanbanBoard } from './KanbanBoard';
import { OpportunityFormModal } from './OpportunityFormModal';
import { OpportunityTable } from './OpportunityTable';

type ViewMode = 'board' | 'table';

export function PipelineWorkspace() {
  const api = useCockpitApi();
  const queryClient = useQueryClient();
  const [viewMode, setViewMode] = useState<ViewMode>('board');
  const [filter, setFilter] = useState<OpportunityFilter>({});
  const [error, setError] = useState<string | null>(null);

  const [showNewCompany, setShowNewCompany] = useState(false);
  const [showNewOpportunity, setShowNewOpportunity] = useState(false);
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

  const companies = companiesQuery.data ?? [];
  const opportunities = opportunitiesQuery.data ?? [];
  const loading = companiesQuery.isLoading || opportunitiesQuery.isLoading;
  const displayError =
    error ??
    (companiesQuery.error instanceof Error
      ? companiesQuery.error.message
      : opportunitiesQuery.error instanceof Error
        ? opportunitiesQuery.error.message
        : null);

  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Pipeline</h2>
          <p className="workspace-subtitle">
            Track active Opportunities across 6 company-agnostic macro Stages and inspect all Outcomes.
          </p>
        </div>
        <div className="workspace-actions">
          <div className="view-toggle" role="group" aria-label="Pipeline View Mode">
            <button
              type="button"
              className={`btn-toggle ${viewMode === 'board' ? 'btn-toggle-active' : ''}`}
              onClick={() => setViewMode('board')}
            >
              Kanban Board
            </button>
            <button
              type="button"
              className={`btn-toggle ${viewMode === 'table' ? 'btn-toggle-active' : ''}`}
              onClick={() => setViewMode('table')}
            >
              Table View
            </button>
          </div>
          <button
            type="button"
            className="btn-primary"
            onClick={() => setShowNewOpportunity(true)}
          >
            + New Opportunity
          </button>
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
          opportunities={opportunities}
          onMoveStage={handleMoveStage}
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
          onRequestNewCompany={() => setShowNewCompany(true)}
          onClose={() => setShowNewOpportunity(false)}
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
