import { useQuery } from '@tanstack/react-query';
import React, { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useCockpitApi } from '../../api/CockpitApiContext';
import type { QuestionBankFilter } from '../../api/types';
import { formatLocalDateTime } from '../../lib/date';

export function QuestionBankPanel() {
  const api = useCockpitApi();

  const [searchQuery, setSearchQuery] = useState('');
  const [category, setCategory] = useState('');

  const filter: QuestionBankFilter = {
    q: searchQuery.trim() || undefined,
    category: category || undefined,
  };

  const allQuestionsQuery = useQuery({
    queryKey: ['question-bank', {}],
    queryFn: () => api.interviewRounds.listQuestionBank({}),
  });

  const filteredQuestionsQuery = useQuery({
    queryKey: ['question-bank', filter],
    queryFn: () => api.interviewRounds.listQuestionBank(filter),
  });

  const categoryOptions = useMemo(() => {
    const set = new Set<string>();
    for (const q of allQuestionsQuery.data ?? []) {
      if (q.category.trim()) {
        set.add(q.category.trim());
      }
    }
    return Array.from(set).sort((a, b) => a.localeCompare(b));
  }, [allQuestionsQuery.data]);

  const entries = filteredQuestionsQuery.data ?? [];

  return (
    <section className="panel-card" aria-label="Interview Question Bank">
      <div className="card-top-row">
        <div>
          <h3>Interview Question Bank ({entries.length})</h3>
          <p className="timestamp-text">
            Real interview questions aggregated across all recorded Debriefs, linked to Company, Opportunity, Interview Round type, and STAR Story used.
          </p>
        </div>
      </div>

      <div className="filter-bar">
        <label className="filter-control">
          Search Questions
          <input
            type="search"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search question, company, role, or story..."
          />
        </label>

        <label className="filter-control">
          Filter by Category
          <select value={category} onChange={(e) => setCategory(e.target.value)}>
            <option value="">All Categories</option>
            {categoryOptions.map((cat) => (
              <option key={cat} value={cat}>
                {cat}
              </option>
            ))}
          </select>
        </label>
      </div>

      {filteredQuestionsQuery.isLoading ? (
        <p>Loading Question Bank...</p>
      ) : entries.length === 0 ? (
        <p className="timestamp-text">
          No Debrief questions recorded yet. Record a Debrief on an Interview Round to populate your Question Bank.
        </p>
      ) : (
        <ul className="timeline-list">
          {entries.map((entry, idx) => (
            <li key={`${entry.round_id}-${idx}`} className="timeline-item">
              <div className="timeline-header">
                <strong>{entry.question}</strong>
                {entry.category && <span className="tech-tag">{entry.category}</span>}
              </div>
              <div className="card-meta">
                <span className="company-pill">{entry.company_name}</span>
                <Link
                  to={`/opportunities/${entry.opportunity_id}`}
                  className="opp-title-link"
                >
                  {entry.opportunity_role_title}
                </Link>
                <span className="stage-badge">{entry.round_type}</span>
                {entry.answered_with_story_title && (
                  <span className="outcome-pill">
                    Story: {entry.answered_with_story_title}
                  </span>
                )}
                <span className="timestamp-text">
                  {formatLocalDateTime(entry.recorded_at)}
                </span>
              </div>
              {entry.notes && <p className="timeline-note">{entry.notes}</p>}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
