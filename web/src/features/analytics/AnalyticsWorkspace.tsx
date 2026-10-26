import { useQuery } from '@tanstack/react-query';
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useCockpitApi } from '../../api/CockpitApiContext';
import { formatLocalDate } from '../../lib/date';

type ActiveTab = 'comparison' | 'funnel';

export function AnalyticsWorkspace() {
  const api = useCockpitApi();
  const [activeTab, setActiveTab] = useState<ActiveTab>('comparison');

  const offersQuery = useQuery({
    queryKey: ['offersComparison'],
    queryFn: () => api.offers.getOfferComparison(),
  });

  const funnelQuery = useQuery({
    queryKey: ['funnelAnalytics'],
    queryFn: () => api.analytics.getFunnelAnalytics(),
  });

  const comparison = offersQuery.data;
  const funnel = funnelQuery.data;

  const totalOpps = funnel?.total_opportunities ?? 0;
  const activeOpps = funnel?.active_opportunities ?? 0;
  const offersCount = comparison?.offers?.length ?? 0;

  return (
    <div className="workspace-container">
      <header className="workspace-header">
        <div>
          <h2 className="workspace-title">Analytics &amp; Offers</h2>
          <p className="workspace-subtitle">
            Side-by-side Offer comparison against your compensation floor and conversion analytics computed from Stage Transitions.
          </p>
        </div>
        <div className="workspace-actions">
          <div className="tab-pill-group" role="tablist">
            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'comparison'}
              className={`btn-tab ${activeTab === 'comparison' ? 'active' : ''}`}
              onClick={() => setActiveTab('comparison')}
            >
              Offers Comparison ({offersCount})
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'funnel'}
              className={`btn-tab ${activeTab === 'funnel' ? 'active' : ''}`}
              onClick={() => setActiveTab('funnel')}
            >
              Funnel Analytics
            </button>
          </div>
        </div>
      </header>

      {/* Offers Comparison View */}
      {activeTab === 'comparison' && (
        <div className="analytics-content">
          <section className="panel-card panel-card-full" aria-label="Compensation Floor Reference">
            <div className="comp-floor-banner">
              <div>
                <span className="metric-label">Candidate Compensation Floor</span>
                <strong className="metric-value">
                  {comparison?.profile_currency || 'USD'}{' '}
                  {comparison?.compensation_floor ? comparison.compensation_floor.toLocaleString() : 'Not Set'}
                </strong>
                {Boolean(comparison?.target_compensation) && (
                  <span className="metric-sub">
                    (Target: {comparison?.profile_currency || 'USD'} {comparison?.target_compensation?.toLocaleString()})
                  </span>
                )}
              </div>
              <Link to="/profile" className="btn-secondary">
                Edit Floor in Profile
              </Link>
            </div>
          </section>

          {offersQuery.isLoading && <p>Loading offer comparisons...</p>}

          {!offersQuery.isLoading && offersCount === 0 && (
            <div className="empty-state">
              <h3>No Offers to Compare</h3>
              <p>
                When you receive structured job offers, record them inside the Opportunity workspace. They will automatically be normalized, annualized, and compared here side by side.
              </p>
              <Link to="/opportunities" className="btn-primary">
                Go to Pipeline
              </Link>
            </div>
          )}

          {!offersQuery.isLoading && offersCount > 0 && (
            <section className="panel-card panel-card-full" aria-label="Side-by-Side Offers Comparison">
              <h3>Side-by-Side Offer Comparison</h3>
              <div className="comparison-table-wrapper">
                <table className="comparison-table">
                  <thead>
                    <tr>
                      <th scope="col">Compensation Component</th>
                      {comparison?.offers.map((o) => (
                        <th key={o.offer.id} scope="col" className="offer-col-header">
                          <div className="offer-header-title">
                            <Link to={`/opportunities/${o.offer.opportunity_id}`} className="company-link">
                              {o.company_name}
                            </Link>
                            <span className="role-subtitle">{o.opportunity_role_title}</span>
                          </div>
                          <span className={`status-pill status-${o.offer.status.toLowerCase()}`}>
                            {o.offer.status}
                          </span>
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <th scope="row">Base Salary</th>
                      {comparison?.offers.map((o) => (
                        <td key={o.offer.id}>
                          {o.offer.currency} {o.offer.base_salary.toLocaleString()}
                        </td>
                      ))}
                    </tr>

                    <tr>
                      <th scope="row">Annualized Bonus</th>
                      {comparison?.offers.map((o) => (
                        <td key={o.offer.id}>
                          <strong>{o.offer.currency} {o.offer.annual_bonus.toLocaleString()}</strong>
                          {o.offer.bonus_amount > 0 ? (
                            <span className="table-subtext">Flat ${o.offer.bonus_amount.toLocaleString()}</span>
                          ) : o.offer.bonus_target_pct > 0 ? (
                            <span className="table-subtext">{o.offer.bonus_target_pct}% target</span>
                          ) : (
                            <span className="table-subtext">—</span>
                          )}
                        </td>
                      ))}
                    </tr>

                    <tr>
                      <th scope="row">Annualized Equity</th>
                      {comparison?.offers.map((o) => (
                        <td key={o.offer.id}>
                          <strong>{o.offer.currency} {o.offer.annual_equity.toLocaleString()}</strong>
                          {o.offer.equity_total_value > 0 ? (
                            <span className="table-subtext">
                              ${o.offer.equity_total_value.toLocaleString()} {o.offer.equity_type} ({o.offer.equity_vesting_years} yrs)
                            </span>
                          ) : (
                            <span className="table-subtext">—</span>
                          )}
                        </td>
                      ))}
                    </tr>

                    <tr>
                      <th scope="row">Sign-On Bonus</th>
                      {comparison?.offers.map((o) => (
                        <td key={o.offer.id}>
                          {o.offer.sign_on_bonus > 0 ? (
                            `${o.offer.currency} ${o.offer.sign_on_bonus.toLocaleString()}`
                          ) : (
                            '—'
                          )}
                        </td>
                      ))}
                    </tr>

                    <tr className="highlight-row">
                      <th scope="row">Recurring Annual Total Comp</th>
                      {comparison?.offers.map((o) => (
                        <td key={o.offer.id} className="highlight-cell">
                          <strong>{o.offer.currency} {o.offer.recurring_total_comp.toLocaleString()}/yr</strong>
                        </td>
                      ))}
                    </tr>

                    <tr>
                      <th scope="row">Year 1 Total Comp</th>
                      {comparison?.offers.map((o) => (
                        <td key={o.offer.id}>
                          <strong>{o.offer.currency} {o.offer.year1_total_comp.toLocaleString()}</strong>
                        </td>
                      ))}
                    </tr>

                    <tr>
                      <th scope="row">Delta vs Comp Floor</th>
                      {comparison?.offers.map((o) => {
                        const delta = o.offer.delta_vs_comp_floor;
                        return (
                          <td
                            key={o.offer.id}
                            className={delta >= 0 ? 'text-positive' : 'text-negative'}
                          >
                            <strong>
                              {delta >= 0 ? '+' : ''}
                              {o.offer.currency} {delta.toLocaleString()}
                            </strong>
                          </td>
                        );
                      })}
                    </tr>

                    <tr>
                      <th scope="row">Decision Deadline</th>
                      {comparison?.offers.map((o) => (
                        <td key={o.offer.id}>
                          {o.offer.deadline_at ? formatLocalDate(o.offer.deadline_at) : '—'}
                        </td>
                      ))}
                    </tr>

                    <tr>
                      <th scope="row">Benefits Notes</th>
                      {comparison?.offers.map((o) => (
                        <td key={o.offer.id} className="notes-cell">
                          {o.offer.benefits_notes || '—'}
                        </td>
                      ))}
                    </tr>

                    <tr>
                      <th scope="row">Negotiation Notes</th>
                      {comparison?.offers.map((o) => (
                        <td key={o.offer.id} className="notes-cell">
                          {o.offer.negotiation_notes || '—'}
                        </td>
                      ))}
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>
          )}
        </div>
      )}

      {/* Funnel Analytics View */}
      {activeTab === 'funnel' && (
        <div className="analytics-content">
          <div className="metrics-summary-row">
            <div className="metric-box">
              <span className="metric-label">Total Opportunities</span>
              <strong className="metric-value">{totalOpps}</strong>
            </div>
            <div className="metric-box">
              <span className="metric-label">Active Opportunities</span>
              <strong className="metric-value">{activeOpps}</strong>
            </div>
            <div className="metric-box">
              <span className="metric-label">Offers Received</span>
              <strong className="metric-value">
                {funnel?.stages?.find((s) => s.stage === 'Offer')?.reached_count ?? 0}
              </strong>
            </div>
          </div>

          {funnelQuery.isLoading && <p>Loading funnel analytics...</p>}

          {!funnelQuery.isLoading && (
            <section className="panel-card panel-card-full" aria-label="Pipeline Funnel Conversion">
              <h3>Stage-by-Stage Pipeline Funnel</h3>
              <p className="panel-hint">
                Conversion rates, drop-offs, and average days spent per Stage computed purely from immutable Stage Transitions.
              </p>

              <div className="funnel-table-wrapper">
                <table className="comparison-table funnel-table">
                  <thead>
                    <tr>
                      <th scope="col">Stage</th>
                      <th scope="col">Opportunities Reached</th>
                      <th scope="col">Conversion from Prev Stage</th>
                      <th scope="col">Drop-Off Count</th>
                      <th scope="col">Drop-Off Outcomes</th>
                      <th scope="col">Avg Days in Stage</th>
                    </tr>
                  </thead>
                  <tbody>
                    {funnel?.stages?.map((st) => (
                      <tr key={st.stage}>
                        <td>
                          <strong>{st.stage}</strong>
                        </td>
                        <td>
                          <span className="stage-reach-pill">{st.reached_count}</span>
                        </td>
                        <td>
                          <div className="conversion-cell">
                            <span className="conversion-pct">
                              {st.conversion_from_previous_pct.toFixed(1)}%
                            </span>
                            <div className="conversion-bar-bg">
                              <div
                                className="conversion-bar-fill"
                                style={{ width: `${Math.min(100, st.conversion_from_previous_pct)}%` }}
                              />
                            </div>
                          </div>
                        </td>
                        <td>
                          {st.drop_off_count > 0 ? (
                            <span className="drop-off-pill">{st.drop_off_count}</span>
                          ) : (
                            '0'
                          )}
                        </td>
                        <td>
                          {Object.entries(st.outcome_breakdown).filter(([k]) => k !== 'Active').length > 0 ? (
                            <div className="outcome-tag-row">
                              {Object.entries(st.outcome_breakdown)
                                .filter(([outcome]) => outcome !== 'Active')
                                .map(([outcome, count]) => (
                                  <span key={outcome} className="outcome-tag">
                                    {outcome}: {count}
                                  </span>
                                ))}
                            </div>
                          ) : (
                            '—'
                          )}
                        </td>
                        <td>
                          <strong>{st.average_days_in_stage.toFixed(1)}</strong> days
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  );
}
