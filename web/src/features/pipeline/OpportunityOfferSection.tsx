import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useState } from 'react';
import { useCockpitApi } from '../../api/CockpitApiContext';
import {
  OFFER_STATUSES,
  type OfferStatus,
  type OfferUpsertInput,
} from '../../api/types';
import { formatLocalDate } from '../../lib/date';

interface OpportunityOfferSectionProps {
  opportunityId: string;
}

export function OpportunityOfferSection({ opportunityId }: OpportunityOfferSectionProps) {
  const api = useCockpitApi();
  const queryClient = useQueryClient();

  const [isEditing, setIsEditing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const offerQuery = useQuery({
    queryKey: ['opportunityOffer', opportunityId],
    queryFn: async () => {
      try {
        return await api.offers.getOffer(opportunityId);
      } catch (err: any) {
        if (err?.status === 404 || err?.message?.includes('not found')) {
          return null;
        }
        throw err;
      }
    },
  });

  const [baseSalary, setBaseSalary] = useState<number>(0);
  const [bonusTargetPct, setBonusTargetPct] = useState<number>(0);
  const [bonusAmount, setBonusAmount] = useState<number>(0);
  const [equityType, setEquityType] = useState('RSU');
  const [equityTotalValue, setEquityTotalValue] = useState<number>(0);
  const [equityVestingYears, setEquityVestingYears] = useState<number>(4);
  const [signOnBonus, setSignOnBonus] = useState<number>(0);
  const [currency, setCurrency] = useState('USD');
  const [deadlineAt, setDeadlineAt] = useState('');
  const [benefitsNotes, setBenefitsNotes] = useState('');
  const [negotiationNotes, setNegotiationNotes] = useState('');
  const [status, setStatus] = useState<OfferStatus>('Pending');

  const startEdit = () => {
    if (offerQuery.data) {
      const o = offerQuery.data.offer;
      setBaseSalary(o.base_salary);
      setBonusTargetPct(o.bonus_target_pct);
      setBonusAmount(o.bonus_amount);
      setEquityType(o.equity_type || 'RSU');
      setEquityTotalValue(o.equity_total_value);
      setEquityVestingYears(o.equity_vesting_years || 4);
      setSignOnBonus(o.sign_on_bonus);
      setCurrency(o.currency || 'USD');
      setDeadlineAt(o.deadline_at ? o.deadline_at.slice(0, 10) : '');
      setBenefitsNotes(o.benefits_notes || '');
      setNegotiationNotes(o.negotiation_notes || '');
      setStatus(o.status || 'Pending');
    } else {
      setBaseSalary(0);
      setBonusTargetPct(0);
      setBonusAmount(0);
      setEquityType('RSU');
      setEquityTotalValue(0);
      setEquityVestingYears(4);
      setSignOnBonus(0);
      setCurrency('USD');
      setDeadlineAt('');
      setBenefitsNotes('');
      setNegotiationNotes('');
      setStatus('Pending');
    }
    setError(null);
    setSuccessMessage(null);
    setIsEditing(true);
  };

  const saveMutation = useMutation({
    mutationFn: (input: OfferUpsertInput) =>
      api.offers.upsertOffer(opportunityId, input),
    onSuccess: (updated) => {
      queryClient.setQueryData(['opportunityOffer', opportunityId], updated);
      queryClient.invalidateQueries({ queryKey: ['opportunityOffer', opportunityId] });
      queryClient.invalidateQueries({ queryKey: ['offersComparison'] });
      setIsEditing(false);
      setSuccessMessage('Offer saved successfully.');
      setError(null);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to save offer');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => api.offers.deleteOffer(opportunityId),
    onSuccess: () => {
      queryClient.setQueryData(['opportunityOffer', opportunityId], null);
      queryClient.invalidateQueries({ queryKey: ['opportunityOffer', opportunityId] });
      queryClient.invalidateQueries({ queryKey: ['offersComparison'] });
      setIsEditing(false);
      setSuccessMessage('Offer removed.');
      setError(null);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to delete offer');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (baseSalary <= 0) {
      setError('Base salary must be greater than zero.');
      return;
    }
    let formattedDeadline: string | undefined = undefined;
    if (deadlineAt) {
      formattedDeadline = new Date(deadlineAt).toISOString();
    }
    saveMutation.mutate({
      base_salary: baseSalary,
      bonus_target_pct: bonusTargetPct,
      bonus_amount: bonusAmount,
      equity_type: equityType,
      equity_total_value: equityTotalValue,
      equity_vesting_years: equityVestingYears || 4,
      sign_on_bonus: signOnBonus,
      currency: currency || 'USD',
      deadline_at: formattedDeadline,
      benefits_notes: benefitsNotes,
      negotiation_notes: negotiationNotes,
      status,
    });
  };

  const offerView = offerQuery.data;
  const offer = offerView?.offer;

  return (
    <section className="panel-card panel-card-full" aria-label="Structured Offer">
      <div className="section-header">
        <div>
          <h3>Structured Offer &amp; Compensation</h3>
          <p className="panel-hint">
            Record and evaluate structured compensation, annualized bonus and equity, decision deadlines, and delta against your compensation floor.
          </p>
        </div>
        {!isEditing && (
          <button
            type="button"
            className="btn-primary"
            onClick={startEdit}
          >
            {offer ? 'Edit Offer' : 'Record Offer'}
          </button>
        )}
      </div>

      {error && (
        <div className="alert-error" role="alert">
          {error}
        </div>
      )}
      {successMessage && <div className="alert-success">{successMessage}</div>}

      {offerQuery.isLoading && <p>Loading offer details...</p>}

      {!isEditing && !offer && !offerQuery.isLoading && (
        <div className="empty-state">
          <p>No offer recorded for this opportunity yet.</p>
          <button
            type="button"
            className="btn-secondary"
            onClick={startEdit}
          >
            Add Structured Offer
          </button>
        </div>
      )}

      {!isEditing && offer && (
        <div className="offer-summary-grid">
          <div className="metrics-summary-row">
            <div className="metric-box">
              <span className="metric-label">Status</span>
              <span className={`status-pill status-${offer.status.toLowerCase()}`}>
                {offer.status}
              </span>
            </div>
            <div className="metric-box">
              <span className="metric-label">Base Salary</span>
              <strong className="metric-value">
                {offer.currency} {offer.base_salary.toLocaleString()}
              </strong>
            </div>
            <div className="metric-box">
              <span className="metric-label">Annualized Bonus</span>
              <strong className="metric-value">
                {offer.currency} {offer.annual_bonus.toLocaleString()}
              </strong>
              {offer.bonus_amount > 0 ? (
                <span className="metric-sub">(Flat ${offer.bonus_amount.toLocaleString()})</span>
              ) : offer.bonus_target_pct > 0 ? (
                <span className="metric-sub">({offer.bonus_target_pct}% target)</span>
              ) : null}
            </div>
            <div className="metric-box">
              <span className="metric-label">Annualized Equity</span>
              <strong className="metric-value">
                {offer.currency} {offer.annual_equity.toLocaleString()}
              </strong>
              {offer.equity_total_value > 0 && (
                <span className="metric-sub">
                  (${offer.equity_total_value.toLocaleString()} / {offer.equity_vesting_years} yrs)
                </span>
              )}
            </div>
            <div className="metric-box highlight-box">
              <span className="metric-label">Recurring Total Comp</span>
              <strong className="metric-value">
                {offer.currency} {offer.recurring_total_comp.toLocaleString()}/yr
              </strong>
            </div>
            <div className="metric-box">
              <span className="metric-label">Year 1 Total Comp</span>
              <strong className="metric-value">
                {offer.currency} {offer.year1_total_comp.toLocaleString()}
              </strong>
              {offer.sign_on_bonus > 0 && (
                <span className="metric-sub">(+${offer.sign_on_bonus.toLocaleString()} sign-on)</span>
              )}
            </div>
            <div className="metric-box">
              <span className="metric-label">Delta vs Floor</span>
              <strong
                className={`metric-value ${
                  offer.delta_vs_comp_floor >= 0 ? 'text-positive' : 'text-negative'
                }`}
              >
                {offer.delta_vs_comp_floor >= 0 ? '+' : ''}
                {offer.currency} {offer.delta_vs_comp_floor.toLocaleString()}
              </strong>
              <span className="metric-sub">
                (Floor: ${offerView.compensation_floor.toLocaleString()})
              </span>
            </div>
          </div>

          <div className="offer-notes-row">
            {offer.deadline_at && (
              <div className="note-block">
                <strong>Decision Deadline:</strong>{' '}
                <span>{formatLocalDate(offer.deadline_at)}</span>
              </div>
            )}
            {offer.benefits_notes && (
              <div className="note-block">
                <strong>Benefits &amp; Perks:</strong>
                <p>{offer.benefits_notes}</p>
              </div>
            )}
            {offer.negotiation_notes && (
              <div className="note-block">
                <strong>Negotiation Log:</strong>
                <p>{offer.negotiation_notes}</p>
              </div>
            )}
          </div>
        </div>
      )}

      {isEditing && (
        <form onSubmit={handleSubmit} className="form-grid">
          <label className="form-field">
            <span>Base Salary *</span>
            <input
              type="number"
              aria-label="Base Salary"
              value={baseSalary || ''}
              onChange={(e) => setBaseSalary(Number(e.target.value))}
              placeholder="e.g. 210000"
              required
            />
          </label>

          <label className="form-field">
            <span>Bonus Target %</span>
            <input
              type="number"
              aria-label="Bonus Target Percentage"
              value={bonusTargetPct || ''}
              onChange={(e) => setBonusTargetPct(Number(e.target.value))}
              placeholder="e.g. 15"
            />
          </label>

          <label className="form-field">
            <span>Explicit Bonus Amount ($ override)</span>
            <input
              type="number"
              aria-label="Bonus Amount"
              value={bonusAmount || ''}
              onChange={(e) => setBonusAmount(Number(e.target.value))}
              placeholder="e.g. 30000"
            />
          </label>

          <label className="form-field">
            <span>Equity Type</span>
            <input
              type="text"
              aria-label="Equity Type"
              value={equityType}
              onChange={(e) => setEquityType(e.target.value)}
              placeholder="RSU, Stock Options"
            />
          </label>

          <label className="form-field">
            <span>Equity Grant Total Value ($)</span>
            <input
              type="number"
              aria-label="Equity Total Value"
              value={equityTotalValue || ''}
              onChange={(e) => setEquityTotalValue(Number(e.target.value))}
              placeholder="e.g. 200000"
            />
          </label>

          <label className="form-field">
            <span>Vesting Years (defaults to 4)</span>
            <input
              type="number"
              aria-label="Equity Vesting Years"
              value={equityVestingYears || ''}
              onChange={(e) => setEquityVestingYears(Number(e.target.value))}
              placeholder="4"
            />
          </label>

          <label className="form-field">
            <span>Sign-On Bonus ($)</span>
            <input
              type="number"
              aria-label="Sign-On Bonus"
              value={signOnBonus || ''}
              onChange={(e) => setSignOnBonus(Number(e.target.value))}
              placeholder="e.g. 25000"
            />
          </label>

          <label className="form-field">
            <span>Currency</span>
            <input
              type="text"
              aria-label="Currency"
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
              placeholder="USD"
            />
          </label>

          <label className="form-field">
            <span>Decision Deadline</span>
            <input
              type="date"
              aria-label="Decision Deadline"
              value={deadlineAt}
              onChange={(e) => setDeadlineAt(e.target.value)}
            />
          </label>

          <label className="form-field">
            <span>Offer Status</span>
            <select
              aria-label="Offer Status"
              value={status}
              onChange={(e) => setStatus(e.target.value as OfferStatus)}
            >
              {OFFER_STATUSES.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </label>

          <label className="form-field form-field-full">
            <span>Benefits &amp; Perks Notes</span>
            <textarea
              rows={2}
              aria-label="Benefits Notes"
              value={benefitsNotes}
              onChange={(e) => setBenefitsNotes(e.target.value)}
              placeholder="Medical/dental coverage, 401(k) match, PTO policy, home office stipend..."
            />
          </label>

          <label className="form-field form-field-full">
            <span>Negotiation Log &amp; Strategy</span>
            <textarea
              rows={2}
              aria-label="Negotiation Notes"
              value={negotiationNotes}
              onChange={(e) => setNegotiationNotes(e.target.value)}
              placeholder="Counter-offer points, competing offers referenced, recruiter replies..."
            />
          </label>

          <div className="form-field-full form-button-row">
            <button type="submit" className="btn-primary">
              Save Offer
            </button>
            <button
              type="button"
              className="btn-secondary"
              onClick={() => setIsEditing(false)}
            >
              Cancel
            </button>
            {offer && (
              <button
                type="button"
                className="btn-danger"
                onClick={() => {
                  if (window.confirm('Are you sure you want to delete this offer?')) {
                    deleteMutation.mutate();
                  }
                }}
              >
                Delete Offer
              </button>
            )}
          </div>
        </form>
      )}
    </section>
  );
}
