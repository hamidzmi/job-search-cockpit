import type {
  JobPostingInput,
  Opportunity,
  OpportunityDetailsInput,
  Priority,
  RemotePolicy,
} from '../../api/types';

export interface OpportunityDetailsDraft {
  roleTitle: string;
  seniority: string;
  location: string;
  remotePolicy: RemotePolicy;
  priority: Priority;
  notes: string;
  sourceUrl: string;
  rawText: string;
  requirementsText: string;
  techStackText: string;
  statedMinComp: string;
  statedMaxComp: string;
  currency: string;
}

export function createEmptyDetailsDraft(
  priority: Priority = 'High',
  remotePolicy: RemotePolicy = 'Remote',
): OpportunityDetailsDraft {
  return {
    roleTitle: '',
    seniority: '',
    location: '',
    remotePolicy,
    priority,
    notes: '',
    sourceUrl: '',
    rawText: '',
    requirementsText: '',
    techStackText: '',
    statedMinComp: '',
    statedMaxComp: '',
    currency: 'USD',
  };
}

export function draftFromOpportunity(opp: Opportunity): OpportunityDetailsDraft {
  return {
    roleTitle: opp.role_title,
    seniority: opp.seniority,
    location: opp.location,
    remotePolicy: opp.remote_policy || 'Remote',
    priority: opp.priority || 'Medium',
    notes: opp.notes,
    sourceUrl: opp.job_posting.source_url,
    rawText: opp.job_posting.raw_text,
    requirementsText: opp.job_posting.requirements.join('\n'),
    techStackText: opp.job_posting.tech_stack_tags.join(', '),
    statedMinComp: opp.job_posting.stated_min_comp
      ? String(opp.job_posting.stated_min_comp)
      : '',
    statedMaxComp: opp.job_posting.stated_max_comp
      ? String(opp.job_posting.stated_max_comp)
      : '',
    currency: opp.job_posting.currency || 'USD',
  };
}

export function buildJobPostingInput(
  draft: OpportunityDetailsDraft,
  includeJobPosting = true,
): JobPostingInput {
  if (!includeJobPosting) {
    return {
      source_url: '',
      raw_text: '',
      requirements: [],
      tech_stack_tags: [],
      stated_min_comp: 0,
      stated_max_comp: 0,
      currency: '',
    };
  }
  const requirements = draft.requirementsText
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean);
  const techStackTags = draft.techStackText
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);

  return {
    source_url: draft.sourceUrl,
    raw_text: draft.rawText,
    requirements,
    tech_stack_tags: techStackTags,
    stated_min_comp: draft.statedMinComp ? Number(draft.statedMinComp) : 0,
    stated_max_comp: draft.statedMaxComp ? Number(draft.statedMaxComp) : 0,
    currency: draft.currency,
  };
}

export function buildOpportunityDetailsInput(
  companyId: string,
  draft: OpportunityDetailsDraft,
  includeJobPosting = true,
): OpportunityDetailsInput {
  return {
    company_id: companyId,
    role_title: draft.roleTitle,
    seniority: draft.seniority,
    location: draft.location,
    remote_policy: draft.remotePolicy,
    priority: draft.priority,
    notes: draft.notes,
    job_posting: buildJobPostingInput(draft, includeJobPosting),
  };
}
