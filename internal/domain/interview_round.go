package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type InterviewRoundType string

const (
	InterviewRoundTypeRecruiterScreen InterviewRoundType = "Recruiter Screen"
	InterviewRoundTypeFounderChat     InterviewRoundType = "Founder Chat"
	InterviewRoundTypeHiringManager   InterviewRoundType = "Hiring Manager"
	InterviewRoundTypeTakeHome        InterviewRoundType = "Take-Home"
	InterviewRoundTypeTechnical       InterviewRoundType = "Technical"
	InterviewRoundTypeSystemDesign    InterviewRoundType = "System Design"
	InterviewRoundTypeBehavioral      InterviewRoundType = "Behavioral"
	InterviewRoundTypeTeamMatch       InterviewRoundType = "Team Match"
	InterviewRoundTypeExecutive       InterviewRoundType = "Executive"
)

func AllInterviewRoundTypes() []InterviewRoundType {
	return []InterviewRoundType{
		InterviewRoundTypeRecruiterScreen,
		InterviewRoundTypeFounderChat,
		InterviewRoundTypeHiringManager,
		InterviewRoundTypeTakeHome,
		InterviewRoundTypeTechnical,
		InterviewRoundTypeSystemDesign,
		InterviewRoundTypeBehavioral,
		InterviewRoundTypeTeamMatch,
		InterviewRoundTypeExecutive,
	}
}

type InterviewRoundStatus string

const (
	InterviewRoundStatusScheduled InterviewRoundStatus = "Scheduled"
	InterviewRoundStatusCompleted InterviewRoundStatus = "Completed"
	InterviewRoundStatusCancelled InterviewRoundStatus = "Cancelled"
)

func AllInterviewRoundStatuses() []InterviewRoundStatus {
	return []InterviewRoundStatus{
		InterviewRoundStatusScheduled,
		InterviewRoundStatusCompleted,
		InterviewRoundStatusCancelled,
	}
}

func (s InterviewRoundStatus) IsValid() bool {
	switch s {
	case InterviewRoundStatusScheduled, InterviewRoundStatusCompleted, InterviewRoundStatusCancelled:
		return true
	default:
		return false
	}
}

type DebriefQuestion struct {
	Question            string
	Category            string
	AnsweredWithStoryID string
	Notes               string
}

type Debrief struct {
	QuestionsAsked  []DebriefQuestion
	WentWell        string
	ToImprove       string
	Signals         string
	ConfidenceScore int
	FollowUpNotes   string
	UpdatedAt       time.Time
}

type DebriefInput struct {
	QuestionsAsked  []DebriefQuestion
	WentWell        string
	ToImprove       string
	Signals         string
	ConfidenceScore int
	FollowUpNotes   string
}

