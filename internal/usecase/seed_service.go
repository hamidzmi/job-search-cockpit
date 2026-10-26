package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/inbound"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type SeedService struct {
	tx              outbound.TxManager
	companyRepo     outbound.CompanyRepository
	oppRepo         outbound.OpportunityRepository
	transitionRepo  outbound.StageTransitionRepository
	profileRepo     outbound.CandidateProfileRepository
	storyRepo       outbound.StoryRepository
	contactRepo     outbound.ContactRepository
	interactionRepo outbound.InteractionRepository
	actionItemRepo  outbound.ActionItemRepository
	roundRepo       outbound.InterviewRoundRepository
	offerRepo       outbound.OfferRepository
	clock           outbound.Clock
}

func NewSeedService(
	tx outbound.TxManager,
	companyRepo outbound.CompanyRepository,
	oppRepo outbound.OpportunityRepository,
	transitionRepo outbound.StageTransitionRepository,
	profileRepo outbound.CandidateProfileRepository,
	storyRepo outbound.StoryRepository,
	contactRepo outbound.ContactRepository,
	interactionRepo outbound.InteractionRepository,
	actionItemRepo outbound.ActionItemRepository,
	roundRepo outbound.InterviewRoundRepository,
	offerRepo outbound.OfferRepository,
	clock outbound.Clock,
) *SeedService {
	return &SeedService{
		tx:              tx,
		companyRepo:     companyRepo,
		oppRepo:         oppRepo,
		transitionRepo:  transitionRepo,
		profileRepo:     profileRepo,
		storyRepo:       storyRepo,
		contactRepo:     contactRepo,
		interactionRepo: interactionRepo,
		actionItemRepo:  actionItemRepo,
		roundRepo:       roundRepo,
		offerRepo:       offerRepo,
		clock:           clock,
	}
}

var _ inbound.SeedUseCase = (*SeedService)(nil)

