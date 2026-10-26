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

export const AI_PROVIDERS = [
  'openrouter',
  'openai',
  'anthropic',
  'gemini',
  'ollama',
] as const;

export type AIProvider = (typeof AI_PROVIDERS)[number];

export interface AISettings {
  provider: AIProvider;
  base_url: string;
  api_key: string;
  has_api_key: boolean;
  is_configured: boolean;
  fast_extraction_model: string;
  reasoning_tailoring_model: string;
  bookmarked_models: string[];
  updated_at: string;
}

export interface AISettingsUpsertInput {
  provider: AIProvider;
  base_url: string;
  api_key: string;
  fast_extraction_model: string;
  reasoning_tailoring_model: string;
  bookmarked_models: string[];
}

export interface OpenRouterModel {
  id: string;
  name: string;
  description: string;
  context_length: number;
  prompt_pricing: string;
  completion_pricing: string;
}

export interface QuickCaptureExtractInput {
  source_url: string;
  raw_text: string;
  model_override?: string;
}

export interface ExtractedJobPostingDraft {
  company_name: string;
  company_domain: string;
  role_title: string;
  seniority: string;
  location: string;
  remote_policy: RemotePolicy | '';
  stated_min_comp: number;
  stated_max_comp: number;
  currency: string;
  tech_stack_tags: string[];
  requirements: string[];
  source_url: string;
  raw_text: string;
  extraction_source: string;
  model_used: string;
  fetch_warning: string;
  assistant_warning: string;
}

export interface QuickCaptureSaveInput {
  company_id?: string;
  company_name: string;
  company_domain: string;
  role_title: string;
  seniority: string;
  location: string;
  remote_policy: RemotePolicy | '';
  priority: Priority;
  stage?: Stage;
  transition_note: string;
  notes: string;
  job_posting: JobPostingInput;
}
