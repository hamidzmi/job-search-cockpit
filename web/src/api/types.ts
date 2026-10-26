export const STAGES = [
  'Discovered',
  'Outreach',
  'Applied',
  'Screening',
  'Interviewing',
  'Offer',
] as const;

export type Stage = (typeof STAGES)[number];

export const OUTCOMES = [
  'Active',
  'On Hold',
  'Won',
  'Declined',
  'Rejected',
  'Withdrawn',
  'Ghosted',
] as const;

export type Outcome = (typeof OUTCOMES)[number];

export const PRIORITIES = ['High', 'Medium', 'Low'] as const;

export type Priority = (typeof PRIORITIES)[number];

export const REMOTE_POLICIES = ['Remote', 'Hybrid', 'On-site'] as const;

export type RemotePolicy = (typeof REMOTE_POLICIES)[number];

export interface Company {
  id: string;
  name: string;
  domain: string;
  headquarters: string;
  remote_policy: string;
  size_stage: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface CompanyUpsertInput {
  name: string;
  domain: string;
  headquarters: string;
  remote_policy: string;
  size_stage: string;
  notes: string;
}

export interface JobPosting {
  source_url: string;
  raw_text: string;
  requirements: string[];
  tech_stack_tags: string[];
  stated_min_comp: number;
  stated_max_comp: number;
  currency: string;
  has_content: boolean;
}

export interface JobPostingInput {
  source_url: string;
  raw_text: string;
  requirements: string[];
  tech_stack_tags: string[];
  stated_min_comp: number;
  stated_max_comp: number;
  currency: string;
}

export interface StageTransition {
  id: string;
  opportunity_id: string;
  from_stage: Stage | '';
  to_stage: Stage;
  from_outcome: Outcome | '';
  to_outcome: Outcome;
  note: string;
  transitioned_at: string;
}

export interface Opportunity {
  id: string;
  company_id: string;
  company_name: string;
  role_title: string;
  seniority: string;
  location: string;
  remote_policy: RemotePolicy | '';
  priority: Priority;
  stage: Stage;
  outcome: Outcome;
  notes: string;
  job_posting: JobPosting;
  stage_transitions?: StageTransition[];
  last_activity_at: string;
  created_at: string;
  updated_at: string;
}

export interface OpportunityDetailsInput {
  company_id: string;
  role_title: string;
  seniority: string;
  location: string;
  remote_policy: RemotePolicy | '';
  priority: Priority;
  notes: string;
  job_posting: JobPostingInput;
}

export type OpportunityUpdateInput = OpportunityDetailsInput;

export interface OpportunityCreateInput extends OpportunityDetailsInput {
  stage?: Stage;
  transition_note: string;
}

export interface OpportunityTransitionInput {
  stage?: Stage;
  outcome?: Outcome;
  note: string;
}

export interface OpportunityFilter {
  stage?: Stage | '';
  outcome?: Outcome | '';
  priority?: Priority | '';
  remote_policy?: RemotePolicy | '';
  company_id?: string;
  q?: string;
}

export interface WorkHistoryEntry {
  company: string;
  title: string;
  start_date: string;
  end_date: string;
  summary: string;
  impact_bullets: string[];
}

export interface CandidateProfile {
  full_name: string;
  headline: string;
  summary: string;
  work_history: WorkHistoryEntry[];
  skills: string[];
  target_role_titles: string[];
  location_preferences: string[];
  remote_preference: string;
  compensation_floor: number;
  target_compensation: number;
  currency: string;
  updated_at: string;
}

export interface CandidateProfileUpsertInput {
  full_name: string;
  headline: string;
  summary: string;
  work_history: WorkHistoryEntry[];
  skills: string[];
  target_role_titles: string[];
  location_preferences: string[];
  remote_preference: string;
  compensation_floor: number;
  target_compensation: number;
  currency: string;
}

export interface Story {
  id: string;
  title: string;
  hook: string;
  situation: string;
  task: string;
  action: string;
  result: string;
  competency_tags: string[];
  skill_tags: string[];
  created_at: string;
  updated_at: string;
}

export interface StoryUpsertInput {
  title: string;
  hook: string;
  situation: string;
  task: string;
  action: string;
  result: string;
  competency_tags: string[];
  skill_tags: string[];
}

export interface StoryFilter {
  q?: string;
  competency_tag?: string;
  skill_tag?: string;
}