func (s *SeedService) SeedSampleData(ctx context.Context) error {
	now := s.clock.Now().UTC()

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		// 1. Candidate Profile
		profile, err := domain.NewCandidateProfile(domain.CandidateProfileInput{
			FullName:            "Alex Rivera",
			Headline:            "Staff Infrastructure Engineer",
			Summary:             "12+ years architecting high-throughput distributed systems, fault-tolerant consensus engines, and cloud platforms.",
			TargetRoleTitles:    []string{"Staff Backend Engineer", "Principal Infrastructure Engineer"},
			LocationPreferences: []string{"San Francisco, CA", "Remote"},
			Skills:              []string{"Go", "Kubernetes", "Distributed Systems", "PostgreSQL", "Kafka", "AWS"},
			CompensationFloor:   220000,
			TargetCompensation:  280000,
			Currency:            "USD",
		}, now)
		if err != nil {
			return err
		}
		if err := s.profileRepo.SaveCandidateProfile(txCtx, profile); err != nil {
			return err
		}

		// 2. Stories in Story Bank
		story1, err := domain.NewStory(uuid.NewString(), domain.StoryInput{
			Title:          "Distributed Cache Migration under 10x Load Spike",
			Hook:           "Reduced P99 tail latency from 450ms to 28ms during Black Friday peak traffic.",
			Situation:      "Legacy cache cluster was hitting memory exhaustion and causing cascading timeout cascades.",
			Task:           "Design and execute zero-downtime cutover to partitioned Redis cluster.",
			Action:         "Implemented dual-write shadow proxy in Go and verified cache coherence over 2 weeks.",
			Result:         "85% reduction in tail latency and zero dropped user requests during 10x traffic spike.",
			CompetencyTags: []string{"Architecture", "System Design"},
			SkillTags:      []string{"Go", "Redis", "Distributed Systems"},
		}, now)
		if err != nil {
			return err
		}
		if err := s.storyRepo.SaveStory(txCtx, story1); err != nil {
			return err
		}

		story2, err := domain.NewStory(uuid.NewString(), domain.StoryInput{
			Title:          "Multi-Region Database Failover Automation",
			Hook:           "Automated cross-region database failover, reducing MTTR from 45 minutes to 40 seconds.",
			Situation:      "Single-region failure risks threatened SLA commitments with enterprise customers.",
			Task:           "Build an automated health check and consensus-driven failover daemon.",
			Action:         "Wrote consensus orchestrator using Raft and automated DNS routing switchover.",
			Result:         "Successfully endured live chaos engineering simulation with zero data loss.",
			CompetencyTags: []string{"Reliability", "Disaster Recovery"},
			SkillTags:      []string{"PostgreSQL", "Raft", "Kubernetes"},
		}, now)
		if err != nil {
			return err
		}
		if err := s.storyRepo.SaveStory(txCtx, story2); err != nil {
			return err
		}

		// 3. Companies
		comp1, err := domain.NewCompany(uuid.NewString(), domain.CompanyInput{
			Name:         "Starlight AI",
			Domain:       "starlight.ai",
			Headquarters: "San Francisco, CA",
			RemotePolicy: string(domain.RemotePolicyRemote),
			SizeStage:    "Series B (120 employees)",
			Notes:        "AI foundation model inference platform.",
		}, now)
		if err != nil {
			return err
		}
		if err := s.companyRepo.SaveCompany(txCtx, comp1); err != nil {
			return err
		}

		comp2, err := domain.NewCompany(uuid.NewString(), domain.CompanyInput{
			Name:         "Acme Cloud",
			Domain:       "acmecloud.io",
			Headquarters: "Seattle, WA",
			RemotePolicy: string(domain.RemotePolicyHybrid),
			SizeStage:    "Public (3,500 employees)",
			Notes:        "Enterprise cloud platform and managed database services.",
		}, now)
		if err != nil {
			return err
		}
		if err := s.companyRepo.SaveCompany(txCtx, comp2); err != nil {
			return err
		}

		comp3, err := domain.NewCompany(uuid.NewString(), domain.CompanyInput{
			Name:         "Quantum Systems",
			Domain:       "quantumsystems.dev",
			Headquarters: "New York, NY",
			RemotePolicy: string(domain.RemotePolicyRemote),
			SizeStage:    "Series C (300 employees)",
			Notes:        "High-performance distributed streaming engine.",
		}, now)
		if err != nil {
			return err
		}
		if err := s.companyRepo.SaveCompany(txCtx, comp3); err != nil {
			return err
		}

		// 4. Contacts
		contact1, err := domain.NewContact(uuid.NewString(), domain.ContactInput{
			Name:        "Sarah Chen",
			CompanyID:   comp1.ID,
			Title:       "Head of Infrastructure",
			Email:       "sarah@starlight.ai",
			LinkedInURL: "https://linkedin.com/in/sarahchen-fake",
			Warmth:      domain.ContactWarmthWarm,
			Notes:       "Met at Systems Conf 2025; values clean architectural boundaries.",
		}, now)
		if err != nil {
			return err
		}
		if err := s.contactRepo.SaveContact(txCtx, contact1); err != nil {
			return err
		}

		contact2, err := domain.NewContact(uuid.NewString(), domain.ContactInput{
			Name:        "Marcus Vance",
			CompanyID:   comp2.ID,
			Title:       "Staff Technical Recruiter",
			Email:       "marcus@acmecloud.io",
			LinkedInURL: "https://linkedin.com/in/marcusvance-fake",
			Warmth:      domain.ContactWarmthAdvocate,
			Notes:       "Highly responsive and championing compensation package.",
		}, now)
		if err != nil {
			return err
		}
		if err := s.contactRepo.SaveContact(txCtx, contact2); err != nil {
			return err
		}

		// 5. Opportunity 1: Starlight AI (Interviewing with repeated story warning and missing debrief)
		opp1ID := uuid.NewString()
		tr1ID := uuid.NewString()
		opp1, tr1Init, err := domain.NewOpportunity(opp1ID, tr1ID, domain.OpportunityInput{
			OpportunityDetailsInput: domain.OpportunityDetailsInput{
				CompanyID:    comp1.ID,
				RoleTitle:    "Staff Infrastructure Engineer",
				Seniority:    "Staff",
				Location:     "Remote",
				RemotePolicy: domain.RemotePolicyRemote,
				Priority:     domain.PriorityHigh,
				Notes:        "High priority opportunity with strong engineering leadership.",
			},
			Stage: domain.StageDiscovered,
		}, now.Add(-18*24*time.Hour))
		if err != nil {
			return err
		}
		if err := s.oppRepo.SaveOpportunity(txCtx, opp1); err != nil {
			return err
		}
		if err := s.transitionRepo.AppendStageTransition(txCtx, tr1Init); err != nil {
			return err
		}

		// Transition opp1 through Outreach, Applied, Screening, Interviewing
		trTimes := []struct {
			stage domain.Stage
			delay int
		}{
			{domain.StageOutreach, 15},
			{domain.StageApplied, 12},
			{domain.StageScreening, 8},
			{domain.StageInterviewing, 4},
		}
		for _, trStep := range trTimes {
			tr, err := opp1.Transition(uuid.NewString(), trStep.stage, domain.OutcomeActive, "Progressed to "+string(trStep.stage), now.Add(-time.Duration(trStep.delay)*24*time.Hour))
			if err != nil {
				return err
			}
			if err := s.oppRepo.SaveOpportunity(txCtx, opp1); err != nil {
				return err
			}
			if err := s.transitionRepo.AppendStageTransition(txCtx, tr); err != nil {
				return err
			}
		}

		// Link contact1 to opp1
		oppContact1, err := domain.NewOpportunityContact(domain.OpportunityContactInput{
			OpportunityID: opp1.ID,
			ContactID:     contact1.ID,
			Role:          domain.StakeholderRoleHiringManager,
		}, now)
		if err != nil {
			return err
		}
		if err := s.contactRepo.SaveOpportunityContact(txCtx, oppContact1); err != nil {
			return err
		}

		// Interview rounds on opp1:
		// Round 1: Recruiter screen, completed, debrief recorded
		r1ID := uuid.NewString()
		r1, err := domain.NewInterviewRound(r1ID, domain.InterviewRoundInput{
			OpportunityID:   opp1.ID,
			RoundOrder:      1,
			RoundType:       domain.InterviewRoundTypeRecruiterScreen,
			Title:           "Initial Recruiter Alignment",
			ScheduledAt:     now.Add(-7 * 24 * time.Hour),
			DurationMinutes: 30,
			Status:          domain.InterviewRoundStatusCompleted,
			PrepNotes:       "Review resume milestones and compensation floor",
			ContactIDs:      []string{contact1.ID},
		}, now.Add(-8*24*time.Hour))
		if err != nil {
			return err
		}
		if err := r1.RecordDebrief(domain.DebriefInput{
			WentWell:        "Strong rapport and positive feedback on distributed systems background",
			ToImprove:       "Clarify timeline expectations earlier",
			Signals:         "Team is actively expanding infra engineering group",
			ConfidenceScore: 4,
			FollowUpNotes:   "Recruiter scheduled technical design within 24h",
		}, now.Add(-7*24*time.Hour)); err != nil {
			return err
		}
		if err := s.roundRepo.SaveInterviewRound(txCtx, r1); err != nil {
			return err
		}

		// Round 2: System design, completed, used Story 1, debrief recorded
		r2ID := uuid.NewString()
		r2, err := domain.NewInterviewRound(r2ID, domain.InterviewRoundInput{
			OpportunityID:   opp1.ID,
			RoundOrder:      2,
			RoundType:       domain.InterviewRoundTypeSystemDesign,
			Title:           "Distributed Storage & Caching Design",
			ScheduledAt:     now.Add(-4 * 24 * time.Hour),
			DurationMinutes: 60,
			Status:          domain.InterviewRoundStatusCompleted,
			PrepNotes:       "Focus on partition strategy, replication, and failover bounds",
			StoryIDs:        []string{story1.ID},
			ContactIDs:      []string{contact1.ID},
		}, now.Add(-5*24*time.Hour))
		if err != nil {
			return err
		}
		if err := r2.RecordDebrief(domain.DebriefInput{
			WentWell:        "Deep dive into cache failure modes was well received",
			ToImprove:       "Draw network topology diagram earlier",
			Signals:         "Interviewer engaged deeply on Go memory management",
			ConfidenceScore: 5,
		}, now.Add(-4*24*time.Hour)); err != nil {
			return err
		}
		if err := s.roundRepo.SaveInterviewRound(txCtx, r2); err != nil {
			return err
		}

		// Round 3: Technical Deep Dive, Scheduled, plans to use Story 1 -> REPEATED STORY WARNING!
		r3ID := uuid.NewString()
		r3, err := domain.NewInterviewRound(r3ID, domain.InterviewRoundInput{
			OpportunityID:   opp1.ID,
			RoundOrder:      3,
			RoundType:       domain.InterviewRoundTypeTechnical,
			Title:           "Architecture Deep Dive with VP",
			ScheduledAt:     now.Add(24 * time.Hour),
			DurationMinutes: 60,
			Status:          domain.InterviewRoundStatusScheduled,
			PrepNotes:       "Discuss multi-tenant isolation and cloud cost reduction",
			StoryIDs:        []string{story1.ID}, // Repeated story from Round 2!
		}, now)
		if err != nil {
			return err
		}
		if err := s.roundRepo.SaveInterviewRound(txCtx, r3); err != nil {
			return err
		}

		// Round 4: Team Match / Culture, Completed, MISSING DEBRIEF!
		r4ID := uuid.NewString()
		r4, err := domain.NewInterviewRound(r4ID, domain.InterviewRoundInput{
			OpportunityID:   opp1.ID,
			RoundOrder:      4,
			RoundType:       domain.InterviewRoundTypeBehavioral,
			Title:           "Team Match & Culture",
			ScheduledAt:     now.Add(-12 * time.Hour),
			DurationMinutes: 45,
			Status:          domain.InterviewRoundStatusCompleted,
			PrepNotes:       "Emphasize cross-functional mentoring",
		}, now.Add(-14*time.Hour))
		if err != nil {
			return err
		}
		if err := s.roundRepo.SaveInterviewRound(txCtx, r4); err != nil {
			return err
		}

		// 6. Opportunity 2: Acme Cloud (Offer stage with active Offer)
		opp2ID := uuid.NewString()
		tr2ID := uuid.NewString()
		opp2, tr2Init, err := domain.NewOpportunity(opp2ID, tr2ID, domain.OpportunityInput{
			OpportunityDetailsInput: domain.OpportunityDetailsInput{
				CompanyID:    comp2.ID,
				RoleTitle:    "Principal Platform Architect",
				Seniority:    "Principal",
				Location:     "Hybrid - Seattle, WA",
				RemotePolicy: domain.RemotePolicyHybrid,
				Priority:     domain.PriorityHigh,
			},
			Stage: domain.StageDiscovered,
		}, now.Add(-25*24*time.Hour))
		if err != nil {
			return err
		}
		if err := s.oppRepo.SaveOpportunity(txCtx, opp2); err != nil {
			return err
		}
		if err := s.transitionRepo.AppendStageTransition(txCtx, tr2Init); err != nil {
			return err
		}

		opp2Stages := []struct {
			stage domain.Stage
			delay int
		}{
			{domain.StageOutreach, 20},
			{domain.StageApplied, 15},
			{domain.StageScreening, 10},
			{domain.StageInterviewing, 5},
			{domain.StageOffer, 2},
		}
		for _, st := range opp2Stages {
			tr, err := opp2.Transition(uuid.NewString(), st.stage, domain.OutcomeActive, "Moved to "+string(st.stage), now.Add(-time.Duration(st.delay)*24*time.Hour))
			if err != nil {
				return err
			}
			if err := s.oppRepo.SaveOpportunity(txCtx, opp2); err != nil {
				return err
			}
			if err := s.transitionRepo.AppendStageTransition(txCtx, tr); err != nil {
				return err
			}
		}

		// Link contact2 to opp2
		oppContact2, err := domain.NewOpportunityContact(domain.OpportunityContactInput{
			OpportunityID: opp2.ID,
			ContactID:     contact2.ID,
			Role:          domain.StakeholderRoleRecruiter,
		}, now)
		if err != nil {
			return err
		}
		if err := s.contactRepo.SaveOpportunityContact(txCtx, oppContact2); err != nil {
			return err
		}

		// Offer on opp2
		offer1, err := domain.NewOffer(uuid.NewString(), domain.OfferInput{
			OpportunityID:      opp2.ID,
			BaseSalary:         235000,
			BonusTargetPct:     15,
			EquityType:         "RSU",
			EquityTotalValue:   240000,
			EquityVestingYears: 4,
			SignOnBonus:        25000,
			Currency:           "USD",
			DeadlineAt:         now.Add(7 * 24 * time.Hour),
			BenefitsNotes:      "100% employer medical/dental, 401k 5% immediate match",
			NegotiationNotes:   "Countered for $30k additional equity; awaiting approval",
			Status:             domain.OfferStatusNegotiating,
		}, now.Add(-2*24*time.Hour))
		if err != nil {
			return err
		}
		if err := s.offerRepo.SaveOffer(txCtx, offer1); err != nil {
			return err
		}

		// 7. Opportunity 3: Quantum Systems (Offer stage with second Offer for side-by-side comparison)
		opp3ID := uuid.NewString()
		tr3ID := uuid.NewString()
		opp3, tr3Init, err := domain.NewOpportunity(opp3ID, tr3ID, domain.OpportunityInput{
			OpportunityDetailsInput: domain.OpportunityDetailsInput{
				CompanyID:    comp3.ID,
				RoleTitle:    "Staff Backend Engineer",
				Seniority:    "Staff",
				Location:     "Remote",
				RemotePolicy: domain.RemotePolicyRemote,
				Priority:     domain.PriorityHigh,
			},
			Stage: domain.StageDiscovered,
		}, now.Add(-14*24*time.Hour))
		if err != nil {
			return err
		}
		if err := s.oppRepo.SaveOpportunity(txCtx, opp3); err != nil {
			return err
		}
		if err := s.transitionRepo.AppendStageTransition(txCtx, tr3Init); err != nil {
			return err
		}

		opp3Stages := []struct {
			stage domain.Stage
			delay int
		}{
			{domain.StageApplied, 10},
			{domain.StageScreening, 7},
			{domain.StageInterviewing, 4},
			{domain.StageOffer, 1},
		}
		for _, st := range opp3Stages {
			tr, err := opp3.Transition(uuid.NewString(), st.stage, domain.OutcomeActive, "Advanced to "+string(st.stage), now.Add(-time.Duration(st.delay)*24*time.Hour))
			if err != nil {
				return err
			}
			if err := s.oppRepo.SaveOpportunity(txCtx, opp3); err != nil {
				return err
			}
			if err := s.transitionRepo.AppendStageTransition(txCtx, tr); err != nil {
				return err
			}
		}

		// Offer on opp3
		offer2, err := domain.NewOffer(uuid.NewString(), domain.OfferInput{
			OpportunityID:      opp3.ID,
			BaseSalary:         215000,
			BonusAmount:        30000,
			EquityType:         "RSU",
			EquityTotalValue:   180000,
			EquityVestingYears: 3,
			SignOnBonus:        15000,
			Currency:           "USD",
			DeadlineAt:         now.Add(10 * 24 * time.Hour),
			BenefitsNotes:      "Full health, $3k home office stipend",
			NegotiationNotes:   "Initial offer received",
			Status:             domain.OfferStatusPending,
		}, now.Add(-24*time.Hour))
		if err != nil {
			return err
		}
		if err := s.offerRepo.SaveOffer(txCtx, offer2); err != nil {
			return err
		}

		// 8. Interactions
		interaction1, err := domain.NewInteraction(uuid.NewString(), domain.InteractionInput{
			ContactID:     contact1.ID,
			OpportunityID: opp1.ID,
			Kind:          domain.InteractionKindEmail,
			Direction:     domain.InteractionDirectionOutbound,
			OccurredAt:    now.Add(-2 * 24 * time.Hour),
			Summary:       "Sent follow-up questions to Sarah regarding infra cluster scale",
		}, now)
		if err != nil {
			return err
		}
		if err := s.interactionRepo.SaveInteraction(txCtx, interaction1); err != nil {
			return err
		}

		interaction2, err := domain.NewInteraction(uuid.NewString(), domain.InteractionInput{
			ContactID:     contact2.ID,
			OpportunityID: opp2.ID,
			Kind:          domain.InteractionKindPhoneCall,
			Direction:     domain.InteractionDirectionInbound,
			OccurredAt:    now.Add(-24 * time.Hour),
			Summary:       "Offer review call with Marcus; walked through bonus structure",
		}, now)
		if err != nil {
			return err
		}
		if err := s.interactionRepo.SaveInteraction(txCtx, interaction2); err != nil {
			return err
		}

		// 9. Action Items
		item1, err := domain.NewActionItem(uuid.NewString(), domain.ActionItemInput{
			Title:         "Prepare system architecture talking points for Starlight AI Round 3",
			OpportunityID: opp1.ID,
			DueAt:         now.Add(18 * time.Hour),
		}, now)
		if err != nil {
			return err
		}
		if err := s.actionItemRepo.SaveActionItem(txCtx, item1); err != nil {
			return err
		}

		item2, err := domain.NewActionItem(uuid.NewString(), domain.ActionItemInput{
			Title:         "Review counter-offer figures with mentor before reply to Marcus",
			OpportunityID: opp2.ID,
			DueAt:         now.Add(48 * time.Hour),
		}, now)
		if err != nil {
			return err
		}
		if err := s.actionItemRepo.SaveActionItem(txCtx, item2); err != nil {
			return err
		}

		return nil
	})
}
