import React from 'react';
import { Link } from 'react-router-dom';
import {
  OUTCOMES,
  PRIORITIES,
  REMOTE_POLICIES,
  STAGES,
  type Company,
  type Opportunity,
  type OpportunityFilter,
  type Outcome,
  type Priority,
  type RemotePolicy,
  type Stage,
} from '../../api/types';
import { formatLocalDateTime } from '../../lib/date';

interface OpportunityTableProps {
  opportunities: Opportunity[];
  companies: Company[];
  filter: OpportunityFilter;
  onFilterChange: (next: OpportunityFilter) => void;
}

export function OpportunityTable({
  opportunities,
  companies,
  filter,
  onFilterChange,
}: OpportunityTableProps) {
  return (
    <section className="table-section" aria-label="Opportunities Filterable Table">
      <div className="filter-bar">
        <label className="filter-control">
          <span>Search</span>
          <input
            type="search"
            aria-label="Search opportunities"
            placeholder="Role, company, tech stack, notes..."
            value={filter.q ?? ''}
            onChange={(e) => onFilterChange({ ...filter, q: e.target.value })}
          />
        </label>
        <label className="filter-control">
          <span>Stage</span>
          <select
            aria-label="Filter by Stage"
            value={filter.stage ?? ''}
            onChange={(e) =>
              onFilterChange({ ...filter, stage: e.target.value as Stage | '' })
            }
          >
            <option value="">All Stages</option>
            {STAGES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </label>
        <label className="filter-control">
          <span>Outcome</span>
          <select
            aria-label="Filter by Outcome"
            value={filter.outcome ?? ''}
            onChange={(e) =>
              onFilterChange({ ...filter, outcome: e.target.value as Outcome | '' })
            }
          >
            <option value="">All Outcomes</option>
            {OUTCOMES.map((o) => (
              <option key={o} value={o}>
                {o}
              </option>
            ))}
          </select>
        </label>
        <label className="filter-control">
          <span>Priority</span>
          <select
            aria-label="Filter by Priority"
            value={filter.priority ?? ''}
            onChange={(e) =>
              onFilterChange({
                ...filter,
                priority: e.target.value as Priority | '',
              })
            }
          >
            <option value="">All Priorities</option>
            {PRIORITIES.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
        </label>
        <label className="filter-control">
          <span>Remote Policy</span>
          <select
            aria-label="Filter by Remote Policy"
            value={filter.remote_policy ?? ''}
            onChange={(e) =>
              onFilterChange({
                ...filter,
                remote_policy: e.target.value as RemotePolicy | '',
              })
            }
          >
            <option value="">All Policies</option>
            {REMOTE_POLICIES.map((policy) => (
              <option key={policy} value={policy}>
                {policy}
              </option>
            ))}
          </select>
        </label>
        <label className="filter-control">
          <span>Company</span>
          <select
            aria-label="Filter by Company"
            value={filter.company_id ?? ''}
            onChange={(e) => onFilterChange({ ...filter, company_id: e.target.value })}
          >
            <option value="">All Companies</option>
            {companies.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </label>
      </div>

      <div className="table-wrapper">
        <table className="cockpit-table">
          <thead>
            <tr>
              <th>Role</th>
              <th>Company</th>
              <th>Stage</th>
              <th>Outcome</th>
              <th>Priority</th>
              <th>Remote Policy</th>
              <th>Job Posting</th>
              <th>Last Activity (Local)</th>
            </tr>
          </thead>
          <tbody>
            {opportunities.length === 0 ? (
              <tr>
                <td colSpan={8} className="empty-table-cell">
                  No matching Opportunities found.
                </td>
              </tr>
            ) : (
              opportunities.map((opp) => (
                <tr key={opp.id}>
                  <td>
                    <Link to={`/opportunities/${opp.id}`} className="table-role-link">
                      {opp.role_title}
                    </Link>
                  </td>
                  <td>{opp.company_name}</td>
                  <td>
                    <span className="stage-badge">{opp.stage}</span>
                  </td>
                  <td>
                    <span
                      className={`outcome-badge outcome-${opp.outcome
                        .toLowerCase()
                        .replace(/\s+/g, '-')}`}
                    >
                      {opp.outcome}
                    </span>
                  </td>
                  <td>{opp.priority}</td>
                  <td>{opp.remote_policy || '—'}</td>
                  <td>{opp.job_posting.has_content ? 'Attached' : 'Unposted'}</td>
                  <td>{formatLocalDateTime(opp.last_activity_at)}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