type InterviewRound struct {
	ID              string
	OpportunityID   string
	RoundOrder      int
	RoundType       InterviewRoundType
	Title           string
	ScheduledAt     time.Time
	DurationMinutes int
	Status          InterviewRoundStatus
	PrepNotes       string
	ContactIDs      []string
	StoryIDs        []string
	Debrief         *Debrief
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type InterviewRoundInput struct {
	OpportunityID   string
	RoundOrder      int
	RoundType       InterviewRoundType
	Title           string
	ScheduledAt     time.Time
	DurationMinutes int
	Status          InterviewRoundStatus
	PrepNotes       string
	ContactIDs      []string
	StoryIDs        []string
}

func NewInterviewRound(id string, in InterviewRoundInput, now time.Time) (InterviewRound, error) {
	if strings.TrimSpace(id) == "" {
		return InterviewRound{}, fmt.Errorf("%w: interview round id is required", ErrValidation)
	}
	oppID := strings.TrimSpace(in.OpportunityID)
	if oppID == "" {
		return InterviewRound{}, fmt.Errorf("%w: opportunity id is required", ErrValidation)
	}
	ts := now.UTC()
	round := InterviewRound{
		ID:            strings.TrimSpace(id),
		OpportunityID: oppID,
		CreatedAt:     ts,
	}
	if err := round.Update(in, ts); err != nil {
		return InterviewRound{}, err
	}
	return round, nil
}

func (r *InterviewRound) Update(in InterviewRoundInput, now time.Time) error {
	if in.RoundOrder <= 0 {
		return fmt.Errorf("%w: round order must be >= 1", ErrValidation)
	}
	roundType := InterviewRoundType(strings.TrimSpace(string(in.RoundType)))
	title := strings.TrimSpace(in.Title)
	if roundType == "" && title == "" {
		return fmt.Errorf("%w: round type or title is required", ErrValidation)
	}
	if roundType == "" {
		roundType = InterviewRoundType(title)
	}
	if title == "" {
		title = string(roundType)
	}
	if in.DurationMinutes < 0 {
		return fmt.Errorf("%w: duration minutes cannot be negative", ErrValidation)
	}
	status := InterviewRoundStatus(strings.TrimSpace(string(in.Status)))
	if status == "" {
		if r.Status != "" {
			status = r.Status
		} else {
			status = InterviewRoundStatusScheduled
		}
	}
	if !status.IsValid() {
		return fmt.Errorf("%w: invalid interview round status %q", ErrValidation, status)
	}

	var scheduledAt time.Time
	if !in.ScheduledAt.IsZero() {
		scheduledAt = in.ScheduledAt.UTC()
	}

	r.RoundOrder = in.RoundOrder
	r.RoundType = roundType
	r.Title = title
	r.ScheduledAt = scheduledAt
	r.DurationMinutes = in.DurationMinutes
	r.Status = status
	r.PrepNotes = strings.TrimSpace(in.PrepNotes)
	r.ContactIDs = normalizeExactIDs(in.ContactIDs)
	r.StoryIDs = normalizeExactIDs(in.StoryIDs)
	r.UpdatedAt = now.UTC()
	return nil
}

func (r *InterviewRound) RecordDebrief(in DebriefInput, now time.Time) error {
	if in.ConfidenceScore < 1 || in.ConfidenceScore > 5 {
		return fmt.Errorf("%w: confidence score must be between 1 and 5", ErrValidation)
	}
	questions := make([]DebriefQuestion, 0, len(in.QuestionsAsked))
	for _, q := range in.QuestionsAsked {
		qText := strings.TrimSpace(q.Question)
		if qText == "" {
			return fmt.Errorf("%w: debrief question text is required", ErrValidation)
		}
		questions = append(questions, DebriefQuestion{
			Question:            qText,
			Category:            strings.TrimSpace(q.Category),
			AnsweredWithStoryID: strings.TrimSpace(q.AnsweredWithStoryID),
			Notes:               strings.TrimSpace(q.Notes),
		})
	}
	ts := now.UTC()
	r.Debrief = &Debrief{
		QuestionsAsked:  questions,
		WentWell:        strings.TrimSpace(in.WentWell),
		ToImprove:       strings.TrimSpace(in.ToImprove),
		Signals:         strings.TrimSpace(in.Signals),
		ConfidenceScore: in.ConfidenceScore,
		FollowUpNotes:   strings.TrimSpace(in.FollowUpNotes),
		UpdatedAt:       ts,
	}
	r.UpdatedAt = ts
	return nil
}

type StoryPriorRoundRef struct {
	RoundID    string
	RoundOrder int
	RoundType  InterviewRoundType
	Title      string
}

// DetectRepeatedStories is a pure domain evaluator that checks each non-cancelled
// InterviewRound's linked StoryIDs against earlier non-cancelled rounds of the
// same Opportunity (ordered by RoundOrder). It returns a map keyed by
// RoundID -> StoryID -> []StoryPriorRoundRef for every story repeated in an
// earlier round of the same loop.
func DetectRepeatedStories(rounds []InterviewRound) map[string]map[string][]StoryPriorRoundRef {
	result := make(map[string]map[string][]StoryPriorRoundRef)
	if len(rounds) == 0 {
		return result
	}

	byOpp := make(map[string][]InterviewRound)
	for _, r := range rounds {
		if r.Status == InterviewRoundStatusCancelled {
			continue
		}
		oppID := strings.TrimSpace(r.OpportunityID)
		byOpp[oppID] = append(byOpp[oppID], r)
	}

	for _, oppRounds := range byOpp {
		sort.SliceStable(oppRounds, func(i, j int) bool {
			if oppRounds[i].RoundOrder != oppRounds[j].RoundOrder {
				return oppRounds[i].RoundOrder < oppRounds[j].RoundOrder
			}
			return oppRounds[i].ID < oppRounds[j].ID
		})

		seenByStory := make(map[string][]StoryPriorRoundRef)
		for _, r := range oppRounds {
			for _, rawStoryID := range r.StoryIDs {
				storyID := strings.TrimSpace(rawStoryID)
				if storyID == "" {
					continue
				}
				if priors := seenByStory[storyID]; len(priors) > 0 {
					if result[r.ID] == nil {
						result[r.ID] = make(map[string][]StoryPriorRoundRef)
					}
					copied := make([]StoryPriorRoundRef, len(priors))
					copy(copied, priors)
					result[r.ID][storyID] = copied
				}
			}

			for _, rawStoryID := range r.StoryIDs {
				storyID := strings.TrimSpace(rawStoryID)
				if storyID == "" {
					continue
				}
				seenByStory[storyID] = append(seenByStory[storyID], StoryPriorRoundRef{
					RoundID:    r.ID,
					RoundOrder: r.RoundOrder,
					RoundType:  r.RoundType,
					Title:      r.Title,
				})
			}
		}
	}

	return result
}

type InterviewRoundContactView struct {
	ContactID   string
	Name        string
	Title       string
	Email       string
	Warmth      ContactWarmth
	CompanyName string
}

type InterviewRoundStoryView struct {
	StoryID             string
	Title               string
	Hook                string
	CompetencyTags      []string
	SkillTags           []string
	UsedInEarlierRounds []StoryPriorRoundRef
}

type DebriefQuestionView struct {
	Question             string
	Category             string
	AnsweredWithStoryID  string
	AnsweredWithStoryTitle string
	Notes                string
}

type DebriefView struct {
	QuestionsAsked  []DebriefQuestionView
	WentWell        string
	ToImprove       string
	Signals         string
	ConfidenceScore int
	FollowUpNotes   string
	UpdatedAt       time.Time
}

type InterviewRoundView struct {
	InterviewRound InterviewRound
	Contacts       []InterviewRoundContactView
	Stories        []InterviewRoundStoryView
	DebriefView    *DebriefView
}

type QuestionBankFilter struct {
	Query      string
	Category   string
	CompanyID  string
	RoundType  InterviewRoundType
	StoryID    string
}

type QuestionBankEntryView struct {
	RoundID              string
	OpportunityID        string
	OpportunityRoleTitle string
	CompanyID            string
	CompanyName          string
	RoundOrder           int
	RoundType            InterviewRoundType
	RoundTitle           string
	Question             string
	Category             string
	AnsweredWithStoryID  string
	AnsweredWithStoryTitle string
	Notes                string
	RecordedAt           time.Time
}

func normalizeExactIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, raw := range ids {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
