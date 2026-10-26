import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useCockpitApi } from '../../api/CockpitApiContext';
import type { Company, CompanyUpsertInput } from '../../api/types';
import { CompanyFormModal } from '../pipeline/CompanyFormModal';

export function NetworkWorkspace() {
  const api = useCockpitApi();
  const queryClient = useQueryClient();
  const [editingCompany, setEditingCompany] = useState<Company | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);

  const companiesQuery = useQuery({
    queryKey: ['companies'],
    queryFn: () => api.companies.listCompanies(),
  });

  const opportunitiesQuery = useQuery({
    queryKey: ['opportunities', {}],
    queryFn: () => api.opportunities.listOpportunities(),
  });

  const saveCompanyMutation = useMutation({
    mutationFn: async (input: CompanyUpsertInput) => {
      if (editingCompany) {
        return api.companies.updateCompany(editingCompany.id, input);
      }
      return api.companies.createCompany(input);
    },
    onSuccess: async () => {
      setEditingCompany(null);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['companies'] }),
        queryClient.invalidateQueries({ queryKey: ['opportunities'] }),
      ]);
    },
  });

  const handleSaveCompany = async (input: CompanyUpsertInput) => {
    await saveCompanyMutation.mutateAsync(input);
  };

  const companies = companiesQuery.data ?? [];
  const opportunities = opportunitiesQuery.data ?? [];
  const loading = companiesQuery.isLoading || opportunitiesQuery.isLoading;

  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Network &amp; Companies</h2>
          <p className="workspace-subtitle">
            Manage target Companies and inspect every Opportunity pursued at each Company.
          </p>
        </div>
        <div className="workspace-actions">
          <button
            type="button"
            className="btn-primary"
            onClick={() => setShowCreateModal(true)}
          >
            + New Company
          </button>
        </div>
      </header>

      {loading ? (
        <p>Loading Companies...</p>
      ) : companies.length === 0 ? (
        <p className="empty-state">No Companies added yet.</p>
      ) : (
        <div className="company-grid">
          {companies.map((c) => {
            const companyOpps = opportunities.filter((o) => o.company_id === c.id);
            return (
              <article key={c.id} className="panel-card">
                <div className="card-top-row">
                  <h3>{c.name}</h3>
                  <button
                    type="button"
                    className="btn-secondary"
                    onClick={() => setEditingCompany(c)}
                  >
                    Edit Company
                  </button>
                </div>
                <div className="card-meta">
                  {c.domain && <span>{c.domain}</span>}
                  {c.headquarters && <span>{c.headquarters}</span>}
                  {c.remote_policy && <span>{c.remote_policy}</span>}
                  {c.size_stage && <span>{c.size_stage}</span>}
                </div>
                {c.notes && <p className="company-notes">{c.notes}</p>}
                <div className="company-opportunities">
                  <h4>Opportunities ({companyOpps.length})</h4>
                  {companyOpps.length === 0 ? (
                    <p className="timestamp-text">No opportunities tracked yet.</p>
                  ) : (
                    <ul>
                      {companyOpps.map((opp) => (
                        <li key={opp.id}>
                          <Link to={`/opportunities/${opp.id}`}>{opp.role_title}</Link>{' '}
                          <span className="stage-badge">{opp.stage}</span>{' '}
                          <span className="outcome-pill">{opp.outcome}</span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </article>
            );
          })}
        </div>
      )}

      {(showCreateModal || editingCompany) && (
        <CompanyFormModal
          initial={editingCompany}
          onClose={() => {
            setShowCreateModal(false);
            setEditingCompany(null);
          }}
          onSubmit={handleSaveCompany}
        />
      )}
    </div>
  );
}
