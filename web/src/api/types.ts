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

export const CONTACT_WARMTHS = ['Cold', 'Warm', 'Advocate'] as const;
export type ContactWarmth = (typeof CONTACT_WARMTHS)[number];

export const STAKEHOLDER_ROLES = [
  'Referrer',
  'Recruiter',
  'Hiring Manager',
  'Interviewer',
  'Champion',
  'Other',
] as const;
export type StakeholderRole = (typeof STAKEHOLDER_ROLES)[number];

export const INTERACTION_KINDS = [
  'Email',
  'DM',
  'Referral Request',
  'Coffee Chat',
  'Phone Call',
  'Other',
] as const;
export type InteractionKind = (typeof INTERACTION_KINDS)[number];

export const INTERACTION_DIRECTIONS = ['Outbound', 'Inbound'] as const;
export type InteractionDirection = (typeof INTERACTION_DIRECTIONS)[number];

export const ACTION_ITEM_STATUSES = ['open', 'completed', 'dismissed'] as const;
export type ActionItemStatus = (typeof ACTION_ITEM_STATUSES)[number];

export interface ContactOpportunityLink {
  opportunity_id: string;
  company_id: string;
  company_name: string;
  role_title: string;
  stage: Stage;
  outcome: Outcome;
  role: StakeholderRole;
  notes: string;
  linked_at: string;
}

export interface Contact {
  id: string;
  company_id: string;
  company_name: string;
  name: string;
  title: string;
  email: string;
  linkedin_url: string;
  warmth: ContactWarmth;
  notes: string;
  linked_opportunities: ContactOpportunityLink[];
  created_at: string;
  updated_at: string;
}

export interface ContactUpsertInput {
  company_id: string;
  name: string;
  title: string;
  email: string;
  linkedin_url: string;
  warmth: ContactWarmth;
  notes: string;
}

export interface ContactFilter {
  company_id?: string;
  warmth?: ContactWarmth | '';
  q?: string;
}

export interface OpportunityContact {
  opportunity_id: string;
  contact_id: string;
  role: StakeholderRole;
  notes: string;
  linked_at: string;
  company_id: string;
  company_name: string;
  name: string;
  title: string;
  email: string;
  linkedin_url: string;
  warmth: ContactWarmth;
  contact_notes: string;
}

export interface OpportunityContactLinkInput {
  contact_id: string;
  role: StakeholderRole;
  notes?: string;
}

export interface Interaction {
  id: string;
  contact_id: string;
  contact_name: string;
  opportunity_id: string;
  opportunity_role_title: string;
  company_name: string;
  kind: InteractionKind;
  direction: InteractionDirection;
  occurred_at: string;
  summary: string;
  created_at: string;
}

export interface FollowUpActionItemInput {
  title: string;
  due_at: string;
  notes?: string;
}

export interface InteractionCreateInput {
  contact_id: string;
  opportunity_id?: string;
  kind: InteractionKind;
  direction: InteractionDirection;
  occurred_at?: string;
  summary: string;
  follow_up?: FollowUpActionItemInput;
}

export interface InteractionFilter {
  contact_id?: string;
  opportunity_id?: string;
}

export interface ActionItem {
  id: string;
  title: string;
  opportunity_id: string;
  opportunity_role_title: string;
  contact_id: string;
  contact_name: string;
  company_name: string;
  due_at: string;
  status: ActionItemStatus;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface ActionItemCreateInput {
  title: string;
  opportunity_id?: string;
  contact_id?: string;
  due_at: string;
  notes?: string;
}

export interface ActionItemUpdateInput {
  title: string;
  opportunity_id?: string;
  contact_id?: string;
  due_at: string;
  status: ActionItemStatus;
  notes?: string;
}

export interface ActionItemFilter {
  status?: ActionItemStatus | '';
  opportunity_id?: string;
  contact_id?: string;
}

export interface LogInteractionResponse {
  interaction: Interaction;
  follow_up_action_item?: ActionItem;
}

export const INTERVIEW_ROUND_TYPES = [
  'Recruiter Screen',
  'Founder Chat',
  'Hiring Manager',
  'Take-Home',
  'Technical',
  'System Design',
  'Behavioral',
  'Team Match',
  'Executive',
] as const;

export type CanonicalInterviewRoundType = (typeof INTERVIEW_ROUND_TYPES)[number];

export const INTERVIEW_ROUND_STATUSES = [
  'Scheduled',
  'Completed',
  'Cancelled',
] as const;

export type InterviewRoundStatus = (typeof INTERVIEW_ROUND_STATUSES)[number];

export interface StoryPriorRoundRef {
  round_id: string;
  round_order: number;
  round_type: string;
  title: string;
}

export interface InterviewRoundContact {
  contact_id: string;
  name: string;
  title: string;
  email: string;
  warmth: ContactWarmth;
  company_name: string;
}

export interface InterviewRoundStory {
  story_id: string;
  title: string;
  hook: string;
  competency_tags: string[];
  skill_tags: string[];
  used_in_earlier_rounds: StoryPriorRoundRef[];
}

export interface DebriefQuestion {
  question: string;
  category: string;
  answered_with_story_id: string;
  answered_with_story_title: string;
  notes: string;
}

export interface DebriefQuestionInput {
  question: string;
  category: string;
  answered_with_story_id: string;
  notes: string;
}

export interface Debrief {
  questions_asked: DebriefQuestion[];
  went_well: string;
  to_improve: string;
  signals: string;
  confidence_score: number;
  follow_up_notes: string;
  updated_at: string;
}

export interface DebriefUpsertInput {
  questions_asked: DebriefQuestionInput[];
  went_well: string;
  to_improve: string;
  signals: string;
  confidence_score: number;
  follow_up_notes: string;
}

export interface InterviewRound {
  id: string;
  opportunity_id: string;
  round_order: number;
  round_type: string;
  title: string;
  scheduled_at: string;
  duration_minutes: number;
  status: InterviewRoundStatus;
  prep_notes: string;
  contact_ids: string[];
  story_ids: string[];
  contacts: InterviewRoundContact[];
  stories: InterviewRoundStory[];
  debrief?: Debrief;
  created_at: string;
  updated_at: string;
}

export interface InterviewRoundUpsertInput {
  round_order: number;
  round_type: string;
  title: string;
  scheduled_at: string;
  duration_minutes: number;
  status: InterviewRoundStatus;
  prep_notes: string;
  contact_ids: string[];
  story_ids: string[];
}

export interface QuestionBankFilter {
  q?: string;
  category?: string;
  company_id?: string;
  round_type?: string;
  story_id?: string;
}

export interface QuestionBankEntry {
  round_id: string;
  opportunity_id: string;
  opportunity_role_title: string;
  company_id: string;
  company_name: string;
  round_order: number;
  round_type: string;
  round_title: string;
  question: string;
  category: string;
  answered_with_story_id: string;
  answered_with_story_title: string;
  notes: string;
  recorded_at: string;
}
