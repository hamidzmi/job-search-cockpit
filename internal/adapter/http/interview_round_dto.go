package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type StoryPriorRoundRefDTO struct {
	RoundID    string `json:"round_id"`
	RoundOrder int    `json:"round_order"`
	RoundType  string `json:"round_type"`
	Title      string `json:"title"`
}

type InterviewRoundContactDTO struct {
	ContactID   string `json:"contact_id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Email       string `json:"email"`
	Warmth      string `json:"warmth"`
	CompanyName string `json:"company_name"`
}

type InterviewRoundStoryDTO struct {
	StoryID             string                  `json:"story_id"`
	Title               string                  `json:"title"`
	Hook                string                  `json:"hook"`
	CompetencyTags      []string                `json:"competency_tags"`
	SkillTags           []string                `json:"skill_tags"`
	UsedInEarlierRounds []StoryPriorRoundRefDTO `json:"used_in_earlier_rounds"`
}

type DebriefQuestionDTO struct {
	Question               string `json:"question"`
	Category               string `json:"category"`
	AnsweredWithStoryID    string `json:"answered_with_story_id"`
	AnsweredWithStoryTitle string `json:"answered_with_story_title"`
	Notes                  string `json:"notes"`
}

type DebriefDTO struct {
	QuestionsAsked  []DebriefQuestionDTO `json:"questions_asked"`
	WentWell        string               `json:"went_well"`
	ToImprove       string               `json:"to_improve"`
	Signals         string               `json:"signals"`
	ConfidenceScore int                  `json:"confidence_score"`
	FollowUpNotes   string               `json:"follow_up_notes"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

type InterviewRoundDTO struct {
	ID              string                     `json:"id"`
	OpportunityID   string                     `json:"opportunity_id"`
	RoundOrder      int                        `json:"round_order"`
	RoundType       string                     `json:"round_type"`
	Title           string                     `json:"title"`
	ScheduledAt     string                     `json:"scheduled_at"`
	DurationMinutes int                        `json:"duration_minutes"`
	Status          string                     `json:"status"`
	PrepNotes       string                     `json:"prep_notes"`
	ContactIDs      []string                   `json:"contact_ids"`
	StoryIDs        []string                   `json:"story_ids"`
	Contacts        []InterviewRoundContactDTO `json:"contacts"`
	Stories         []InterviewRoundStoryDTO   `json:"stories"`
	Debrief         *DebriefDTO                `json:"debrief,omitempty"`
	CreatedAt       time.Time                  `json:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at"`
}

type InterviewRoundUpsertRequest struct {
	RoundOrder      int      `json:"round_order"`
	RoundType       string   `json:"round_type"`
	Title           string   `json:"title"`
	ScheduledAt     string   `json:"scheduled_at"`
	DurationMinutes int      `json:"duration_minutes"`
	Status          string   `json:"status"`
	PrepNotes       string   `json:"prep_notes"`
	ContactIDs      []string `json:"contact_ids"`
	StoryIDs        []string `json:"story_ids"`
}

func (r InterviewRoundUpsertRequest) toDomainInput() (domain.InterviewRoundInput, error) {
	scheduledAt, err := parseFlexibleTimestamp(r.ScheduledAt)
	if err != nil {
		return domain.InterviewRoundInput{}, err
	}
	return domain.InterviewRoundInput{
		RoundOrder:      r.RoundOrder,
		RoundType:       domain.InterviewRoundType(r.RoundType),
		Title:           r.Title,
		ScheduledAt:     scheduledAt,
		DurationMinutes: r.DurationMinutes,
		Status:          domain.InterviewRoundStatus(r.Status),
		PrepNotes:       r.PrepNotes,
		ContactIDs:      r.ContactIDs,
		StoryIDs:        r.StoryIDs,
	}, nil
}

type DebriefQuestionRequest struct {
	Question            string `json:"question"`
	Category            string `json:"category"`
	AnsweredWithStoryID string `json:"answered_with_story_id"`
	Notes               string `json:"notes"`
}

type DebriefUpsertRequest struct {
	QuestionsAsked  []DebriefQuestionRequest `json:"questions_asked"`
	WentWell        string                   `json:"went_well"`
	ToImprove       string                   `json:"to_improve"`
	Signals         string                   `json:"signals"`
	ConfidenceScore int                      `json:"confidence_score"`
	FollowUpNotes   string                   `json:"follow_up_notes"`
}

func (r DebriefUpsertRequest) toDomainInput() domain.DebriefInput {
	questions := make([]domain.DebriefQuestion, 0, len(r.QuestionsAsked))
	for _, q := range r.QuestionsAsked {
		questions = append(questions, domain.DebriefQuestion{
			Question:            q.Question,
			Category:            q.Category,
			AnsweredWithStoryID: q.AnsweredWithStoryID,
			Notes:               q.Notes,
		})
	}
	return domain.DebriefInput{
		QuestionsAsked:  questions,
		WentWell:        r.WentWell,
		ToImprove:       r.ToImprove,
		Signals:         r.Signals,
		ConfidenceScore: r.ConfidenceScore,
		FollowUpNotes:   r.FollowUpNotes,
	}
}

type QuestionBankEntryDTO struct {
	RoundID                string    `json:"round_id"`
	OpportunityID          string    `json:"opportunity_id"`
	OpportunityRoleTitle   string    `json:"opportunity_role_title"`
	CompanyID              string    `json:"company_id"`
	CompanyName            string    `json:"company_name"`
	RoundOrder             int       `json:"round_order"`
	RoundType              string    `json:"round_type"`
	RoundTitle             string    `json:"round_title"`
	Question               string    `json:"question"`
	Category               string    `json:"category"`
	AnsweredWithStoryID    string    `json:"answered_with_story_id"`
	AnsweredWithStoryTitle string    `json:"answered_with_story_title"`
	Notes                  string    `json:"notes"`
	RecordedAt             time.Time `json:"recorded_at"`
}

func toInterviewRoundDTO(v domain.InterviewRoundView) InterviewRoundDTO {
	r := v.InterviewRound
	var scheduledAtStr string
	if !r.ScheduledAt.IsZero() {
		scheduledAtStr = r.ScheduledAt.UTC().Format(time.RFC3339)
	}

	contactIDs := r.ContactIDs
	if contactIDs == nil {
		contactIDs = []string{}
	}
	storyIDs := r.StoryIDs
	if storyIDs == nil {
		storyIDs = []string{}
	}

	contacts := make([]InterviewRoundContactDTO, 0, len(v.Contacts))
	for _, c := range v.Contacts {
		contacts = append(contacts, InterviewRoundContactDTO{
			ContactID:   c.ContactID,
			Name:        c.Name,
			Title:       c.Title,
			Email:       c.Email,
			Warmth:      string(c.Warmth),
			CompanyName: c.CompanyName,
		})
	}

	stories := make([]InterviewRoundStoryDTO, 0, len(v.Stories))
	for _, st := range v.Stories {
		compTags := st.CompetencyTags
		if compTags == nil {
			compTags = []string{}
		}
		skillTags := st.SkillTags
		if skillTags == nil {
			skillTags = []string{}
		}
		priors := make([]StoryPriorRoundRefDTO, 0, len(st.UsedInEarlierRounds))
		for _, p := range st.UsedInEarlierRounds {
			priors = append(priors, StoryPriorRoundRefDTO{
				RoundID:    p.RoundID,
				RoundOrder: p.RoundOrder,
				RoundType:  string(p.RoundType),
				Title:      p.Title,
			})
		}
		stories = append(stories, InterviewRoundStoryDTO{
			StoryID:             st.StoryID,
			Title:               st.Title,
			Hook:                st.Hook,
			CompetencyTags:      compTags,
			SkillTags:           skillTags,
			UsedInEarlierRounds: priors,
		})
	}

	var debriefDTO *DebriefDTO
	if v.DebriefView != nil {
		qs := make([]DebriefQuestionDTO, 0, len(v.DebriefView.QuestionsAsked))
		for _, q := range v.DebriefView.QuestionsAsked {
			qs = append(qs, DebriefQuestionDTO{
				Question:               q.Question,
				Category:               q.Category,
				AnsweredWithStoryID:    q.AnsweredWithStoryID,
				AnsweredWithStoryTitle: q.AnsweredWithStoryTitle,
				Notes:                  q.Notes,
			})
		}
		debriefDTO = &DebriefDTO{
			QuestionsAsked:  qs,
			WentWell:        v.DebriefView.WentWell,
			ToImprove:       v.DebriefView.ToImprove,
			Signals:         v.DebriefView.Signals,
			ConfidenceScore: v.DebriefView.ConfidenceScore,
			FollowUpNotes:   v.DebriefView.FollowUpNotes,
			UpdatedAt:       v.DebriefView.UpdatedAt,
		}
	}

	return InterviewRoundDTO{
		ID:              r.ID,
		OpportunityID:   r.OpportunityID,
		RoundOrder:      r.RoundOrder,
		RoundType:       string(r.RoundType),
		Title:           r.Title,
		ScheduledAt:     scheduledAtStr,
		DurationMinutes: r.DurationMinutes,
		Status:          string(r.Status),
		PrepNotes:       r.PrepNotes,
		ContactIDs:      contactIDs,
		StoryIDs:        storyIDs,
		Contacts:        contacts,
		Stories:         stories,
		Debrief:         debriefDTO,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func toQuestionBankEntryDTO(v domain.QuestionBankEntryView) QuestionBankEntryDTO {
	return QuestionBankEntryDTO{
		RoundID:                v.RoundID,
		OpportunityID:          v.OpportunityID,
		OpportunityRoleTitle:   v.OpportunityRoleTitle,
		CompanyID:              v.CompanyID,
		CompanyName:            v.CompanyName,
		RoundOrder:             v.RoundOrder,
		RoundType:              string(v.RoundType),
		RoundTitle:             v.RoundTitle,
		Question:               v.Question,
		Category:               v.Category,
		AnsweredWithStoryID:    v.AnsweredWithStoryID,
		AnsweredWithStoryTitle: v.AnsweredWithStoryTitle,
		Notes:                  v.Notes,
		RecordedAt:             v.RecordedAt,
	}
}
